package botmeans

import (

	// "sync"
	"fmt"
	"testing"
)

func TestActionExecute(t *testing.T) {
	s := ""

	interruptSuccess := true

	needToError := false
	handlersProvider := func(id string) (ActionHandler, bool) {
		switch id {
		case "cmd1":
			return func(context ActionContextInterface) {
				if needToError {
					context.Error(AbortedContextError{"ffuuuu"})
					interruptSuccess = false
				}
				type SD struct {
					Stage int
				}
				sd := SD{}
				context.Session().GetData(&sd)
				s += fmt.Sprintf("handler%v", sd.Stage)
				sd.Stage++
				context.Session().SetData(sd)
				context.Output().Create("", struct{}{})
				if sd.Stage == 3 {
					context.Finish()
					needToError = true
				}
			}, true
		case "":
			return func(ActionContextInterface) {

			}, true
		}

		return nil, false
	}
	out := make(chan Executer)
	session := &Session{SessionBase: SessionBase{42, "fuuu", 24, false, false}}
	sender := &Sender{session: session, msgFactory: func() BotMessageInterface { return &BotMessage{} }}
	go ActionFactory(
		session,
		actionExecuterFactoryConfig{
			func() string { return "cmd1" },
			func() []ArgInterface { return []ArgInterface{Arg{"/cmd1"}, Arg{"ffuuu"}, Arg{9.75}} },
			func() BotMessageInterface { return &BotMessage{} },
		},
		sender,
		out,
		handlersProvider,
	)
	action := (<-out).(*Action)
	// t.Logf("%+v", action)
	action.Args()
	action.Id()
	action.SourceMessage()
	action.Output()

	action.Execute()
	action.Execute()
	if action.LastCommand != "cmd1" {
		t.Fail()
	}

	action.Execute()
	if action.LastCommand != "" {
		t.Fail()
	}
	action.Execute()

	if interruptSuccess == false {
		t.Fail()
	}
	if s != "handler0handler1handler2" {
		t.Log(s, "should be", "handler0handler1handler2")
		t.Fail()
	}
	// if len(sender.outputMessages) != 3 {
	// 	t.Log(len(sender.outputMessages), "should be", 3)
	// 	t.Fail()
	// }
}
