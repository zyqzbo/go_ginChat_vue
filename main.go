package main

import (
	"goChat/Router"
	"goChat/utils"
)

func main() {
	utils.IntConfig()
	utils.InitMysSQL()
	utils.InitRedis()

	r := router.GetRouter()
	r.Run() //  如果8080端口被占用可以改为：r.Run("8081")
}
