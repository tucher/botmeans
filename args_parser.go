package botmeans

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
)

//ArgInterface defines the arg type, which is used to pass parsed args through the context
type ArgInterface interface {
	String() (string, bool)
	Float() (float64, bool)
	Mention() (SessionInterface, bool)
	NewSession() bool
	LeftSession() bool
}

//Arg is a ArgInterface implementation
type Arg struct {
	arg interface{}
}

//String treats the arg as string
func (a Arg) String() (string, bool) {
	val, ok := a.arg.(string)
	return val, ok
}

//Float treats the arg as float
func (a Arg) Float() (float64, bool) {
	val, ok := a.arg.(float64)
	return val, ok
}

//Mention treats the arg as SessionInterface
func (a Arg) Mention() (SessionInterface, bool) {
	val, ok := a.arg.(SessionInterface)
	return val, ok
}

//NewSession treats the arg as flag if the session in arg is new
func (a Arg) NewSession() bool {
	val, ok := a.arg.(SessionInterface)
	if ok && val.IsNew() {
		return true
	}
	return false
}

//LeftSession treats the arg as flag if the session in arg is left
func (a Arg) LeftSession() bool {
	val, ok := a.arg.(SessionInterface)
	if ok && val.IsLeft() {
		return true
	}
	return false
}

//CommandAliaser converts any text to cmd and args
type CommandAliaser func(string) (string, []ArgInterface, bool)

//ArgsParser parses arguments from Update
func ArgsParser(tgUpdate tgbotapi.Update, sessionFactory SessionFactory, aliaser CommandAliaser) []ArgInterface {
	text := ""
	type mention struct {
		t string
		s SessionInterface
	}
	mentions := []mention{}

	switch {
	case tgUpdate.Message != nil:
		text = tgUpdate.Message.Text

		if tgUpdate.Message.NewChatMember != nil {
			s, _ := sessionFactory(SessionBase{int64(tgUpdate.Message.NewChatMember.ID), tgUpdate.Message.NewChatMember.UserName, tgUpdate.Message.Chat.ID, true, false})
			return []ArgInterface{Arg{s}}
		}
		if tgUpdate.Message.LeftChatMember != nil {
			s, _ := sessionFactory(SessionBase{int64(tgUpdate.Message.LeftChatMember.ID), tgUpdate.Message.LeftChatMember.UserName, tgUpdate.Message.Chat.ID, false, true})
			return []ArgInterface{Arg{s}}
		}
		if tgUpdate.Message.Entities != nil {
			for _, ent := range *tgUpdate.Message.Entities {
				switch {
				case ent.Type == "text_mention" && ent.User != nil:

					if s, err := sessionFactory(SessionBase{int64(ent.User.ID), ent.User.UserName, tgUpdate.Message.Chat.ID, false, false}); err != nil {

						mentions = append(mentions, mention{string([]rune(text)[ent.Offset : ent.Offset+ent.Length]), s})
					}

				case ent.Type == "mention":
					userName := string([]rune(text)[ent.Offset+1 : ent.Offset+ent.Length])
					if s, err := sessionFactory(SessionBase{0, userName, tgUpdate.Message.Chat.ID, false, false}); err != nil {
						mentions = append(mentions, mention{string([]rune(text)[ent.Offset : ent.Offset+ent.Length]), s})
					}
				}
			}
		}
	case tgUpdate.CallbackQuery != nil:
		text = tgUpdate.CallbackQuery.Data
	}

	if _, args, ok := aliaser(text); ok {
		return args
	}
	retArgs := []ArgInterface{}

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
			retArgs = append(retArgs, Arg{mentions[mfound].s})
		} else if val, ok := strconv.ParseFloat(str, 64); ok == nil {
			retArgs = append(retArgs, Arg{val})
		} else {
			retArgs = append(retArgs, Arg{str})
		}
	}

	return retArgs
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
