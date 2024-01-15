package db

import (
	"diceGame/config"
	"diceGame/utils"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MysqlConn *gorm.DB

func InitMysql() error {
	if config.CFG.Mysql.Addr == "" {
		utils.Logger.Warn("not find mysql config")
		return nil
	}
	u := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.CFG.Mysql.UserName,
		config.CFG.Mysql.Password,
		config.CFG.Mysql.Addr,
		config.CFG.Mysql.DB,
	)
	utils.Logger.Debug(u)

	db, err := gorm.Open(mysql.Open(u), &gorm.Config{})
	if err != nil {
		utils.Logger.Error(err)
		return err
	}
	sqlDb, err := db.DB()
	if err != nil {
		utils.Logger.Error(err)
		return err
	}

	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)

	if config.CFG.Model == "debug" {
		db.Debug()
	}
	MysqlConn = db
	return nil
}
