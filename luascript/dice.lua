
local go = go
local json = require("json")
local roomMod = require("__room")
local palyerMod = require("__player")
__NewTimer = go.NewTimer
__CancelTimer = go.CancelTimer
__TimerLastTime = go.TimerLastTime
__PlayerOut = go.userOut
__CloseThisVm = go.closeVM
__CloseWS = go.closeWS
GameServerSent = go.GameServerSent
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



