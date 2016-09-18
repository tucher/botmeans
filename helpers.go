package botmeans

import (
	// 	"encoding/json"
	// 	"fmt"
	// 	"github.com/go-telegram-bot-api/telegram-bot-api"
	// 	"io/ioutil"
	"reflect"
	// 	"strings"
	// 	"text/template"
	// "log"
)

type Identifiable interface {
	Id() int64
}

type UserIdentifier interface {
	UserId() int64
}

func (ui *MeansBot) Find(val ...Identifiable) (err error) {
	for _, v := range val {
		if err = ui.db.Where("id=?", v.Id()).First(v).Error; err != nil {
			return
		}
	}
	return
}

func (ui *MeansBot) FindSession(id int64) SessionInterface {
	r := Session{}
	if err := ui.db.Where("id=?", id).First(&r).Error; err != nil {
		return &r
	}
	return nil
}

//GetNeighborSessions returns all sessions that are in the same chat as given session
func (ui *MeansBot) GetChatSessions(session SessionInterface) (ret []SessionInterface) {

	sessions := []*Session{}
	ui.db.Where("telegram_chat_id=?", session.ChatId()).Find(&sessions)
	for _, s := range sessions {
		ret = append(ret, s)
	}
	return
}

//SetChatLocale sets the locale of all sessions in this chat
func (ui *MeansBot) SetChatLocale(session SessionInterface, locale string) {
	session.SetLocale(locale)
	sL := []Session{}
	ui.db.Where("telegram_chat_id=?", session.ChatId()).Find(&sL)
	for _, s := range sL {
		s.SetLocale(locale)
		s.db = ui.db
		s.Save()
	}

}

// //GetSessionsByUserData returns all sessions in this chat that have given value in UserData field
// func (ui *MeansBot) GetSessionsByUserData(filters map[string]interface{}) (ret []TelegramUserSession) {
// 	query := ui.db.Model(TelegramUserSession{})
// 	for k, v := range filters {
// 		query = query.Where("(user_data->> ?)::text = ?::text", k, v)
// 	}
// 	query.Find(&ret)
// 	return
// }

//GetSessionsByTelegramUserID returns all sessions with given Telegram User ID
func (ui *MeansBot) GetUserSessions(session SessionInterface) (ret []SessionInterface) {
	s := []*Session{}
	ui.db.Where("telegram_user_id=?", session.UserId()).Find(&s)
	for _, ses := range s {
		ret = append(ret, ses)
	}
	return
}

// //GetBotMessagesByChatAndMsgID returns Bot message from session's chat with given MsgID
// func (ui *MeansBot) GetBotMessagesByChatAndMsgID(session *TelegramUserSession, msgID int64) (ret TelegramBotMessage) {
// 	ui.db.Where("telegram_chat_id=? and telegram_msg_id=?", session.TelegramChatID, msgID).First(&ret)
// 	return
// }

//GetBotMessagesByChatAndType returns Bot messages from session's chat with given UserData type
func (ui *MeansBot) GetBotMessagesByChatAndType(session SessionInterface, typeSample interface{}) (ret []BotMessageInterface) {
	t := reflect.TypeOf(typeSample).Name()
	msgs := []*BotMessage{}
	ui.db.Where("telegram_chat_id=?", session.ChatId()).
		Where("jsonb_exists(user_data, ?::text)", t).Find(&msgs)
	for _, m := range msgs {
		ret = append(ret, m)
	}
	return
}

// //FindBotMessagesByChatAndUserData returns Bot messages from session's chat with given UserData type and given values in UserData
// func (ui *MeansBot) FindBotMessagesByChatAndUserData(session TelegramUserSession, typeSample interface{}, filters map[string]interface{}) (ret []TelegramBotMessage) {
// 	t := reflect.TypeOf(typeSample).Name()
// 	query := ui.db.Where("telegram_chat_id=?", session.TelegramChatID).
// 		Where("(user_data->>'Type')::text = ?::text", t)
// 	for k, v := range filters {
// 		query = query.Where("(user_data->'UserData'->> ?)::text = ?::text", k, v)
// 	}
// 	query.Find(&ret)
// 	return
// }

// //CreateSession creates the new session for given credentials
// func (ui *MeansBot) CreateSession(ses *TelegramUserSession) (session *TelegramUserSession, err error) {
// 	session, err = ui.findOrCreateSession(ses.TelegramChatID, ses.TelegramUserID, ses.TelegramUserName)
// 	return
// }
