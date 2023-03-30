package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"goChat/docs"
	"goChat/service"
)

func GetRouter() *gin.Engine {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // ginSwagger的使用 在swagger后面的所有路径

	// 静态资源
	r.Static("asset", "asset")
	r.LoadHTMLGlob("views/**/*") // gin框架的LoadHTMLGlob去扫描html网友模版

	// 首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	r.POST("/searchFriends", service.SearchFriends)

	// 用户模块
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/login", service.Login)

	// 发送消息
	r.GET("/user/sendMsg", service.SendMsg)
	// 发送消息
	r.GET("/user/sendUserMsg", service.SendUserMsg)
	// 上传文件
	r.POST("/attach/upload", service.Upload)
	// 添加好友
	r.POST("/contact/addFriend", service.AddFriend)
	// 创建群
	r.POST("/contact/createCommunity", service.CreateCommunity)
	// 加载群列表
	r.POST("/contact/loadCommunity", service.LoadCommunity)
	// 加群
	r.POST("contact/joinGroup", service.JoinGroup)
	// 加载redis缓存
	r.POST("/user/redisMsg", service.RedisMsg)
	return r
}
