local poker = {
	NONE   = 0,
	SINGLE = 1,
	DOUBLE = 2,
	THREE  = 3,
	BOOM   = 4,
	JOKERBOOM = 5,

	SINGLE_STRAIGHT = 6,
	DOUBLE_STRAIGHT = 7,

	THREE_STRAIGHT = 8,
	THREE_ONE = 9,
	THREE_TWO = 10,
	THREE_STRAIGHT_WITH_SINGLE = 11,
	THREE_STRAIGHT_WITH_DOUBLE = 12,

	FOUR_WITH_TWO_SINGLE = 13,
	FOUR_WITH_TWO_DOUBLE = 14,

}


local BASECARDTYPE = {
	JOKER = 5
}

local BASECARDPOINT = {
	TWO = 15
}

local CARDS = require("testLuaGame.cards")


local function TableLen(tab)
	local ret = 0
	for _,_ in ipairs(tab) do
		ret = ret + 1
	end
	return ret
end
poker.TableLen = TableLen

local function TableClone(tab)
	local ret = {}
	for k,v in pairs(tab) do
		if type(v) == "table" then
			ret[k] = TableClone(v)
		else
			ret[k] = v
		end
	end
	return ret
end
poker.TableClone = TableClone

function poker.isJOKERBOOM(cards, len)
	local len = len or TableLen(cards)
	if len ~= 2 then
		return false
	end
	for i=1,2 do
		if CARDS[cards[i]].typ ~= BASECARDTYPE.JOKER then
			return false
		end
	end
	return true
end

function poker.isDOUBLE(cards, len)
	local len = len or TableLen(cards)
	if len ~= 2 then
		return false
	end
	for i=2,2 do
		if CARDS[cards[i]].point ~= CARDS[cards[1]].point then
			return false
		end
	end
	return true
end

function poker.isTHREE(cards, len)
	local len = len or TableLen(cards)
	if len ~= 3 then
		return false
	end
	for i=2,3 do
		if CARDS[cards[i]].point ~= CARDS[cards[1]].point then
			return false
		end
	end
	return true
end

function poker.isBOOM(cards, len)
	local len = len or TableLen(cards)
	if len ~= 4 then
		return false
	end
	for i=2,4 do
		if CARDS[cards[i]].point ~= CARDS[cards[1]].point then
			return false
		end
	end
	return true
end

local sortByPoint = function (a, b)
	return CARDS[a].point < CARDS[b].point
end

function poker.isTHREE_ONE(cards, len)
	local len = len or TableLen(cards)
	if len ~= 4 then
		return false
	end
	local tmpTab = TableClone(cards)
	table.sort(tmpTab, sortByPoint)

	local point
	if CARDS[tmpTab[1]].point == CARDS[tmpTab[2]].point then
		point = CARDS[tmpTab[1]].point
		if CARDS[tmpTab[2]].point ~= CARDS[tmpTab[3]].point 
			or CARDS[tmpTab[3]].point == CARDS[tmpTab[4]].point then
			return false
		end
	else
		for i=3,4 do
			if CARDS[tmpTab[i]].point ~= CARDS[tmpTab[2]].point then
				return false
			end
		end

		point = CARDS[tmpTab[3]].point
	end
	return true, point
end

function poker.isSINGLE_STRAIGHT(cards, len)
	local len = len or TableLen(cards)
	if len < 5 then
		return false
	end
	for i=1,len do
		if CARDS[cards[i]].point == BASECARDPOINT.TWO then
			return false
		end
		if CARDS[cards[i]].typ == BASECARDTYPE.JOKER then
			return false
		end
	end

	local tmpTab = TableClone(cards)
	table.sort(tmpTab, sortByPoint)
	local basePoint = CARDS[tmpTab[1]].point
	for i=2,len do
		if CARDS[tmpTab[i]].point ~= (basePoint + i - 1) then
			return false
		end
	end
	return true, CARDS[tmpTab[len]].point
end

