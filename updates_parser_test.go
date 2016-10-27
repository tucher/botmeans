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
		s interface{},
		getters actionExecuterFactoryConfig,
		out chan Executer,
	) {
		if session, ok := s.(SessionInterface); ok {
			ActionFactory(
				session,
				getters,
				func(senderSession) SenderInterface {
					return &Sender{session: session, msgFactory: func() BotMessageInterface { return &BotMessage{} }}
				},
				out,
				handlersProvider,
			)
		}
	}

	mutex := sync.Mutex{}
	sessionIdFlags := make(map[string]struct{})

	sessionFactory := func(base SessionBase) (interface{}, error) {
		stringID := fmt.Sprintf("%v:%v:%v", base.TelegramChatID, base.TelegramUserID, base.TelegramUserName)
		mutex.Lock()
		_, ok := sessionIdFlags[stringID]
		sessionIdFlags[stringID] = struct{}{}
		mutex.Unlock()
		return &Session{SessionBase: base, db: nil, isNew: !ok}, nil
	}

	aliaser := func(string) (string, Args, bool) { return "", args{}, false }

	argsParser := func(tgUpdate tgbotapi.Update) (r Args) {
		r = ArgsParser(tgUpdate, sessionFactory, aliaser)
		// t.Logf("%+v", r)
		return
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
		parserConfig{
			sessionFactory,
			actionFactory,
			botMsgFactory,
			cmdParser,
			argsParser,
		},
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
					session: &Session{SessionBase: SessionBase{42, "fuuu", 24, false, false}, isNew: true},
					getters: actionExecuterFactoryConfig{
						cmdGetter:  func() string { return "" },
						argsGetter: func() Args { return args{[]arg{arg{}}, "session"} },
					},
				},
				&Action{
					session: &Session{SessionBase: SessionBase{42, "fuuu", 24, false, false}, isNew: true},
					getters: actionExecuterFactoryConfig{
						cmdGetter:  func() string { return "" },
						argsGetter: func() Args { return args{[]arg{arg{"blabla"}}, "blabla"} },
					},
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
					session: &Session{SessionBase: SessionBase{42, "fuuu", 24, false, false}},
					getters: actionExecuterFactoryConfig{
						cmdGetter:  func() string { return "cmd1" },
						argsGetter: func() Args { return args{[]arg{arg{"/cmd1"}}, "/cmd1"} },
					},
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
					session: &Session{SessionBase: SessionBase{42, "fuuu", 24, false, false}},
					getters: actionExecuterFactoryConfig{
						cmdGetter:  func() string { return "cmd1" },
						argsGetter: func() Args { return args{[]arg{arg{"/cmd1"}, arg{"ffuuu"}, arg{"9.75"}}, "/cmd1 ffuuu 9.75"} },
					},
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
					if action.getters.cmdGetter() != testEntry.result[lastIndex].getters.cmdGetter() {
						fail = true
						t.Log("Wrong cmd", action.getters.cmdGetter(), testEntry.result[lastIndex].getters.cmdGetter())
					}
					if action.getters.argsGetter().Count() != testEntry.result[lastIndex].getters.argsGetter().Count() {
						fail = true
						t.Log("Wrong args len", action.getters.argsGetter().Raw(), testEntry.result[lastIndex].getters.argsGetter().Raw())
					} else {
						for i := range make([]int, action.getters.argsGetter().Count()) {
							_, ok := action.getters.argsGetter().At(i).NewSession()
							a1, _ := action.getters.argsGetter().At(i).String()
							a2, _ := testEntry.result[lastIndex].getters.argsGetter().At(i).String()
							if a1 != a2 && ok != true {
								fail = true
								t.Log("Wrong arg: ", a1, a2)
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
