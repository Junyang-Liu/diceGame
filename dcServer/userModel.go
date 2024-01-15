package dcserver

import "gorm.io/gorm"

type User struct {
	Id    int    `gorm:"primary_key"`
	Name  string `gorm:"name"`
	Coins int    `gorm:"coins"`
}

func (User) TableName() string {
	return "user"
}

func (this *User) Delete(db *gorm.DB) error {
	return nil
}
