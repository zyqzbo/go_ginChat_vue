package main

import (
	"goChat/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:zyq4836..@tcp(127.0.0.1:3306)/ginchat_db?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("err:" + err.Error())
	}
	//db.AutoMigrate(&models.UserBasic{})
	db.AutoMigrate(&models.Message{})
	db.AutoMigrate(&models.Contact{})
	db.AutoMigrate(&models.GroupBasic{})

	//user := &models.UserBasic{}
	//user.Name = "zyq"
	//db.Create(user)
	//
	//db.Model(user).Update("Password", "1234")
}
