package botmeans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"strings"
	"text/template"
)

func getTemplater() *template.Template {
	return template.New("msg").Funcs(
		template.FuncMap{"underscore": func(s string) string { return strings.Replace(strings.TrimSpace(s), " ", "_", -1) }},
	)
}

//MessageTemplate defines the structure of the message template
type MessageTemplate struct {
	ParseMode     string
	Keyboard      map[string][][]MessageButton
	Template      map[string]string
	ReplyKeyboard map[string][][]MessageButton
}

//MessageButton represents a button in Telegram UI
type MessageButton struct {
	Text    string
	Command string
	Args    string
}

func renderFromTemplate(
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
