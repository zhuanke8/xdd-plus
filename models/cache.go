package models

import "github.com/beego/beego/v2/core/logs"

type Cache struct {
	key      string `gorm:"column:Key;primaryKey"`
	Type     string
	value    string
	ActiveAt string
}

func getCache(key string) (value string) {
	u := &Cache{}
	//format := "2006-01-02 15:04:05"
	err := db.Where("key = ? and active_at < ?", key, Date()).First(&u).Error
	if err == nil {
		return u.value
	} else {
		return ""
	}
}

func saveCache(key string, value string) (flag bool) {
	u := &Cache{}
	err := db.Where("key = ? and active_at = ?", key, Date()).First(&u).Error
	if err == nil {
		logs.Info("为空不报错")
		if u.value != "" {
			db.Where("ID = ?", u.key).Updates(&Cache{
				value: value,
			})
			return true
		} else {
			u.ActiveAt = Date()
			u.key = key
			u.value = value
			begin := db.Begin()
			begin.Create(u)
			begin.Commit()
			return true
		}
	} else {
		return false
	}

}
