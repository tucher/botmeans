package botmeans

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type BotMessageInterface interface {
	PersistentSaver
	DataGetSetter
	Sendable
	Id() int64
	CallbackID() string
	SetID(int64)
}

type BotMessage struct {
	ID             int64  `sql:"index;unique"`
	TelegramMsgID  int64  `sql:"index"`
	TelegramChatID int64  `sql:"index"`
	UserData       string `sql:"type:jsonb"`
	db             *gorm.DB
	callbackID     string
	Timestamp      time.Time
}

//Set sets internal UserData field to JSON representation of given value.
//Automatically saves information for the value's type, so you don't need to care about it.
//Just use the same types of value for SetData and GetData
func (botMessage *BotMessage) SetData(value interface{}) {
	botMessage.UserData = serialize(botMessage.UserData, value)
}

//GetData extracts internal UserData field to given value
func (botMessage *BotMessage) GetData(value interface{}) {
	deserialize(botMessage.UserData, value)
}

func (botMessage *BotMessage) Save() error {
	if botMessage.db != nil {
		return botMessage.db.Save(botMessage).Error
	} else {
		return fmt.Errorf("db not set")
	}
}

func BotMessageInitDB(db *gorm.DB) {
	db.AutoMigrate(&BotMessage{})
}

func (botMessage *BotMessage) Send() bool {
	botMessage.Save()
	return false
}

func (botMessage *BotMessage) Id() int64 {
	return botMessage.TelegramMsgID
}

func (botMessage *BotMessage) SetID(i int64) {
	botMessage.TelegramMsgID = i
}

func (botMessage *BotMessage) CallbackID() string {
	return botMessage.callbackID
}

func BotMessageDBLoader(TelegramChatID int64, TelegramMsgID int64, CallbackID string, db *gorm.DB) BotMessageInterface {
	ret := &BotMessage{}
	db.Where("telegram_chat_id=? and telegram_msg_id=?", TelegramChatID, TelegramMsgID).Find(ret)
	ret.db = db
	ret.callbackID = CallbackID
	ret.TelegramChatID = TelegramChatID
	ret.TelegramMsgID = TelegramMsgID
	return ret
}

func NewBotMessage(TelegramChatID int64, db *gorm.DB) BotMessageInterface {
	ret := &BotMessage{}
	ret.db = db
	ret.TelegramChatID = TelegramChatID
	ret.UserData = "{}"
	return ret
}
