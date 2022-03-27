package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID       int
	Number   int `gorm:"unique"`
	Class    string
	ActiveAt time.Time
	Coin     float64
    Womail   string
}
func ClearCoin(uid int) float64 {
	var u User
	if db.Where("number = ?", uid).First(&u).Error != nil {
		return 0
	}
	db.Model(u).Updates(map[string]interface{}{
		"coin": gorm.Expr(fmt.Sprintf("%d",1)),
	})
	u.Coin=1
	return u.Coin
}
func AdddCoin(uid int , num float64) float64 {
	var u User
	if db.Where("number = ?", uid).First(&u).Error != nil {
		return 0
	}
	db.Model(u).Updates(map[string]interface{}{
		"coin": gorm.Expr(fmt.Sprintf("coin+%d",num)),
	})
	u.Coin+=num
	return u.Coin
}
func AddCoin(uid int) float64 {
	var u User
	if db.Where("number = ?", uid).First(&u).Error != nil {
		return 0
	}
	db.Model(u).Updates(map[string]interface{}{
		"coin": gorm.Expr("coin+1"),
	})
	u.Coin++
	return u.Coin
}

func RemCoin(uid int, num float64) float64 {
	var u User
	db.Where("number = ?", uid).First(&u)
	db.Model(u).Updates(map[string]interface{}{
		"coin": gorm.Expr(fmt.Sprintf("coin-%f", num)),
	})
	u.Coin -= num
	return u.Coin
}

func GetCoin(uid int) float64 {
	var u User
	db.Where("number = ?", uid).First(&u)
	return u.Coin
}
