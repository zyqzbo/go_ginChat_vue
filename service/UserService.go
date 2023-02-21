package service

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"goChat/models"
	"goChat/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// @Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code", "message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()
	c.JSON(200, gin.H{
		"code":    0, // 0：成功 -1：失败
		"data":    data,
		"message": "获取列表成功",
	})
}

// swagger的使用

// CreateUser
// @Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确定密码"
// @param phone query string false "手机号码"
// @param email query string false "邮箱"
// @Success 200 {string} json{"code", "message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	// 通过Query方法获取参数
	//user.Name = c.Query("name")
	//password := c.Query("password")
	//repassword := c.Query("repassword")
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("repassword")
	user.Phone = c.Query("phone")
	user.Email = c.Query("email")
	salt := fmt.Sprintf("%06d", rand.Int31()) // 生成随机数
	//data := models.FindUserByName(user.Name)
	//fmt.Println(user.Name, "<<<<<<<<<", password, repassword)
	//data = models.FindUserByPhone(user.Phone)
	//data = models.FindUserByEmail(user.Email)
	//fmt.Println("data.Nam:", data.Name)
	////fmt.Println("user.Name", user.Name)
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{
			"code":    0, // 0：成功 -1：失败
			"data":    user,
			"message": "用户已注册",
		})
		return
	}
	//if data.Phone != "" {
	//	c.JSON(-1, gin.H{
	//		"message": "该号码用户已注册！",
	//	})
	//	return
	//}
	//if data.Email != "" {
	//	c.JSON(-1, gin.H{
	//		"message": "该号码用户已注册！",
	//	})
	//	return
	//}
	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1, // 0：成功 -1：失败
			"data":    user,
			"message": "两次密码不一致",
		})
		return
	}

	//user.Password = password
	user.Password = utils.MakePassword(password, salt) // 给密码加入随机数进行加密
	user.Salt = salt
	models.CrateUser(user)
	c.JSON(200, gin.H{
		"code":    0, // 0：成功 -1：失败
		"data":    user,
		"message": "新增用户成功",
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @param id query string false "id"
// @Success 200 {string} json{"code", "message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id")) // 把uint类型的id转换成字符串类型
	user.ID = uint(id)                   // 添加到数据库的时候再把类型转回来
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0, // 0：成功 -1：失败
		"data":    user,
		"message": "删除用户成功",
	})
}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code", "message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id")) // 把uint类型的id转换成字符串类型
	user.ID = uint(id)                      // 添加到数据库的时候再把类型转回来
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")

	_, err := govalidator.ValidateStruct(user) // govalidator 的结构体验证方法
	if err != nil {
		c.JSON(200, gin.H{
			"code":    -1, // 0：成功 -1：失败
			"data":    user,
			"message": "修改参数不匹配",
		})
		return
	}
	models.UpdateUser(user)
	c.JSON(200, gin.H{
		"code":    0, // 0：成功 -1：失败
		"data":    user,
		"message": "修改用户成功",
	})
}

// 登陆
// @Summary 登陆
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code", "message"}
// @Router /user/login [post]

func Login(c *gin.Context) {
	data := models.UserBasic{}
	//name := c.Query("name")
	//password := c.Query("password")
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	user := models.FindUserByName(name)
	fmt.Println("user:", &user)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0：成功 -1：失败
			"data":    data,
			"message": "该用户不存在",
		})
		return
	}

	validPassword := utils.ValidPassword(password, user.Salt, user.Password) // 校验密码
	if !validPassword {
		c.JSON(200, gin.H{
			"code":    -1, // 0：成功 -1：失败
			"data":    data,
			"message": "密码错误",
		})
		return
	}

	pwd := utils.MakePassword(password, user.Salt)
	data = models.FindUserByNameAndPwd(name, pwd)
	c.JSON(http.StatusOK, gin.H{
		"code":    0, // 0：成功 -1：失败
		"data":    data,
		"message": "登陆成功",
	})
}

// 防止跨域的虚请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) { // 用redis发送消息
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(ws)
	MsgHandler(ws, c)
}
func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	msg, err := utils.Subscribe(c, utils.PublishKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("发送消息：", msg)
	tm := time.Now().Format("2006-01-02 15:04:05")
	m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
	ws.WriteMessage(1, []byte(m))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func SendUserMsg(c *gin.Context) { // 发送消息
	models.Chat(c.Writer, c.Request)
}

func SearchFriends(c *gin.Context) { // 发送消息
	id, _ := strconv.Atoi(c.Request.FormValue("userId")) // 获取参数欧转换成uint类型
	users := models.SearchFriend(uint(id))

	//c.JSON(200, gin.H{
	//	"code":    0, // 0：成功 -1：失败
	//	"data":    users,
	//	"message": "查询好友列表成功",
	//})
	utils.RespOKList(c.Writer, users, len(users))
}

func AddFriend(c *gin.Context) { // 添加好友
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))     // 获取参数欧转换成uint类型
	targetId, _ := strconv.Atoi(c.Request.FormValue("targetId")) // 获取参数欧转换成uint类型
	code, msg := models.AddFriend(uint(userId), uint(targetId))

	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}

}
