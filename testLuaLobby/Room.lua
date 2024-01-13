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
			game_addr = msg.game_addr or "",
			connet_time = os.time(),
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


	if message.fast_label == ENUM.FAST_LABEL_NEW_ROOM then
		print("FAST_LABEL_NEW_ROOM")
		if GLOBAL_GAME_ROOM[message.room_id]
			and GLOBAL_GAME_ROOM[message.room_id].waiting_create == true then
			GLOBAL_GAME_ROOM[message.room_id].waiting_create = false
		end
		print("get room:", json.encode(GLOBAL_GAME_ROOM[message.room_id]))
		local uid = msg.uid
		if uid then
			local p = GLOBAL_USER_INFO[uid]
			if p then
				p:Send("sys",
					{msg="create room success", room_id = message.room_id})
			else
				print("not find uid:", uid)
			end
		end
		return
	end


	print("not handle this message:", data.message)
	return
end


local StartRoomId = 666660
function Room:NewGameRoom(game_id, u_id)

	StartRoomId = StartRoomId + 1
	local data = {
		fast_label = ENUM.FAST_LABEL_NEW_ROOM,
		room_id = StartRoomId,
		Msg = {uid=u_id}

	}
	local WS = GLOBAL_GAME[game_id][1].__WS
	GameServerSent(WS, json.encode(data))
	GLOBAL_GAME_ROOM[StartRoomId] = {
		creat_by = u_id,
		waiting_create = true,
		game_id = game_id,
		game_connet_time = GLOBAL_GAME[game_id][1].connet_time
	}
	return true
end


function Room:SendUserToGame(gameInfo, player, room_id)
	local data = {
		fast_label = ENUM.FAST_LABEL_NEW_USER,
		room_id = room_id,
		Msg = {
			uid = player.id,
			name = player.name,
			img = player.img,
			room_id = room_id,
		}
	}
	local WS = gameInfo.__WS
	GameServerSent(WS, json.encode(data))
end


local function GetGameInfoByGameRoom(roomInfo)
	local gameTab = GLOBAL_GAME[roomInfo.game_id]
	if gameTab then
		for _,gameInfo in ipairs(gameTab) do
			if gameInfo.connet_time == roomInfo.game_connet_time then
				return gameInfo
			end
		end
	end
end
function Room:EnterGameRoom(player, room_id)
	if not GLOBAL_GAME_ROOM[room_id] then
		player:Send("sys", {msg="not exist room_id:"..room_id})
		return
	end

	local roomInfo = GLOBAL_GAME_ROOM[room_id]
	local gameInfo = GetGameInfoByGameRoom(roomInfo)
	if gameInfo then
		GLOBAL_USER_INFO[player.id].inRoomId = room_id
		self:SendUserToGame(gameInfo, player, room_id)

		local ret = {
			addr = gameInfo.game_addr,
			room_id = room_id
		}
		return ret
	else
		player:Send("sys", {msg="not find gameInfo, room_id:"..room_id})
	end
end


function Room:PlayerIn(player)
	print("lobby PlayerIn id:", player.id)
	if GLOBAL_USER_INFO[player.id] and GLOBAL_USER_INFO[player.id].inRoomId then
		player:Send("sys", {msg = "welcome to the lobby", inRoomId = GLOBAL_USER_INFO[player.id].inRoomId})
	else
		GLOBAL_USER_INFO[player.id] = player
		player:Send("sys", {msg = "welcome to the lobby"})
	end
end
































