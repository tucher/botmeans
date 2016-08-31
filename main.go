// Package botmeans provides a framework for creation of complex Telegram bots with rich behaviour
package botmeans

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

//MeansBot is a body of botmeans framework instance. Can be filled with actions(NewAction)
type MeansBot struct {
	bot        *tgbotapi.BotAPI
	db         *gorm.DB
	Logger     *log.Logger
	telegramID int64
	netConfig  NetConfig
	tlgConfig  TelegramConfig

	actions map[string]*Action

	NewSessionCallback      func(session *TelegramUserSession)
	AfterNewSessionCallback func(session *TelegramUserSession)
	OnUserLeft              func(session *TelegramUserSession)
	templateDir             string

	templateFunctions template.FuncMap

	cmdAliases map[string]string
	stopChan   chan struct{}
}

//NetConfig is a MeansBot network config for using with New function
type NetConfig struct {
	ListenIP   string
	ListenPort int16
}

//TelegramConfig is a MeansBot telegram API config for using with New function
type TelegramConfig struct {
	BotToken    string
	WebhookHost string
	SSLCertFile string
}

//New creates new MeansBot instance
func New(DB *gorm.DB, netConfig NetConfig, tlgConfig TelegramConfig) (*MeansBot, error) {
	if DB == nil {
		return &MeansBot{}, fmt.Errorf("No db connection given")
	}
	bot, err := tgbotapi.NewBotAPI(tlgConfig.BotToken)
	if err != nil {
		return &MeansBot{}, err
	}

	botID, _ := strconv.ParseInt(strings.Split(tlgConfig.BotToken, ":")[0], 10, 64)

	ret := &MeansBot{
		bot:               bot,
		db:                DB,
		Logger:            log.New(os.Stdout, "MeansBot: ", log.Llongfile|log.Ldate|log.Ltime),
		telegramID:        botID,
		netConfig:         netConfig,
		actions:           make(map[string]*Action),
		tlgConfig:         tlgConfig,
		templateFunctions: template.FuncMap{"underscore": func(s string) string { return strings.Replace(strings.TrimSpace(s), " ", "_", -1) }},
		cmdAliases:        make(map[string]string),
		stopChan:          make(chan struct{}),
	}
	ret.NewSessionCallback = func(session *TelegramUserSession) {
		ret.Logger.Printf("New session: %+v", session)
	}
	ret.AfterNewSessionCallback = func(session *TelegramUserSession) {
		ret.Logger.Printf("After New session: %+v", session)
	}
	ret.OnUserLeft = func(session *TelegramUserSession) {
		ret.Logger.Printf("User left: %+v", session)
	}

	ret.init()
	return ret, nil
}

func (ui *MeansBot) init() {
	ui.bot.RemoveWebhook()
	_, err := ui.bot.SetWebhook(tgbotapi.NewWebhookWithCert(fmt.Sprintf("https://%v:8443/%v", ui.tlgConfig.WebhookHost, ui.bot.Token),
		ui.tlgConfig.SSLCertFile))
	if err != nil {
		ui.Logger.Println("Failed to set webhook", err)
	}

	ui.db.AutoMigrate(&TelegramUserSession{})
	ui.db.AutoMigrate(&TelegramBotMessage{})
}
func (ui *MeansBot) SetTemplatesDir(path string) {
	ui.templateDir = path
	if path != "" {
		ui.cmdAliases, _ = parseCmdAliasesFromTemplates(path)
	}
}

//NewAction creates new Action with given command. Can be filled with handlers by Handler function
func (ui *MeansBot) NewAction(cmd string) *Action {
	ui.actions[cmd] = &Action{cmd: cmd}
	return ui.actions[cmd]
}

//Handler appends new handler to this Action
func (command *Action) Handler(h cmdHandler) *Action {
	command.handlers = append(command.handlers, h)
	return command
}

//Stop function stops the bot, which was started by Run
func (ui *MeansBot) Stop() {
	ui.stopChan <- struct{}{}
}

