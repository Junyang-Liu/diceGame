local json = require("json")
local ENUM = require("testLuaLobby.ENUM")

GLOBAL_GAME = GLOBAL_GAME or {}
GLOBAL_GAME_ROOM = GLOBAL_GAME_ROOM or {}
GLOBAL_USER_INFO = GLOBAL_USER_INFO or {}


local function sortGameInfo(a, b)
	return a.priority > b.priority
end

function Room:GameServerCall(data)
	local addr = data.addr
	local __WS = data.__WS
	local message, err = json.decode(data.message)
	if err then
		print("err:", err)
		return
	end

	local msg = message.msg or {}
	if message.fast_label == ENUM.FAST_LABEL_NONE then
		print("nothing to do message:", data.message)
		return
	end

	if message.fast_label == ENUM.FAST_LABEL_NEW_SERVER then
		print("FAST_LABEL_NEW_SERVER")
		if message.game_id == 0 then
			print("new game server but game_id 0")
			return
		end
		GLOBAL_GAME[message.game_id] = GLOBAL_GAME[message.game_id] or {}

		local gameInfo = {
			addr = addr,
			__WS = __WS,
			game_id = message.game_id,
			priority = message.priority,
			game_addr = msg.game_addr or ""
		}
		table.insert(GLOBAL_GAME[message.game_id], gameInfo)
		table.sort(GLOBAL_GAME[message.game_id], sortGameInfo)
		return
	end


	if message.fast_label == ENUM.FAST_LABEL_USER_LEAVE then
		print("FAST_LABEL_USER_LEAVE")
		if GLOBAL_USER_INFO[msg.uid] then
			GLOBAL_USER_INFO[msg.uid].imRoom = nil
		end
		return
	end


	if message.fast_label == ENUM.FAST_LABEL_ROOM_START then
		print("FAST_LABEL_ROOM_START")
		if GLOBAL_GAME_ROOM[message.room_id] then
			GLOBAL_GAME_ROOM[message.room_id].playing = true
		end
		return
	end


	if message.fast_label == ENUM.FAST_LABEL_ROOM_END then
		print("FAST_LABEL_ROOM_END")
		if GLOBAL_GAME_ROOM[message.room_id] then
			GLOBAL_GAME_ROOM[message.room_id].playing = false
		end
		print("get result:", json.encode(msg))
		return
	end


	print("not handle this message:", data.message)
	return
end


function Room:PlayerIn(player)
	print("PlayerIn id:", player.id)
	GLOBAL_USER_INFO[player.id] = player
	player:Send("sys", "welcome to the lobby")
end
































