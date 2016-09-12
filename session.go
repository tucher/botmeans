package botmeans

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"time"
)

type SessionBase struct {
	TelegramUserID   int64  `sql:"index"`
	TelegramUserName string `sql:"index"`
	TelegramChatID   int64  `sql:"index"`
	isNew            bool
	isLeft           bool
}

type Session struct {
	SessionBase
	ID        int64  `sql:"index;unique"`
	UserData  string `sql:"type:jsonb"`
	db        *gorm.DB
	FirstName string
	LastName  string
	ChatName  string
	CreatedAt time.Time
}

func (session *Session) IsNew() bool {
	return session.isNew
}

func (session *Session) IsLeft() bool {
	return session.isLeft
}

func (session *Session) ChatId() int64 {
	return session.TelegramChatID
}

//SetData sets internal UserData field to JSON representation of given value
func (session *Session) SetData(value interface{}) {
	session.UserData = serialize(session.UserData, value)
}

//GetData extracts internal UserData field to given value
func (session *Session) GetData(value interface{}) {
	deserialize(session.UserData, value)
}

//GetData extracts internal UserData field to given value
func (session *Session) Save() error {
	if session.db != nil {
		return session.db.Save(session).Error
	} else {
		return fmt.Errorf("db not set")
	}
}

func (session *Session) Locale() string {
	type Locale string

	var lo Locale
	session.GetData(&lo)
	return string(lo)
}

func (session *Session) String() string {

	return fmt.Sprintf("UserID: %v, UserName: %v, ChatID: %v, New: %v, Left: %v, Data: %v, Name: %v %v, Locale: %v",
		session.TelegramUserID,
		session.TelegramUserName,
		session.TelegramChatID,
		session.isNew,
		session.isLeft,
		session.UserData,
		session.FirstName,
		session.LastName,
		session.Locale(),
	)
}

func SessionInitDB(db *gorm.DB) {
	db.AutoMigrate(&Session{})
}
func SessionLoader(base SessionBase, db *gorm.DB, BotName string, BotID int64, api *tgbotapi.BotAPI) (SessionInterface, error) {
	TelegramUserID := base.TelegramUserID
	TelegramUserName := base.TelegramUserName
	TelegramChatID := base.TelegramChatID
	if TelegramUserID == 0 && TelegramUserName == "" {
		return nil, fmt.Errorf("Invalid session IDs")
	}
	//TODO!
	if TelegramUserID == BotID || TelegramUserName == BotName {
		return nil, fmt.Errorf("Cannot create the session for myself")
	}
	session := &Session{}
	session.db = db
	found := !db.Where("((telegram_user_id=? and telegram_user_id!=0) or (telegram_user_name=? and telegram_user_name!='')) and telegram_chat_id=?", TelegramUserID, TelegramUserName, TelegramChatID).
		First(session).RecordNotFound()
	err := fmt.Errorf("Unknown")
	if api != nil && (!found || session.FirstName == "" && session.LastName == "") {
		if chatMember, err := api.GetChatMember(tgbotapi.ChatConfigWithUser{TelegramChatID, "", int(TelegramUserID)}); err == nil {
			session.FirstName = chatMember.User.FirstName
			session.LastName = chatMember.User.LastName
		}
	}
	if !found {
		session.isNew = true
		session.TelegramChatID = TelegramChatID
		session.TelegramUserID = TelegramUserID
		session.TelegramUserName = TelegramUserName
		session.CreatedAt = time.Now()
		if api != nil {
			if chat, err := api.GetChat(tgbotapi.ChatConfig{ChatID: session.TelegramChatID}); err == nil {
				session.ChatName = chat.Title
			}
		}
		session.UserData = "{}"
		err = nil

	}

	return session, err
}
