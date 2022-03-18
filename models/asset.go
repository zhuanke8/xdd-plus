package models

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/buger/jsonparser"
)

type Asset struct {
	Nickname string
	Bean     struct {
		Total         int
		TodayIn       int
		TodayOut      int
		YestodayIn    int
		YestodayOut   int
		XDTodayIn     int
		XDTodayOut    int
		XDYestodayIn  int
		XDYestodayOut int
		ToExpire      []int
	}
	RedPacket struct {
		Total      float64
		ToExpire   float64
		ToExpireJd float64
		ToExpireJx float64
		ToExpireJs float64
		ToExpireJk float64
		Jd         float64
		Jx         float64
		Js         float64
		Jk         float64
	}
	Other struct {
		JsCoin   float64
		NcStatus float64
		McStatus float64
	}
}

var Int = func(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

var Float64 = func(s string) float64 {
	i, _ := strconv.ParseFloat(s, 64)
	return i
}

func DailyAssetsPush() {
	for _, ck := range GetJdCookies() {
		if (ck.QQ != 0 && Config.QQID != 0 && SendQQ != nil) || ck.PushPlus != "" {
			msg := ck.Query()

			if ck.QQ != 0 && Config.QQID != 0 && SendQQ != nil {
				SendQQ(int64(ck.QQ), msg)
			}
			if ck.PushPlus != "" {
				pushPlus(ck.PushPlus, msg)
			}
			time.Sleep(time.Second * 60)
		}
	}
}

func CompletePush() {
	for _, ck := range GetJdCookies() {
		if (ck.QQ != 0 && Config.QQID != 0 && SendQQ != nil) || ck.PushPlus != "" {
			flag := false
			var msg1 []string
			var fruit = make(chan string)
			var pet = make(chan string)
			cookie := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
			go initFarm(cookie, fruit)
			if strings.Contains(<-fruit, "已可领取") {
				flag = true
				msg1 = append(msg1, ck.Nickname+"您的农场无门槛红包已经成熟，请尽快领取\r\n 【东东农场】京东->我的->东东农场,完成是京东红包,可以用于京东app的任意商品")
			}
			time.Sleep(time.Second * 30)
			go initPetTown(cookie, pet)
			if strings.Contains(<-pet, "已可领取") {
				flag = true
				msg1 = append(msg1, ck.Nickname+"您的萌宠无门槛红包已经成熟，请尽快领取\r\n 【东东萌宠】京东->我的->东东萌宠,完成是京东红包,可以用于京东app的任意商品")
			}
			if flag {
				if ck.QQ != 0 && Config.QQID != 0 && SendQQ != nil {
					SendQQ(int64(ck.QQ), strings.Join(msg1, "\n"))
				}
				if ck.PushPlus != "" {
					pushPlus(ck.PushPlus, strings.Join(msg1, "\n"))
				}
			}
			time.Sleep(time.Second * 30)
		}
	}
}

func (ck *JdCookie) Query1() string {
	name := "jd_bean_change_new.js"
	envs := []Env{{Name: "pins", Value: "&" + ck.PtPin}}
	msg := runTask(&Task{Path: name, Envs: envs}, &Sender{})
	//log.Info(msg)
	if !strings.Contains(msg, "cookies") {
		msg = regexp.MustCompile(`^(.+\s+){3}|\s*.+\s*$|.*东东工厂.*\s*`).ReplaceAllString(msg, "")
		msg = fmt.Sprintf("账号昵称：%s\n绑定QQ: %v\n用户等级：%v\n等级名称：%v\n优先级: %v\n%s", ck.Nickname, ck.QQ, ck.UserLevel, ck.LevelName, ck.Priority, msg)
	} else if CookieOK(ck) {
		msg = fmt.Sprintf("查询失败\n账号: %s\n备注: %s\n%s", ck.PtPin, ck.Note, msg)
	} else {
		msg = fmt.Sprintf("失效账号\n账号: %s\n备注: %s", ck.PtPin, ck.Note)
	}
	return msg
}

func (ck *JdCookie) Query() string {
	parse, _ := time.Parse("2006-01-02", ck.CreateAt)
	t, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	f := t.Sub(parse).Hours() / 24
	i, _ := strconv.Atoi(fmt.Sprintf("%1.0f", f))
	msgs := []string{
		fmt.Sprintf("账号昵称：%s", ck.Nickname),
		fmt.Sprintf("您已挂机：%d天", i),
	}
	if ck.Note != "" {
		msgs = append(msgs, fmt.Sprintf("账号备注：%s", ck.Note))
	}
	asset := Asset{}
	if CookieOK(ck) {
		//msgs = append(msgs, fmt.Sprintf("优先级：%v", ck.Priority))
		//msgs = append(msgs, fmt.Sprintf("用户等级：%v", ck.UserLevel))
		//msgs = append(msgs, fmt.Sprintf("等级名称：%v", ck.LevelName))

		cookie := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
		if !strings.Contains(cookie, "open") {
			if ck.UpdateAt != "" {
				parse1, _ := time.Parse("2006-01-02", ck.UpdateAt)
				f := t.Sub(parse1).Hours() / 24
				i, _ := strconv.Atoi(fmt.Sprintf("%1.0f", f))
				msgs = append(msgs, fmt.Sprintf("您距离失效还有：%d天", 28-i))
			}
		}
		var rpc = make(chan []RedList)
		var fruit = make(chan string)
		var pet = make(chan string)
		var gold = make(chan int64)
		var egg = make(chan int64)
		var tyt = make(chan string)
		var mmc = make(chan int64)
		var zjb = make(chan int64)
		var xgc = make(chan string)
		var dsy = make(chan string)
		var jxzz = make(chan string)
		go redPacket(cookie, rpc)
		go initFarm(cookie, fruit)
		go initPetTown(cookie, pet)
		go jsGold(cookie, gold)
		go jxncEgg(cookie, egg)
		go tytCoupon(cookie, tyt)
		go mmCoin(cookie, mmc)
		go jdzz(cookie, zjb)
		go jxgc(cookie, xgc)
		go jdsy(cookie, dsy)
		go jingxiangzhi(cookie, jxzz)
		msgs = append(msgs, fmt.Sprintf("京享值：%s", <-jxzz))
		today := time.Now().Local().Format("2006-01-02")
		yestoday := time.Now().Local().Add(-time.Hour * 24).Format("2006-01-02")
		today1 := time.Now().Local().Format("2006/01/02")
		yestoday1 := time.Now().Local().Add(-time.Hour * 24).Format("2006/01/02")
		page := 1
		end := false
		jds := getJingXiBeanDeatil(cookie)
		if jds == nil {
			msgs = append(msgs, "喜豆加载中，请耐心等待")
		}
		for _, jd := range jds {
			amount := jd.Amount
			if strings.Contains(jd.Createdate, today1) {
				if amount > 0 {
					asset.Bean.XDTodayIn += amount
				} else {
					asset.Bean.XDTodayOut += -amount
				}
			} else if strings.Contains(jd.Createdate, yestoday1) {
				if amount > 0 {
					asset.Bean.XDYestodayIn += amount
				} else {
					asset.Bean.XDYestodayOut += -amount
				}
			}
		}
		for {
			if end {
				msgs = append(msgs, []string{
					fmt.Sprintf("昨日收入：%d京豆,%d喜豆", asset.Bean.YestodayIn, asset.Bean.XDYestodayIn),
					//fmt.Sprintf("昨日支出：%d京豆", asset.Bean.YestodayOut),
					fmt.Sprintf("今日收入：%d京豆,%d喜豆", asset.Bean.TodayIn, asset.Bean.XDTodayIn),
					//fmt.Sprintf("今日支出：%d京豆", asset.Bean.TodayOut),
				}...)
				break
			}
			bds := getJingBeanBalanceDetail(page, cookie)
			if bds == nil {
				end = true
				msgs = append(msgs, "京豆加载中，请耐心等待")
				break
			}
			for _, bd := range bds {
				amount := Int(bd.Amount)
				if strings.Contains(bd.Date, today) {
					if amount > 0 {
						asset.Bean.TodayIn += amount
					} else {
						asset.Bean.TodayOut += -amount
					}
				} else if strings.Contains(bd.Date, yestoday) {
					if amount > 0 {
						asset.Bean.YestodayIn += amount
					} else {
						asset.Bean.YestodayOut += -amount
					}
				} else {
					end = true
					break
				}
			}
			page++
		}
		//logs.Info(ck.BeanNum)
		xd, s := getXd(cookie)
		msgs = append(msgs, fmt.Sprintf("当前京豆：%s京豆,%s喜豆", s, xd))
		ysd := int(time.Now().Add(24 * time.Hour).Unix())
		if rps := <-rpc; len(rps) != 0 {
			for _, rp := range rps {
				b := Float64(rp.Balance)
				asset.RedPacket.Total += b
				if strings.Contains(rp.ActivityName, "京喜") || strings.Contains(rp.OrgLimitStr, "京喜") {
					asset.RedPacket.Jx += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJx += b
						asset.RedPacket.ToExpire += b
					}
				} else if strings.Contains(rp.ActivityName, "极速版") {
					asset.RedPacket.Js += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJs += b
						asset.RedPacket.ToExpire += b
					}

				} else if strings.Contains(rp.ActivityName, "京东健康") {
					asset.RedPacket.Jk += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJk += b
						asset.RedPacket.ToExpire += b
					}
				} else {
					asset.RedPacket.Jd += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJd += b
						asset.RedPacket.ToExpire += b
					}
				}
			}
			e := func(m float64) string {
				if m > 0 {
					return fmt.Sprintf(`(今日过期%.2f)`, m)
				}
				return ""
			}
			msgs = append(msgs, []string{
				fmt.Sprintf("所有红包：%.2f%s元🧧", asset.RedPacket.Total, e(asset.RedPacket.ToExpire)),
				fmt.Sprintf("京喜红包：%.2f%s元", asset.RedPacket.Jx, e(asset.RedPacket.ToExpireJx)),
				fmt.Sprintf("极速红包：%.2f%s元", asset.RedPacket.Js, e(asset.RedPacket.ToExpireJs)),
				//fmt.Sprintf("健康红包：%.2f%s元", asset.RedPacket.Jk, e(asset.RedPacket.ToExpireJk)),
				fmt.Sprintf("京东红包：%.2f%s元", asset.RedPacket.Jd, e(asset.RedPacket.ToExpireJd)),
			}...)
		} else {
			msgs = append(msgs, "暂无红包数据🧧")
		}
		msgs = append(msgs, fmt.Sprintf("东东农场：%s", <-fruit))
		msgs = append(msgs, fmt.Sprintf("东东萌宠：%s", <-pet))
		msgs = append(msgs, fmt.Sprintf("京喜工厂：%s", <-xgc))
		msgs = append(msgs, fmt.Sprintf("京东试用：%s", <-dsy))
		gn := <-gold
		msgs = append(msgs, fmt.Sprintf("极速金币：%d(≈%.2f元)💰", gn, float64(gn)/10000))
		zjbn := <-zjb
		if zjbn != 0 {
			msgs = append(msgs, fmt.Sprintf("京东赚赚：%d金币(≈%.2f元)💰", zjbn, float64(zjbn)/10000))
		} else {
			msgs = append(msgs, fmt.Sprintf("京东赚赚：暂无数据"))
		}
		mmcCoin := <-mmc
		if mmcCoin != 0 {
			msgs = append(msgs, fmt.Sprintf("京东秒杀：%d秒秒币(≈%.2f元)💰", mmcCoin, float64(mmcCoin)/1000))
		} else {
			msgs = append(msgs, fmt.Sprintf("京东秒杀：暂无数据"))
		}
		msgs = append(msgs, fmt.Sprintf("推一推券：%s", <-tyt))
		msgs = append(msgs, fmt.Sprintf("惊喜牧场：%d枚鸡蛋🥚", <-egg))

	} else {
		msgs = append(msgs, []string{
			"提醒：该账号已过期，请重新登录",
		}...)
	}
	ck.PtPin, _ = url.QueryUnescape(ck.PtPin)
	return strings.Join(msgs, "\n")
}

