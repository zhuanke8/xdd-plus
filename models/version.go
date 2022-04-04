package models

import (
	"errors"
	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var version = "v3.2"
var describe = "最终稳定版"
var AppName = "xdd"
var pname = regexp.MustCompile(`/([^/\s]+)`).FindStringSubmatch(os.Args[0])[1]

func initVersion() {
	Config.Version = version
	//logs.Info("检查更新" + version)
	//value, err := httplib.Get(GhProxy + "https://raw.githubusercontent.com/764763903a/xdd-plus/main/models/version.go").String()
	//if err != nil {
	//	logs.Info("更新版本的失败")
	//} else {
	//	// name := AppName + "_" + runtime.GOOS + "_" + runtime.GOARCH
	//	if match := regexp.MustCompile(`var version = "(\d{10})"`).FindStringSubmatch(value); len(match) != 0 {
	//		des := regexp.MustCompile(`var describe = "([^"]+)"`).FindStringSubmatch(value)
	//		if len(des) != 0 {
	//			describe = des[1]
	//		}
	//		if match[1] > version {
	//			err := Update(&Sender{})
	//			if err != nil {
	//				logs.Warn("更新失败,", err)
	//				return
	//			}
	//			(&JdCookie{}).Push("小滴滴更新：" + describe)
	//			Daemon()
	//		}
	//	}
	//}
}

func Exists(path string) bool {

	_, err := os.Stat(path) //os.Stat获取文件信息

	if err != nil {

		if os.IsExist(err) {

			return true

		}

		return false

	}

	return true

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
			//检查更新文件是否存在
			exists := Exists(ExecPath + "/run.sh")
			if exists {
				rtn, err := exec.Command("sh", "-c", "."+ExecPath+" /run.sh").Output()
				if err != nil {
					return errors.New("小滴滴拉取代码失败：" + err.Error())
				}
				t := string(rtn)
				if !strings.Contains(t, "错误") {
					sender.Reply("小滴滴拉取代码成功")
					os.Chmod(ExecPath+"/xdd", 0777)
				}
			} else {
				return errors.New("更新文件不存在")
			}
		}
		return nil
	}
}
