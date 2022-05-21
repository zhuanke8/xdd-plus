package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"

	"github.com/beego/beego/v2/server/web"
	"github.com/cdle/xdd/controllers"
	"github.com/cdle/xdd/models"
	"github.com/cdle/xdd/qbot"
)

var theme = ""

func main() {
	go func() {
		models.Save <- &models.JdCookie{}
	}()
	web.Get("/count", func(ctx *context.Context) {
		ctx.WriteString(models.Count())
	})
	web.Get("/", func(ctx *context.Context) {
		if models.Config.Theme == "" {
			models.Config.Theme = models.GhProxy + "https://ghproxy.com/https://raw.githubusercontent.com/764763903a/xdd-plus/main/theme/admin.html"
		}
		if theme != "" {
			ctx.WriteString(theme)
			return
		}
		if strings.Contains(models.Config.Theme, "http") {
			logs.Info("下载最新主题")
			s, _ := httplib.Get(models.Config.Theme).String()
			if s != "" {
				theme = s
				ctx.WriteString(s)
				return
			}
			logs.Warn("主题下载失败，使用默认主题")
		}
		f, err := os.Open(models.Config.Theme)
		if err == nil {
			d, _ := ioutil.ReadAll(f)
			theme = string(d)
			ctx.WriteString(string(d))
			return
		}
	})

	web.Router("/api/login/qrcode", &controllers.LoginController{}, "get:GetQrcode")
	web.Router("/api/login/qrcode.png", &controllers.LoginController{}, "get:GetQrcode")
	web.Router("/api/login/qrcode1", &controllers.LoginController{}, "get:GetQrcode1")
	web.Router("/api/login/query", &controllers.LoginController{}, "get:Query")
	web.Router("/api/login/cookie", &controllers.LoginController{}, "get:Cookie")
	web.Router("/api/login/admin", &controllers.LoginController{}, "post:IsAdmin")
	web.Router("/api/login/cklogin", &controllers.LoginController{}, "post:CkLogin")
	web.Router("/api/login/smslogin", &controllers.LoginController{}, "post:SMSLogin")
	web.Router("/api/account", &controllers.AccountController{}, "get:List")
	web.Router("/api/account", &controllers.AccountController{}, "post:CreateOrUpdate")
	web.Router("/admin", &controllers.AccountController{}, "get:Admin")
	web.Router("/admin", &controllers.AccountController{}, "post:Admin")

	if models.Config.Static == "" {
		models.Config.Static = "./static"
	}
	web.BConfig.WebConfig.StaticDir["/static"] = models.Config.Static
	web.BConfig.AppName = models.AppName
	web.BConfig.WebConfig.AutoRender = false
	web.BConfig.CopyRequestBody = true
	web.BConfig.WebConfig.Session.SessionOn = true
	web.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600
	web.BConfig.WebConfig.Session.SessionName = models.AppName
	go func() {
		time.Sleep(time.Second * 4)
		(&models.JdCookie{}).Push(fmt.Sprintf("小滴滴已启动，版本号:%s", models.Config.Version))

	}()
	if models.Config.QQID != 0 && models.Config.OpenQQ == "" {
		go qbot.Main()
	} else {
		logs.Info("不启动QQ")
	}
	web.Run()
}
