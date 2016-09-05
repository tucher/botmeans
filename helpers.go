package botmeans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"reflect"
	"strings"
	"text/template"
)

func (ui MeansBot) findOrCreateSession(TelegramChatID, TelegramUserID int64, TelegramUserName string) (*TelegramUserSession, error) {
	if TelegramUserID == 0 && TelegramUserName == "" {
		return nil, fmt.Errorf("Invalid session IDs")
	}
	//TODO!
	if TelegramUserID == ui.telegramID || TelegramUserName == ui.BotName {
		return nil, fmt.Errorf("Cannot create the session for myself")
	}
	session := &TelegramUserSession{}
	found := !ui.db.Where("((telegram_user_id=? and telegram_user_id!=0) or (telegram_user_name=? and telegram_user_name!='')) and telegram_chat_id=?", TelegramUserID, TelegramUserName, TelegramChatID).
		First(session).RecordNotFound()
	err := fmt.Errorf("Unknown")
	if !found || session.FirstName == "" && session.LastName == "" {
		if chatMember, err := ui.bot.GetChatMember(tgbotapi.ChatConfigWithUser{TelegramChatID, "", int(TelegramUserID)}); err == nil {
			session.FirstName = chatMember.User.FirstName
			session.LastName = chatMember.User.LastName
		} else {
			ui.Logger.Println(err)
		}
	}
	if !found {
		session.TelegramChatID = TelegramChatID
		session.TelegramUserID = TelegramUserID
		session.TelegramUserName = TelegramUserName

		if chat, err := ui.bot.GetChat(tgbotapi.ChatConfig{ChatID: session.TelegramChatID}); err == nil {
			session.ChatName = chat.Title
		}
		session.UserData = "{}"
		err = ui.db.Save(session).Error
		ui.NewSessionCallback(session)
		err = ui.db.Save(session).Error
		ui.AfterNewSessionCallback(session)
	} else {
		err = nil
	}
	// ui.Logger.Printf("%+v", session)
	return session, err
}

func (ui *MeansBot) sendMessage(session *TelegramUserSession, msg *renderedOutputMessage) error {
	toSent := tgbotapi.NewMessage(session.TelegramChatID, msg.text)
	toSent.ParseMode = msg.ParseMode
	if msg.replyKbdMarkup != nil {
		toSent.ReplyMarkup = *msg.replyKbdMarkup
	}

	if msg.inlineKbdMarkup != nil {
		toSent.ReplyMarkup = *msg.inlineKbdMarkup
	}
	if sentMsg, err := ui.bot.Send(toSent); err == nil {
		telegramBotMessage := TelegramBotMessage{
			OwnerUserID:    session.TelegramUserID,
			TelegramChatID: session.TelegramChatID,
			TelegramMsgID:  int64(sentMsg.MessageID),
		}
		telegramBotMessage.SetData(msg.msg.UserData)
		ui.db.Save(&telegramBotMessage)
	} else {
		return err
	}
	return nil
}

func (ui *MeansBot) editMessage(session *TelegramUserSession, msg *renderedOutputMessage) error {
	editConfig := tgbotapi.NewEditMessageText(session.TelegramChatID, int(msg.msg.TelegramMsgIDToEdit), msg.text)

	if msg.inlineKbdMarkup != nil {
		editConfig.ReplyMarkup = msg.inlineKbdMarkup
	}
	editConfig.ParseMode = msg.ParseMode
	if _, err := ui.bot.Send(editConfig); err == nil {

		telegramBotMessage := TelegramBotMessage{
			TelegramMsgID:  msg.msg.TelegramMsgIDToEdit,
			TelegramChatID: session.TelegramChatID,
		}
		if !ui.db.Where(telegramBotMessage).First(&telegramBotMessage).RecordNotFound() {
			telegramBotMessage.SetData(msg.msg.UserData)
			ui.db.Save(&telegramBotMessage)
		}
		return nil
	} else {
		return err
	}
}

func readMsgTemplate(path string) (ret MessageTemplate, err error) {
	c := []byte{}
	c, err = ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Template not found: %v", path)
		return
	}
	c = []byte(strings.Replace(string(c), "\n", "", -1))
	err = json.Unmarshal(c, &ret)
	return
}

func (msg *renderedOutputMessage) renderFromTemplate(templateDir string, locale string, templ *template.Template) error {
	msgTemplate, err := readMsgTemplate(templateDir + "/" + msg.msg.TemplateName + ".json")
	if err != nil {
		return err
	}
	msg.ParseMode = msgTemplate.ParseMode

	if _, ok := msgTemplate.Template[locale]; !ok {
		locale = ""
	}
	msg.text, _ = renderText(msgTemplate.Template[locale], msg.msg.UserData, templ)
	msg.inlineKbdMarkup = createInlineKeyboard(msgTemplate.Keyboard[locale])
	msg.replyKbdMarkup = createReplyKeyboard(msgTemplate.ReplyKeyboard[locale])
	return nil
}

func createInlineKeyboard(buttons [][]MessageButton) *tgbotapi.InlineKeyboardMarkup {
	var inlineKbdMarkup tgbotapi.InlineKeyboardMarkup
	if len(buttons) > 0 {
		kbdRows := [][]tgbotapi.InlineKeyboardButton{}

		for _, row := range buttons {
			r := []tgbotapi.InlineKeyboardButton{}
			for _, btnData := range row {
				r = append(r, tgbotapi.NewInlineKeyboardButtonData(btnData.Text, btnData.Command+" "+btnData.Args))
			}
			kbdRows = append(kbdRows, r)
		}
		inlineKbdMarkup = tgbotapi.NewInlineKeyboardMarkup(kbdRows...)
	} else {
		return nil
	}
	return &inlineKbdMarkup
}