func getXd(cookie string) (string, string) {
	req := httplib.Get(fmt.Sprintf("https://m.jingxi.com/activeapi/querybeanamount?_=%t&sceneval=2&g_login_type=1&g_ty=ls", time.Now().UnixMilli()))
	log.Info(time.Now().UnixMilli())
	req.Header("User-Agent", "jdpingou;android;5.5.0;11;network/wifi;model/M2102K1C;appBuild/18299;partner/lcjx11;session/110;pap/JA2019_3111789;brand/Xiaomi;Mozilla/5.0 (Linux; Android 11; M2102K1C Build/RKQ1.201112.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/92.0.4515.159 Mobile Safari/537.36")
	req.Header("Host", "m.jingxi.com")
	req.Header("Accept", "*/*")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header("Referer", "https://st.jingxi.com/")
	req.Header("Cookie", cookie)
	resp, _ := req.Bytes()
	xibean, err := jsonparser.GetInt(resp, "data", "xibean")
	if err != nil {
		log.Info(err)
		return "喜豆加载中", "京豆加载中"
	}
	jingbean, err := jsonparser.GetInt(resp, "data", "jingbean")
	if err != nil {
		log.Info(err)
		return "喜豆加载中", "京豆加载中"
	}
	return strconv.FormatInt(xibean, 10), strconv.FormatInt(jingbean, 10)
}

