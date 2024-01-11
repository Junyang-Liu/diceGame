
local json = require("json")
local POKER = require("testLuaGame.poker")

function Room:PlayerIn(Player)

	Room:PlayerSeatDown(Player)
end

local seatMod = require("testLuaGame.Seat")
local Cards = require("testLuaGame.cards")

GAMESTATUS = {
	READY = 1,
	Banking = 2,
	GAMING = 3,
}

function Room:Init()
	self.Seats = self.Seats or {}
	for i=1, 2 do
		self.Seats[i] = seatMod.new(self)
		self.Seats[i].IDX = i
	end

	self.maxRound = 2
	self.round = 0
	self.status = GAMESTATUS.READY
	self.opSeat = nil
	self.roundData = {}

end

function Room:CheckPlayerBack(p)
	for i=1, 2 do
		if self.Seats[i].player and self.Seats[i].player.id == p.id then
			self.Seats[i].player = p
			p.seat = self.Seats[i]
			p:Send("inGameRsp", {msg="welcome to my game!", code=0, mySeatIdx = p.seat.IDX, status = self.status})
			self:BroadCast("SeatsStatus", {msg="player back", uid = p.id, seat=i})
			return true
		end
	end
end

function Room:SendRoomStatus(seat)
	seat:Send("GameStatus", {op= self.opSeat and self.opSeat.IDX or nil, status= self.status})
end

function Room:PlayerSeatDown(player)
	if self:CheckPlayerBack(player) then
		return
	end

	for i=1, 2 do
		if not self.Seats[i].player then
			self.Seats[i].player = player
			player.seat = self.Seats[i]
			player:Send("inGameRsp", {msg="welcome to my game!", code=0, mySeatIdx = player.seat.IDX, status = self.status})
			self:BroadCast("SeatsStatus", {msg="new player seat down", uid = player.id, seat=i})
			break
		end
	end
end

function Room:BroadCast(...)
	for _,seat in ipairs(self.Seats) do
		seat:Send(...)
	end
end

function Room:BroadCastExcept(s, ...)
	for _,seat in ipairs(self.Seats) do
		if seat ~= s then
			seat:Send(...)
		end
	end
end

function Room:CheckStart()
	local readySeats = 0
	for i,v in ipairs(self.Seats) do
		if v.ready then
			readySeats = readySeats + 1
		end
	end
	if readySeats == 2 then
		return true
	end
	return false
end

function Room:WaitingReady()
	return self.status == GAMESTATUS.READY
end

function Room:Banking()
	return self.status == GAMESTATUS.Banking
end

function Room:Gaming()
	return self.status > GAMESTATUS.READY
end

function Room:Shuffle()
	-- if not self.allCards or not next(self.allCards) then
		self.allCards = {}
		for _,v in pairs(Cards) do
			table.insert(self.allCards, v.id)
		end
	-- end

	self.shuffleCards = {}
	math.randomseed(os.time())
	for i=54,1,-1 do
		local idx = math.random(1, i)
		table.insert(self.shuffleCards, table.remove(self.allCards, idx))
	end
end

function Room:InitSeatCards(num)
	local num = num or 17
	local startIdx = 0
	for seatIdx=1,2 do
		self.Seats[seatIdx].handCards = {}
		for i=1,num do
			self.Seats[seatIdx].handCards[i] = self.shuffleCards[startIdx+i]
			self.Seats[seatIdx]:SortHandCards()
			self.Seats[seatIdx].sourceHandCards = POKER.TableClone(self.Seats[seatIdx].handCards)
		end
		startIdx =  startIdx + 17
	end
end

function Room:InitBankerCards()
	self.bankerCards = {}
	for i=0,2 do
		table.insert(self.bankerCards, self.shuffleCards[54-i])
	end
end

function Room:SendCards()
	for _,v in ipairs(self.Seats) do
		v:SendHandCards()
	end
end

function Room:BroadCastBankerCards()
	self:BroadCast("bankerCards", self.bankerCards)
end

function Room:BankerCardsToSeat(s)
	for i,v in ipairs(self.bankerCards) do
		table.insert(s.handCards, v)
	end
	s.sourceHandCards = POKER.TableClone(s.handCards)
	s:SortHandCards()
	s:SendHandCards()
end

function Room:StartNewRound()
	self.status = GAMESTATUS.Banking
	self.round = self.round + 1
	if self.round == 1 then
		lobby.StartPlay()
	end
	self.opSeat = self.nextRoundFirstOpSeatIDX and self.Seats[self.nextRoundFirstOpSeatIDX] or self.Seats[1]
	self.nextRoundFirstOpSeatIDX = self.opSeat.IDX

	self:Shuffle()
	self:InitSeatCards()
	self:InitBankerCards()
	self:BroadCast("GameStatus", {msg= "new round start", round= self.round, status=self.status, op=self.opSeat.IDX, time=10000})
	self:SendCards()
	-- self:BroadCastBankerCards()
	self:NewTimer(10000, self.opSeat.player.DoBanker, self.opSeat.player, {banker = 1})
end

