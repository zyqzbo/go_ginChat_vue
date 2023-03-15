# go_ginChat_vue
聊天社交通讯项目，前后端不分离，前端用的是vue2+

配置环境：

go mod init go_ginChat_vue


运行代码：

打包main并且运行生成的main文件：

go build main.go && ./main

运行项目：

go build main.go


刷新go.mod的依赖目录：

go mod tidy

依赖包：

go get github.com/spf13/viper


swag：是用来前后端调试api接口的：

go get -u github.com/swaggo/swag/cmd/swag

修改了代码之后初始化一下才能看到修改的效果：

swag init

go get -u github.com/swaggo/gin-swagger

go get -u github.com/swaggo/files


Md5直接引入的 go包里面的

项目Redis 的引入：

go mod init github.com/my/repo

go get github.com/go-redis/redis/v8

本机下载redis：

brew install redis

开启redis服务：

redis-server

通讯websocket：

go get github.com/gorilla/websocket

set：

go get gopkg.in/fatih/set.v0
