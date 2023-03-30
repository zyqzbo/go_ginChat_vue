package models

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"goChat/utils"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Message struct { // 消息字段
	gorm.Model
	UserId     int64  // 发送者x
	TargetId   int64  // 接受者
	Type       int    // 发送类型 群聊 广播
	Media      int    // 消息类型 表情包 文字 图片 音频
	Content    string // 消息内容
	CreateTime uint64 // 创建时间
	ReadTime   uint64 // 读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int // 其他数字统计
}

func (table *Message) TableName() string { // 设置数据库表的名字
	return "message"
}

type Node struct { // 创建一个node节点去负责管理所有容器、监控/上报所有Pod的运行状态。
	Conn          *websocket.Conn //为客户端与服务器建立的TCP链接
	Addr          string          //客户端地址
	FirstTime     uint64          //首次连接时间
	HeartbeatTime uint64          //心跳时间
	LoginTime     uint64          //登录时间
	DataQueue     chan []byte     //消息
	GroupSets     set.Interface   //好友 / 群
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// 需要：发送者ID，收受者ID，消息；类型，发送的内容，发送类型

func Chat(writer http.ResponseWriter, request *http.Request) {
	// 1、获取参数 并且 校验token 等合法性
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//context := query.Get("context")
	isValida := true // checkToke() TODO
	conn, err := (&websocket.Upgrader{
		// token 校验
		CheckOrigin: func(r *http.Request) bool {
			return isValida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 2、获取连接Conn
	node := &Node{
		Conn:      conn,
		Addr:      conn.RemoteAddr().String(), //客户端地址
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}
	// 3、用户关系
	// 4、userId 跟 node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//5、完成发送的逻辑uin
	go sendProc(node)
	//6、完成接受的逻辑
	go revProc(node)

	//7.加入在线用户到缓存
	SetUserOnlineInfo("online_"+Id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
	//sendMsg(userId, []byte("欢迎进入聊天系统"))
}

func sendProc(node *Node) {
	for {
		//当前handler阻塞监听管道的消息，一旦两个管道有一个有值，就会执行select
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws] sendProc >>>> msg: ", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data) // 写
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func revProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage() // 读取内容
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(data) // 发送给对方
		broadMsh(data) // todo 将消息广播到局域网
		//fmt.Println("[ws] revProc <<<<", string(data))
	}
}

var updSendChan chan []byte = make(chan []byte, 1024)

func broadMsh(data []byte) {
	updSendChan <- data
}

func init() {
	go updSendProc()
	go udpRecvProc()
	//fmt.Println("init goroutine")
}

// 完成upd数据发送协程
func updSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255), // 路由网关地址
		Port: 3000,
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close() // 发送完之后断开发送连接
	for {
		select {
		case data := <-updSendChan:
			//fmt.Println("udpSendProc data:", string(data))
			_, err := con.Write(data) // 写
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 完成upd数据接受协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero, // 任意ip
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
	for { //循环一直从conn中读取，然后输出到终端
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("udpRecvProc data: ", string(buf[0:n]))
		dispatch(buf[0:n]) // 把消息内容发送到 后端调度做逻辑处理
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{} // 实例化message结构体
	msg.CreateTime = uint64(time.Now().Unix())
	// 把消息内容data 是json字符串类型转为结构体类型（键值对） （json字符串[转]结构体）方便下面获取msg.type的值
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("msg.Type:", msg.Type)
	switch msg.Type {
	case 1: // 私信
		fmt.Println("dispatch data:", string(data))
		sendMsg(msg.TargetId, data) // 传入消息类型、消息内容
	//err := utils.Publish(ctx, utils.PublishKey, "xxx")
	case 2: // 群发
		//sendGroupMsg(msg.TargetId, data) // 发送的群ID，消息内容
		//case 3: // 广播
		//	sendAllMsg()
		//case 4:
		//
	}
}

// 发送消息
func sendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId] // 通过userId来绑定node是谁发的
	rwLocker.RUnlock()            // 解锁
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.Rdb.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err) // 没找到
	}
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> userID:", userId, "msg", string(msg))
			node.DataQueue <- msg
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	res, err := utils.Rdb.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("res:", res)
	score := float64(cap(res)) + 1                                       // 按分数加1（时间去排序）
	data, err := utils.Rdb.ZAdd(ctx, key, &redis.Z{score, msg}).Result() // jsonMsg
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
}

var ctx context.Context

func init() {
	ctx = context.Background()
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息")
	//userIds := SearchUserByGroupId(uint(targetId))
	//for i := 0; i < len(userIds); i++ {
	//	// 排除自己的
	//	if targetId != int64(userIds[i]) {
	//		sendMsg(int64(userIds[i]), msg)
	//	}
	//}
}

func SearchUserByGroupId(u uint) interface{} {
	return nil
}

// JoinGroup 加群
func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	//contact.TargetId = comId
	contact.Type = 2
	community := Community{}

	utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没找到群"
	}
	utils.DB.Where("owner_id=? and target_id=? and type = 2", userId, comId).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功!"
	}
}

//获取缓存里面的消息

func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	//node, ok := clientMap[userIdA]
	rwLocker.RUnlock()
	//jsonMsg := Message{}
	//json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	//key = "msg_" + userIdStr + "_" + targetIdStr
	//rels, err := utils.Rdb.ZRevRange(ctx, key, 0, -1).Result() //根据score倒叙

	//if err != nil {
	//	fmt.Println(err)
	//}
	var rels []string
	var err error
	if isRev {
		rels, err = utils.Rdb.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Rdb.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) //没有找到
	}
	// 发送推送消息
	/**
	// 后台通过websoket 推送消息
	for _, val := range rels {
		fmt.Println("sendMsg >>> userID: ", userIdA, "  msg:", val)
		node.DataQueue <- []byte(val)
	}**/
	return rels
}