func jingxiangzhi(cookie string, state chan string) {
	req := httplib.Get(`https://wxapp.m.jd.com/kwxhome/myJd/home.json?&useGuideModule=0&bizId=&brandId=&fromType=wxapp&timestamp=` + fmt.Sprint(time.Now().Unix()))
	req.Header("User-Agent", ua)
	req.Header("Cookie", cookie)
	data, _ := req.Bytes()
	jxzz, _ := jsonparser.GetString(data, "user", "uclass")
	jxzz = strings.Replace(jxzz, "京享值", "", -1)
	state <- jxzz
}

func jdsy(cookie string, desc chan string) {
	type AutoGenerated struct {
		Success bool        `json:"success"`
		Code    interface{} `json:"code"`
		BsCode  interface{} `json:"bsCode"`
		Message interface{} `json:"message"`
		Data    struct {
			List []struct {
				ActID     int         `json:"actId"`
				ReportID  interface{} `json:"reportId"`
				OrderID   interface{} `json:"orderId"`
				ApplyTime int64       `json:"applyTime"`
				Status    int         `json:"status"`
				TrialImg  string      `json:"trialImg"`
				TrialName string      `json:"trialName"`
				Text      struct {
					ID   int    `json:"id"`
					Text string `json:"text"`
				} `json:"text"`
				LeftTime      int `json:"leftTime"`
				TryButtonList []struct {
					ID     int         `json:"id"`
					Text   string      `json:"text"`
					Schema interface{} `json:"schema"`
				} `json:"tryButtonList"`
				Bottom           interface{}   `json:"bottom"`
				ActType          int           `json:"actType"`
				TaskType         int           `json:"taskType"`
				SkuID            string        `json:"skuId"`
				OrderState       interface{}   `json:"orderState"`
				Tag              []interface{} `json:"tag"`
				OrderAmount      interface{}   `json:"orderAmount"`
				EndTime          int64         `json:"endTime"`
				SupplierDelivery bool          `json:"supplierDelivery"`
			} `json:"list"`
			PageSize int   `json:"pageSize"`
			Page     int   `json:"page"`
			SysDate  int64 `json:"sysDate"`
		} `json:"data"`
	}
	rt := ""
	warn := make(chan string)
	go func() {
		req := httplib.Post("https://api.m.jd.com/client.action")
		req.Header("Host", "api.m.jd.com")
		req.Header("Content-Type", "application/x-www-form-urlencoded")
		req.Header("Origin", "https://prodev.m.jd.com")
		// req.Header("Accept-Encoding", "gzip, deflate, br")
		req.Header("Cookie", cookie)
		req.Header("Connection", "keep-alive")
		req.Header("Accept", "application/json, text/plain, */*")
		req.Header("User-Agent", ua)
		// req.Header("Referer", "https://prodev.m.jd.com/mall/active/2Y2YgUu1Xbbv8AfN7TAHhNqfQrAV/index.html?tttparams=eliIVi1eyJncHNfYXJlYSI6IjEyXzkzOV8yMzY4M181NjE4NCIsInByc3RhdGUiOiIwIiwidW5fYXJlYSI6IjEyXzkzOV8yMzY4M181NjE4NCIsIm1vZGVsIjoiaVBob25lMTAsMiIsImdMYXQiOiIzMy4yOTQ5MyIsImdMbmciOiIxMjAuMTQ4MjMyIiwibG5nIjoiMTIwLjE1MDMyMyIsImxhdCI6IjMzLjI5NTUzNi7J9&sid=a2e1c9b3b215a2337517cd01f2e04cbw&un_area=12_939_23683_56184")
		// req.Header("Content-Length", "332")
		// req.Header("Accept-Language", "zh-cn")
		req.Body(`appid=newtry&functionId=try_MyTrials&uuid=3345ad3d16ab2153c69f8ca91cd3e931b06a3bb8&clientVersion=10.2.7&client=wh5&osVersion=14.7.1&area=12_939_23683_56184&networkType=wifi&body=%7B%22geo%22%3A%7B%22lng%22%3A121.15326252577907%2C%22lat%22%3A34.295611038697575%7D%2C%22page%22%3A1%2C%22selected%22%3A2%2C%22previewTime%22%3A%22%22%7D`)
		//appid=newtry&functionId=try_MyTrials&uuid=3345ad3d16ab2153c69f8ca91cd3e931b06a3bb8&clientVersion=10.2.7&client=wh5&osVersion=14.7.1&area=12_939_23683_56184&networkType=wifi&body=%7B%22geo%22%3A%7B%22lng%22%3A121.15326252577907%2C%22lat%22%3A34.295611038697575%7D%2C%22page%22%3A1%2C%22selected%22%3A1%2C%22previewTime%22%3A%22%22%7D
		data, _ := req.Bytes()
		// fmt.Println(string(data))
		a := &AutoGenerated{}
		json.Unmarshal(data, a)
		// fmt.Println(a)
		for _, v := range a.Data.List {
			if len(v.TryButtonList) == 2 {
				if v.TryButtonList[0].ID <= 2 {
					warn <- "你有一个商品待领取，详情：" + v.TrialName + "👆"
					return
				}
			}
		}
		warn <- ""
	}()
	if rt == "" {
		req := httplib.Post("https://api.m.jd.com/client.action")
		req.Header("Host", "api.m.jd.com")
		req.Header("Content-Type", "application/x-www-form-urlencoded")
		req.Header("Origin", "https://prodev.m.jd.com")
		// req.Header("Accept-Encoding", "gzip, deflate, br")
		req.Header("Cookie", cookie)
		req.Header("Connection", "keep-alive")
		req.Header("Accept", "application/json, text/plain, */*")
		req.Header("User-Agent", ua)
		// req.Header("Referer", "https://prodev.m.jd.com/mall/active/2Y2YgUu1Xbbv8AfN7TAHhNqfQrAV/index.html?tttparams=eliIVi1eyJncHNfYXJlYSI6IjEyXzkzOV8yMzY4M181NjE4NCIsInByc3RhdGUiOiIwIiwidW5fYXJlYSI6IjEyXzkzOV8yMzY4M181NjE4NCIsIm1vZGVsIjoiaVBob25lMTAsMiIsImdMYXQiOiIzMy4yOTQ5MyIsImdMbmciOiIxMjAuMTQ4MjMyIiwibG5nIjoiMTIwLjE1MDMyMyIsImxhdCI6IjMzLjI5NTUzNi7J9&sid=a2e1c9b3b215a2337517cd01f2e04cbw&un_area=12_939_23683_56184")
		// req.Header("Content-Length", "332")
		// req.Header("Accept-Language", "zh-cn")
		req.Body(`appid=newtry&functionId=try_MyTrials&uuid=3345ad3d16ab2153c69f8ca91cd3e931b06a3bb8&clientVersion=10.2.7&client=wh5&osVersion=14.7.1&area=12_939_23683_56184&networkType=wifi&body=%7B%22geo%22%3A%7B%22lng%22%3A121.15326252577907%2C%22lat%22%3A34.295611038697575%7D%2C%22page%22%3A1%2C%22selected%22%3A1%2C%22previewTime%22%3A%22%22%7D`)

		data, _ := req.Bytes()
		// fmt.Println(string(data))
		a := &AutoGenerated{}
		json.Unmarshal(data, a)
		// fmt.Println(a)
		rt = fmt.Sprintf("%d件商品申请中", len(a.Data.List))
	}
	if xx := <-warn; xx != "" {
		rt = xx
	}
	desc <- rt
}