function Room:NextOp()
	self.opSeat = self:GetNextOpSeat()
end

function Room:GetNextOpSeat()
	local nextIndex = self.opSeat.IDX + 1
	if nextIndex > 2 then
		nextIndex = 1
	end
	print("nextIndex:", nextIndex)
	return self.Seats[nextIndex]
end


function Room:SetBanker(seat)
	if self.bankerSeat then
		seat:Send("sys", {msg="has a bankerSeat!"})
		return
	end

	if self.opSeat ~= seat then
		seat:Send("sys", {msg="not your turn!"})
		return
	end
	self:CancelTimer()
	self.bankerSeat = seat
	self.lastOutCardSeatIDX = seat.IDX
	seat:Send("opBankerRsp", {msg= "success", code=0})
	self:BroadCastExcept(seat, "BankerSeat", {seat=seat.IDX})
	self:BroadCastBankerCards()
	self:BankerCardsToSeat(seat)
	self:RoundPlaying()
end


function Room:NotBanker(seat)
	if self.opSeat ~= seat then
		seat:Send("sys", {msg="not your turn!"})
		return
	end

	if self.nextRoundFirstOpSeatIDX == seat.IDX and self.BankHasPassOne then
		seat:Send("sys", {msg="no one do banking , this seat must do"})
		return
	end
	self:CancelTimer()
	seat:Send("opBankerRsp", {msg= "success", code=0})
	self.BankHasPassOne = true
	self:NextOp()
	self:BroadCast("GameStatus", {msg= "banking", status=self.status, op=self.opSeat.IDX, time=10000})
	self:NewTimer(10000, self.opSeat.player.DoBanker, self.opSeat.player, {banker = 1})
	return true
end

function Room:RoundPlaying()
	self.status = GAMESTATUS.GAMING
	self:BroadCast("GameStatus", {status=self.status, op=self.opSeat.IDX, time=10000})
	self.opSeat = self.bankerSeat
end

function Room:DoOutCards(seat, outcards)
	if self.opSeat ~= seat then
		seat:Send("sys", {msg="not your turn!"})
		return
	end

	for _,v in ipairs(outcards) do
		local has = false
		for _,handcard in ipairs(seat.handCards) do
			if handcard == v then
				has = true
				break
			end
		end
		if has == false then
			seat:Send("sys", {msg = "dont have this card: " .. v})
		end
	end
	if self.nowOutCardInfo and self.lastOutCardSeatIDX ~= seat.IDX then
		local nowOutCardInfo = self.nowOutCardInfo
		local max, t, p, len = POKER.MaxThen(nowOutCardInfo.outcardTyp, 
									nowOutCardInfo.outcardPoint,
									nowOutCardInfo.outcardLen,
									outcards)
		if not max then
			seat:Send("sys", {msg="not max then", yourdata=outcards})
			return
		end

		self.nowOutCardInfo.outcards = outcards
		self.nowOutCardInfo.outcardTyp = t
		self.nowOutCardInfo.outcardLen = len
		self.nowOutCardInfo.outcardPoint = p
	else
		self.nowOutCardInfo = {
			outcards = {},
			outcardTyp = 0,
			outcardLen = 0,
			outcardPoint = 0,
		}
		local t, p= POKER.GetCardTypeAndPoint(outcards)
		if not t then
			seat:Send("sys", {msg="not a card type", yourdata=outcards})
			return
		end
		self.nowOutCardInfo.outcards = outcards
		self.nowOutCardInfo.outcardTyp = t
		self.nowOutCardInfo.outcardLen = POKER.TableLen(outcards)
		self.nowOutCardInfo.outcardPoint = p
	end
	self.opSeat:RemoveHandCard(outcards)
	self.lastOutCardSeatIDX = self.opSeat.IDX
	self:BroadCast("newOutCardInfo", self.nowOutCardInfo)

	if self.opSeat:Finish() then
		self:RoundEnd()
		return true
	end
	self:NextOp()
	self:BroadCast("GameStatus", {status=self.status, op=self.opSeat.IDX, time=10000})
	return true
end

function Room:DoPass(seat)
	if self:Gaming() then
		if self.lastOutCardSeatIDX ~= seat.IDX then
			self:NextOp()
			return true
		end
	end
	return false
end

function Room:RoundEnd()
	self.status = GAMESTATUS.READY
	self:BroadCast("GameStatus", {msg = "round finish", status= self.status})

	if self.maxRound == self.round then
		self:FinishGame()
		self:Destroy()
		return
	end

	self.bankerSeat = nil
	self.BankHasPassOne = nil
	self.nextRoundFirstOpSeatIDX = self.opSeat.IDX
	self.opSeat = nil
	self.nowOutCardInfo = nil
	for _,v in ipairs(self.Seats) do
		v.ready = false
	end
end

function Room:FinishGame()
	self:BroadCast("sys", {msg = "game finish"})
	local result = [[{"msg":"nothing result"}]]
	lobby.EndPlay(result)
end

function Room:Destroy()
	print("yep destroy!")
	self:destroy()
end























