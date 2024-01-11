
function Player:DoReady()
	print("Player DoReady")
	if self.seat then
		self.seat:DoReady()
	end
end

function Player:DoBanker(data)
	print("Player DoBanker")
	if not data or data.banker ~= 1 then
		self.seat:DoNotBanker()
		return
	end
	self.seat:DoBanker()
end

function Player:DoPass()
	if self.seat then
		self.seat:DoPass()
	end
end

function Player:DoOpCards(data)
	if self.seat then
		self.seat:DoOutCards(data.outCards)
	end
end

function Player:opGameTest(data)
	self:Send("opGameTestRsp", {msg="yes", code=0, data=data})
end

function Player:Leave(data)
	print("Player Leave")
	if self.seat then
		self.seat:PlayerLeave()
	end
end


Player.line_opGameTest = Player.opGameTest
Player.line_opLeave = Player.Leave
Player.line_opReady = Player.DoReady
Player.line_opBanker = Player.DoBanker
Player.line_opPass = Player.DoPass
Player.line_opCards = Player.DoOpCards
function Player:OP(line, data)
	if type(self["line_" .. line]) == "function" then
		self["line_" .. line](self, data)
	else
		self:Send(line, {msg="not suport", line=line, code=-1, yourdata = data})
	end
end