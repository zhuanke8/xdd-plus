package models

import (
	"errors"
	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	"os"
	"regexp"
	"strings"
)

var version = "v4.0"
var describe = "情人节特别版"
var AppName = "xdd"
var pname = regexp.MustCompile(`/([^/\s]+)`).FindStringSubmatch(os.Args[0])[1]

func initVersion() {
	Config.Version = version
}

func Update(sender *Sender) error {
	logs.Info("检查更新" + version)
	sender.Reply("小滴滴开始检查更新")
	value, err := httplib.Get("http://xdd.smxy.xyz/version").String()
	if err != nil {
		return errors.New("获取版本号失败")
	} else {
		if strings.Contains(Config.Version, value) {
			return errors.New("小滴滴已是最新版啦")
		} else {
			sender.Reply("小滴滴开始更新程序")

		}
		return nil
	}
}
