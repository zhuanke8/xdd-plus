package models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
)

var SendQQ = func(a int64, b interface{}) {

}
var SendQQGroup = func(a int64, b int64, c interface{}) {

}

var ListenQQPrivateMessage = func(uid int64, msg string) {
	SendQQ(uid, handleMessage(msg, "qq", int(uid)))
}

var ListenQQGroupMessage = func(gid int64, uid int64, msg string) {
	if gid == Config.QQGroupID {
		if Config.QbotPublicMode {
			SendQQGroup(gid, uid, handleMessage(msg, "qqg", int(uid), int(gid)))
		} else {
			SendQQ(uid, handleMessage(msg, "qq", int(uid)))
		}
	}
}

var replies = map[string]string{}

func InitReplies() {
	f, err := os.Open(ExecPath + "/conf/reply.php")
	if err == nil {
		defer f.Close()
		data, _ := ioutil.ReadAll(f)
		ss := regexp.MustCompile("`([^`]+)`\\s*=>\\s*`([^`]+)`").FindAllStringSubmatch(string(data), -1)
		for _, s := range ss {
			replies[s[1]] = s[2]
		}
	}
	if _, ok := replies["壁纸"]; !ok {
		replies["壁纸"] = "https://acg.toubiec.cn/random.php"
	}
}

