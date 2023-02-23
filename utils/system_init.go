package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var (
	DB  *gorm.DB
	Rdb *redis.Client
)
var ctx = context.Background()// 返回一个go的空上下文,作为redis的初始化时的参数使用

func IntConfig() {
	viper.SetConfigName("app")    // 扫描文件名
	viper.AddConfigPath("config") // 扫描的文件夹
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("config app", viper.Get("app"))
	fmt.Println("config mysql", viper.Get("mysql"))
}

func InitMysSQL() {
	// 自定义日志模版 打印SQL语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // 慢SQL阀值
			LogLevel:      logger.Info, // 级别
			Colorful:      true,        // 彩色
		},
	)

	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	//if err != nil {
	//	panic("err:" + err.Error())
	//}

	//user := models.UserBasic{}
	//DB.Find(&user)
	//fmt.Println("user", user)
}

func InitRedis() { // redis的连接
	Rdb = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.DB"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConn"),
	})
	//result, err := Red.Ping().Result()
	err := Rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := Rdb.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)
}

const (
	PublishKey = "websocket"
)

// Publish 发布消息到Redis
func Publish(ctx context.Context, channel string, msg string) error {
	var err error
	fmt.Println("Publish", msg)
	err = Rdb.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return err
}

// Subscribe 订阅Redis 消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Rdb.Subscribe(ctx, channel)
	fmt.Println("Subscribe ctx", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Subscribe", msg.Payload)
	return msg.Payload, err //msg.Payload 把消息转为字符串
}
