package botmeans

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

//OutMsgFactoryInterface allows users to create or edit messages inside ActionHandlers
type OutMsgFactoryInterface interface {
	Create(templateName string, Data interface{}) error
	CreateWithCustomReplyKeyboard(templateName string, Data interface{}, kdb [][]MessageButton) error
	Edit(msg BotMessageInterface, templateName string, Data interface{}) error
	Notify(BotMessageInterface, string, bool)
	SimpleText(text string) error
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

type Localizer interface {
	Locale() string
}

type senderFactory func(s senderSession) SenderInterface

type senderSession interface {
	ChatIdentifier
	Localizer
}

//Sender implements SenderInterface
type Sender struct {
	msgFactory  func() BotMessageInterface
	session     senderSession
	bot         *tgbotapi.BotAPI
	templateDir string
}

//Create creates new telegram message from template
func (f *Sender) Create(templateName string, Data interface{}) error {
	botMsg := f.msgFactory()
	botMsg.SetData(Data)

	params, err := renderFromTemplate(f.templateDir, templateName, f.session.Locale(), Data)
	if err != nil {
		return err
	}
	toSent := tgbotapi.NewMessage(f.session.ChatId(), params.text)
	toSent.ParseMode = params.ParseMode
	if params.replyKbdMarkup != nil {
		toSent.ReplyMarkup = *params.replyKbdMarkup
	}
	if params.replyKbdRemove != nil {
		toSent.ReplyMarkup = params.replyKbdRemove
	}

	if params.inlineKbdMarkup != nil {
		toSent.ReplyMarkup = *params.inlineKbdMarkup
	}
	if f.bot != nil {
		if sentMsg, err := f.bot.Send(toSent); err == nil {
			botMsg.SetID(int64(sentMsg.MessageID))
		} else {
			return err
		}
	}

	botMsg.Save()
	return nil
}

//Create creates new telegram message from template using custom reply keyboard
func (f *Sender) CreateWithCustomReplyKeyboard(templateName string, Data interface{}, kbd [][]MessageButton) error {
	botMsg := f.msgFactory()
	botMsg.SetData(Data)

	params, err := renderFromTemplate(f.templateDir, templateName, f.session.Locale(), Data)
	if err != nil {
		return err
	}

	toSent := tgbotapi.NewMessage(f.session.ChatId(), params.text)
	toSent.ParseMode = params.ParseMode

	if params.replyKbdMarkup != nil {
		params.replyKbdMarkup.Keyboard = append(params.replyKbdMarkup.Keyboard, createReplyKeyboard(kbd).Keyboard...)
		toSent.ReplyMarkup = *params.replyKbdMarkup
	} else {
		toSent.ReplyMarkup = createReplyKeyboard(kbd)
	}

	if params.inlineKbdMarkup != nil {
		toSent.ReplyMarkup = *params.inlineKbdMarkup
	}
	if f.bot != nil {
		if sentMsg, err := f.bot.Send(toSent); err == nil {
			botMsg.SetID(int64(sentMsg.MessageID))
		} else {
			return nil
		}
	}

	botMsg.Save()
	return nil
}

//SimpleText creates new telegram message with given text
func (f *Sender) SimpleText(text string) error {
	botMsg := f.msgFactory()
	toSent := tgbotapi.NewMessage(f.session.ChatId(), text)
	if f.bot != nil {
		if sentMsg, err := f.bot.Send(toSent); err == nil {
			botMsg.SetID(int64(sentMsg.MessageID))
		} else {
			return err
		}
	}
	botMsg.Save()
	return nil
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
func (f *Sender) Edit(msg BotMessageInterface, templateName string, Data interface{}) error {
	msg.SetData(Data)
	params, err := renderFromTemplate(f.templateDir, templateName, f.session.Locale(), Data)
	if err != nil {
		return err
	}
	editConfig := tgbotapi.NewEditMessageText(f.session.ChatId(), int(msg.Id()), params.text)

	if params.inlineKbdMarkup != nil {
		editConfig.ReplyMarkup = params.inlineKbdMarkup
	}
	editConfig.ParseMode = params.ParseMode

	if f.bot != nil {
		f.bot.Send(editConfig)
	}
	msg.Save()
	return nil
}
