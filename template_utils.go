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
	"unicode"
)

func getTemplater() *template.Template {
	return template.New("msg").Funcs(
		template.FuncMap{
			"underscore": func(s string) string {
				return strings.Map(
					func(r rune) rune {
						if unicode.IsLetter(r) || unicode.IsNumber(r) {
							return r
						}
						return '_'
					},
					strings.TrimSpace(s),
				)
			},
		},
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

type tgMsgParams struct {
	ParseMode       string
	text            string
	inlineKbdMarkup *tgbotapi.InlineKeyboardMarkup
	replyKbdMarkup  *tgbotapi.ReplyKeyboardMarkup
	replyKbdHide    *tgbotapi.ReplyKeyboardHide
}

func renderFromTemplate(
	templateDir string,
	templateName string,
	locale string,
	Data interface{},
) tgMsgParams {
	ret := tgMsgParams{}
	templ := getTemplater()
	msgTemplate, err := readMsgTemplate(templateDir + "/" + templateName + ".json")
	if err != nil {
		return ret
	}
	ret.ParseMode = msgTemplate.ParseMode
	if _, ok := msgTemplate.Template[locale]; !ok {
		locale = ""
	}
	ret.text, err = renderText(msgTemplate.Template[locale], Data, templ)
	ret.inlineKbdMarkup = createInlineKeyboard(msgTemplate.Keyboard[locale])
	ret.replyKbdMarkup = createReplyKeyboard(msgTemplate.ReplyKeyboard[locale])
	if len(msgTemplate.ReplyKeyboard[locale]) == 0 {
		h := tgbotapi.NewHideKeyboard(true)
		ret.replyKbdHide = &h
	}
	return ret
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
		inlineKbdMarkup.InlineKeyboard = make([][]tgbotapi.InlineKeyboardButton, 0)
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
		replyKbdMarkup.Keyboard = make([][]tgbotapi.KeyboardButton, 0)
	}
	return &replyKbdMarkup
}

func renderText(templateText string, data interface{}, templ *template.Template) (text string, err error) {
	// log.Printf("\n%+v", data)
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