//Run starts updates handling
func (ui *MeansBot) Run() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := ui.bot.ListenForWebhook("/" + ui.bot.Token)
	go http.ListenAndServe(fmt.Sprintf("%v:%v", ui.netConfig.ListenIP, ui.netConfig.ListenPort), nil)

	updatesChanMap := make(map[int64]chan tgbotapi.Update)
	handlerClosedChan := make(chan int64)

	for {
		select {
		case update := <-updates:
			chatID := int64(0)

			switch {
			case update.Message != nil:
				chatID = update.Message.Chat.ID
			case update.CallbackQuery != nil:
				chatID = update.CallbackQuery.Message.Chat.ID
			}
			// ui.Logger.Println("Update #", update.UpdateID)
			var chatChan chan tgbotapi.Update
			ok := false
			if chatChan, ok = updatesChanMap[chatID]; !ok {
				chatChan = make(chan tgbotapi.Update)
				updatesChanMap[chatID] = chatChan
				go func(ch chan tgbotapi.Update, chatID int64) {
					defer func() {
						if r := recover(); r != nil {
							ui.Logger.Println("Recovered ", r)
							handlerClosedChan <- chatID
						}
					}()
					// t := time.Now()
					// defer func() { ui.Logger.Println("Finished handler gouroutine for chat", time.Since(t)) }()
					exitSignaller := time.After(10 * time.Second)
					for {
						select {
						case update := <-ch:
							// ui.Logger.Println("Processing update #", update.UpdateID)
							// t := time.Now()
							ui.handleUpdate(&update)
							// ui.Logger.Println("Finished HandleUpdate in ", time.Since(t))
							exitSignaller = time.After(10 * time.Second)
						case <-exitSignaller:
							handlerClosedChan <- chatID
							return
						}
					}
				}(chatChan, chatID)
			}
			go func() { chatChan <- update }()
		case id := <-handlerClosedChan:
			delete(updatesChanMap, id)
		case <-ui.stopChan:
			return nil
		}
	}
}

//HandlerError can be used to return errors, which should be displayed in Telegram client popup
type HandlerError interface {
	Error() string
	Feedback() string //Message to display in popup
}

func (ui *MeansBot) parseMentions(msg *tgbotapi.Message) (mentionedSessions []*TelegramUserSession) {
	if msg.Entities == nil {
		return
	}
	for _, ent := range *msg.Entities {
		switch ent.Type {
		case "text_mention":
			if ent.User != nil {
				if abonentSession, err := ui.findOrCreateSession(msg.Chat.ID,
					int64(ent.User.ID),
					ent.User.UserName); err == nil {
					mentionedSessions = append(mentionedSessions, abonentSession)
				} else {
					ui.Logger.Println(err)
				}
			}
		case "mention":
			userName := msg.Text[ent.Offset+1 : ent.Offset+ent.Length]
			if abonentSession, err := ui.findOrCreateSession(msg.Chat.ID,
				0,
				userName); err == nil {
				mentionedSessions = append(mentionedSessions, abonentSession)
			} else {
				ui.Logger.Println(err)
			}

		}
	}
	return
}

func parseCmdAndArgs(text string) (cmd string, args []string) {
	if len(text) == 0 {
		return
	}
	splitted := []string{}
	for _, a := range strings.Split(text, " ") {
		trimmed := strings.TrimSpace(a)
		if len(trimmed) > 0 {
			splitted = append(splitted, trimmed)
		}
	}

	if splitted[0][0] == '/' {
		cmd = strings.Split(splitted[0][1:], "@")[0]
		args = splitted[1:]
	} else {
		args = splitted
	}
	return
}

func (ui *MeansBot) contextFromUpdate(upd *tgbotapi.Update) *CommandContext {
	cmd := ""
	cmdArgs := []string{}
	from := &tgbotapi.User{}
	msg := &tgbotapi.Message{}
	err := fmt.Errorf("error")
	var sourceMessage *TelegramBotMessage
	switch {
	case upd.Message != nil:
		from = upd.Message.From
		msg = upd.Message
		err = nil
		if cmdFromAlias, ok := ui.cmdAliases[upd.Message.Text]; ok {
			cmd, cmdArgs = parseCmdAndArgs(cmdFromAlias)
		} else {
			cmd, cmdArgs = parseCmdAndArgs(upd.Message.Text)
		}
	case upd.InlineQuery != nil, upd.ChosenInlineResult != nil, upd.EditedMessage != nil:
		err = fmt.Errorf("Unsupported update type")
	case upd.CallbackQuery != nil:
		from = upd.CallbackQuery.From
		msg = upd.CallbackQuery.Message
		cmd, cmdArgs = parseCmdAndArgs(upd.CallbackQuery.Data)
		srcMsg := TelegramBotMessage{TelegramMsgID: int64(msg.MessageID), TelegramChatID: int64(msg.Chat.ID)}
		ui.db.Where(srcMsg).Find(&srcMsg)
		sourceMessage = &srcMsg
		err = nil
	}
	mentionedSessions := ui.parseMentions(msg)
	if err != nil {
		ui.Logger.Printf("%v, %+v", err, upd)
		return nil
	}
	session, err := ui.findOrCreateSession(msg.Chat.ID, int64(from.ID), from.UserName)
	if err != nil {
		ui.Logger.Println("cannot find session:", msg.Chat.ID, int64(from.ID), from.UserName, err)
		return nil
	}

	if upd.Message != nil && upd.Message.LeftChatMember != nil {
		leftMember := upd.Message.LeftChatMember
		if int64(leftMember.ID) == ui.telegramID {
			ui.Logger.Printf("I've left the group %v!", msg.Chat.ID)
		} else {
			ui.OnUserLeft(session)
			//TODO remove consumptions, products and transfers
		}
	}
	if upd.Message != nil && upd.Message.NewChatMember != nil {
		newMember := upd.Message.NewChatMember
		if int64(newMember.ID) == ui.telegramID {
			ui.Logger.Printf("I've joined the group %v!", upd.Message.Chat.ID)
		} else {
			// ui.findOrCreateSession(msg.Chat.ID, int64(newMember.ID), newMember.UserName)
		}
	}

	context := &CommandContext{
		Session:           session,
		Text:              msg.Text,
		MentionedSessions: mentionedSessions,
		Args:              cmdArgs,
		ui:                ui,
		SourceMessage:     sourceMessage,
		cmd:               cmd,
	}
	return context
}

