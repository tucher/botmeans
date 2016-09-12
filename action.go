package botmeans

//ActionHandler defines the type of handler function
type ActionHandler func(context ActionContextInterface)

//ActionHandlersProvider returns ActionHandler for given command
type ActionHandlersProvider func(id string) (ActionHandler, bool)

//ActionFactory generates Executers
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
		out <- &Action{
			session:             session,
			handlersProvider:    handlersProvider,
			sourceMessageGetter: func() (r BotMessageInterface) { return },
			argsGetter:          func() []ArgInterface { return append([]ArgInterface{Arg{session}}, argsGetter()...) },
			cmdGetter:           func() string { return "" },
			sender:              sender,
		}
	}
	if _, ok := handlersProvider(cmdGetter()); ok == true {

		ret := &Action{
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

//Action provides the context for the user command
type Action struct {
	session     SessionInterface
	LastCommand string

	cmdGetter           func() string
	argsGetter          func() []ArgInterface
	sourceMessageGetter func() BotMessageInterface
	handlersProvider    ActionHandlersProvider
	sender              SenderInterface
	err                 interface{}
}

//Execute implements Execute for BotMachine
func (a *Action) Execute() {
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

	// a.sender.Send()
}

//Id returns id based on chat id
func (a *Action) Id() int64 {
	return a.session.ChatId()
}

//Args allow user to access command  args inside ActionHandler through the Context()
func (a *Action) Args() []ArgInterface {
	return a.argsGetter()
}

//Error allow user to terminate ActionHandler through the Context()
func (a *Action) Error(e interface{}) {
	a.err = e
	a.LastCommand = ""
	panic(AbortedContextError{e})
}

//Session allow user to access the session inside ActionHandler through the Context()
func (a *Action) Session() SessionInterface {
	return a.session
}

//SourceMessage allow user to access the session inside ActionHandler through the Context()
func (a *Action) SourceMessage() BotMessageInterface {
	return a.sourceMessageGetter()
}

//Output allow user to access the OutMsgFactoryInterface inside ActionHandler through the Context()
func (a *Action) Output() OutMsgFactoryInterface {
	return a.sender
}

//Finish allow user to access finish command processing inside ActionHandler through the Context()
func (a *Action) Finish() {
	a.LastCommand = ""
}

//ActionContextInterface defines the context for ActionHandler
type ActionContextInterface interface {
	Args() []ArgInterface
	Output() OutMsgFactoryInterface
	Error(interface{})
	Session() SessionInterface
	SourceMessage() BotMessageInterface
	Finish()
}

//AbortedContextError is used to distinguish aborted context from other panics
type AbortedContextError struct {
	content interface{}
}

//DataGetSetter defines interface for saving/loading custom data inside the object
type DataGetSetter interface {
	SetData(value interface{})
	GetData(value interface{})
}

//PersistentSaver can save itself to permanent storage
type PersistentSaver interface {
	Save() error
}
