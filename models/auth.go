package models

import (
	"crypto/md5"
	"encoding/hex"
	"net"
)

func GetLocalMac() (mac string) {
	// 获取本机的MAC地址
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("设备码获取错误，请联系管理员 " + err.Error())
	}
	mac = string(interfaces[0].HardwareAddr)
	//for _, inter := range interfaces {
	//	fmt.Println(inter.Name)
	//	mac := inter.HardwareAddr //获取本机MAC地址
	//}
	return GetAuthKey(mac)
}

func GetAuthKey(token string) (key string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(token))
	data := md5Ctx.Sum(nil)
	return hex.EncodeToString(data)
}
