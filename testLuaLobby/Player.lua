
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
		self:Send("newGameRoomRsp", "game_id require")
		return
	end
end


Player.line_opGameTest = Player.opGameTest
Player.line_getAllGame = Player.getAllGame
Player.line_newGameRoom = Player.newGameRoom

function Player:OP(line, data)
	if type(self["line_" .. line]) == "function" then
		self["line_" .. line](self, data)
	else
		self:Send(line, {msg="not suport", line=line, code=-1, yourdata = data})
	end
end