var handleMessage = func(msgs ...interface{}) interface{} {
	msg := msgs[0].(string)
	args := strings.Split(msg, " ")
	head := args[0]
	contents := args[1:]
	sender := &Sender{
		UserID:   msgs[2].(int),
		Type:     msgs[1].(string),
		Contents: contents,
	}
	if len(msgs) >= 4 {
		sender.ChatID = msgs[3].(int)
	}
	if sender.Type == "tgg" {
		sender.MessageID = msgs[4].(int)
		sender.Username = msgs[5].(string)
		sender.ReplySenderUserID = msgs[6].(int)
	}
	if sender.UserID == Config.TelegramUserID || sender.UserID == int(Config.QQID) {
		sender.IsAdmin = true
	}
	if sender.IsAdmin == false {
		if IsUserAdmin(strconv.Itoa(sender.UserID)) {
			sender.IsAdmin = true
		}
	}
	for i := range codeSignals {
		for j := range codeSignals[i].Command {
			if codeSignals[i].Command[j] == head {
				return func() interface{} {
					if codeSignals[i].Admin && !sender.IsAdmin {
						return "你没有权限操作"
					}
					return codeSignals[i].Handle(sender)
				}()
			}
		}
	}
	switch msg {
	default:
		{ //沃邮箱
			ss := regexp.MustCompile(`https://nyan.mail.*3D`).FindStringSubmatch(msg)
			if len(ss) > 0 {
				var u User
				if db.Where("number = ?", sender.UserID).First(&u).Error != nil {
					return 0
				}
				db.Model(u).Updates(map[string]interface{}{
					"womail": ss[0],
				})
				sender.Reply(fmt.Sprintf("沃邮箱提交成功!"))
				return nil
			}
		}
		{
			if strings.Contains(msg, "口令") {
				rsp := httplib.Post("http://jd.zack.xin/api/jd/ulink.php")
				rsp.Param("url", msg)
				rsp.Param("type", "hy")
				//rsp.Body(fmt.Sprintf(`url=%s&type=hy`, msg))
				data, err := rsp.Response()

				if err != nil {
					return "口令转换失败"
				}
				body, _ := ioutil.ReadAll(data.Body)
				if strings.Contains(string(body), "口令转换失败") {
					return "口令转换失败"
				} else {
					return string(body)
				}
			}
		}
		{
			ss := regexp.MustCompile(`pin=([^;=\s]+);wskey=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)
			if len(ss) > 0 {
				for _, s := range ss {
					wkey := "pin=" + s[1] + ";wskey=" + s[2] + ";"
					//rsp := cmd(fmt.Sprintf(`python3 test.py "%s"`, wkey), &Sender{})
					rsp, err := getKey(wkey)
					if err != nil {
						logs.Error(err)
					}
					if strings.Contains(rsp, "fake_") {
						logs.Error("wskey错误")
						sender.Reply(fmt.Sprintf("wskey错误 除京东APP皆不可用"))
					} else {
						ptKey := FetchJdCookieValue("pt_key", rsp)
						ptPin := FetchJdCookieValue("pt_pin", rsp)
						ck := JdCookie{
							PtPin: ptPin,
							PtKey: ptKey,
							WsKey: s[2],
						}
						if CookieOK(&ck) {

							if sender.IsQQ() {
								ck.QQ = sender.UserID
							} else if sender.IsTG() {
								ck.Telegram = sender.UserID
							}
							if nck, err := GetJdCookie(ck.PtPin); err == nil {
								nck.InPool(ck.PtKey)
								if nck.WsKey == "" || len(nck.WsKey) == 0 {
									if sender.IsQQ() {
										ck.Update(QQ, ck.QQ)
									}
									nck.Update(WsKey, ck.WsKey)
									msg := fmt.Sprintf("写入WsKey，并更新账号%s", ck.PtPin)
									sender.Reply(fmt.Sprintf(msg))
									(&JdCookie{}).Push(msg)
									logs.Info(msg)
								} else {
									if nck.WsKey == ck.WsKey {
										msg := fmt.Sprintf("重复写入")
										sender.Reply(fmt.Sprintf(msg))
										(&JdCookie{}).Push(msg)
										logs.Info(msg)
									} else {
										nck.Updates(JdCookie{
											WsKey: ck.WsKey,
										})
										msg := fmt.Sprintf("更新WsKey，并更新账号%s", ck.PtPin)
										sender.Reply(fmt.Sprintf(msg))
										(&JdCookie{}).Push(msg)
										logs.Info(msg)
									}
								}

							} else {
								NewJdCookie(&ck)

								msg := fmt.Sprintf("添加账号，账号名:%s", ck.PtPin)

								if sender.IsQQ() {
									ck.Update(QQ, ck.QQ)
								}

								sender.Reply(fmt.Sprintf(msg))
								sender.Reply(ck.Query())
								(&JdCookie{}).Push(msg)
							}
						}
						go func() {
							Save <- &JdCookie{}
						}()
						return nil
					}
				}
			}
		}
		{ //
			ss := regexp.MustCompile(`pt_key=([^;=\s]+);pt_pin=([^;=\s]+)`).FindAllStringSubmatch(msg, -1)

			if len(ss) > 0 {

				xyb := 0
				for _, s := range ss {
					ck := JdCookie{
						PtKey: s[1],
						PtPin: s[2],
					}
					xyb++
					if sender.IsQQ() {
						ck.QQ = sender.UserID
					} else if sender.IsTG() {
						ck.Telegram = sender.UserID
					}
					if HasKey(ck.PtKey) {
						sender.Reply(fmt.Sprintf("重复提交"))
					} else {
						if nck, err := GetJdCookie(ck.PtPin); err == nil {
							nck.InPool(ck.PtKey)
							msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
							(&JdCookie{}).Push(msg)
							logs.Info(msg)
						} else {
							if Cdle {
								ck.Hack = True
							}
							NewJdCookie(&ck)
							msg := fmt.Sprintf("添加账号，%s", ck.PtPin)
							sender.Reply(fmt.Sprintf("很棒，许愿币+1，余额%d", AddCoin(sender.UserID)))
							logs.Info(msg)
						}
					}

				}
				go func() {
					Save <- &JdCookie{}
				}()
				return nil
			}
		}
		{
			//dyj
			inviterId := regexp.MustCompile(`inviterId=(\S+)(&|&amp;)helpType`).FindStringSubmatch(msg)
			redEnvelopeId := regexp.MustCompile(`redEnvelopeId=(\S+)(&|&amp;)inviterId`).FindStringSubmatch(msg)
			if len(inviterId) > 0 && len(redEnvelopeId) > 0 {
				if !sender.IsAdmin {
					sender.Reply("仅管理员可用")
				} else {

					sender.Reply(fmt.Sprintf("大赢家开始，管理员通道"))
					num := startdyj(inviterId[1], redEnvelopeId[1])
					sender.Reply(fmt.Sprintf("助力完成，助力成功：%d个", num))
					//runTask(&Task{Path: "xdd_fcdyj.js", Envs: []Env{
					//	{Name: "djyinviter", Value: inviterId[1]},
					//	{Name: "djyredEnvelopeId", Value: redEnvelopeId[1]},
					//}}, sender)
				}
				return nil
			}

		}
		{
			//k1k
			ss := regexp.MustCompile(`launchid=(\S+)(&|&amp;)ptag`).FindStringSubmatch(msg)
			if len(ss) > 0 {
				if !sender.IsAdmin {
					sender.Reply("仅管理员可用")
				} else {
					sender.Reply(fmt.Sprintf("砍价开始，管理员通道"))
					runTask(&Task{Path: "jd_kanjia.js", Envs: []Env{
						{Name: "launchid", Value: ss[1]},
					}}, sender)
				}
				return nil
			}
		}
		{ //tyt
			ss := regexp.MustCompile(`packetId=(\S+)(&|&amp;)currentActId`).FindStringSubmatch(msg)
			log.Info(ss)
			if len(ss) > 0 {
				if !sender.IsAdmin {
					coin := GetCoin(sender.UserID)
					if coin < Config.Tyt {
						return fmt.Sprintf("推一推需要%d个互助值", Config.Tyt)
					}
					RemCoin(sender.UserID, 8)
					sender.Reply(fmt.Sprintf("推一推即将开始，已扣除%d个互助值", Config.Tyt))
				} else {
					sender.Reply(fmt.Sprintf("推一推即将开始，已扣除%d个互助值，管理员通道", Config.Tyt))
				}

				runTask(&Task{Path: "jd_tyt.js", Envs: []Env{
					{Name: "tytpacketId", Value: ss[1]},
				}}, sender)
				return nil
			}
		}
		{
			if strings.Contains(msg, "pt_key") {
				ptKey := FetchJdCookieValue("pt_key", msg)
				ptPin := FetchJdCookieValue("pt_pin", msg)
				if len(ptPin) > 0 && len(ptKey) > 0 {
					ck := JdCookie{
						PtKey: ptKey,
						PtPin: ptPin,
					}
					if CookieOK(&ck) {
						if sender.IsQQ() {
							ck.QQ = sender.UserID
						} else if sender.IsTG() {
							ck.Telegram = sender.UserID
						}
						if HasKey(ck.PtKey) {
							sender.Reply(fmt.Sprintf("重复提交"))
						} else {
							if nck, err := GetJdCookie(ck.PtPin); err == nil {
								nck.InPool(ck.PtKey)
								msg := fmt.Sprintf("更新账号，%s", ck.PtPin)
								if sender.IsQQ() {
									ck.Update(QQ, ck.QQ)
								}
								sender.Reply(fmt.Sprintf(msg))
								(&JdCookie{}).Push(msg)
								logs.Info(msg)
							} else {
								if Cdle {
									ck.Hack = True
								}
								NewJdCookie(&ck)
								msg := fmt.Sprintf("添加账号，账号名:%s", ck.PtPin)
								if sender.IsQQ() {
									ck.Update(QQ, ck.QQ)
								}
								sender.Reply(fmt.Sprintf(msg))
								sender.Reply(ck.Query())
								(&JdCookie{}).Push(msg)
								logs.Info(msg)
							}
						}
					} else {
						sender.Reply(fmt.Sprintf("无效"))
					}
				}
				go func() {
					Save <- &JdCookie{}
				}()
				return nil
			}
		}
		{
			o := findShareCode(msg)
			if o != "" {
				return "导入互助码成功"
			}
		}
		for k, v := range replies {
			if regexp.MustCompile(k).FindString(msg) != "" {
				if strings.Contains(msg, "妹") && time.Now().Unix()%10 == 0 {
					v = "https://pics4.baidu.com/feed/d833c895d143ad4bfee5f874cfdcbfa9a60f069b.jpeg?token=8a8a0e1e20d4626cd31c0b838d9e4c1a"
				}
				if regexp.MustCompile(`^https{0,1}://[^\x{4e00}-\x{9fa5}\n\r\s]{3,}$`).FindString(v) != "" {
					url := v
					rsp, err := httplib.Get(url).Response()
					if err != nil {
						return nil
					}
					ctp := rsp.Header.Get("content-type")
					if ctp == "" {
						rsp.Header.Get("Content-Type")
					}
					if strings.Contains(ctp, "text") || strings.Contains(ctp, "json") {
						data, _ := ioutil.ReadAll(rsp.Body)
						return string(data)
					}
					return rsp
				}
				return v
			}
		}
	}
	return nil
}

