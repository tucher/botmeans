package botmeans

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"sync"
	"testing"
)

func TestUpdatesParser(t *testing.T) {

	handlersProvider := func(id string) (ActionHandler, bool) {
		switch id {
		case "cmd1":
			return func(ActionContextInterface) {}, true
		case "":
			return func(ActionContextInterface) {}, true
		}

		return nil, false
	}
	actionFactory := func(
		session SessionInterface,
		cmdGetter func() string,
		argsGetter func() []ArgInterface,
		sourceMessageGetter func() BotMessageInterface,
		out chan Executer,
	) {
		ActionFactory(
			session,
			cmdGetter,
			argsGetter,
			sourceMessageGetter,
			&Sender{session: session, msgFactory: func() BotMessageInterface { return &BotMessage{} }},
			out,
			handlersProvider,
		)
	}

	mutex := sync.Mutex{}
	sessionIdFlags := make(map[string]struct{})

	sessionFactory := func(base SessionBase) (SessionInterface, error) {
		stringID := fmt.Sprintf("%v:%v:%v", base.TelegramChatID, base.TelegramUserID, base.TelegramUserName)
		mutex.Lock()
		_, ok := sessionIdFlags[stringID]
		sessionIdFlags[stringID] = struct{}{}
		mutex.Unlock()
		base.isNew = !ok
		return &Session{SessionBase: base, db: nil}, nil
	}

	aliaser := func(string) (string, []ArgInterface, bool) { return "", []ArgInterface{}, false }

	argsParser := func(tgUpdate tgbotapi.Update) []ArgInterface {
		return ArgsParser(tgUpdate, sessionFactory, aliaser)
	}

	cmdParser := func(tgUpdate tgbotapi.Update) string {
		return CmdParser(tgUpdate, aliaser)
	}

	botMsgFactory := func(chatID int64, msgId int64, callbackID string) BotMessageInterface {
		return &BotMessage{TelegramChatID: chatID, TelegramMsgID: msgId, callbackID: callbackID}
	}

	updatesChan := make(chan tgbotapi.Update)

	actionsChan := createTGUpdatesParser(
		updatesChan,
		sessionFactory,
		actionFactory,
		botMsgFactory,
		cmdParser,
		argsParser,
	)
	type TestEntry struct {
		tgUpdate tgbotapi.Update
		result   []*Action
	}
	testData := []TestEntry{
		TestEntry{
			tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{ID: 42, UserName: "fuuu"},
					Chat: &tgbotapi.Chat{ID: 24, Title: "Chat2"},
					Text: "blabla",
				},
			},
			[]*Action{
				&Action{
					session:    &Session{SessionBase: SessionBase{42, "fuuu", 24, true, false}},
					cmdGetter:  func() string { return "" },
					argsGetter: func() []ArgInterface { return []ArgInterface{Arg{}, Arg{"blabla"}} },
				},
				&Action{
					session:    &Session{SessionBase: SessionBase{42, "fuuu", 24, true, false}},
					cmdGetter:  func() string { return "" },
					argsGetter: func() []ArgInterface { return []ArgInterface{Arg{"blabla"}} },
				},
			},
		},
		TestEntry{
			tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{ID: 42, UserName: "fuuu"},
					Chat: &tgbotapi.Chat{ID: 24, Title: "Chat2"},
					Text: "/cmd1",
				},
			},
			[]*Action{
				&Action{
					session:    &Session{SessionBase: SessionBase{42, "fuuu", 24, false, false}},
					cmdGetter:  func() string { return "cmd1" },
					argsGetter: func() []ArgInterface { return []ArgInterface{Arg{"/cmd1"}} },
				},
			},
		},
		TestEntry{
			tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{ID: 42, UserName: "fuuu"},
					Chat: &tgbotapi.Chat{ID: 24, Title: "Chat2"},
					Text: "/cmd1 ffuuu 9.75",
				},
			},
			[]*Action{
				&Action{
					session:    &Session{SessionBase: SessionBase{42, "fuuu", 24, false, false}},
					cmdGetter:  func() string { return "cmd1" },
					argsGetter: func() []ArgInterface { return []ArgInterface{Arg{"/cmd1"}, Arg{"ffuuu"}, Arg{9.75}} },
				},
			},
		},
	}
	fail := false

	for _, testEntry := range testData {

		updatesChan <- testEntry.tgUpdate
		lastIndex := 0

		for a := range actionsChan {

			if lastIndex == len(testEntry.result) {
				fail = true
				t.Log("Too many actions spawned")
				for _ = range actionsChan {
				}
				break
			}
			switch {
			case a == nil && testEntry.result[lastIndex] != nil:
				t.Log("Should not be nil")
				fail = true
			case a != nil && testEntry.result[lastIndex] == nil:
				t.Log("Should be nil")
				fail = true
			case a != nil && testEntry.result[lastIndex] != nil:
				if action, ok := a.(*Action); ok {
					if *(action.session.(*Session)) != *testEntry.result[lastIndex].session.(*Session) {
						fail = true
						t.Log("Session content is wrong: ", *(action.session.(*Session)), *testEntry.result[lastIndex].session.(*Session))
					}
					if action.cmdGetter() != testEntry.result[lastIndex].cmdGetter() {
						fail = true
						t.Log("Wrong cmd", action.cmdGetter(), testEntry.result[lastIndex].cmdGetter())
					}
					if len(action.argsGetter()) != len(testEntry.result[lastIndex].argsGetter()) {
						fail = true
						t.Log("Wrong args len", action.argsGetter(), testEntry.result[lastIndex].argsGetter())
					} else {
						for i := range action.argsGetter() {
							if action.argsGetter()[i] != testEntry.result[lastIndex].argsGetter()[i] && action.argsGetter()[i].NewSession() != true {
								fail = true
								t.Log("Wrong arg: ", action.argsGetter()[i], testEntry.result[lastIndex].argsGetter()[i])
								break
							}
						}
					}
				} else {
					fail = true
					t.Log("Should be Action type")
				}
			}
			lastIndex++
			if lastIndex == len(testEntry.result) {
				break
			}
		}
		if lastIndex != len(testEntry.result) {
			fail = true
			t.Log("Too less actions spawned")
		}
	}
	close(updatesChan)

	if fail {
		t.Fail()
	}
}
