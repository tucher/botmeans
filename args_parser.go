package botmeans

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	// "log"
	// "reflect"
	"strconv"
	"strings"
)

//Arg defines the arg type, which is used to pass parsed args through the context
type Arg interface {
	String() (string, bool)
	Float() (float64, bool)
	Mention() (SessionInterface, bool)
	NewSession() (SessionInterface, bool)
	LeftSession() (SessionInterface, bool)
	ComeSession() (SessionInterface, bool)
}

//arg is a Arg implementation
type arg struct {
	arg interface{}
}
type args struct {
	a   []arg
	raw string
}

func (a args) At(index int) Arg {
	if len(a.a) > index {
		return a.a[index]
	}
	return arg{}
}
func (a args) Count() int {
	return len(a.a)
}

func (a args) Raw() string {
	return a.raw
}

type Args interface {
	At(int) Arg
	Count() int
	Raw() string
}

//String treats the arg as string
func (a arg) String() (string, bool) {
	val, ok := a.arg.(string)
	return val, ok
}

//Float treats the arg as float
func (a arg) Float() (float64, bool) {

	if val, ok := a.arg.(string); ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
		return 0, false
	}
	// log.Println(reflect.TypeOf(a.arg))
	return 0, false
}

//Mention treats the arg as SessionInterface
func (a arg) Mention() (SessionInterface, bool) {
	if val, ok := a.arg.(mention); ok {
		s, ok := val.s.(SessionInterface)
		return s, ok
	}
	return nil, false
}

//NewSession treats the arg as flag if the session in arg is new
func (a arg) NewSession() (SessionInterface, bool) {
	val, ok := a.arg.(SessionInterface)
	if ok && val.IsNew() {
		return val, true
	}
	return nil, false
}

//LeftSession treats the arg as flag if the session in arg is left
func (a arg) LeftSession() (SessionInterface, bool) {
	val, ok := a.arg.(SessionInterface)
	if ok && val.HasLeft() {
		return val, true
	}
	return nil, false
}

//LeftSession treats the arg as flag if the session in arg is left
func (a arg) ComeSession() (SessionInterface, bool) {
	val, ok := a.arg.(SessionInterface)
	if ok && val.HasCome() {
		return val, true
	}
	return nil, false
}

//CommandAliaser converts any text to cmd and args
type CommandAliaser func(string) (string, Args, bool)

type mention struct {
	t string
	s interface{}
}

func extractMentions(tgUpdate tgbotapi.Update, sessionFactory SessionFactory) (mentions []mention) {
	if tgUpdate.Message.Entities != nil {
		for _, ent := range *tgUpdate.Message.Entities {
			switch {
			case ent.Type == "text_mention" && ent.User != nil:

				if s, err := sessionFactory(SessionBase{int64(ent.User.ID), ent.User.UserName, tgUpdate.Message.Chat.ID, false, false}); err == nil {

					mentions = append(mentions, mention{string([]rune(tgUpdate.Message.Text)[ent.Offset : ent.Offset+ent.Length]), s})
				}

			case ent.Type == "mention":
				userName := string([]rune(tgUpdate.Message.Text)[ent.Offset+1 : ent.Offset+ent.Length])
				if s, err := sessionFactory(SessionBase{0, userName, tgUpdate.Message.Chat.ID, false, false}); err == nil {
					mentions = append(mentions, mention{string([]rune(tgUpdate.Message.Text)[ent.Offset : ent.Offset+ent.Length]), s})
				}
			}
		}
	}
	return
}

//ArgsParser parses arguments from Update
func ArgsParser(tgUpdate tgbotapi.Update, sessionFactory SessionFactory, aliaser CommandAliaser) Args {
	text := ""

	mentions := []mention{}

	switch {
	case tgUpdate.Message != nil:
		text = tgUpdate.Message.Text

		if tgUpdate.Message.NewChatMember != nil {
			if s, err := sessionFactory(SessionBase{int64(tgUpdate.Message.NewChatMember.ID), tgUpdate.Message.NewChatMember.UserName, tgUpdate.Message.Chat.ID, true, false}); err == nil {
				return args{[]arg{arg{s}}, ""}
			}
		}
		if tgUpdate.Message.LeftChatMember != nil {
			if s, err := sessionFactory(SessionBase{int64(tgUpdate.Message.LeftChatMember.ID), tgUpdate.Message.LeftChatMember.UserName, tgUpdate.Message.Chat.ID, false, true}); err == nil {
				return args{[]arg{arg{s}}, ""}
			}
		}
		mentions = extractMentions(tgUpdate, sessionFactory)
	case tgUpdate.CallbackQuery != nil:
		text = tgUpdate.CallbackQuery.Data
	}

	if _, args, ok := aliaser(text); ok {
		return args
	}
	retArgs := []arg{}

	splitted := []string{}
	for _, a := range strings.Split(text, " ") {
		trimmed := strings.TrimSpace(a)
		if len(trimmed) > 0 {
			splitted = append(splitted, trimmed)
		}
	}

	for _, str := range splitted {
		mfound := -1
		for i, menti := range mentions {
			if menti.t == str {
				mfound = i
				break
			}
		}
		if mfound != -1 {
			retArgs = append(retArgs, arg{mentions[mfound]})
		} else {
			retArgs = append(retArgs, arg{str})
		}
	}

	return args{retArgs, text}
}

//CmdParser parses command from Update
func CmdParser(tgUpdate tgbotapi.Update, aliaser CommandAliaser) string {
	text := ""

	switch {
	case tgUpdate.Message != nil:
		if tgUpdate.Message.NewChatMember != nil || tgUpdate.Message.LeftChatMember != nil {
			return ""
		}
		text = tgUpdate.Message.Text
	case tgUpdate.CallbackQuery != nil:
		text = tgUpdate.CallbackQuery.Data
	}

	if cmd, _, ok := aliaser(text); ok {
		return cmd
	}

	splitted := strings.Split(strings.TrimSpace(text), " ")

	if len(splitted) == 0 || len(splitted[0]) == 0 {
		return ""
	}

	cmd := ""
	if splitted[0][0] == '/' {
		cmd = strings.Split(splitted[0][1:], "@")[0]
	}

	return cmd
}