type BeanDetail struct {
	Date         string `json:"date"`
	Amount       string `json:"amount"`
	EventMassage string `json:"eventMassage"`
}

func getJingBeanBalanceDetail(page int, cookie string) []BeanDetail {
	type AutoGenerated struct {
		Code       string       `json:"code"`
		DetailList []BeanDetail `json:"detailList"`
	}
	a := AutoGenerated{}
	req := httplib.Post(`https://api.m.jd.com/client.action?functionId=getJingBeanBalanceDetail`)
	req.Header("User-Agent", ua)
	req.Header("Host", "api.m.jd.com")
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Header("Cookie", cookie)
	req.Body(fmt.Sprintf(`body={"pageSize": "20", "page": "%d"}&appid=ld`, page))
	data, err := req.Bytes()
	if err != nil {
		return nil
	}
	json.Unmarshal(data, &a)
	return a.DetailList
}

type JingXiBeanDetails struct {
	Detail []JingXiDetail `json:"detail"`
	Ret    int            `json:"ret"`
	Retmsg string         `json:"retmsg"`
}

type JingXiDetail struct {
	Amount      int    `json:"amount"`
	Createdate  string `json:"createdate"`
	Visibleinfo string `json:"visibleinfo"`
}

func getJingXiBeanDeatil(cookie string) []JingXiDetail {
	req := httplib.Get(fmt.Sprintf("https://m.jingxi.com/activeapi/queryuserjingdoudetail?_=%t&sceneval=2&g_login_type=1&g_ty=ls&pagesize=15&type=16", time.Now().UnixMilli()))
	req.Header("User-Agent", "jdpingou;android;5.5.0;11;network/wifi;model/M2102K1C;appBuild/18299;partner/lcjx11;session/110;pap/JA2019_3111789;brand/Xiaomi;Mozilla/5.0 (Linux; Android 11; M2102K1C Build/RKQ1.201112.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/92.0.4515.159 Mobile Safari/537.36")
	req.Header("Host", "m.jingxi.com")
	req.Header("Accept", "*/*")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header("Referer", "https://st.jingxi.com/")
	req.Header("Cookie", cookie)
	resp, _ := req.Bytes()
	a := JingXiBeanDetails{}
	json.Unmarshal(resp, &a)
	return a.Detail
}

type RedList struct {
	ActivityName string `json:"activityName"`
	Balance      string `json:"balance"`
	BeginTime    int    `json:"beginTime"`
	DelayRemark  string `json:"delayRemark"`
	Discount     string `json:"discount"`
	EndTime      int    `json:"endTime"`
	HbID         string `json:"hbId"`
	HbState      int    `json:"hbState"`
	IsDelay      bool   `json:"isDelay"`
	OrgLimitStr  string `json:"orgLimitStr"`
}

func redPacket(cookie string, rpc chan []RedList) {
	type UseRedInfo struct {
		Count   int       `json:"count"`
		RedList []RedList `json:"redList"`
	}
	type Data struct {
		AvaiCount      int        `json:"avaiCount"`
		Balance        string     `json:"balance"`
		CountdownTime  string     `json:"countdownTime"`
		ExpiredBalance string     `json:"expiredBalance"`
		ServerCurrTime int        `json:"serverCurrTime"`
		UseRedInfo     UseRedInfo `json:"useRedInfo"`
	}
	type AutoGenerated struct {
		Data    Data   `json:"data"`
		Errcode int    `json:"errcode"`
		Msg     string `json:"msg"`
	}
	a := AutoGenerated{}
	req := httplib.Get(`https://m.jingxi.com/user/info/QueryUserRedEnvelopesV2?type=1&orgFlag=JD_PinGou_New&page=1&cashRedType=1&redBalanceFlag=1&channel=1&_=` + fmt.Sprint(time.Now().Unix()) + `&sceneval=2&g_login_type=1&g_ty=ls`)
	req.Header("User-Agent", ua)
	req.Header("Host", "m.jingxi.com")
	req.Header("Accept", "*/*")
	req.Header("Connection", "keep-alive")
	req.Header("Accept-Language", "zh-cn")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Referer", "https://st.jingxi.com/my/redpacket.shtml?newPg=App")
	req.Header("Cookie", cookie)
	data, _ := req.Bytes()
	json.Unmarshal(data, &a)
	rpc <- a.Data.UseRedInfo.RedList
}

