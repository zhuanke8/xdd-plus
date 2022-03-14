package models

import (
	"github.com/beego/beego/v2/adapter/logs"
	"github.com/robfig/cron/v3"
)

var c *cron.Cron

func initCron() {
	c = cron.New()
	if Config.DailyAssetPushCron != "" {
		_, err := c.AddFunc(Config.DailyAssetPushCron, DailyAssetsPush)
		if err != nil {
			logs.Warn("资产推送任务失败：%v", err)
		} else {
			logs.Info("资产推送任务就绪")
		}

		c.AddFunc(Config.DailyCompletePush, CompletePush)
		//c.AddFunc("3 */1 * * *", initVersion)
		//c.AddFunc("40 */1 * * *", GitPullAll)
	}
	//logs.Info("0 " + strconv.Itoa(rand.Intn(59)) + " 0/" + strconv.Itoa(Config.Later) + " * * ?" + "调试推送时间")
	c.AddFunc("0 8-20/5 * * ?", initCookie)

	c.Start()
}
