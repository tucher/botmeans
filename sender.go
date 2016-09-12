package botmeans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	// "log"
	"strings"
	"text/template"
)

type SenderInterface interface {
	OutMsgFactoryInterface
	Sendable
}

type OutMsgFactoryInterface interface {
	Create(templateName string, Data interface{})
	Edit(msg BotMessageInterface, templateName string, Data interface{})
	Notify(BotMessageInterface, string, bool)
	SimpleText(text string)
}

type Sendable interface {
	Send() bool
}

type notification struct {
	id        string
	showAlert bool
	text      string
}

type Sender struct {
	msgFactory     func() BotMessageInterface
	session        SessionInterface
	outputMessages []msgMeta
	notifications  []notification
	bot            *tgbotapi.BotAPI
	templ          *template.Template
	templateDir    string
}

type msgMeta struct {
	message BotMessageInterface
	edit    bool
}

func (f *Sender) Create(templateName string, Data interface{}) {
	botMsg := f.msgFactory()
	botMsg.SetData(Data)

	ParseMode, text, inlineKbdMarkup, replyKbdMarkup := renderFromTemplate2(f.templateDir, templateName, f.session.Locale(), Data, f.templ)

	toSent := tgbotapi.NewMessage(f.session.ChatId(), text)
	toSent.ParseMode = ParseMode
	if replyKbdMarkup != nil {
		toSent.ReplyMarkup = *replyKbdMarkup
	}

	if inlineKbdMarkup != nil {
		toSent.ReplyMarkup = *inlineKbdMarkup
	}
	if f.bot != nil {
		if sentMsg, err := f.bot.Send(toSent); err == nil {
			botMsg.SetID(int64(sentMsg.MessageID))
		}
	}

	f.outputMessages = append(f.outputMessages, msgMeta{botMsg, false})
}

func (f *Sender) SimpleText(text string) {
	botMsg := f.msgFactory()
	toSent := tgbotapi.NewMessage(f.session.ChatId(), text)
	if f.bot != nil {
		if sentMsg, err := f.bot.Send(toSent); err == nil {
			botMsg.SetID(int64(sentMsg.MessageID))
		}
	}

	f.outputMessages = append(f.outputMessages, msgMeta{botMsg, false})
}

func (f *Sender) Notify(msg BotMessageInterface, callbackNotification string, showAlert bool) {
	f.notifications = append(f.notifications, notification{msg.CallbackID(), showAlert, callbackNotification})
}

func (f *Sender) Edit(msg BotMessageInterface, templateName string, Data interface{}) {
	msg.SetData(Data)
	ParseMode, text, inlineKbdMarkup, _ := renderFromTemplate2(f.templateDir, templateName, f.session.Locale(), Data, f.templ)

	editConfig := tgbotapi.NewEditMessageText(f.session.ChatId(), int(msg.Id()), text)

	if inlineKbdMarkup != nil {
		editConfig.ReplyMarkup = inlineKbdMarkup
	}
	editConfig.ParseMode = ParseMode

	if f.bot != nil {
		f.bot.Send(editConfig)
	}

	f.outputMessages = append(f.outputMessages, msgMeta{msg, true})
}

func (f *Sender) Send() bool {
	for _, msgMeta := range f.outputMessages {
		msgMeta.message.Save()
	}

	for _, n := range f.notifications {
		f.bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
			CallbackQueryID: n.id,
			ShowAlert:       n.showAlert,
			Text:            n.text,
		})
	}
	return true
}

func renderFromTemplate2(
	templateDir string,
	templateName string,
	locale string,
	Data interface{},
	templ *template.Template,
) (string, string, *tgbotapi.InlineKeyboardMarkup, *tgbotapi.ReplyKeyboardMarkup) {
	msgTemplate, err := readMsgTemplate(templateDir + "/" + templateName + ".json")
	if err != nil {
		return "", "", nil, nil
	}
	ParseMode := msgTemplate.ParseMode
	if _, ok := msgTemplate.Template[locale]; !ok {
		locale = ""
	}
	text, _ := renderText(msgTemplate.Template[locale], Data, templ)
	inlineKbdMarkup := createInlineKeyboard(msgTemplate.Keyboard[locale])
	replyKbdMarkup := createReplyKeyboard(msgTemplate.ReplyKeyboard[locale])
	return ParseMode, text, inlineKbdMarkup, replyKbdMarkup
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