func initFarm(cookie string, state chan string) {
	type RightUpResouces struct {
		AdvertID string `json:"advertId"`
		Name     string `json:"name"`
		AppImage string `json:"appImage"`
		AppLink  string `json:"appLink"`
		CxyImage string `json:"cxyImage"`
		CxyLink  string `json:"cxyLink"`
		Type     string `json:"type"`
		OpenLink bool   `json:"openLink"`
	}
	type TurntableInit struct {
		TimeState int `json:"timeState"`
	}
	type MengchongResouce struct {
		AdvertID string `json:"advertId"`
		Name     string `json:"name"`
		AppImage string `json:"appImage"`
		AppLink  string `json:"appLink"`
		CxyImage string `json:"cxyImage"`
		CxyLink  string `json:"cxyLink"`
		Type     string `json:"type"`
		OpenLink bool   `json:"openLink"`
	}
	type GUIDPopupTask struct {
		GUIDPopupTask string `json:"guidPopupTask"`
	}
	type IosConfigResouces struct {
		AdvertID string `json:"advertId"`
		Name     string `json:"name"`
		AppImage string `json:"appImage"`
		AppLink  string `json:"appLink"`
		CxyImage string `json:"cxyImage"`
		CxyLink  string `json:"cxyLink"`
		Type     string `json:"type"`
		OpenLink bool   `json:"openLink"`
	}
	type TodayGotWaterGoalTask struct {
		CanPop bool `json:"canPop"`
	}
	type LeftUpResouces struct {
		AdvertID string `json:"advertId"`
		Name     string `json:"name"`
		AppImage string `json:"appImage"`
		AppLink  string `json:"appLink"`
		CxyImage string `json:"cxyImage"`
		CxyLink  string `json:"cxyLink"`
		Type     string `json:"type"`
		OpenLink bool   `json:"openLink"`
	}
	type RightDownResouces struct {
		AdvertID string `json:"advertId"`
		Name     string `json:"name"`
		AppImage string `json:"appImage"`
		AppLink  string `json:"appLink"`
		CxyImage string `json:"cxyImage"`
		CxyLink  string `json:"cxyLink"`
		Type     string `json:"type"`
		OpenLink bool   `json:"openLink"`
	}
	type FarmUserPro struct {
		TotalEnergy     int    `json:"totalEnergy"`
		TreeState       int    `json:"treeState"`
		CreateTime      int64  `json:"createTime"`
		TreeEnergy      int    `json:"treeEnergy"`
		TreeTotalEnergy int    `json:"treeTotalEnergy"`
		ShareCode       string `json:"shareCode"`
		WinTimes        int    `json:"winTimes"`
		NickName        string `json:"nickName"`
		CouponKey       string `json:"couponKey"`
		CouponID        string `json:"couponId"`
		CouponEndTime   int64  `json:"couponEndTime"`
		Type            string `json:"type"`
		SimpleName      string `json:"simpleName"`
		Name            string `json:"name"`
		GoodsImage      string `json:"goodsImage"`
		SkuID           string `json:"skuId"`
		LastLoginDate   int64  `json:"lastLoginDate"`
		NewOldState     int    `json:"newOldState"`
		OldMarkComplete int    `json:"oldMarkComplete"`
		CommonState     int    `json:"commonState"`
		PrizeLevel      int    `json:"prizeLevel"`
	}
	type LeftDownResouces struct {
		AdvertID string `json:"advertId"`
		Name     string `json:"name"`
		AppImage string `json:"appImage"`
		AppLink  string `json:"appLink"`
		CxyImage string `json:"cxyImage"`
		CxyLink  string `json:"cxyLink"`
		Type     string `json:"type"`
		OpenLink bool   `json:"openLink"`
	}
	type LoadFriend struct {
		Code            string      `json:"code"`
		StatisticsTimes interface{} `json:"statisticsTimes"`
		SysTime         int64       `json:"sysTime"`
		Message         interface{} `json:"message"`
		FirstAddUser    bool        `json:"firstAddUser"`
	}
	type AutoGenerated struct {
		Code                  string                `json:"code"`
		RightUpResouces       RightUpResouces       `json:"rightUpResouces"`
		TurntableInit         TurntableInit         `json:"turntableInit"`
		IosShieldConfig       interface{}           `json:"iosShieldConfig"`
		MengchongResouce      MengchongResouce      `json:"mengchongResouce"`
		ClockInGotWater       bool                  `json:"clockInGotWater"`
		GUIDPopupTask         GUIDPopupTask         `json:"guidPopupTask"`
		ToFruitEnergy         int                   `json:"toFruitEnergy"`
		StatisticsTimes       interface{}           `json:"statisticsTimes"`
		SysTime               int64                 `json:"sysTime"`
		CanHongbaoContineUse  bool                  `json:"canHongbaoContineUse"`
		ToFlowTimes           int                   `json:"toFlowTimes"`
		IosConfigResouces     IosConfigResouces     `json:"iosConfigResouces"`
		TodayGotWaterGoalTask TodayGotWaterGoalTask `json:"todayGotWaterGoalTask"`
		LeftUpResouces        LeftUpResouces        `json:"leftUpResouces"`
		MinSupportAPPVersion  string                `json:"minSupportAPPVersion"`
		LowFreqStatus         int                   `json:"lowFreqStatus"`
		FunCollectionHasLimit bool                  `json:"funCollectionHasLimit"`
		Message               interface{}           `json:"message"`
		TreeState             int                   `json:"treeState"`
		RightDownResouces     RightDownResouces     `json:"rightDownResouces"`
		IconFirstPurchaseInit bool                  `json:"iconFirstPurchaseInit"`
		ToFlowEnergy          int                   `json:"toFlowEnergy"`
		FarmUserPro           FarmUserPro           `json:"farmUserPro"`
		RetainPopupLimit      int                   `json:"retainPopupLimit"`
		ToBeginEnergy         int                   `json:"toBeginEnergy"`
		LeftDownResouces      LeftDownResouces      `json:"leftDownResouces"`
		EnableSign            bool                  `json:"enableSign"`
		LoadFriend            LoadFriend            `json:"loadFriend"`
		HadCompleteXgTask     bool                  `json:"hadCompleteXgTask"`
		OldUserIntervalTimes  []int                 `json:"oldUserIntervalTimes"`
		ToFruitTimes          int                   `json:"toFruitTimes"`
		OldUserSendWater      []string              `json:"oldUserSendWater"`
	}
	a := AutoGenerated{}
	req := httplib.Post(`https://api.m.jd.com/client.action?functionId=initForFarm`)
	req.Header("accept", "*/*")
	req.Header("accept-encoding", "gzip, deflate, br")
	req.Header("accept-language", "zh-CN,zh;q=0.9")
	req.Header("cache-control", "no-cache")
	req.Header("cookie", cookie)
	req.Header("origin", "https://home.m.jd.com")
	req.Header("pragma", "no-cache")
	req.Header("referer", "https://home.m.jd.com/myJd/newhome.action")
	req.Header("sec-fetch-dest", "empty")
	req.Header("sec-fetch-mode", "cors")
	req.Header("sec-fetch-site", "same-site")
	req.Header("User-Agent", ua)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Body(`body={"version":4}&appid=wh5&clientVersion=9.1.0`)
	data, _ := req.Bytes()
	json.Unmarshal(data, &a)

	rt := a.FarmUserPro.Name
	if rt == "" {
		rt = "数据异常"
	} else {
		if a.TreeState == 2 || a.TreeState == 3 {
			rt += "已可领取⏰"
		} else if a.TreeState == 1 {
			rt += fmt.Sprintf("种植中，进度%.2f%%🍒", 100*float64(a.FarmUserPro.TreeEnergy)/float64(a.FarmUserPro.TreeTotalEnergy))
		} else if a.TreeState == 0 {
			rt = "您忘了种植新的水果⏰"
		}
	}
	state <- rt
}