/*
 url: `https://api.m.jd.com/?functionId=openRedEnvelopeInteract&body={"linkId":"PFbUR7wtwUcQ860Sn8WRfw","redEnvelopeId":"${ redEnvelopeId }","inviter":"${ inviter }","helpType":"1"}&t=1626363029817&appid=activities_platform&clientVersion=3.5.0`,

            headers: {
                "Origin": "https://618redpacket.jd.com",
                "Host": "api.m.jd.com",
                "User-Agent": "User-Agent:jdapp;android;10.2.2;11;2623632613665613-4636264326366343;model/V2141A;addressid/1294647027;aid/b26b1fe1dcb4bc64;oaid/0b3ff6566ee75f3d558fd06149d16d310473ed980a032fe0228878ebe5092edb;osVer/30;appBuild/91077;partner/vivo;eufv/1;jdSupportDarkMode/0;Mozilla/5.0 (Linux; Android 11; V2141A Build/RP1A.200720.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/77.0.3865.120 MQQBrowser/6.2 TBS/045714 Mobile Safari/537.36",
                "Cookie": cookie,
            }
*/

func startdyj(ine string, red string) (num int) {
	i := 0
	cks := GetJdCookies()
	for i := range cks {
		cookie := "pt_key=" + cks[i].PtKey + ";pt_pin=" + cks[i].PtPin + ";"
		logs.Info(cookie)
		sprintf := fmt.Sprintf(`https://api.m.jd.com/client.action?functionId=openRedEnvelopeInteract&body={"linkId":"PFbUR7wtwUcQ860Sn8WRfw","redEnvelopeId":"%s","inviter":"%s","helpType":"1"}&t=1626363029817&appid=activities_platform&clientVersion=3.5.0`, red, ine)
		logs.Info(sprintf)
		req := httplib.Get(sprintf)
		req.Header("User-Agent", ua)
		req.Header("Host", "api.m.jd.com")
		req.Header("Accept", "application/json, text/plain, */*")
		req.Header("Connection", "keep-alive")
		req.Header("Accept-Language", "zh-cn")
		req.Header("Accept-Encoding", "gzip, deflate, br")
		req.Header("Origin", "https://618redpacket.jd.com")
		req.Header("Cookie", cookie)
		data, _ := req.String()
		if strings.Contains(data, "恭喜帮好友助力成功") {
			i++
		} else {
			i++
			logs.Info(data)
			logs.Info("火爆了")
		}
		if i == 10 {
			return
		}
	}
	return i
}

func FetchJdCookieValue(key string, cookies string) string {
	match := regexp.MustCompile(key + `=([^;]*);{0,1}`).FindStringSubmatch(cookies)
	if len(match) == 2 {
		return match[1]
	} else {
		return ""
	}
}
