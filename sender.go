package botmeans

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

//OutMsgFactoryInterface allows users to create or edit messages inside ActionHandlers
type OutMsgFactoryInterface interface {
	Create(templateName string, Data interface{})
	Edit(msg BotMessageInterface, templateName string, Data interface{})
	Notify(BotMessageInterface, string, bool)
	SimpleText(text string)
}

//SenderInterface is the abstraction for the Sender
type SenderInterface interface {
	OutMsgFactoryInterface
	// Sendable
}

//Sendable defines something that can be Send
type Sendable interface {
	Send() bool
}

//Sender implements SenderInterface
type Sender struct {
	msgFactory  func() BotMessageInterface
	session     SessionInterface
	bot         *tgbotapi.BotAPI
	templateDir string
}

//Create creates new telegram message from template
func (f *Sender) Create(templateName string, Data interface{}) {
	botMsg := f.msgFactory()
	botMsg.SetData(Data)

	ParseMode, text, inlineKbdMarkup, replyKbdMarkup := renderFromTemplate(f.templateDir, templateName, f.session.Locale(), Data)

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

	botMsg.Save()
}

//SimpleText creates new telegram message with given text
func (f *Sender) SimpleText(text string) {
	botMsg := f.msgFactory()
	toSent := tgbotapi.NewMessage(f.session.ChatId(), text)
	if f.bot != nil {
		if sentMsg, err := f.bot.Send(toSent); err == nil {
			botMsg.SetID(int64(sentMsg.MessageID))
		}
	}
	botMsg.Save()
}

//Notify creates notification for callback queries
func (f *Sender) Notify(msg BotMessageInterface, callbackNotification string, showAlert bool) {
	f.bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: msg.CallbackID(),
		ShowAlert:       showAlert,
		Text:            callbackNotification,
	})
}

//Edit allows to edit existing messages
func (f *Sender) Edit(msg BotMessageInterface, templateName string, Data interface{}) {
	msg.SetData(Data)
	ParseMode, text, inlineKbdMarkup, _ := renderFromTemplate(f.templateDir, templateName, f.session.Locale(), Data)

	editConfig := tgbotapi.NewEditMessageText(f.session.ChatId(), int(msg.Id()), text)

	if inlineKbdMarkup != nil {
		editConfig.ReplyMarkup = inlineKbdMarkup
	}
	editConfig.ParseMode = ParseMode

	if f.bot != nil {
		f.bot.Send(editConfig)
	}
	msg.Save()
}
