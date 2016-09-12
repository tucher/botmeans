package botmeans

// import (
// 	"encoding/json"
// 	"github.com/go-telegram-bot-api/telegram-bot-api"
// 	"reflect"
// 	"strings"
// )

// //TelegramUserSession represents the user in chat.
// type TelegramUserSession struct {
// 	ID             int64 `sql:"index;unique"`
// 	TelegramUserID int64 `sql:"index"`
// 	TelegramChatID int64 `sql:"index"`
// 	LastCommand    string
// 	//Used internally for determining which handler should be executed next time
// 	HandlerIndex     int
// 	TelegramUserName string `sql:"index"`
// 	//Storage for any user data in JSON format
// 	UserData  string `sql:"type:jsonb"`
// 	FirstName string
// 	LastName  string
// 	ChatName  string
// 	Locale    string
// }

// //SetData sets internal UserData field to JSON representation of given value
// func (session *TelegramUserSession) SetData(value interface{}) {
// 	d, _ := json.Marshal(value)
// 	session.UserData = string(d)
// }

// //GetData extracts internal UserData field to given value
// func (session *TelegramUserSession) GetData(value interface{}) {
// 	json.Unmarshal([]byte(session.UserData), value)
// }

// //GetUserName creates displayable user name
// func (session *TelegramUserSession) GetUserName() (r string) {
// 	r = strings.TrimSpace(session.FirstName + " " + session.LastName)
// 	if r == "" {
// 		r = session.TelegramUserName
// 	}
// 	return
// }

// //TelegramBotMessage represents the message sent by the bot.
// type TelegramBotMessage struct {
// 	ID             int64 `sql:"index;unique"`
// 	TelegramMsgID  int64 `sql:"index"`
// 	TelegramChatID int64 `sql:"index"`
// 	OwnerUserID    int64 `sql:"index"`
// 	//Internal storage for user data
// 	UserData string `sql:"type:jsonb"`
// }

// //SetData sets internal UserData field to JSON representation of given value.
// //Automatically saves information for the value's type, so you don't need to care about it.
// //Just use the same types of value for SetData and GetData
// func (botMessage *TelegramBotMessage) SetData(value interface{}) {
// 	if value == nil {
// 		botMessage.UserData = "{}"
// 		return
// 	}
// 	type botMessageUserDataWrapper struct {
// 		UserData interface{}
// 		Type     string
// 	}
// 	d := botMessageUserDataWrapper{
// 		UserData: value,
// 		Type:     reflect.TypeOf(value).Name(),
// 	}
// 	str, _ := json.Marshal(d)
// 	botMessage.UserData = string(str)
// }

// //GetData extracts internal UserData field to given value
// func (botMessage *TelegramBotMessage) GetData(value interface{}) {
// 	if value == nil {
// 		return
// 	}
// 	// t := reflect.TypeOf(value).Name()
// 	// log.Println(t)
// 	d := struct {
// 		Type     string
// 		UserData json.RawMessage
// 	}{}
// 	json.Unmarshal([]byte(botMessage.UserData), &d)
// 	json.Unmarshal(d.UserData, value)

// }

// //OutputMessage represents a message, which should be sent by the botmeans.
// //It is returned from handlers and can be used to send messages manually with SendMessages
// //If no error occures, several TelegramBotMessage will be created or modified
// //Keyboards, ParseMode and template text will be filled from the template file
// //The template is rendered with UserData as data source
// type OutputMessage struct {
// 	//can be used to edit all messages which user data satisfies this filter
// 	EditSelector map[string]interface{}
// 	//set to message id, if you don't want to use EditSelector
// 	TelegramMsgIDToEdit int64

// 	//the content of this field will be saved in the resulting TelegramBotMessage.
// 	//Also it is used for template rendering
// 	UserData interface{}

// 	//template will be searched in TemplateDir as $TemplateName$.json.
// 	TemplateName string
// }

// type renderedOutputMessage struct {
// 	msg             OutputMessage
// 	ParseMode       string
// 	text            string
// 	replyKbdMarkup  *tgbotapi.ReplyKeyboardMarkup
// 	inlineKbdMarkup *tgbotapi.InlineKeyboardMarkup
// }
// type cmdHandler func(context *CommandContext) ([]OutputMessage, error)

// //CommandContext represents the context, which is passed to user-defined handlers
// type CommandContext struct {
// 	//session of the user, who triggered the event
// 	Session *TelegramUserSession
// 	//all sessions of mentioned users
// 	MentionedSessions []*TelegramUserSession
// 	//message text
// 	Text string
// 	//bot. can be used to call different methods to work with sessions and  messages
// 	ui *MeansBot
// 	//arduments are made be splitting the message text
// 	Args []string
// 	//related message. it is the new message if the user posted a new message or the existing message if user
// 	SourceMessage *TelegramBotMessage
// 	cmd           string
// }

// //Action represents the chain of handlers, which execution begins when user sends command cmd
// type Action struct {
// 	cmd      string
// 	handlers []cmdHandler
// }

// type MessageTemplate struct {
// 	ParseMode     string
// 	Keyboard      map[string][][]MessageButton
// 	Template      map[string]string
// 	ReplyKeyboard map[string][][]MessageButton
// }