func initPetTown(cookie string, state chan string) {
	type ResourceList struct {
		AdvertID string `json:"advertId"`
		ImageURL string `json:"imageUrl"`
		Link     string `json:"link"`
		ShopID   string `json:"shopId"`
	}
	type PetPlaceInfoList struct {
		Place  int `json:"place"`
		Energy int `json:"energy"`
	}
	type PetInfo struct {
		AdvertID     string `json:"advertId"`
		NickName     string `json:"nickName"`
		IconURL      string `json:"iconUrl"`
		ClickIconURL string `json:"clickIconUrl"`
		FeedGifURL   string `json:"feedGifUrl"`
		HomePetImage string `json:"homePetImage"`
		CrossBallURL string `json:"crossBallUrl"`
		RunURL       string `json:"runUrl"`
		TickleURL    string `json:"tickleUrl"`
	}
	type GoodsInfo struct {
		GoodsName        string `json:"goodsName"`
		GoodsURL         string `json:"goodsUrl"`
		GoodsID          string `json:"goodsId"`
		ExchangeMedalNum int    `json:"exchangeMedalNum"`
		ActivityID       string `json:"activityId"`
		ActivityIds      string `json:"activityIds"`
	}
	type Result struct {
		ShareCode              string             `json:"shareCode"`
		HisHbFlag              bool               `json:"hisHbFlag"`
		MasterHelpPeoples      []interface{}      `json:"masterHelpPeoples"`
		HelpSwitchOn           bool               `json:"helpSwitchOn"`
		UserStatus             int                `json:"userStatus"`
		TotalEnergy            int                `json:"totalEnergy"`
		MasterInvitePeoples    []interface{}      `json:"masterInvitePeoples"`
		ShareTo                string             `json:"shareTo"`
		PetSportStatus         int                `json:"petSportStatus"`
		UserImage              string             `json:"userImage"`
		MasterHelpReward       int                `json:"masterHelpReward"`
		ShowHongBaoExchangePop bool               `json:"showHongBaoExchangePop"`
		ShowNeedCollectPop     bool               `json:"showNeedCollectPop"`
		PetSportReward         string             `json:"petSportReward"`
		NewhandBubble          bool               `json:"newhandBubble"`
		ResourceList           []ResourceList     `json:"resourceList"`
		ProjectBubble          bool               `json:"projectBubble"`
		MasterInvitePop        bool               `json:"masterInvitePop"`
		MasterInviteReward     int                `json:"masterInviteReward"`
		MedalNum               int                `json:"medalNum"`
		MasterHelpPop          bool               `json:"masterHelpPop"`
		MeetDays               int                `json:"meetDays"`
		PetPlaceInfoList       []PetPlaceInfoList `json:"petPlaceInfoList"`
		MedalPercent           float64            `json:"medalPercent"`
		CharitableSwitchOn     bool               `json:"charitableSwitchOn"`
		PetInfo                PetInfo            `json:"petInfo"`
		NeedCollectEnergy      int                `json:"needCollectEnergy"`
		FoodAmount             int                `json:"foodAmount"`
		InviteCode             string             `json:"inviteCode"`
		RulesURL               string             `json:"rulesUrl"`
		PetStatus              int                `json:"petStatus"`
		GoodsInfo              GoodsInfo          `json:"goodsInfo"`
	}
	type AutoGenerated struct {
		Code       string `json:"code"`
		ResultCode string `json:"resultCode"`
		Message    string `json:"message"`
		Result     Result `json:"result"`
	}
	a := AutoGenerated{}
	req := httplib.Post(`https://api.m.jd.com/client.action?functionId=initPetTown`)
	req.Header("Host", "api.m.jd.com")
	req.Header("User-Agent", ua)
	req.Header("cookie", cookie)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Body(`body={}&appid=wh5&loginWQBiz=pet-town&clientVersion=9.0.4`)
	data, _ := req.Bytes()
	json.Unmarshal(data, &a)
	rt := ""
	if a.Code == "0" && a.ResultCode == "0" && a.Message == "success" {
		if a.Result.UserStatus == 0 {
			rt = "请手动开启活动⏰"
		} else if a.Result.GoodsInfo.GoodsName == "" {
			rt = "你忘了选购新的商品⏰"
		} else if a.Result.PetStatus == 5 {
			rt = a.Result.GoodsInfo.GoodsName + "已可领取⏰"
		} else if a.Result.PetStatus == 6 {
			rt = a.Result.GoodsInfo.GoodsName + "未继续领养新的物品⏰"
		} else {
			rt = a.Result.GoodsInfo.GoodsName + fmt.Sprintf("领养中，进度%.2f%%，勋章%d/%d🐶", a.Result.MedalPercent, a.Result.MedalNum, a.Result.GoodsInfo.ExchangeMedalNum)
		}
	} else {
		rt = "数据异常"
	}
	state <- rt
}

func jsGold(cookie string, state chan int64) { //

	type BalanceVO struct {
		CashBalance       string `json:"cashBalance"`
		EstimatedAmount   string `json:"estimatedAmount"`
		ExchangeGold      string `json:"exchangeGold"`
		FormatGoldBalance string `json:"formatGoldBalance"`
		GoldBalance       int    `json:"goldBalance"`
	}
	type Gears struct {
		Amount         string `json:"amount"`
		ExchangeAmount string `json:"exchangeAmount"`
		Order          int    `json:"order"`
		Status         int    `json:"status"`
		Type           int    `json:"type"`
	}
	type Data struct {
		Advertise      string    `json:"advertise"`
		BalanceVO      BalanceVO `json:"balanceVO"`
		Gears          []Gears   `json:"gears"`
		IsGetCoupon    bool      `json:"isGetCoupon"`
		IsGetCouponEid bool      `json:"isGetCouponEid"`
		IsLogin        bool      `json:"isLogin"`
		NewPeople      bool      `json:"newPeople"`
	}
	type AutoGenerated struct {
		Code      int    `json:"code"`
		Data      Data   `json:"data"`
		IsSuccess bool   `json:"isSuccess"`
		Message   string `json:"message"`
		RequestID string `json:"requestId"`
	}
	a := AutoGenerated{}
	req := httplib.Post(`https://api.m.jd.com?functionId=MyAssetsService.execute&appid=market-task-h5`)
	req.Header("Accept", "application/json, text/plain, */*")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Cookie", cookie)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Header("Origin", "https://gold.jd.com")
	req.Header("Host", "api.m.jd.com")
	req.Header("Connection", "keep-alive")
	req.Header("User-Agent", ua)
	req.Header("Referer", "https://gold.jd.com/")
	req.Body(`functionId=MyAssetsService.execute&body={"method":"goldShopPage","data":{"channel":1}}&_t=` + fmt.Sprint(time.Now().Unix()) + `&appid=market-task-h5;`)
	data, _ := req.Bytes()
	json.Unmarshal(data, &a)
	state <- int64(a.Data.BalanceVO.GoldBalance)
}