function poker.isTHREE_TWO(cards, len)
	local len = len or TableLen(cards)
	if len ~= 5 then
		return false
	end

	local tmpTab = TableClone(cards)
	table.sort(tmpTab, sortByPoint)
	local basePoint = CARDS[tmpTab[1]].point
	local getcardnum = 1
	local point
	for i=2,len do
		if CARDS[tmpTab[i]].point ~= basePoint then
			if getcardnum == 2 or getcardnum == 3 then
				basePoint = CARDS[tmpTab[i]].point
				if getcardnum == 3 then
					point = basePoint
				end
				getcardnum = 1
			else
				return false
			end
		else
			getcardnum = getcardnum + 1
		end
	end
	return true, point or basePoint
end

function poker.hasBoom(cards, len)
	local len = len or TableLen(cards)
	if len < 4 then
		return false
	end

	local tmpTab = TableClone(cards)
	table.sort(tmpTab, sortByPoint)

	local boomPoint = CARDS[tmpTab[1]].point 
	for i=2,len do
		if CARDS[tmpTab[i]].point ~= boomPoint then
			if (len - i + 1) < 4 then
				return false
			else
				boomPoint = CARDS[tmpTab[i]].point
			end
		end
	end

	return true, boomPoint
end

function poker.isFOUR_WITH_TWO_SINGLE(cards, len)
	local len = len or TableLen(cards)
	if len ~= 6 then
		return false
	end
	local is, point = poker.hasBoom(cards, len)
	if not is then
		return false
	end
	return true, point
end

function poker.removeMaxPointBoom(tmpTab, len)
	local len = len or TableLen(tmpTab)
	if len < 4 then
		return false
	end

	table.sort(tmpTab, sortByPoint)

	local boomPoint = CARDS[tmpTab[len]].point
	local getcardnum = 1
	local boomIdx
	for i=len-1,1,-1 do
		if CARDS[tmpTab[i]].point ~= boomPoint then
			if i < 4 then
				return false
			end
			getcardnum = 1
			boomPoint = CARDS[tmpTab[i]].point
		else
			getcardnum = getcardnum + 1
			if getcardnum == 4 then
				boomIdx = i
				break
			end
		end
	end

	if not boomIdx then
		return false
	end

	for _= 1,4 do
		table.remove(tmpTab, boomIdx)
	end
	return true, boomPoint
end


function poker.removeMaxDouble(tmpTab, len)
	local len = len or TableLen(tmpTab)
	if len < 2 then
		return false
	end
	table.sort(tmpTab, sortByPoint)

	local doublePoint = CARDS[tmpTab[len]].point
	local getcardnum = 1
	local doubleIdx
	for i=len-1,1,-1 do
		if CARDS[tmpTab[i]].point ~= doublePoint then
			if i < 2 then
				return false
			end
			getcardnum = 1
			doublePoint = CARDS[tmpTab[i]].point
		else
			getcardnum = getcardnum + 1
			if getcardnum == 2 then
				doubleIdx = i
				break
			end
		end
	end


	if not doubleIdx then
		return false
	end

	for _= 1,2 do
		table.remove(tmpTab, doubleIdx)
	end
	return true
end

function poker.isFOUR_WITH_TWO_DOUBLE(cards, len)
	local len = len or TableLen(cards)
	if len ~= 8 then
		return false
	end
	local is, point = poker.hasBoom(cards, len)
	if not is then
		return false
	end

	local tmpTab = TableClone(cards)
	local is, point = poker.removeMaxPointBoom(tmpTab, len)
	if not is then
		return false
	end

	len = TableLen(tmpTab)
	if not poker.removeMaxDouble(tmpTab, len) then
		return false
	end

	len = TableLen(tmpTab)
	if not poker.removeMaxDouble(tmpTab, len) then
		return false
	end
	return true, point
end

function poker.isDOUBLE_STRAIGHT(cards, len)
	local len = len or TableLen(cards)
	if len < 6 or (len%2) ~= 0 then
		return false
	end

	for i=1,len do
		if CARDS[cards[i]].point == BASECARDPOINT.TWO then
			return false
		end
		if CARDS[cards[i]].typ == BASECARDTYPE.JOKER then
			return false
		end
	end

	local tmpTab = TableClone(cards)
	table.sort(tmpTab, sortByPoint)

	local pointCards = {}
	for i=3,len,2 do
		if CARDS[tmpTab[i]].point ~= (CARDS[tmpTab[1]].point + (i-1)/2 ) then
			return false
		end
	end
	return true, CARDS[tmpTab[len]].point
