package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goChat/utils"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

func Upload(c *gin.Context) { // 图片上传
	w := c.Writer
	req := c.Request
	srcFile, header, err := req.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
	}
	suffix := ".png" // 随便初始化一个文件后缀名
	ofilName := header.Filename
	tem := strings.Split(ofilName, ".") // 把图片名字以 . 为界限切割为两份放在一个数组里 [zyq jpg] 索引0为zyq 索引1为.
	if len(tem) > 1 {
		suffix = "." + tem[len(tem)-1] // 相当于截取的是文件后缀的格式
	}
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix) // 当前时间 + 四位数的随机数 + 截取的文件后缀
	dstFile, err := os.Create("./asset/upload/" + fileName)
	if err != nil {
		utils.RespFail(w, err.Error())
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		utils.RespFail(w, err.Error())
	}
	url := "./asset/upload/" + fileName
	utils.RespOK(w, url, "发送图片成功")
}
