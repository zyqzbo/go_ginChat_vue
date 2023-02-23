package models

import (
	"fmt"
	"goChat/utils"
	"gorm.io/gorm"
	"time"
)

type UserBasic struct {
	gorm.Model
	Name          string
	Password      string
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Identity      string // 唯一表识
	ClientIp      string // 设备id
	ClientPort    string // 客户端口
	Salt          string
	LoginTime     Time   // 登陆时间
	HeartbeatTime Time   // 心跳
	LoginOutTime  Time   `gorm:"column:login_out_time" json:"login_out_time"` // 下线时间
	IsLogout      bool   // 是否下线
	DeviceInfo    string // 设备信息
}

func (table *UserBasic) TableName() string { // 设置数据库表的名字
	return "user_basic"
}

func GetUserList() []*UserBasic { // 加载所有用户 返回一个数组类型的集合体相当于[{},{},{}]
	data := make([]*UserBasic, 10)
	utils.DB.Find(&data)
	return data
}

func CrateUser(user UserBasic) *gorm.DB { // 创建用户
	return utils.DB.Create(&user)
}

func DeleteUser(user UserBasic) *gorm.DB { // 删除用户
	return utils.DB.Delete(&user)
}

func UpdateUser(user UserBasic) *gorm.DB { // 修改用户
	return utils.DB.Model(&user).Updates(UserBasic{
		Name:     user.Name,
		Password: user.Password,
		Phone:    user.Phone,
		Email:    user.Email,
	})
}

// 验证注册的时候重复注册

func FindUserByName(name string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ?", name).First(&user)
	return user
}

func FindUserByPhone(phone string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("phone = ?", phone).First(&user)
	return user
}

func FindUserByEmail(email string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("email = ?", email).First(&user)
	return user
}

// 登陆的时候调用 通过传惨：账号和密码 获取整条数据字段

func FindUserByNameAndPwd(name string, password string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ? and password = ?", name, password).First(&user)
	// token加密
	str := fmt.Sprintf("%d", time.Now().Unix()) // 获取系统当前时间
	temp := utils.MD5Encode(str)
	utils.DB.Model(&user).Where("id = ?", user.ID).Update("identity", temp) // 通过查询id去改变identity的值
	return user
}

func FindById(id uint) UserBasic { //通过id来查找用户
	user := UserBasic{}
	utils.DB.Where("id = ?", id).First(&user)
	return user
}
