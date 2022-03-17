package models

import (
	"fmt"
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
		ids := "XDD" + id.String()
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
	if err == nil {
		if u.Use != true {
			var user = &User{}
			err := db.Where("Number = ?", use).First(&user).Error
			if err != nil {
				db.Create(&User{
					Class:    "qq",
					Number:   use,
					Coin:     u.Value,
					ActiveAt: time.Now(),
					Womail:   "",
				})
			} else {
				user.Coin += u.Value
				db.Where("Number = ?", use).Updates(user)
			}
			u.UseBy = use
			u.Use = true
			db.Where("Token = ?", id).Updates(u)
			return fmt.Sprintf("使用成功,积分增加%d", u.Value)
		} else {
			return "卡密已被使用"
		}
	}
	return "查无此卡"
}
