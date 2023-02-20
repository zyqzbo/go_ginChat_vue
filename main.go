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
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")  如果改的话可以：r.Run("8081")
}
