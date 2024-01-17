package dcserver

import "gorm.io/gorm"

type Lobby struct {
	Id    int    `gorm:"primary_key"`
	Token string `gorm:"token"`
}

func (Lobby) TableName() string {
	return "lobby"
}

func (this *Lobby) Delete(db *gorm.DB) error {
	return nil
}