end

function poker.isTHREE_STRAIGHT(cards, len)
	local len = len or TableLen(cards)
	if len < 6 or (len%3) ~= 0 then
		return false
	end

	for i=1,len do
		if CARDS[cards[i]].point == BASECARDPOINT.TWO then
			return false
		end
		if CARDS[cards[i]].typ == BASECARDTYPE.JOKER then
			return false
		end
	end

	local tmpTab = TableClone(cards)
	table.sort(tmpTab, sortByPoint)

	local pointCards = {}
	for i=4,len,3 do
		if CARDS[tmpTab[i]].point ~= (CARDS[tmpTab[1]].point + (i-1)/3 ) then
			return false
		end
	end
	return true, CARDS[tmpTab[len]].point
end

function poker.removeMaxThreeStraight(tmpTab, len)
	local len = len or TableLen(tmpTab)
	if len < 6 then
		return false
	end

	table.sort(tmpTab, sortByPoint)
	local removeIdxs = {}
	local removeNum = 0
	local removingPoint = CARDS[tmpTab[len]].point
	local removing
	for i=len-1,1,-1 do
		if CARDS[tmpTab[i]].point ~= BASECARDPOINT.TWO 
			and CARDS[tmpTab[i]].typ ~= BASECARDTYPE.JOKER then
			if removing then
				if CARDS[tmpTab[i]].point == removingPoint then
					removing = removing + 1
					if removing == 2 then
						table.insert(removeIdxs, i)
						removeNum = removeNum + 1
					end
				else
					removingPoint = CARDS[tmpTab[i]].point
					removing = nil
				end
			else
				if CARDS[tmpTab[i]].point ~= removingPoint then
					removingPoint = CARDS[tmpTab[i]].point
					removing = nil
				else
					removing = 1
				end
			end
		end
	end

	if removeNum < 2 then
		return false
	end

	local point
	for i=1,removeNum do
		local idx = removeIdxs[i]
		local nextIdx = removeIdxs[i+1]
		local needSkip = false
		if nextIdx then
			if CARDS[tmpTab[nextIdx]].point ~= CARDS[tmpTab[idx]].point -1 then
				needSkip = true
			end
		end

		if not needSkip then
			for _=1,3 do
				if not point then
					point = CARDS[tmpTab[idx]].point
				end
				table.remove(tmpTab, removeIdxs[i])
			end
		end
	end

	return true, point
end

function poker.isTHREE_STRAIGHT_WITH_DOUBLE(cards, len)
	local len = len or TableLen(cards)
	if len < 10 then
		return false
	end

	local tmpTab = TableClone(cards)

	local boomNum = 0
	local tmpLen = len
	while(poker.removeMaxPointBoom(tmpTab, tmpLen))
	do
		boomNum = boomNum + 1
		tmpLen = TableLen(tmpTab)
	end

	local is, point = poker.removeMaxThreeStraight(tmpTab, tmpLen)
	if not is then
		return false
	end

	local doubleNum = 0
	tmpLen = TableLen(tmpTab)
	while(poker.removeMaxDouble(tmpTab, tmpLen))
	do
		doubleNum = doubleNum + 1
		tmpLen = TableLen(tmpTab)
	end

	local totalDoubleNum = doubleNum + boomNum*2
	if (len - totalDoubleNum*2)/3 ~= totalDoubleNum then
		return false
	end

	return true, point
end