func (ui *MeansBot) handleNewAction(a *Action, context *CommandContext) (messagesToSent []OutputMessage, handlerError error) {
	if context.Session.LastCommand != context.cmd {
		context.Session.HandlerIndex = 0
	}
	if len(a.handlers) == 0 {
		handlerError = fmt.Errorf("No handlers for action: ", context.cmd)
	} else {
		context.Session.LastCommand = context.cmd
		if messagesToSent, handlerError = a.handlers[context.Session.HandlerIndex](context); handlerError == nil {
			context.Session.HandlerIndex += 1
			if context.Session.HandlerIndex >= len(a.handlers) {
				context.Session.HandlerIndex = 0
				context.Session.LastCommand = ""
			}
		}
	}
	return
}

func (ui *MeansBot) handleActionContinue(a *Action, context *CommandContext) (messagesToSent []OutputMessage, handlerError error) {
	if len(a.handlers) == 0 {
		handlerError = fmt.Errorf("No handlers for action: ", context.cmd)
	} else if messagesToSent, handlerError = a.handlers[context.Session.HandlerIndex](context); handlerError == nil {
		context.Session.HandlerIndex += 1
		if context.Session.HandlerIndex >= len(a.handlers) {
			context.Session.HandlerIndex = 0
			context.Session.LastCommand = ""
		}
	}

	return
}
func (ui *MeansBot) handleUpdate(upd *tgbotapi.Update) {
	context := ui.contextFromUpdate(upd)
	if context == nil {
		return
	}
	messagesToSent := []OutputMessage{}
	var handlerError interface{}
	if a, ok := ui.actions[context.cmd]; ok == true {
		messagesToSent, handlerError = ui.handleNewAction(a, context)
	} else if a, ok := ui.actions[context.Session.LastCommand]; ok == true {
		messagesToSent, handlerError = ui.handleActionContinue(a, context)
	} else {
		//ui.Logger.Printf("Cannot find command handler for last command %+v for session %+v", context.Session.LastCommand, context.Session)
	}

	ui.db.Save(&context.Session)
	if upd.CallbackQuery != nil {
		if feedBackErr, ok := handlerError.(HandlerError); ok == true {
			ui.bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
				CallbackQueryID: upd.CallbackQuery.ID,
				ShowAlert:       true,
				Text:            feedBackErr.Feedback(),
			})
		} else {
			ui.bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
				CallbackQueryID: upd.CallbackQuery.ID,
				ShowAlert:       false,
				Text:            "",
			})
		}
	}
	ui.SendMessages(context.Session, messagesToSent)
}

//SendMessages sends messages from the given session
func (ui *MeansBot) SendMessages(session *TelegramUserSession, messagesToSent []OutputMessage) {
	for _, msg := range messagesToSent {
		renderedMsg := renderedOutputMessage{msg: msg}
		if err := renderedMsg.renderFromTemplate(ui.templateDir, session.Locale, template.New("msg").Funcs(ui.templateFunctions)); err != nil {
			ui.Logger.Println(err)
			continue
		}

		// t := time.Now()
		if msg.TelegramMsgIDToEdit == 0 && msg.EditSelector == nil {
			if err := ui.sendMessage(session, &renderedMsg); err != nil {
				ui.Logger.Printf("Error sending message, %v, %+v", err, msg)
			}
			// ui.Logger.Println("New message sent in ", time.Since(t))
		} else {
			toEdit := []renderedOutputMessage{}
			if msg.EditSelector != nil {
				msgs := ui.FindBotMessagesByChatAndUserData(*session, msg.UserData, msg.EditSelector)
				for _, m := range msgs {
					o := renderedMsg
					o.msg.TelegramMsgIDToEdit = m.TelegramMsgID
					toEdit = append(toEdit, o)
				}
			} else {
				toEdit = append(toEdit, renderedMsg)
			}
			for _, m := range toEdit {
				if err := ui.editMessage(session, &m); err != nil {
					ui.Logger.Println(err)
				}
				// ui.Logger.Println("Message edited in ", time.Since(t))
			}
		}
		// ui.Logger.Println("Messages sent in ", time.Since(t))
	}
}

type handlerErrorImpl struct {
	msg string
}

func (i handlerErrorImpl) Feedback() string { return i.msg }
func (i handlerErrorImpl) Error() string    { return i.msg }

//NewFeedbackError creates new error
func NewFeedbackError(msg string) HandlerError {
	return handlerErrorImpl{msg}
}
