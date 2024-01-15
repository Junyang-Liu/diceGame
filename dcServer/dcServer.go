package dcserver

import (
	"diceGame/config"
	"diceGame/db"
	"diceGame/utils"
	"fmt"

	"github.com/gin-gonic/gin"

	"crypto/md5"
	"strconv"
	"time"

	Lutils "diceGame/utils"
)

var DcGIN *gin.Engine

func InitServer() {
	initModels()
	initDcHttpServer()
}

func initModels() {
	db.MysqlConn.AutoMigrate(
		&User{},
		&Lobby{},
	)
}

func initDcHttpServer() {
	ginEN := gin.Default()
	ginEN.Use(gin.Recovery())

	groupInternal := ginEN.Group("/dc")
	{
		groupInternal.Use(internalMiddleware())

		userRouter := groupInternal.Group("/user")
		{
			userRouter.GET("/test", userTest)
		}
	}
	go ginEN.Run(config.CFG.DC.Addr)
}

func internalMiddleware() gin.HandlerFunc {
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
				"err":  err.Error(),
			}
			c.JSON(403, ret)
			c.AbortWithStatus(403)
			return
		}

		secret := config.CFG.DC.Secret
		source := []byte(fmt.Sprintf("/dc%s%s", t, secret))
		md5Byte := md5.Sum(source)
		Lutils.Logger.Debugf("s:%s, md5Byte:%x", s, md5Byte[:])
		if s != fmt.Sprintf("%x", md5Byte[:]) {
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