function poker.getAllThreeIdx(tmpTab, len)
	local ret = {}
	local len = len or TableLen(tmpTab)
	if len < 3 then
		return ret
	end

	table.sort(tmpTab, sortByPoint)
	local removeIdxs = {}
	local removingPoint = CARDS[tmpTab[len]].point
	local removing

	for i=len-1,1,-1 do
		if CARDS[tmpTab[i]].point ~= BASECARDPOINT.TWO 
			and CARDS[tmpTab[i]].typ ~= BASECARDTYPE.JOKER then
			if removing then
				if CARDS[tmpTab[i]].point == removingPoint then
					removing = removing + 1
					if removing == 2 then
						table.insert(removeIdxs, i)
					end
				else
					removingPoint = CARDS[tmpTab[i]].point
					removing = nil
				end
			else
				if CARDS[tmpTab[i]].point ~= removingPoint then
					removingPoint = CARDS[tmpTab[i]].point
					removing = nil
				else
					removing = 1
				end
			end
		end
	end

	return removeIdxs
end

function poker.getSpliteThreeIdx(tmpTab, len, allThreeIdx, idxLen)
	local ret = {}
	local len = len or TableLen(tmpTab)
	if len < 3 then
		return ret
	end

	local idxLen = idxLen or TableLen(allThreeIdx)
	if idxLen < 1 then
		return ret
	end

	table.sort(tmpTab, sortByPoint)
	local maxPoint = CARDS[tmpTab[allThreeIdx[1]]].point
	local lastIdx = 1
	for i=1,idxLen do
		local nextIdx = allThreeIdx[i+1]
		if nextIdx then
			if CARDS[tmpTab[nextIdx]].point ~= (maxPoint - i) then
				for j=lastIdx, i do
					if j ~= lastIdx and j ~= i then
						local tmp = {}
						for get=lastIdx,j do
							table.insert(tmp, allThreeIdx[get])
						end
						table.insert(ret, tmp)
					end
				end
				for j=lastIdx, i do
					if j ~= i then
						local tmp = {}
						for get=j, i do
							table.insert(tmp, allThreeIdx[get])
						end
						table.insert(ret, tmp)
					end
				end
				lastIdx = i
			end
		else
			for j=lastIdx, i do
				if j ~= lastIdx and j ~= i then
					local tmp = {}
					for get=lastIdx,j do
						table.insert(tmp, allThreeIdx[get])
					end
					table.insert(ret, tmp)
				end
			end
			for j=lastIdx, i do
				if j ~= i then
					local tmp = {}
					for get=j, i do
						table.insert(tmp, allThreeIdx[get])
					end
					table.insert(ret, tmp)
				end
			end
		end
	end
	for i,v in ipairs(ret) do
		for j,k in ipairs(v) do
		end
	end
	return ret
end

function poker.isTHREE_STRAIGHT_WITH_SINGLE(cards, len)
	local len = len or TableLen(cards)
	if len < 8 or (len%2) ~= 0then
		return false
	end

	local tmpTab = TableClone(cards)
	local allThreeIdx = poker.getAllThreeIdx(tmpTab, len)
	local allThreeIdxLen = TableLen(allThreeIdx)

	-- simple
	if TableLen(allThreeIdx) > 1 then
		local simpleMaxPoint = CARDS[tmpTab[allThreeIdx[1]]].point
		local hasSimple = true
		for i=1, allThreeIdxLen do
			local nextIdx = allThreeIdx[i+1]
			if nextIdx then
				if CARDS[tmpTab[nextIdx]].point ~= (simpleMaxPoint - i) then
					hasSimple = false
					break
				end
			end
		end
		if hasSimple then
			local point
			for i=1,allThreeIdxLen do
				for _=1,3 do
					if not point then
						point = CARDS[tmpTab[allThreeIdx[i]]].point
					end
					table.remove(tmpTab, allThreeIdx[i])
				end
			end
			local aliveNum = TableLen(tmpTab)
			if aliveNum == allThreeIdxLen then
				return true, point
			end
		end
	else
		-- none THREE_STRAIGHT
		return false
	end

	-- complex
	tmpTab = TableClone(cards)
	local spliteThreeIdx = poker.getSpliteThreeIdx(tmpTab, len, allThreeIdx, allThreeIdxLen)
	local spliteLen = TableLen(spliteThreeIdx)
	if spliteLen > 1 then
		local point
		for i=1,spliteLen do
			tmpTab = TableClone(cards)
			table.sort(tmpTab, sortByPoint)
			local threeIdx = spliteThreeIdx[i]
			local doTimes = 0
			for _,v in ipairs(threeIdx) do
				for _=1,3 do
					if not point then
						point = CARDS[tmpTab[v]].point
					end
					table.remove(tmpTab, v)
				end
				doTimes = doTimes + 1
			end
			local aliveNum = TableLen(tmpTab)
			if aliveNum == doTimes then
				return true, point
			end
			point = nil
		end
	end

	return false
