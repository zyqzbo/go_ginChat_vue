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

func AddFriend(userId uint, targetId uint) (int, string) {
	user := UserBasic{}
	if targetId != 0 {
		user = FindById(targetId)
		if user.Salt != "" {
			if userId == user.ID {
				return -1, "不能添加自己"
			}
			contact0 := Contact{}
			utils.DB.Where("owner_id = ? and target_id = ? and type = 1", userId, targetId).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "不能重复添加"
			}
			//在事务提交或回滚之前使用提供的上下文。
			//如果上下文被取消，则 sql 包将回滚事务。
			//如果取消提供给BeginTx 的上下文，Tx.Commit 将返回错误。
			tx := utils.DB.Begin() // 添加事务 解决双向添加好友，如果一方添加失败会自动回滚
			// 事务一旦开始, 不论什么异常最终都会Rollback
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()
			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetId
			contact.OwnerId = userId
			contact.Type = 1
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			contact1 := Contact{}
			contact1.OwnerId = targetId
			contact1.TargetId = userId
			contact1.Type = 1
			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			tx.Commit()
			return 0, "添加好友成功"
		}
		return -1, "没有找到此用户"
	}
	return -1, "好友id不能为空"
}
