package models

import (
	"github.com/beego/beego/v2/client/httplib"
	"github.com/robfig/cron/v3"
	"strconv"
)

func Auth() {
	c = cron.New()

	c.AddFunc("0 8-20/5 * * ?", GetAuthKey)

	c.Start()
}

func GetAuthKey() {
	//String qqNum,String master,String uid
	post := httplib.Post("http://auth.smxy.xyz/user/auth1")
	post.Param("qqNum", strconv.FormatInt(Config.QQID, 10))
	post.Param("master", Config.Master)
	post.Param("uid", string(Config.QQGroupID))
	post.Bytes()
}