func createReplyKeyboard(buttons [][]MessageButton) *tgbotapi.ReplyKeyboardMarkup {
	var replyKbdMarkup tgbotapi.ReplyKeyboardMarkup
	if len(buttons) > 0 {
		kbdRows := [][]tgbotapi.KeyboardButton{}

		for _, row := range buttons {
			r := []tgbotapi.KeyboardButton{}
			for _, btnData := range row {
				r = append(r, tgbotapi.NewKeyboardButton(btnData.Text))
			}
			kbdRows = append(kbdRows, r)
		}
		replyKbdMarkup = tgbotapi.NewReplyKeyboard(kbdRows...)
		replyKbdMarkup.OneTimeKeyboard = true
	} else {
		return nil
	}
	return &replyKbdMarkup
}

func renderText(templateText string, data interface{}, templ *template.Template) (text string, err error) {
	if t, e := templ.Parse(templateText); err != nil {
		err = e
	} else {
		bfr := &bytes.Buffer{}
		if e := t.Execute(bfr, data); err != nil {
			err = e
			text = templateText
		} else {
			text = bfr.String()
		}
	}
	return
}

//GetNeighborSessions returns all sessions that are in the same chat as given session
func (ui *MeansBot) GetNeighborSessions(session *TelegramUserSession) (ret []TelegramUserSession) {
	ui.db.Where("telegram_chat_id=?", session.TelegramChatID).Find(&ret)
	return
}

//SetChatLocale sets the locale of all sessions in this chat
func (ui *MeansBot) SetChatLocale(session *TelegramUserSession, locale string) {
	session.Locale = locale
	ui.db.Table("telegram_user_sessions").Where("telegram_chat_id=?", session.TelegramChatID).
		Updates(map[string]interface{}{"locale": locale})
}

//GetSessionsByUserData returns all sessions in this chat that have given value in UserData field
func (ui *MeansBot) GetSessionsByUserData(filters map[string]interface{}) (ret []TelegramUserSession) {
	query := ui.db.Model(TelegramUserSession{})
	for k, v := range filters {
		query = query.Where("(user_data->> ?)::text = ?::text", k, v)
	}
	query.Find(&ret)
	return
}

//GetSessionsByTelegramUserID returns all sessions with given Telegram User ID
func (ui *MeansBot) GetSessionsByTelegramUser(session *TelegramUserSession) (ret []TelegramUserSession) {
	ui.db.Where("telegram_user_id=? or (telegram_user_name=? and telegram_user_name != '')", session.TelegramUserID, session.TelegramUserName).Find(&ret)
	return
}

//GetBotMessagesByChatAndMsgID returns Bot message from session's chat with given MsgID
func (ui *MeansBot) GetBotMessagesByChatAndMsgID(session *TelegramUserSession, msgID int64) (ret TelegramBotMessage) {
	ui.db.Where("telegram_chat_id=? and telegram_msg_id=?", session.TelegramChatID, msgID).First(&ret)
	return
}

//GetBotMessagesByChatAndType returns Bot messages from session's chat with given UserData type
func (ui *MeansBot) GetBotMessagesByChatAndType(session *TelegramUserSession, typeSample interface{}) (ret []TelegramBotMessage) {
	t := reflect.TypeOf(typeSample).Name()
	ui.db.Where("telegram_chat_id=?", session.TelegramChatID).
		Where("(user_data->>'Type')::text = ?::text", t).Find(&ret)
	return
}

//FindBotMessagesByChatAndUserData returns Bot messages from session's chat with given UserData type and given values in UserData
func (ui *MeansBot) FindBotMessagesByChatAndUserData(session TelegramUserSession, typeSample interface{}, filters map[string]interface{}) (ret []TelegramBotMessage) {
	t := reflect.TypeOf(typeSample).Name()
	query := ui.db.Where("telegram_chat_id=?", session.TelegramChatID).
		Where("(user_data->>'Type')::text = ?::text", t)
	for k, v := range filters {
		query = query.Where("(user_data->'UserData'->> ?)::text = ?::text", k, v)
	}
	query.Find(&ret)
	return
}

//CreateSession creates the new session for given credentials
func (ui *MeansBot) CreateSession(ses *TelegramUserSession) (session *TelegramUserSession, err error) {
	session, err = ui.findOrCreateSession(ses.TelegramChatID, ses.TelegramUserID, ses.TelegramUserName)
	return
}

func parseCmdAliasesFromTemplates(dir string) (ret map[string]string, err error) {
	ret = make(map[string]string)
	path := dir + "/"
	files, e := ioutil.ReadDir(path)
	if err != nil {
		err = e
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if c, err := ioutil.ReadFile(path + file.Name()); err == nil {
			c = []byte(strings.Replace(string(c), "\n", "", -1))

			template := MessageTemplate{}

			err := json.Unmarshal(c, &template)
			if err != nil {
				//ui.Logger.Println("Error in template ", file.Name())
				continue
			}
			for _, k := range template.ReplyKeyboard {
				for _, r := range k {
					for _, b := range r {
						ret[b.Text] = b.Command + " " + b.Args
					}
				}
			}
		}
	}
	return
}
