package telegram

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//ChatIdentifier defines something that knows which chat it belongs to
type ChatIdentifier interface {
	ChatId() int64
}

//SessionFactory creates the session from given session base
type SessionFactory func(base SessionBase) (SessionInterface, error)

//ActionExecuterFactory creates Executers from given session, cmd, args and source message
type ActionExecuterFactory func(
	SessionBase,
	SessionFactory,
	actionExecuterFactoryConfig,
	chan Executer,
)

//BotMessageFactory loads bot message from given chat id, msg id and callback id
type BotMessageFactory func(int64, int64, string) BotMessageInterface

//CmdParserFunc returns command for the given update
type CmdParserFunc func(tgbotapi.Update) string

//ArgsParserFunc returns args of command for the given update
type ArgsParserFunc func(tgbotapi.Update) Args

type parserConfig struct {
	sessionFactory        SessionFactory
	actionExecuterFactory ActionExecuterFactory
	botMessageFactory     BotMessageFactory
	cmdParser             CmdParserFunc
	argsParser            ArgsParserFunc
}

func createTGUpdatesParser(
	cmdQueueChan chan Executer,
	tgUpdateChan <-chan tgbotapi.Update,
	pC parserConfig,
) {
	wg := sync.WaitGroup{}
	for tgUpdate := range tgUpdateChan {
		wg.Add(1)
		go func() {
			var chatId, userId int64
			var username string
			var msg *tgbotapi.Message
			var msgId int64
			var callbackID string

			switch {
			case tgUpdate.Message != nil:
				chatId = tgUpdate.Message.Chat.ID
				userId = int64(tgUpdate.Message.From.ID)
				username = tgUpdate.Message.From.UserName
				msg = tgUpdate.Message
			case tgUpdate.CallbackQuery != nil:
				chatId = tgUpdate.CallbackQuery.Message.Chat.ID
				userId = int64(tgUpdate.CallbackQuery.From.ID)
				username = tgUpdate.CallbackQuery.From.UserName
				msg = tgUpdate.CallbackQuery.Message
				msgId = int64(msg.MessageID)
				callbackID = tgUpdate.CallbackQuery.ID
			case tgUpdate.EditedMessage != nil:
				chatId = tgUpdate.EditedMessage.Chat.ID
				userId = int64(tgUpdate.EditedMessage.From.ID)
				username = tgUpdate.EditedMessage.From.UserName
				msg = tgUpdate.EditedMessage
			case tgUpdate.InlineQuery != nil:
			case tgUpdate.ChosenInlineResult != nil:

			}

			pC.actionExecuterFactory(
				SessionBase{TelegramUserID: userId, TelegramUserName: username, TelegramChatID: chatId},
				pC.sessionFactory,
				actionExecuterFactoryConfig{
					func() string { return pC.cmdParser(tgUpdate) },
					func() Args { return pC.argsParser(tgUpdate) },
					func() BotMessageInterface { return pC.botMessageFactory(chatId, msgId, callbackID) },
				},
				cmdQueueChan)

			wg.Done()
		}()
	}
	wg.Wait()
}
