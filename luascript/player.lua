
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

-- return _M