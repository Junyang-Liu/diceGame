
function initLobby()
    print("initLobby")

    require("testLuaLobby.Player")
    require("testLuaLobby.Room")


    print("rdb.Set:", rdb.Set("key_test", "val_test"))
    print("rdb.Get:", rdb.Get("key_test"))
    print("rdb.Get:", rdb.Get("key_test_not_exist"))
end

initLobby()