package botmeans

import (
	"time"

	"github.com/jinzhu/gorm"
)

//SessionBase passes core session identifiers
type Session struct {
	ID              int64  `sql:"index;unique"`
	UserID          int64  `sql:"index"`
	MessangerID     string `sql:"index"`
	MessangerUserID string `sql:"index"`
	MessangerChatID string `sql:"index"`

	hasCome bool
	hasLeft bool

	messenger MessengerAdapter
	user      User

	MessengerData string `sql:"type:jsonb"`
	CreatedAt     time.Time
	isNew         bool
}

//IsNew should return true if the session has not been saved yet
func (session *Session) IsNew() bool {
	return session.isNew
}

//HasLeft returns true if the user has gone from chat
func (session *Session) HasLeft() bool {
	return session.hasLeft
}

//HasCome returns true if the user has come to chat
func (session *Session) HasCome() bool {
	return session.hasCome
}

//IsOneToOne should return true if the session represents one-to-one chat with bot
func (session *Session) IsOneToOne() bool {
	return messenger.IsOneToOne(session)
}

//SetData sets internal UserData field to JSON representation of given value

//Locale returns the locale for this user
func (session *Session) Locale() string {
	type Locale string

	var lo Locale
	session.GetData(&lo)
	return string(lo)
}

func (session *Session) SetLocale(locale string) {
	type Locale string
	var lo Locale = Locale(locale)
	session.SetData(lo)
}

func (session *Session) Messenger() MessengerAdapter {
	return session.messenger
}

//SessionInitDB creates sql table for Session
func SessionInitDB(db *gorm.DB) {
	db.AutoMigrate(&Session{})
}

//SessionInterface defines the user session
type SessionInterface interface {
	DataGetSetter
	IsNew() bool
	HasLeft() bool
	HasCome() bool
	Locale() string
	SetLocale(string)
	IsOneToOne() bool
	Messenger() MessengerAdapter
	User() *User
}
