package models

import (
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

type Key struct {
	Expiration time.Time
	Token      string
	Use        bool
	UseBy      int
	Value      int
}

func createKey(num int, value int) string {
	var str []string
	for i := 0; i < num; i++ {
		id := uuid.NewV4()
		ids := id.String()
		var u Key
		u = Key{
			Expiration: time.Now(),
			Token:      ids,
			Use:        false,
			Value:      value,
		}
		if err := db.Create(&u).Error; err != nil {
			return err.Error()
		} else {
			str = append(str, ids)
		}
	}
	return strings.Join(str, "\n")
}

func useKey(id string, use int) string {
	var u Key
	err := db.Where("Token = ?", id).First(&u).Error
	if err != nil {
		if u.Use != true {
			var user = &User{}
			err := db.Where("Number = ?", use).First(&user).Error
			if err == nil {
				user.Coin = u.Value
				db.Create(user)
			}
			u.UseBy = use
			db.Updates(u).Where(u.Token)
			return "充值成功"
		} else {
			return "卡密已被使用"
		}
	}
	return "查无此卡"
}