func jxncEgg(cookie string, state chan int64) {
	req := httplib.Get("https://m.jingxi.com/jxmc/queryservice/GetHomePageInfo?channel=7&sceneid=1001&activeid=null&activekey=null&isgift=1&isquerypicksite=1&_stk=activeid%2Cactivekey%2Cchannel%2Cisgift%2Cisquerypicksite%2Csceneid&_ste=1&h5st=20210818211830955%3B4408816258824161%3B10028%3Btk01w8db21b2130ny2eg0siAPpNQgBqjGzYfuG6IP7Z%2BAOB40BiqLQ%2Blglfi540AB%2FaQrTduHbnk61ngEeKn813gFeRD%3Bd9a0b833bf99a29ed726cbffa07ba955cc27d1ff7d2d55552878fc18fc667929&_=1629292710957&sceneval=2&g_login_type=1&g_ty=ls")
	req.Header("User-Agent", ua)
	req.Header("Host", "m.jingxi.com")
	req.Header("Accept", "*/*")
	req.Header("Connection", "keep-alive")
	req.Header("Accept-Language", "zh-cn")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Referer", "https://st.jingxi.com/pingou/jxmc/index.html?nativeConfig=%7B%22immersion%22%3A1%2C%22toColor%22%3A%22%23e62e0f%22%7D&;__mcwvt=sjcp&ptag=7155.9.95")
	req.Header("Cookie", cookie)
	data, _ := req.Bytes()

	egg, _ := jsonparser.GetInt(data, "data", "eggcnt")
	state <- egg
}

func tytCoupon(cookie string, state chan string) {

	type DiscountInfo struct {
		High string        `json:"high"`
		Info []interface{} `json:"info"`
	}
	type ExtInfo struct {
		Num5              string `json:"5"`
		Num12             string `json:"12"`
		Num16             string `json:"16"`
		Num21             string `json:"21"`
		Num52             string `json:"52"`
		Num54             string `json:"54"`
		Num74             string `json:"74"`
		BusinessLabel     string `json:"business_label"`
		LimitOrganization string `json:"limit_organization"`
		UserLabel         string `json:"user_label"`
	}
	type Useable struct {
		AreaDesc         string        `json:"areaDesc"`
		AreaType         int           `json:"areaType"`
		Batchid          string        `json:"batchid"`
		BeanNumForPerson int           `json:"beanNumForPerson"`
		BeanNumForPlat   int           `json:"beanNumForPlat"`
		BeginTime        string        `json:"beginTime"`
		CanBeSell        bool          `json:"canBeSell"`
		CanBeShare       bool          `json:"canBeShare"`
		CompleteTime     string        `json:"completeTime"`
		CouponKind       int           `json:"couponKind"`
		CouponStyle      int           `json:"couponStyle"`
		CouponTitle      string        `json:"couponTitle"`
		Couponid         string        `json:"couponid"`
		Coupontype       int           `json:"coupontype"`
		CreateTime       string        `json:"createTime"`
		Discount         string        `json:"discount"`
		DiscountInfo     DiscountInfo  `json:"discountInfo"`
		EndTime          string        `json:"endTime"`
		ExpireType       int           `json:"expireType"`
		ExtInfo          ExtInfo       `json:"extInfo"`
		HourCoupon       int           `json:"hourCoupon"`
		IsOverlay        int           `json:"isOverlay"`
		LimitStr         string        `json:"limitStr"`
		LinkStr          string        `json:"linkStr"`
		OperateTime      string        `json:"operateTime"`
		OrderID          string        `json:"orderId"`
		OverlayDesc      string        `json:"overlayDesc"`
		PassKey          string        `json:"passKey"`
		Pin              string        `json:"pin"`
		PlatFormInfo     string        `json:"platFormInfo"`
		Platform         int           `json:"platform"`
		PlatformDetails  []interface{} `json:"platformDetails"`
		PwdKey           string        `json:"pwdKey"`
		Quota            string        `json:"quota"`
		SellID           string        `json:"sellId"`
		ShareID          string        `json:"shareId"`
		ShopID           string        `json:"shopId"`
		ShopName         string        `json:"shopName"`
		State            int           `json:"state"`
		UseTime          string        `json:"useTime"`
		VenderID         string        `json:"venderId"`
	}
	type Coupon struct {
		Curtimestamp           int       `json:"curtimestamp"`
		ExpiredCount           int       `json:"expired_count"`
		IsHideBaiTiaoInJxWxapp int       `json:"isHideBaiTiaoInJxWxapp"`
		IsHideMailInWxapp      int       `json:"isHideMailInWxapp"`
		Useable                []Useable `json:"useable"`
		UseableCount           int       `json:"useable_count"`
		UsedCount              int       `json:"used_count"`
	}
	type AutoGenerated struct {
		Coupon    Coupon `json:"coupon"`
		ErrMsg    string `json:"errMsg"`
		ErrorCode int    `json:"errorCode"`
		HasNext   int    `json:"hasNext"`
		Jdpin     string `json:"jdpin"`
		State     int    `json:"state"`
		Uin       string `json:"uin"`
	}
	a := AutoGenerated{}
	req := httplib.Get(`https://m.jingxi.com/activeapi/queryjdcouponlistwithfinance?state=1&wxadd=1&filterswitch=1&_=1629296270692&sceneval=2&g_login_type=1&callback=jsonpCBKB&g_ty=ls`)
	req.Header("Accept", "*/*")
	req.Header("Connection", "keep-alive")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Cookie", cookie)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Header("Host", "m.jingxi.com")

	req.Header("User-Agent", ua)
	req.Header("Referer", "https://st.jingxi.com/my/coupon/jx.shtml?sceneval=2&ptag=7155.1.18")
	data, _ := req.Bytes()
	res := regexp.MustCompile(`jsonpCBKB[(](.*)\s+[)];}catch`).FindSubmatch(data)
	rt := "暂无数据"
	if len(res) > 0 {
		json.Unmarshal(res[1], &a)
		num := 0
		toexp := 0
		tm := int(time.Now().Unix() * 1000)
		for _, cp := range a.Coupon.Useable {
			if strings.Contains(cp.CouponTitle, "推推5.01") {
				num++
				if Int(cp.EndTime) < tm {
					toexp++
				}
			}
		}
		if num == 0 {
			rt = "无优惠券"
		} else {
			rt = fmt.Sprintf("%d张5元优惠券", num)
			if toexp > 0 {
				rt += fmt.Sprintf("(今天将过期%d张)⏰", toexp)
			} else {
				rt += "🎰"
			}
		}
	}
	state <- rt
}

func mmCoin(cookie string, state chan int64) {
	req := httplib.Post(`https://api.m.jd.com/client.action`)
	req.Header("Host", "api.m.jd.com")
	req.Header("Accept", "application/json, text/plain, */*")
	req.Header("Origin", "https://h5.m.jd.com")

	req.Header("User-Agent", ua)
	req.Header("cookie", cookie)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Body(`uuid=3245ad3d16ab2153c69f9ca91cd2e931b06a3bb8&clientVersion=10.1.0&client=wh5&osVersion=&area=&networkType=wifi&functionId=homePageV2&body=%7B%7D&appid=SecKill2020`)
	data, _ := req.Bytes()
	mmc, _ := jsonparser.GetInt(data, "result", "assignment", "assignmentPoints")
	state <- mmc
}

