
-- 1. never pcall xpcall dofile here before this bug fix (https://github.com/yuin/gopher-lua/issues/464)
-- 2. Room is ready before init.lua run
-- 3. when a new player get in this Room, function Room:PlayerIn(Player) will be called
-- 4. when a player operate this game, function Player:OP(line, data) will be called
-- 5. use Player:Send(line, data) for sending msg to client
-- 6. use require("json"), offer json.encode json.decode function
-- 7. use Room:NewTimer(Millisecond, function, ...) to create a timer and cache it in this Room,
-- 		a duplicate calling will cover the timer exist,
-- 		Room:CancelTimer to cancel this room's timer,
-- 		Room:TimerLast() to get Millisecond the timer run,
-- 		Room:ExistTimer() to check exist a timer,
--		Same as Player:NewTimer and Player:CancelTimer and Player:TimerLast and Player:ExistTimer
-- 8. use Room:PlayerOut(player.id) to clear one player
-- 9. use Room:destroy() to clear all players in this room, and close this room
-- 10. use lobby.StartPlay to notice lobby server this room is start palying
-- 11. use lobby.EndPlay(result) to notice lobby server this room is finish palying, and sent string `result` to lobby

local function genCars( fileName)
	local fileName = fileName or "testLuaGame/cards.lua"
	local carFile = io.open(fileName, "w+")
	carFile:write("local cards = {\n")

	for typ=1,5 do
		if typ < 5 then
			for point=1,13 do
				local point = point
				local idx = point
				if point == 1 then
					point = 14
				end
				if point == 2 then
					point = 15
				end
				if idx < 10 then
					idx = "0" .. idx
				end
				idx = typ .. idx
				idx = tonumber(idx)
				local str = "[" .. idx .. "] = " .. "{ id = " .. idx .. ", point = " .. point .. ", typ = " .. typ .. " },\n"
				carFile:write(str)
			end
		else
			for point=1,2 do
				local idx = "0" .. point
				idx = typ .. idx
				idx = tonumber(idx)
				local str = "[" .. idx .. "] = " .. "{ id = " .. idx .. ", point = " .. point+15 .. ", typ = " .. typ .. " },\n"
				carFile:write(str)
			end
		end
	end
	carFile:write("}\n")
	carFile:write("setmetatable(cards, {__newindex = function(...)return end})\n")
	carFile:write("return cards\n")

	carFile:close()
end


local function tt( ... )
	print(...)
end

local function initMyGame()
	print("initMyGame !!!")

	require("testLuaGame.poker")
	require("testLuaGame.Seat")
	require("testLuaGame.Player")
	require("testLuaGame.Room")

	Room:Init()
	Room:NewTimer(1000, tt, "1000 timer")
	Room:CancelTimer()
	Room:NewTimer(2000, tt, "2000 timer")
	print(Room:TimerLast())

	local cardsFile = io.open("testLuaGame/cards.lua", "r")
	if not cardsFile then
		genCars("testLuaGame/cards.lua")
	end

	print("initMyGame finish !!!")
end

initMyGame()

-- genCars("cards.lua")














