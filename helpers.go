package botmeans

// import (
// 	"encoding/json"
// 	"fmt"
// 	"github.com/go-telegram-bot-api/telegram-bot-api"
// 	"io/ioutil"
// 	"reflect"
// 	"strings"
// 	"text/template"
// )

// //GetNeighborSessions returns all sessions that are in the same chat as given session
// func (ui *MeansBot) GetNeighborSessions(session *TelegramUserSession) (ret []TelegramUserSession) {
// 	ui.db.Where("telegram_chat_id=?", session.TelegramChatID).Find(&ret)
// 	return
// }

// //SetChatLocale sets the locale of all sessions in this chat
// func (ui *MeansBot) SetChatLocale(session *TelegramUserSession, locale string) {
// 	session.Locale = locale
// 	ui.db.Table("telegram_user_sessions").Where("telegram_chat_id=?", session.TelegramChatID).
// 		Updates(map[string]interface{}{"locale": locale})
// }

// //GetSessionsByUserData returns all sessions in this chat that have given value in UserData field
// func (ui *MeansBot) GetSessionsByUserData(filters map[string]interface{}) (ret []TelegramUserSession) {
// 	query := ui.db.Model(TelegramUserSession{})
// 	for k, v := range filters {
// 		query = query.Where("(user_data->> ?)::text = ?::text", k, v)
// 	}
// 	query.Find(&ret)
// 	return
// }

// //GetSessionsByTelegramUserID returns all sessions with given Telegram User ID
// func (ui *MeansBot) GetSessionsByTelegramUser(session *TelegramUserSession) (ret []TelegramUserSession) {
// 	ui.db.Where("telegram_user_id=? or (telegram_user_name=? and telegram_user_name != '')", session.TelegramUserID, session.TelegramUserName).Find(&ret)
// 	return
// }

// //GetBotMessagesByChatAndMsgID returns Bot message from session's chat with given MsgID
// func (ui *MeansBot) GetBotMessagesByChatAndMsgID(session *TelegramUserSession, msgID int64) (ret TelegramBotMessage) {
// 	ui.db.Where("telegram_chat_id=? and telegram_msg_id=?", session.TelegramChatID, msgID).First(&ret)
// 	return
// }

// //GetBotMessagesByChatAndType returns Bot messages from session's chat with given UserData type
// func (ui *MeansBot) GetBotMessagesByChatAndType(session *TelegramUserSession, typeSample interface{}) (ret []TelegramBotMessage) {
// 	t := reflect.TypeOf(typeSample).Name()
// 	ui.db.Where("telegram_chat_id=?", session.TelegramChatID).
// 		Where("(user_data->>'Type')::text = ?::text", t).Find(&ret)
// 	return
// }

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
