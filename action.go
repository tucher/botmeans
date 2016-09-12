package botmeans

type ActionHandler func(context ActionContextInterface)

type ActionHandlersProvider func(id string) (ActionHandler, bool)

func ActionFactory(
	session SessionInterface,
	cmdGetter func() string,
	argsGetter func() []ArgInterface,
	sourceMessageGetter func() BotMessageInterface,
	sender SenderInterface,
	out chan Executer,
	handlersProvider ActionHandlersProvider,
) {
	if session.IsNew() {
		out <- &Action2{
			session:             session,
			handlersProvider:    handlersProvider,
			sourceMessageGetter: func() (r BotMessageInterface) { return },
			argsGetter:          func() []ArgInterface { return append([]ArgInterface{Arg{session}}, argsGetter()...) },
			cmdGetter:           func() string { return "" },
			sender:              sender,
		}
	}
	if _, ok := handlersProvider(cmdGetter()); ok == true {

		ret := &Action2{
			session:             session,
			handlersProvider:    handlersProvider,
			sourceMessageGetter: sourceMessageGetter,
			argsGetter:          argsGetter,
			cmdGetter:           cmdGetter,
			sender:              sender,
		}
		session.GetData(ret)
		out <- ret
	} else {
		out <- nil
	}

}

type Action2 struct {
	session     SessionInterface
	LastCommand string

	cmdGetter           func() string
	argsGetter          func() []ArgInterface
	sourceMessageGetter func() BotMessageInterface
	handlersProvider    ActionHandlersProvider
	sender              SenderInterface
	err                 interface{}
}

func (a *Action2) Execute() {
	defer func() {
		r := recover()
		if _, ok := r.(AbortedContextError); !ok {
			if r != nil {
				panic(r)
			}
		} else {

		}
	}()
	ok := false
	cmd := a.cmdGetter()
	if _, ok = a.handlersProvider(cmd); ok == true && cmd != "" {
		a.LastCommand = cmd
	} else if _, ok = a.handlersProvider(a.LastCommand); ok == true {

	}
	if !ok {
		return
	}
	handler, _ := a.handlersProvider(a.LastCommand)
	handler(a)
	a.session.SetData(*a)
	a.session.Save()

	a.sender.Send()
}

func (a *Action2) Id() int64 {
	return a.session.ChatId()
}

func (a *Action2) Args() []ArgInterface {
	return a.argsGetter()
}

func (a *Action2) Error(e interface{}) {
	a.err = e
	a.LastCommand = ""
	panic(AbortedContextError{e})
}

func (a *Action2) Session() SessionInterface {
	return a.session
}

func (a *Action2) SourceMessage() BotMessageInterface {
	return a.sourceMessageGetter()
}

func (a *Action2) Output() OutMsgFactoryInterface {
	return a.sender
}

func (a *Action2) Finish() {
	a.LastCommand = ""
}

type ActionContextInterface interface {
	Args() []ArgInterface
	Output() OutMsgFactoryInterface
	Error(interface{})
	Session() SessionInterface
	SourceMessage() BotMessageInterface
	Finish()
}

type AbortedContextError struct {
	content interface{}
}

type DataGetSetter interface {
	SetData(value interface{})
	GetData(value interface{})
}

type PersistentSaver interface {
	Save() error
}
