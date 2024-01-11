 APP=GameEngine
 luaFile=./luaScript/luaFile.go
 luaFilePath=./luascript

build: clean gen
	go build -o ${APP} main.go

run: gen
	go run -race main.go

clean:
	go clean

gen:
	echo "package luascript\n" > ${luaFile}

	echo "var Player=\`" >> ${luaFile}
	cat ${luaFilePath}/player.lua >> ${luaFile}
	echo "\`\n" >> ${luaFile}

	echo "var Room=\`" >> ${luaFile}
	cat ${luaFilePath}/room.lua >> ${luaFile}
	echo "\`\n" >> ${luaFile}

	echo "var Dice=\`" >> ${luaFile}
	cat ${luaFilePath}/dice.lua >> ${luaFile}
	echo "\`\n" >> ${luaFile}