end


function poker.GetCardTypeAndPoint(cards)
	local len = TableLen(cards)

	-- SINGLE finish
	if len == 1 then
		return poker.SINGLE, CARDS[cards[1]].point
	end

	-- DOUBLE finish
	-- JOKERBOOM finish
	if len == 2 then
		if poker.isJOKERBOOM(cards, len) then
			return poker.JOKERBOOM, CARDS[502].point
		end
		if poker.isDOUBLE(cards, len) then
			return poker.DOUBLE, CARDS[cards[1]].point
		end
	end

	-- THREE finish
	if len == 3 then
		if poker.isTHREE(cards, len) then
			return poker.THREE, CARDS[cards[1]].point
		end
	end

	-- BOOM finish
	-- THREE_ONE finish
	if len == 4 then
		if poker.isBOOM(cards, len) then
			return poker.BOOM, CARDS[cards[1]].point
		end

		local is, point = poker.isTHREE_ONE(cards, len)
		if is then
			return poker.THREE_ONE, point
		end
	end

	-- SINGLE_STRAIGHT finish
	if len >= 5 then
		local is, point = poker.isSINGLE_STRAIGHT(cards, len)
		if is then
			return poker.SINGLE_STRAIGHT, point
		end
	end

	-- THREE_TWO finsh
	if len == 5 then
		local is, point = poker.isTHREE_TWO(cards, len)
		if is then
			return poker.THREE_TWO, point
		end
	end

	-- FOUR_WITH_TWO_SINGLE finish
	if len == 6 then
		local is, point = poker.isFOUR_WITH_TWO_SINGLE(cards, len)
		if is then
			return poker.FOUR_WITH_TWO_SINGLE, point
		end
	end

	-- FOUR_WITH_TWO_DOUBLE finish
	if len == 8 then
		local is, point = poker.isFOUR_WITH_TWO_DOUBLE(cards, len)
		if is then
			return poker.FOUR_WITH_TWO_DOUBLE, point
		end
	end

	-- DOUBLE_STRAIGHT finish
	if (len%2) == 0 then
		local is, point = poker.isDOUBLE_STRAIGHT(cards, len)
		if is then
			return poker.DOUBLE_STRAIGHT, point
		end
	end

	-- THREE_STRAIGHT finish
	if (len%3) == 0 then
		local is, point = poker.isTHREE_STRAIGHT(cards, len)
		if is then
			return poker.THREE_STRAIGHT, point
		end
	end

	-- THREE_STRAIGHT_WITH_DOUBLE finish
	if (len >=10) then
		local is, point = poker.isTHREE_STRAIGHT_WITH_DOUBLE(cards, len)
		if is then
			return poker.THREE_STRAIGHT_WITH_DOUBLE, point
		end
	end

	-- THREE_STRAIGHT_WITH_SINGLE finish
	if (len >=8) then
		local is, point = poker.isTHREE_STRAIGHT_WITH_SINGLE(cards, len)
		if is then
			return poker.THREE_STRAIGHT_WITH_SINGLE, point
		end
	end

	return poker.NONE
end

function poker.MaxThen(typ, point, sourceLen, cards)
	if typ == poker.JOKERBOOM then
		return false
	end

	local len = TableLen(cards)
	local t,p = poker.GetCardTypeAndPoint(cards)
	if t == poker.NONE then
		return false
	end
	if t ~= typ then
		if t == poker.BOOM or t == poker.JOKERBOOM then
			return true, t, p, len
		end
	else
		if t == poker.JOKERBOOM then
			return true, t, p, len
		end
		if t == poker.BOOM then
			if p > point then
				return true, t, p, len
			else
				return false
			end
		end
		if sourceLen == len then
			if p > point then
				return true, t, p, len
			end
		end
	end

	return false
