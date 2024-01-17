package dcserver

import "gorm.io/gorm"

type User struct {
	Id     int    `gorm:"primary_key" json:"uid"`
	Name   string `gorm:"name" json:"name"`
	ImgUrl string `gorm:"img" json:"img"`
}

func (User) TableName() string {
	return "user"
}

func (this *User) Delete(db *gorm.DB) error {
	return nil
}
