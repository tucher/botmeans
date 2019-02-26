package botmeans

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	ID       int64  `sql:"index;unique"`
	UserData string `sql:"type:jsonb"`
}

//SetData sets internal UserData field to JSON representation of given value
func (user *User) SetData(db *gorm.DB, value interface{}) {

	s := User{}
	if session.db.Where("id=?", session.ID).First(&s).Error == nil {
		user.UserData = s.UserData
	}

	user.UserData = serialize(user.UserData, value)
	user.Save()
}

//GetData extracts internal UserData field to given value
func (user *User) GetData(db *gorm.DB, value interface{}) {
	deserialize(user.UserData, value)
}

func (user *User) Save(db *gorm.DB) error {
	if err := db.Save(user).Error; err == nil {
		return nil
	} else {
		return err
	}
}

//SessionInitDB creates sql table for Session
func UserInitDB(db *gorm.DB) {
	db.AutoMigrate(&User{})
}
