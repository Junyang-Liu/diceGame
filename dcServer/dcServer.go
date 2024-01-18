package dcserver

import (
	"diceGame/config"
	"diceGame/db"
	"diceGame/utils"

	"github.com/gin-gonic/gin"

	"strconv"
	"time"

	Lutils "diceGame/utils"
)

var DcGIN *gin.Engine

func InitServer() {
	initModels()
	initDcHttpServer()
	if config.CFG.Model == "debug" {
		initSomeTestData()
	}
}
func initSomeTestData() {
	utils.Logger.Debug("initSomeTestData")
	uid := 100000
	for {
		uid++
		utils.Logger.Debugf("uid:%d", uid)
		err := db.MysqlConn.FirstOrCreate(&User{Id: uid}).Error
		if err != nil {
			utils.Logger.Warn(err)
		}
		if uid == 100010 {
			break
		}
	}
	err := db.MysqlConn.FirstOrCreate(&Lobby{Id: 11}).Error
	if err != nil {
		utils.Logger.Warn(err)
	}
}

func initModels() {
	err := db.MysqlConn.AutoMigrate(
		&User{},
		&Lobby{},
	)
	if err != nil {
		utils.Logger.Error(err)
	}
}

func initDcHttpServer() {
	ginEN := gin.Default()
	ginEN.Use(gin.Recovery())

	groupInternal := ginEN.Group("/dc")
	{
		groupInternal.Use(InternalMiddleware())

		userRouter := groupInternal.Group("/user")
		{
			userRouter.GET("/test", userTest)
			userRouter.GET("/:uid", userGet)
		}
	}
	go ginEN.Run(config.CFG.DC.Addr)
}

func InternalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s, _ := c.GetQuery("sign")
		t, _ := c.GetQuery("time")
		if s == "" || t == "" {
			ret := map[string]any{
				"msg":  "not sign or not time",
				"code": -1,
			}
			c.JSON(403, ret)
			c.AbortWithStatus(403)
			return
		}

		timestamp, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			ret := map[string]any{
				"msg":  "time not right",
				"code": -1,
				"err":  err.Error(),
			}
			c.JSON(403, ret)
			c.AbortWithStatus(403)
			return
		}

		paraTime := time.Unix(timestamp, 0).UTC()
		now := time.Now()
		utils.Logger.Debugf("paraTime:%s, now:%s", paraTime, now)
		if now.After(paraTime) {
			ret := map[string]any{
				"msg":  "time expire",
				"code": -1,
			}
			c.JSON(403, ret)
			c.AbortWithStatus(403)
			return
		}

		secret := config.CFG.DC.Secret
		token := utils.GenToken("/dc", t, secret)
		Lutils.Logger.Debugf("s:%s, md5Byte:%s", s, token)
		if s != token {
			ret := map[string]any{
				"msg":  "sign not right",
				"code": -1,
			}
			c.JSON(403, ret)
			c.AbortWithStatus(403)
			return
		}

		c.Next()
	}
}

func userTest(c *gin.Context) {
	c.JSON(200, "{}")
}

func userGet(c *gin.Context) {
	utils.Logger.Debug("userGet")
	uid, _ := strconv.Atoi(c.Param("uid"))
	user := User{Id: uid}
	err := db.MysqlConn.First(&user).Error
	if err != nil {
		utils.Logger.Debug(err)
		ret := map[string]any{
			"msg":  "not find one",
			"err":  err.Error(),
			"code": -1,
		}
		c.JSON(200, ret)
		return
	}
	if user.Id == 0 {
		ret := map[string]any{
			"msg":  "not find one",
			"code": -1,
		}
		c.JSON(200, ret)
		return
	}
	c.JSON(200, user)
}
