package main

import (
	"github.com/gin-gonic/gin"
	"goChat/service"
)

func GetRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/index", service.GetIndex)

	return r
}
