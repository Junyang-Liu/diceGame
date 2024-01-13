
local json = require("json")

GLOBAL_GAME = GLOBAL_GAME or {}
GLOBAL_GAME_ROOM = GLOBAL_GAME_ROOM or {}
GLOBAL_USER_INFO = GLOBAL_USER_INFO or {}

function Player:opGameTest(data)
	self:Send("opGameTestRsp", {msg="yes", code=0, data=data})
end

function Player:getAllGame()
	local ret = {}
	for gameId in pairs(GLOBAL_GAME) do
		table.insert(ret, gameId)
	end
	self:Send("getAllGameRsp", ret)
end

function Player:newGameRoom(data)
	local gameId = data.game_id
	if not gameId then
		self:Send("newGameRoomRsp", {msg = "game_id require", code = -1})
		return
	end

	if not GLOBAL_GAME[gameId] then
		self:Send("newGameRoomRsp", {msg = "game_id not exist", code = -1})
		return
	end

	if Room:NewGameRoom(gameId, self.id) then
		self:Send("newGameRoomRsp", {msg = "newGameRoom success, waiting the game login info", code = 0})
		return
	else
		self:Send("newGameRoomRsp", {msg = "newGameRoom faild", code = -1})
		return
	end
end

function Player:enterGameRoom(data)
	local roomId = data.room_id
	if not roomId then
		self:Send("enterGameRoomRsp", {msg = "room_id require", code = -1})
		return
	end

	local ret = Room:EnterGameRoom(self, roomId)
	if ret then
		self:Send("enterGameRoomRsp", {msg = "enterGameRoom success, socket close usually ", code = 0, data = ret})
		self:Offline()
	else
		self:Send("enterGameRoomRsp", {msg = "enterGameRoom faild", code = -1})
	end
end


Player.line_opGameTest = Player.opGameTest
Player.line_getAllGame = Player.getAllGame
Player.line_newGameRoom = Player.newGameRoom
Player.line_enterGameRoom = Player.enterGameRoom

function Player:OP(line, data)
	if type(self["line_" .. line]) == "function" then
		self["line_" .. line](self, data)
	else
		self:Send(line, {msg="not suport", line=line, code=-1, yourdata = data})
	end
end