end


function test()
	local t,p = poker.GetCardTypeAndPoint({111})
	assert(t == poker.SINGLE, "faild " .. "SINGLE")
	assert(p == 11, "faild " .. "SINGLE")

	local t,p = poker.GetCardTypeAndPoint({111, 311})
	assert(t == poker.DOUBLE, "faild " .. "DOUBLE")
	assert(p == 11, "faild " .. "SINGLE")


	local t,p = poker.GetCardTypeAndPoint({111, 311, 211})
	assert(t == poker.THREE, "faild " .. "THREE")
	assert(p == 11, "faild " .. "THREE")

	local t,p = poker.GetCardTypeAndPoint({111, 311, 211, 411})
	assert(t == poker.BOOM, "faild " .. "BOOM")
	assert(p == 11, "faild " .. "BOOM")

	local t,p = poker.GetCardTypeAndPoint({501, 502})
	assert(t == poker.JOKERBOOM, "faild " .. "JOKERBOOM")
	assert(p == 17, "faild " .. "JOKERBOOM")

	local t,p = poker.GetCardTypeAndPoint({103, 204, 205, 306, 107})
	assert(t == poker.SINGLE_STRAIGHT, "faild " .. "SINGLE_STRAIGHT")
	assert(p == 7, "faild " .. "SINGLE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({103, 204, 205, 306, 107, 408})
	assert(t == poker.SINGLE_STRAIGHT, "faild " .. "SINGLE_STRAIGHT")
	assert(p == 8, "faild " .. "SINGLE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({110, 111, 212, 313, 401})
	assert(t == poker.SINGLE_STRAIGHT, "faild " .. "SINGLE_STRAIGHT")
	assert(p == 14, "faild " .. "SINGLE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({110, 111, 212, 313, 401, 402})
	assert(t == poker.NONE, "faild " .. "SINGLE_STRAIGHT")
	assert(p == nil, "faild " .. "SINGLE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({111, 212, 313, 401})
	assert(t == poker.NONE, "faild " .. "SINGLE_STRAIGHT")
	assert(p == nil, "faild " .. "SINGLE_STRAIGHT")

	local t,p = poker.GetCardTypeAndPoint({211, 111, 212, 312, 413, 213})
	assert(t == poker.DOUBLE_STRAIGHT, "faild " .. "DOUBLE_STRAIGHT")
	assert(p == 13, "faild " .. "DOUBLE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({211, 111, 212, 312, 413, 213, 201, 301})
	assert(t == poker.DOUBLE_STRAIGHT, "faild " .. "DOUBLE_STRAIGHT")
	assert(p == 14, "faild " .. "DOUBLE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({413, 213, 201, 301, 202, 302})
	assert(t == poker.NONE, "faild " .. "DOUBLE_STRAIGHT")
	assert(p == nil, "faild " .. "DOUBLE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({413, 213, 201, 301})
	assert(t == poker.NONE, "faild " .. "DOUBLE_STRAIGHT")
	assert(p == nil, "faild " .. "DOUBLE_STRAIGHT")

	local t,p = poker.GetCardTypeAndPoint({110, 310, 210, 309})
	assert(t == poker.THREE_ONE, "faild " .. "THREE_ONE")
	assert(p == 10, "faild " .. "THREE_ONE")

	local t,p = poker.GetCardTypeAndPoint({110, 310, 210, 309, 409})
	assert(t == poker.THREE_TWO, "faild " .. "THREE_TWO")
	assert(p == 10, "faild " .. "THREE_TWO")
	local t,p = poker.GetCardTypeAndPoint({110, 310, 210, 309, 408})
	assert(t == poker.NONE, "faild " .. "THREE_TWO")
	assert(p == nil, "faild " .. "THREE_TWO")

	local t,p = poker.GetCardTypeAndPoint({110, 310, 210, 410, 309, 409})
	assert(t == poker.FOUR_WITH_TWO_SINGLE, "faild " .. "FOUR_WITH_TWO_SINGLE")
	assert(p == 10, "faild " .. "FOUR_WITH_TWO_SINGLE")
	local t,p = poker.GetCardTypeAndPoint({110, 310, 210, 410, 308, 409})
	assert(t == poker.FOUR_WITH_TWO_SINGLE, "faild " .. "FOUR_WITH_TWO_SINGLE")
	assert(p == 10, "faild " .. "FOUR_WITH_TWO_SINGLE")

	local t,p = poker.GetCardTypeAndPoint({110, 310, 210, 410, 309, 409, 306, 206})
	assert(t == poker.FOUR_WITH_TWO_DOUBLE, "faild " .. "FOUR_WITH_TWO_DOUBLE")
	assert(p == 10, "faild " .. "FOUR_WITH_TWO_DOUBLE")

	local t,p = poker.GetCardTypeAndPoint({110, 310, 210, 309, 409, 209, 211, 111, 311})
	assert(t == poker.THREE_STRAIGHT, "faild " .. "THREE_STRAIGHT")
	assert(p == 11, "faild " .. "THREE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({310, 410, 210, 211, 111, 311})
	assert(t == poker.THREE_STRAIGHT, "faild " .. "THREE_STRAIGHT")
	assert(p == 11, "faild " .. "THREE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({313, 413, 213, 201, 101, 301})
	assert(t == poker.THREE_STRAIGHT, "faild " .. "THREE_STRAIGHT")
	assert(p == 14, "faild " .. "THREE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({302, 402, 202, 201, 101, 301})
	assert(t == poker.NONE, "faild " .. "THREE_STRAIGHT")
	assert(p == nil, "faild " .. "THREE_STRAIGHT")
	local t,p = poker.GetCardTypeAndPoint({305, 405, 205, 207, 107, 307})
	assert(t == poker.NONE, "faild " .. "THREE_STRAIGHT")
	assert(p == nil, "faild " .. "THREE_STRAIGHT")

	local t,p = poker.GetCardTypeAndPoint({103, 203, 208, 108, 312, 212, 110, 310, 210, 309, 409, 209, 211, 111, 311})
	assert(t == poker.THREE_STRAIGHT_WITH_DOUBLE, "faild " .. "THREE_STRAIGHT_WITH_DOUBLE")
	assert(p == 11, "faild " .. "THREE_STRAIGHT_WITH_DOUBLE")
	local t,p = poker.GetCardTypeAndPoint({103, 203, 208, 108, 307, 207, 113, 313, 213, 301, 401, 201, 212, 112, 312})
	assert(t == poker.THREE_STRAIGHT_WITH_DOUBLE, "faild " .. "THREE_STRAIGHT_WITH_DOUBLE")
	assert(p == 14, "faild " .. "THREE_STRAIGHT_WITH_DOUBLE")
	local t,p = poker.GetCardTypeAndPoint({101, 201, 301, 102, 202, 302, 113, 213, 313, 309, 409, 207, 307, 111, 311})
	assert(t == poker.NONE, "faild " .. "THREE_STRAIGHT_WITH_DOUBLE")
	assert(p == nil, "faild " .. "THREE_STRAIGHT_WITH_DOUBLE")

	local t,p = poker.GetCardTypeAndPoint({101, 201, 301, 113, 213, 313, 309, 409})
	assert(t == poker.THREE_STRAIGHT_WITH_SINGLE, "faild " .. "THREE_STRAIGHT_WITH_SINGLE")
	assert(p == 14, "faild " .. "THREE_STRAIGHT_WITH_SINGLE")
	local t,p = poker.GetCardTypeAndPoint({101, 201, 301, 401, 113, 213, 313, 112, 212, 312, 111, 211, 311, 110, 210, 310})
	assert(t == poker.THREE_STRAIGHT_WITH_SINGLE, "faild " .. "THREE_STRAIGHT_WITH_SINGLE")
	assert(p == 14, "faild " .. "THREE_STRAIGHT_WITH_SINGLE")

end

-- test()

























return poker