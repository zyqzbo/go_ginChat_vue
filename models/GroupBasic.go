package models

import "gorm.io/gorm"

type GroupBasic struct {
	gorm.Model
	Name    string
	OwnerId uint   // 谁的关系信息
	Icon    string // 对应的谁
	Type    int    // 对应的类型 0 1 3
	Desc    string
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}
