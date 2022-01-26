package models

import "time"

type Cache struct {
	key      string `gorm:"column:Key;primaryKey"`
	Type     string
	value    string
	ActiveAt string
}

func getCache(key string) (value string) {
	u := &Cache{}
	format := "2006-01-02 15:04:05"
	err := db.Where("key = ? and active_at < ?", key, time.Now().Format(format)).First(&u).Error
	if err == nil {
		return u.value
	} else {
		return ""
	}
}

//func saveCache(cache Cache) (flag bool) {
//	u := &Cache{}
//	err := db.Where("key = ? and active_at < ?", key, time.Now()).First(&u).Error
//	if err == nil {
//		if u.ActiveAt < time.Now() {
//			db.Where("ID = ?", u.ID).Updates(&Limit{
//				Num: u.Num + 1,
//			})
//			return true
//		} else {
//			return false
//		}
//	} else {
//		return u.value
//	}
//	begin := db.Begin()
//	begin.Create(cache)
//	commit := begin.Commit()
//
//}
