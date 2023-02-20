package models

import (
	"goChat/utils"
	"gorm.io/gorm"
)

type Contact struct { // 人员关系
	gorm.Model
	OwnerId  uint // 谁的关系信息 （本人的id）
	TargetId uint // 对应的谁 (对应的好友id标识，是OwnerId的好友)
	Type     int  // 对应的类型 1 2 3
	Desc     string
}

func (table *Contact) TableName() string { // 设置数据库表的名字
	return "contact"
}

func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type = 1", userId).Find(&contacts) // 通过这两个字段关联 来查询
	for _, val := range contacts {
		//fmt.Println("<<<<<<<<<<<<<", val)
		objIds = append(objIds, uint64(val.TargetId))
	}

	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users) // 通过多个的id来查询对应的用户并且返回
	return users
}
