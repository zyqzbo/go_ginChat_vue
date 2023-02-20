package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func Md5Encode(data string) string { // 小写
	h := md5.New()
	h.Write([]byte(data))
	tempStr := h.Sum(nil)
	return hex.EncodeToString(tempStr)
}

func MD5Encode(data string) string { // 大写
	return strings.ToUpper(Md5Encode(data))
}

func MakePassword(plainpwd, salt string) string { // 随机数加密
	return Md5Encode(plainpwd + salt)
}

func ValidPassword(plainpwd, salt string, password string) bool { // 随机数解密
	return Md5Encode(plainpwd+salt) == password
}
