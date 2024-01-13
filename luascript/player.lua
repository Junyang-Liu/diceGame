
local json = require("json")
local function send(self, line, data)
    local str
    if type(data) == "string" then
        str = data
    else
        str = json.encode(data)
    end
    if self.__WS then
        go.send(self.__WS, line, str)
    end
end

local function OP(self, line, data)
	print("OP: ", self.id, "line: ",line , "data:", data)
    print("this is test, got to rewrite Player:OP function!")
    data.msg = "this is test msg"

    local str = json.encode(data)
    if self.__WS then
        go.send(self.__WS, line, str)
    end
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

local function Offline(self)
    if self.__WS then
        __CloseWS(self.__WS)
        self.__WS = nil
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
        Offline = Offline,
}

local function New(room, uid, WS) 
	local ret = {
        id = uid,
        uid = uid,
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