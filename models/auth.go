package models

import (
	"github.com/beego/beego/v2/client/httplib"
	"github.com/buger/jsonparser"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
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
	log.Info(Config.Master)
	post.Param("uid", strconv.FormatInt(Config.QQGroupID, 10))
	post.Bytes()
}

func getAuthFlag() {
	post := httplib.Post("http://auth.smxy.xyz/user/authFlag")
	post.Param("qqNum", strconv.FormatInt(Config.QQID, 10))
	s, _ := post.Bytes()
	getString, _ := jsonparser.GetString(s, "data")
	log.Info(getString)
}
