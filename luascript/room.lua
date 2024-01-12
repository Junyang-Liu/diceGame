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

local function PlayerIn(self, player)
	print("PlayerIn Room id:", self.id,  "uid:", player.uid)
    print("this is test, got to rewrite Room:PlayerIn function!")
	local timer = __NewTimer(1000, broadcast, self, "fromInGame", { msg="this id test timer msg! player " .. (player.uid and player.uid or "nil") .. " get in this room "..  (self and self.id or "nil")})
	__NewTimer(2000, broadcast, self, "fromInGame", { msg="this id test timer msg! player " .. (player.uid and player.uid or "nil") .. " get in this room "..  (self and self.id or "nil")})
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
-- return _M