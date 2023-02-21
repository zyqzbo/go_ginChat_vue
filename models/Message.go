package models

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type Message struct { // 消息字段
	gorm.Model
	FormId   int64  // 发送者
	TargetId int64  // 接受者
	Type     int    // 发送类型 群聊 广播
	Media    int    // 消息类型 表情包 文字 图片 音频
	Content  string // 消息内容
	Pic      string
	Url      string
	Desc     string
	Amount   int // 其他数字统计
}

func (table *Message) TableName() string { // 设置数据库表的名字
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
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

	sendMsg(userId, []byte("欢迎进入聊天系统"))
}

func sendProc(node *Node) {
	for {
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
	defer con.Close()
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
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("udpRecvProc data: ", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg) // 转json串
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("msg.Type:", msg.Type)
	switch msg.Type {
	case 1: // 私信
		//fmt.Println("dispatch data:", string(data))
		sendMsg(msg.TargetId, data)
		//case 2: // 群发
		//	sendGroupMsg()
		//case 3: // 广播
		//	sendAllMsg()
		//case 4:
		//
	}
}

func sendMsg(userId int64, msg []byte) {
	fmt.Println("sendMsg >>> userID:", userId, "msg", string(msg))
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock() // 解锁
	if ok {
		node.DataQueue <- msg
	}
}
