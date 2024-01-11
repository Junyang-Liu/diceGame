# diceGame: Distributed architecture game server  <!-- omit from toc -->

diceGame is a Lua game engine, basic Lua VM [gopher-lua](https://github.com/yuin/gopher-lua). It provides Lua APIs that allows you focus on game-play process.
Every single game `Room` owns a Lua VM and player's requests are sent in order to it. As you can see, there's nothing to worry about concurrency or network connectivity.
Check here for [how it works](#how-it-works)

- [Build And Run](#build-and-run)
- [Lua APIs](#lua-apis)
    - [**`Room Create`** lua start](#room-create-lua-start)
    - [**`Room Destroy`** Room:destroy()](#room-destroy-roomdestroy)
    - [**`Player Entry`** Room:PlayerIn(player)](#player-entry-roomplayerinplayer)
    - [**`Player Clear`** Room:PlayerOut(player.id)](#player-clear-roomplayeroutplayerid)
    - [**`Player Request`** Player:OP(line, data)](#player-request-playeropline-data)
    - [**`Player Response`** Player:Send(line, data)](#player-response-playersendline-data)
    - [**`Timer Create`** Room:NewTimer(Millisecond, function, ...)](#timer-create-roomnewtimermillisecond-function-)
    - [**`Timer Exist`** Room:ExistTimer()](#timer-exist-roomexisttimer)
    - [**`Timer Left`** Room:TimerLast()](#timer-left-roomtimerlast)
    - [**`Timer Cancel`** Room:CancelTimer()](#timer-cancel-roomcanceltimer)
    - [**`Timer Create`** Player:NewTimer(Millisecond, function, ...)](#timer-create-playernewtimermillisecond-function-)
    - [**`Timer Exist`** Player:ExistTimer()](#timer-exist-playerexisttimer)
    - [**`Timer Left`** Player:TimerLast()](#timer-left-playertimerlast)
    - [**`Timer Cancel`** Player:CancelTimer()](#timer-cancel-playercanceltimer)
    - [**`lobby StartPlay`** lobby.StartPlay()](#lobby-startplay-lobbystartplay)
    - [**`lobby EndPlay`** lobby.EndPlay(result)](#lobby-endplay-lobbyendplayresult)
- [Configure](#configure)
    - [game server](#game-server)
    - [lobby server](#lobby-server)
    - [data center server](#data-center-server)
- [How it works](#how-it-works)


## Build And Run
```bash
make build
./GameEngine -c ./go.yaml
```


## Lua APIs
A game means a global Lua table `Room` and every player created based on a Lua table `Player` in a Lua Vm. Write `Room:your_function()` or `Player:your_function()` to create your game features.

#### **`Room Create`** lua start
Lua file from the [configure lua start](#configure) path `for example ./testLuaGame/init.lua` will be call when a new game `Room` was created.
A new `Room` always create by a websocket request from lobby server, data :`{"op":"newGame","type":1}`

####  **`Room Destroy`** Room:destroy()
Function `Room:destroy()` for close this game Room. It will call all the [**`Player Clear`** Room:PlayerOut(player.id)](#player-clear-roomplayeroutplayerid) to clear all players in this `Room` before closing the Lua Vm.

#### **`Player Entry`** Room:PlayerIn(player)
Function `Room:PlayerIn(player)` will be call when a new player entry this room. Rewrite it for store the player or anything else. Arg player for a table based on `Player`

#### **`Player Clear`** Room:PlayerOut(player.id)
Function `Room:PlayerOut(player.id)` for clearing a player, will remove the player's data in game engine, including the socket connect, user info cache.

#### **`Player Request`** Player:OP(line, data)
Function `Player:OP(line, data)` will be call when a player game request receive. Arg line for the game protocol; Arg data for the request data.

#### **`Player Response`** Player:Send(line, data)
Function `Player:Send(line, data)` for sending data to the player client. Arg line for the game protocol; Arg data for the request data.

#### **`Timer Create`** Room:NewTimer(Millisecond, function, ...)
Function `Room:NewTimer(Millisecond, function, ...)` for creating a new timer on the Room. Arg Millisecond for the timer runs at millisecond after now; Arg function for the Lua function going to call; Arg ... for the Arg function's call args. A `Room` only cache one timer, a duplicate calling will cover the timer exist.
Same as [**`Timer Create`** Player:NewTimer(Millisecond, function, ...)](#timer-create-playernewtimermillisecond-function)

#### **`Timer Exist`** Room:ExistTimer()
Function `Room:ExistTimer()` returns true when a timer exist, returns false otherwise. Same as [**`Timer Exist`** Player:ExistTimer()](#timer-left-playertimerlast)

#### **`Timer Left`** Room:TimerLast()
Function `Room:TimerLast()` return millisecond left the timer exist this `Room`, returns 0 when there's none timer. Same as [**`Timer Left`** Player:TimerLast()](#timer-left-playertimerlast)

#### **`Timer Cancel`** Room:CancelTimer()
Function `Room:CancelTimer()` to cancel the timer. Same as [**`Timer Cancel`** Player:CancelTimer()](#timer-cancel-playercanceltimer)

#### **`Timer Create`** Player:NewTimer(Millisecond, function, ...)
Function `Player:NewTimer(Millisecond, function, ...)` for creating a new timer on the Room. Arg Millisecond for the timer runs at millisecond after now; Arg function for the Lua function going to call; Arg ... for the Arg function's call args. A `Player` only cache one timer, a duplicate calling will cover the timer exist.
Same as [**`Timer Create`** Room:NewTimer(Millisecond, function, ...)](#timer-create-roomnewtimermillisecond-function)

#### **`Timer Exist`** Player:ExistTimer()
Function `Player:ExistTimer()` returns true when a timer exist, returns false otherwise. Same as [**`Timer Exist`** Room:ExistTimer()](#timer-exist-roomexisttimer)

#### **`Timer Left`** Player:TimerLast()
Function `Room:TimerLast()` return millisecond left the timer exist this `Room`, returns 0 when there's none timer. Same as [**`Timer Left`** Room:TimerLast()](#timer-left-roomtimerlast)

#### **`Timer Cancel`** Player:CancelTimer()
Function `Player:CancelTimer()` to cancel the timer. Same as [**`Timer Cancel`** Room:CancelTimer()](#timer-cancel-roomcanceltimer)

#### **`lobby StartPlay`** lobby.StartPlay()
Function `lobby.StartPlay()` to notice the lobby server this `Room` is start gaming. Only works in game server.

#### **`lobby EndPlay`** lobby.EndPlay(result)
Function `lobby.EndPlay(result)` to notice the lobby server this `Room` is finish gaming, and send json string `result` to lobby. Only works in game server.


## Configure

#### game server
```yml
server:                             # game server config
    addr: :8080                         # listen address
    game_id: 101                        # this game server's id
    priority: 1                         # in same game id, lobby create new room to game server by priority
    lobby_addr: localhost:8081          # lobby server's address

lua:                                # run lua a file when a new game was created by engine
    start: "./testLuaGame/init.lua"     # lua file path to run

model: debug                        # engine run model, create a user when a login request receive,
                                        # instead of getting it from the lobby server if it is set to "debug"
log: info                           # engine log level
```

#### lobby server
```yml
lobby:                             # lobby server config
    addr: :8080                         # listen address
    lobby_id: 11                        # this lobby server's id

redis:                              # engine redis client config
    addr: localhost:6379                # address
    username: default                   # username
    password: ""                        # password
    db: 0                               # db

lua:                                # run lua a file when a new game was created by engine
    start: "./testLuaLobby/init.lua"     # lua file path to run

model: debug                        # engine run model, create a user when a login request receive,
                                        # instead of getting it from the data center server if it is set to "debug"
log: info                           # engine log level
```

#### data center server
```yml
```


## How it works