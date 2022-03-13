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
