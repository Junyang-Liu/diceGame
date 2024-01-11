package server

import (
	"container/list"
	"diceGame/utils"
	"errors"
	"fmt"
	"sync"
	"time"

	rbtMod "github.com/HuKeping/rbtree"
	lua "github.com/yuin/gopher-lua"
)

// TODO some Mutex are unessential
var G_VM_LOCK sync.Mutex
var RuningVm map[int]*VmRunMsg

var G_RBT_LOCK sync.Mutex
var GlobalRBT map[int]*rbtMod.Rbtree

func init() {
	RuningVm = make(map[int]*VmRunMsg)
}

type VmRunMsg struct {
	L       *lua.LState
	RoomId  *int
	Onclose bool
	Runing  bool

	MsgQueLock  sync.Mutex
	MsgQue      *list.List
	TimerRBtree *rbtMod.Rbtree
}

type RunMsg struct {
	isLobby bool
	uid     *int
	op      *string
	data    *any
}

func CacheNewRunQue(roomId int, L *lua.LState) error {
	G_VM_LOCK.Lock()
	if _, ok := RuningVm[roomId]; ok {
		G_VM_LOCK.Unlock()
		return errors.New(fmt.Sprintf("RuningVm exist roomId: %d", roomId))
	}
	runVm := &VmRunMsg{L: L, MsgQue: new(list.List), RoomId: &roomId,
		Onclose: false, Runing: false, TimerRBtree: NewTimerRBtre()}

	RuningVm[roomId] = runVm
	G_VM_LOCK.Unlock()

	setRunVmUserdata(runVm)

	if runVm.Runing == false {
		runVm.Runing = true
		go Runing(runVm)
	}
	return nil
}

func CloseOneRunQue(roomId int) {
	G_VM_LOCK.Lock()
	if runVm, ok := RuningVm[roomId]; ok {
		runVm.Onclose = true
	}
	G_VM_LOCK.Unlock()
	return
}

func CacheRunMsg(roomId int, runMsg *RunMsg) error {
	G_VM_LOCK.Lock()
	defer G_VM_LOCK.Unlock()
	if runVm, ok := RuningVm[roomId]; ok {
		runVm.MsgQueLock.Lock()
		runVm.MsgQue.PushFront(runMsg)
		runVm.MsgQueLock.Unlock()

		return nil
	}
	return errors.New(fmt.Sprintf("RuningVm not exist roomId: %d", roomId))
}

func setRunVmUserdata(runVm *VmRunMsg) {
	L := runVm.L
	ud := L.NewUserData()
	ud.Value = runVm

	L.SetGlobal("VM_RUN_MSG", ud)
}

func ClearRuningVm(runVm *VmRunMsg) {
	G_VM_LOCK.Lock()
	defer G_VM_LOCK.Unlock()
	// runVm.MsgQueLock.Unlock()
	RuningVm[*runVm.RoomId] = nil
	ReMoveGameVm(*runVm.RoomId)
}

func Runing(runVm *VmRunMsg) {
	for {
		if runVm.Onclose == true {
			utils.Logger.Debug(fmt.Sprintf("runVm close: %+v", runVm))
			ClearRuningVm(runVm)
			return
		}

		RunTimers(runVm.L)
		RunMsgs(runVm)
	}
}

func RunMsgs(runVm *VmRunMsg) {
	runVm.MsgQueLock.Lock()
	if runVm.MsgQue.Len() > 0 {
		elm := runVm.MsgQue.Back()
		runVm.MsgQue.Remove(elm)
		runVm.MsgQueLock.Unlock()

		runMsg := elm.Value.(*RunMsg)
		utils.Logger.Debugf(fmt.Sprintf("Runing runMsg: %+v", *runMsg))

		L := runVm.L
		if L != nil {
			if !runMsg.isLobby {
				uid := runMsg.uid
				op := runMsg.op
				data := *runMsg.data

				callData := utils.MapToTable(data.(map[string]interface{}))
				CallLua(L, *uid, "playerOP", *op, callData)
			} else {
				LobbyCallLua(L, *runMsg.op)
			}
		} else {
			utils.Logger.Error("RunMsgs L nil")
		}
	} else {
		runVm.MsgQueLock.Unlock()

		// runtime.Gosched()
		time.Sleep(50 * time.Millisecond)
	}
}

func RunTimers(L *lua.LState) {
	VM_RUN_MSG := L.GetGlobal("VM_RUN_MSG")
	if VM_RUN_MSG.Type() != lua.LTUserData {
		utils.Logger.Error("VM_RUN_MSG not userdata, something wrong")
		return
	}

	runVm := VM_RUN_MSG.(*lua.LUserData).Value.(*VmRunMsg)
	if runVm.TimerRBtree.Len() < 1 {
		return
	}

	maxRun := 1000
	now := time.Now()
	for {
		maxRun--
		timer := runVm.TimerRBtree.Min().(*TimerInfo)
		if timer.at.Before(now) {
			if timer.cancel != true {
				TimerCallLua(L, timer.lf, timer.params)
			}
			runVm.TimerRBtree.Delete(timer)

			if runVm.TimerRBtree.Len() < 1 {
				return
			}
		} else {
			return
		}
		if maxRun < 0 {
			utils.Logger.Error("RunTimers maxRun 1000!!!")
			return
		}
	}
}

func ByTime(a, b any) int {
	aAsserted := a.(*TimerInfo).at
	bAsserted := b.(*TimerInfo).at

	switch {
	case aAsserted.After(bAsserted):
		return 1
	case aAsserted.Before(bAsserted):
		return -1
	default:
		return 0
	}
}

func NewTimerRBtre() *rbtMod.Rbtree {
	return rbtMod.New()
}

type TimerInfo struct {
	cancel bool
	at     time.Time
	lf     *lua.LFunction
	params []lua.LValue
}

func (x *TimerInfo) Less(than rbtMod.Item) bool {
	return x.at.Before(than.(*TimerInfo).at)
}

func SetOneTimer(L *lua.LState, mills int, lf *lua.LFunction, params []lua.LValue) error {
	utils.Logger.Debug("SetOneTimer!!!")

	VM_RUN_MSG := L.GetGlobal("VM_RUN_MSG")
	if VM_RUN_MSG.Type() != lua.LTUserData {
		return errors.New("VM_RUN_MSG not userdata, something wrong")
	}

	runVm := VM_RUN_MSG.(*lua.LUserData).Value.(*VmRunMsg)

	now := time.Now()
	timer := &TimerInfo{at: now.Add(time.Duration(mills) * time.Millisecond), lf: lf, params: params}
	utils.Logger.Debugf("now: %s", now.String())
	utils.Logger.Debugf("time: %s", timer.at.String())
	runVm.TimerRBtree.Insert(timer)

	udata := L.NewUserData()
	udata.Value = timer
	utils.Logger.Debugf(fmt.Sprintf("timer: %v", timer))

	L.Push(udata)
	return nil
}
