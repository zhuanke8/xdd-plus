package models

import (
	"github.com/beego/beego/v2/client/httplib"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

var test2 = func(string) {

}

func init() {
	killp()
	for _, arg := range os.Args {
		if arg == "-d" {
			Daemon()
		}
	}
	ExecPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	logs.Info("当前%s", ExecPath)
	initConfig()
	initNolan()
	initDB()
	go initVersion()
	//go initUserAgent()
	initContainer()
	initHandle()
	initCron()
	go initTgBot()
	InitReplies()
	initTask()
	//initNolan()
	//initRepos()
	intiSky()
}

func initNolan() {
	s, _ := httplib.Get("https://update.smxy.xyz/qq.txt").String()
	logs.Info(s)
	logs.Info(Config.QQID)
	if strings.Contains(s, string(Config.QQID)) {
		Config.VIP = true
		logs.Info("VIP验证成功")
	}

}