func jdzz(cookie string, state chan int64) { //
	req := httplib.Get(`https://api.m.jd.com/client.action?functionId=interactTaskIndex&body={}&client=wh5&clientVersion=9.1.0`)
	req.Header("Host", "api.m.jd.com")
	req.Header("Accept-Language", "zh-cn")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Referer", "http://wq.jd.com/wxapp/pages/hd-interaction/index/index")
	req.Header("User-Agent", ua)
	req.Header("cookie", cookie)
	req.Header("Content-Type", "application/json")
	data, _ := req.Bytes()
	mmc, _ := jsonparser.GetString(data, "data", "totalNum")
	state <- int64(Int(mmc))
}

func jxgc(cookie string, state chan string) {
	type FactoryList struct {
		Name string `json:"name"`
	}
	type ProductionList struct {
		CommodityDimId   int `json:"commodityDimId"`
		InvestedElectric int `json:"investedElectric"`
		NeedElectric     int `json:"needElectric"`
		Status           int `json:"status"`
		ProductionId     int `json:"ProductionId"`
		ExchangeStatus   int `json:"exchangeStatus"`
	}
	type CommodityList struct {
		Name string `json:"name"`
	}
	type Data struct {
		FactoryList    []FactoryList    `json:"factoryList"`
		ProductionList []ProductionList `json:"productionList"`
		CommodityList  []CommodityList  `json:"commodityList"`
	}
	type AutoGenerated struct {
		Ret  int    `json:"ret"`
		Data Data   `json:"data"`
		Msg  string `json:"msg"`
	}
	a := AutoGenerated{}
	a1 := AutoGenerated{}
	_stk := "_time%2CmaterialTuanId%2CmaterialTuanPin%2Cpin%2CsharePin%2CshareType%2Csource%2Czone"
	req := jxGcFuncName(cookie, "userinfo/GetUserInfo?zone=dream_factory&pin=&sharePin=&shareType=&materialTuanPin=&materialTuanId=&source=", _stk)
	data, _ := req.Bytes()
	json.Unmarshal(data, &a)
	msg := "获取失败"
	if a.Ret == 0 {
		if len(a.Data.FactoryList) > 0 && len(a.Data.ProductionList) > 0 {
			production := a.Data.ProductionList[0]
			xq := jxGcFuncName(cookie, fmt.Sprintf("diminfo/GetCommodityDetails?zone=dream_factory&commodityId=%d", a.Data.ProductionList[0].CommodityDimId), _stk)
			dataInfo, _ := xq.Bytes()
			json.Unmarshal(dataInfo, &a1)
			name := ""
			if a1.Data.CommodityList != nil {
				name = a1.Data.CommodityList[0].Name
			}
			msg = fmt.Sprintf(`%s ,进度: %.2f`, name, (float64(production.InvestedElectric)/float64(production.NeedElectric))*100)
			if production.InvestedElectric >= production.NeedElectric {
				if production.ExchangeStatus == 1 {
					msg = name + `,已经可兑换，请手动兑换`
				}
				if production.ExchangeStatus == 3 {
					if time.Now().Hour() == 9 {
						msg = name + `,兑换已超时，请选择新商品进行制造`
					}
				}
			} else {
				msg = msg + fmt.Sprintf(`,预计:%.2f天可兑换`, float64((production.NeedElectric-production.InvestedElectric)/(2*60*60*24)))
			}
			if production.Status == 3 {
				msg = name + ",已经超时失效, 请选择新商品进行制造"
			}
		} else {
			if len(a.Data.FactoryList) == 0 {
				msg = "当前未开始生产商品,请手动去京东APP->游戏与互动->查看更多->京喜工厂 开启活动"
				// $.msg($.name, '【提示】', `京东账号${$.index}[${$.nickName}]京喜工厂活动未开始\n请手动去京东APP->游戏与互动->查看更多->京喜工厂 开启活动`);
			} else if len(a.Data.FactoryList) > 0 && len(a.Data.ProductionList) == 0 {
				msg = "当前未开始生产商品,请手动去京东APP->游戏与互动->查看更多->京喜工厂 开启活动"
			}
		}
	}
	state <- msg
}

func jxGcFuncName(cookie string, body string, _stk string) *httplib.BeegoHTTPRequest {
	now := time.Now()
	duration, _ := time.ParseDuration("48h")
	req := httplib.Get(fmt.Sprintf(`https://m.jingxi.com/dreamfactory/%s?zone=dream_factory&pin=&sharePin=&shareType=&materialTuanPin=&materialTuanId=&source=&sceneval=2&g_login_type=1&_time=%s&_=%s&_ste=1&_stk=%s`, body, fmt.Sprint(now.Unix()), fmt.Sprint(now.Add(duration).Unix()), _stk))
	req.Header("Host", "api.m.jd.com")
	req.Header("Accept-Language", "zh-cn")
	req.Header("Accept-Encoding", "gzip, deflate, br")
	req.Header("Referer", "http://wq.jd.com/wxapp/pages/hd-interaction/index/index")
	req.Header("User-Agent", ua)
	req.Header("cookie", cookie)
	req.Header("Content-Type", "application/json")
	return req
}

// func jxgc() {
// 	req := httplib.Get(fmt.Sprintf(`https://m.jingxi.com/dreamfactory/userinfo/GetUserInfo?zone=dream_factory&pin=&sharePin=&shareType=&materialTuanPin=&materialTuanId=&source=&sceneval=2&g_login_type=1&_time=${Date.now()}&_=${Date.now() + 2}&_ste=1`))
// 	req.Header("Host", "api.m.jd.com")
// 	req.Header("Accept-Language", "zh-cn")
// 	req.Header("Accept-Encoding", "gzip, deflate, br")
// 	req.Header("Referer", "http://wq.jd.com/wxapp/pages/hd-interaction/index/index")
// 	req.Header("User-Agent", ua)
// 	req.Header("cookie", cookie)
// 	req.Header("Content-Type", "application/json")
// 	data, _ := req.Bytes()
// }

// // 惊喜的Taskurl
// function jxTaskurl(functionId, body = '', stk) {
// 	let url = `https://m.jingxi.com/dreamfactory/${functionId}?zone=dream_factory&${body}&sceneval=2&g_login_type=1&_time=${Date.now()}&_=${Date.now() + 2}&_ste=1`
// 	url += `&h5st=${decrypt(Date.now(), stk, '', url)}`
// 	if (stk) {
// 	    url += `&_stk=${encodeURIComponent(stk)}`;
// 	}
// 	return {
// 	    url,
// 	    headers: {
// 		   'Cookie': cookie,
// 		   'Host': 'm.jingxi.com',
// 		   'Accept': '*/*',
// 		   'Connection': 'keep-alive',
// 		   'User-Agent': functionId === 'AssistFriend' ? "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36" : 'jdpingou',
// 		   'Accept-Language': 'zh-cn',
// 		   'Referer': 'https://wqsd.jd.com/pingou/dream_factory/index.html',
// 		   'Accept-Encoding': 'gzip, deflate, br',
// 	    }
// 	}
//  }
