package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
)

func GetLocalMac() (mac string) {
	// 获取本机的MAC地址
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}
	for _, inter := range interfaces {
		fmt.Println(inter.Name)
		mac := inter.HardwareAddr //获取本机MAC地址
		fmt.Println("MAC = ", mac)
	}
	return GetAuthKey(mac)
}

func GetAuthKey(token string) (key string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(token))
	data := md5Ctx.Sum(nil)
	return hex.EncodeToString(data)
}
