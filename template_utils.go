package botmeans

import (
	"strings"
	"text/template"
)

func getTemplater() *template.Template {
	return template.New("msg").Funcs(
		template.FuncMap{"underscore": func(s string) string { return strings.Replace(strings.TrimSpace(s), " ", "_", -1) }},
	)
}

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
