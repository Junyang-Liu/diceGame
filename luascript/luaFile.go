package luascript

var Player=`

local json = require("json")
local function send(self, line, data)
    local str = json.encode(data)
    go.send(self.__WS, line, str)
end

local function OP(self, line, data)
	print("OP: ", self.id, "line: ",line , "data:", data)
    print("this is test, got to rewrite Player:OP function!")
    data.msg = "this is test msg"

    local str = json.encode(data)
    go.send(self.__WS, line, str)
end

local function NewTimer(self, ...)
	local timer = __NewTimer(...)
	self.__timer = timer
end

local function CancelTimer(self)
    if self.__timer then
        __CancelTimer(self.__timer)
    end
end

local function ExistTimer(self)
    if self.__timer then
        return true
    end
    return false
end

local function TimerLast( self )
    if self.__timer then
        return __TimerLastTime(self.__timer)
    end
end


Player = {
        -- data
        room = Room,

        -- func
        Send = send,
        OP = OP,
        NewTimer = NewTimer,
        ExistTimer = ExistTimer,
        CancelTimer = CancelTimer,
        TimerLast = TimerLast,
}

local function New(room, uid, WS) 
	local ret = {
        id = uid,
        __WS = WS
    }
	setmetatable(ret, {__index = Player})
	return ret
end


local _M = {
	New = New
}

package.loaded["__player"] = _M

-- return _M`

var Room=`
local playerMod = require("__player")

local function broadcast(self, ...)
	for _,v in pairs(self.__players) do
		v:Send(...)
	end
end

local function addPlayer(self, uId, WS)
	local first = true
	if self.__players[uId] then
		first = false
	end
	self.__players[uId] = playerMod.New(self, uId, WS)
	return self.__players[uId], first
end

local function PlayerIn(self, uid)
	print("PlayerIn Room id:", self.id,  "uid:", uid)
    print("this is test, got to rewrite Room:PlayerIn function!")
	local timer = __NewTimer(1000, broadcast, self, "fromInGame", { msg="this id test timer msg! player " .. (uid and uid or "nil") .. " get in this room "..  (self and self.id or "nil")})
	__NewTimer(2000, broadcast, self, "fromInGame", { msg="this id test timer msg! player " .. (uid and uid or "nil") .. " get in this room "..  (self and self.id or "nil")})
	__CancelTimer(timer)

end

local function PlayerOut(self, uid)
	if self.__players[uid] then
		__PlayerOut(uid)
		self.__players[uid] = nil
	end
end

local function NewTimer(self, ...)
	local timer = __NewTimer(...)
	self.__timer = timer
end

local function CancelTimer(self)
	if self.__timer then
		__CancelTimer(self.__timer)
		self.__timer = nil
	end
end

local function ExistTimer(self)
    if self.__timer then
        return true
    end
    return false
end

local function TimerLast( self )
    if self.__timer then
        return __TimerLastTime(self.__timer)
    end
end

local function destroy(self)
	self:CancelTimer()
	for k,v in pairs(self.__players) do
		v:CancelTimer()
		self:PlayerOut(k)
	end
	__CloseThisVm()
end

local function New(roomId, uId)
	local Room = {
		-- data
		id = roomId,
		createBy = uId,
		-- maxPlayerNum = 2,
		__players = {},
		opPlayer = nil,
		-- __timer = nil,

		-- func
		PlayerIn = PlayerIn,
		PlayerOut = PlayerOut,
		NewTimer = NewTimer,
		ExistTimer = ExistTimer,
		CancelTimer = CancelTimer,
		TimerLast = TimerLast,
		destroy = destroy,


	}
	local ret = {}
	setmetatable(ret, {__index = Room})
	return ret
end


local _M = {
	New = New,
	addPlayer = addPlayer,
}

package.loaded["__room"] = _M
-- return _M`

var Dice=`

local go = go
local json = require("json")
local roomMod = require("__room")
local palyerMod = require("__player")
__NewTimer = go.NewTimer
__CancelTimer = go.CancelTimer
__TimerLastTime = go.TimerLastTime
__PlayerOut = go.userOut
__CloseThisVm = go.closeVM
Room = nil
G_ROOM_ID = nil

function newRoom(roomId, uId)
	if Room then
		print([[存在房间了]])
		return -1
	end
	Room = roomMod.New(roomId, uId)
	G_ROOM_ID = roomId

	print("newRoom succees roomId:", roomId)
	return 0
end

function userIn(uid, line, data)
	local roomId = data.roomId
	local uId = data.uId
	if not Room then
		print("userIn fail, Room nil")
		return -1
	end
	if Room.id ~= roomId then
		print("userIn fail already rooomId:", roomId, "Room.id:", Room.id)
		return -1
	end
	local player, isFirst = roomMod.addPlayer(Room, uId, data.__WS)
	Room:PlayerIn(player)
	print("userIn succees roomId:", roomId, "uId:", uId, "isFirst:", isFirst)

	return 0
end

function playerOP(uid, line, data)
	if Room.__players and Room.__players[uid] then
		Room.__players[uid]:OP(line, data)
		return 0
	end
	print("[lua]: playerOP faild uid:", uid, "line:", line, "has follow uid")
	for k,_ in pairs(Room.__players) do
		print(k)
	end
	return -1
end



`

