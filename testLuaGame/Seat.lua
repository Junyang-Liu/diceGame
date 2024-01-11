local Seat = {}

local Cards = require("testLuaGame.cards")
local POKER = require("testLuaGame.poker")

function Seat:Send(...)
	if self.player then
		self.player:Send(...)
	end
end

function Seat:DoReady()
	if not self.Room:WaitingReady() then
		self:Send("opReadyRsp", {msg="faild", code = -1, status = self.Room.status})
		return
	end

	self.ready = true
	self:Send("opReadyRsp", {msg="success", code = 0})
	self.Room:BroadCastExcept(self, "DoReady", {msg="do ready", seat= self.IDX})

	if self.Room:CheckStart() then
		self.Room:StartNewRound()
	end
end

function Seat:DoBanker()
	if not self.Room:Banking()then
		self:Send("sys", {msg="room not Banking"})
		return
	end
	self.Room:SetBanker(self)
end

function Seat:DoNotBanker()
	if not self.Room:Banking()then
		self:Send("sys", {msg="room not Banking"})
		return
	end
	self.Room:NotBanker(self)
end

function Seat:DoPass()
	if not self.Room:Gaming() then
		self:Send("opPassRsp", {msg="faild", code = -1, status=self.Room.status})
		return
	end

	if self.Room.opSeat ~= self then
		self:Send("opPassRsp", {msg="faild", code = -1 , op=self.Room.opSeat.IDX})
		return
	end

	if not self.Room:DoPass(self) then
		self:Send("opPassRsp", {msg="faild",code = -1})
	else
		self:Send("opPassRsp", {msg="success",code = 0})
		self.Room:BroadCastExcept(self, "DoPass", {msg="do DoPass", seat= self.IDX})
		self.Room:BroadCast("GameStatus", {status=self.Room.status, op=self.Room.opSeat.IDX, time=10000})
	end
end

function Seat:DoOutCards(outCards)
	if not self.Room:Gaming() then
		self:Send("opOutCardsRsp", {msg="faild", code = -1, status=self.Room.status})
		return
	end

	if self.Room.opSeat ~= self then
		self:Send("opOutCardsRsp", {msg="faild", code = -1 , op=self.Room.opSeat.IDX})
		return
	end

	if self.Room:DoOutCards(self, outCards) then
		self:Send("opOutCardsRsp", {code=0})
	else
		self:Send("opOutCardsRsp", {msg="faild", code = -1})
	end
end

function Seat:PlayerLeave()
	if not self.Room:WaitingReady() then
		self:Send("opLeaveRsp", {code=-1})
		return
	end

	self:Send("opLeaveRsp", {code=1})
	self.Room:PlayerOut(self.player.id)
	self.player = nil
end

function Seat:SendHandCards()
	self:Send("handCards", self.handCards)
end

local sortFunc = function (a, b)
	if Cards[a].point == Cards[b].point then
		return Cards[a].typ > Cards[b].typ
	end

	return Cards[a].point > Cards[b].point
end

function Seat:SortHandCards()
	if next(self.handCards) then
		table.sort( self.handCards, sortFunc )
	end
end

function Seat:RemoveHandCard(cards)
	local idx = {}
	for i,v in ipairs(self.handCards) do
		for _,card in ipairs(cards) do
			if card == v then
				table.insert(idx, i)
			end
		end
	end
	-- reverse
	table.sort(idx, function (a, b)
		return a > b
	end)

	for _,v in ipairs(idx) do
		table.remove(self.handCards, v)
	end
	-- for i,v in ipairs(self.handCards) do
	-- 	print("handCards", i,v)
	-- end
end

function Seat:Finish()
	return POKER.TableLen(self.handCards) == 0
end


local function new(Room)
	local ret = {
		Room = Room
	}
	setmetatable(ret, {__index = Seat})
	return ret
end



local _M = {
	new = new

}
return _M
