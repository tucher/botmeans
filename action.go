package botmeans

//ActionHandler defines the type of handler function
type ActionHandler func(context ActionContextInterface)

//ActionHandlersProvider returns ActionHandler for given command
type ActionHandlersProvider func(id string) (ActionHandler, bool)

type ChatSession interface {
	Identifiable
	ChatIdentifier
	UserIdentifier
	DataGetSetter
	PersistentSaver
	UserName() string
	ChatTitle() string
	IsOneToOne() bool
	SetLocale(string)
	Locale() string
}

type ActionSessionInterface interface {
	DataGetSetter
	PersistentSaver
	ChatIdentifier
	IsNew() bool
	Locale() string
}

type actionExecuterFactoryConfig struct {
	cmdGetter       func() string
	argsGetter      func() Args
	sourceMsgGetter func() BotMessageInterface
}

//ActionFactory generates Executers
func ActionFactory(
	sessionBase SessionBase,
	sessionFactory SessionFactory,
	getters actionExecuterFactoryConfig,
	senderFactory senderFactory,
	out chan Executer,
	handlersProvider ActionHandlersProvider,
) {
	session, err := sessionFactory(sessionBase)
	if err != nil {
		return
	}
	if session.IsNew() {

		out <- &Action{
			session:          session,
			sessionFactory:   sessionFactory,
			handlersProvider: handlersProvider,
			getters: actionExecuterFactoryConfig{
				cmdGetter:       func() string { return "" },
				argsGetter:      func() Args { return args{[]arg{arg{session}}, ""} },
				sourceMsgGetter: func() (r BotMessageInterface) { return },
			},
			senderFactory: senderFactory,
			execChan:      out,
		}
	}
	//if _, ok := handlersProvider(getters.cmdGetter()); ok == true {

	ret := &Action{
		session:          session,
		sessionFactory:   sessionFactory,
		handlersProvider: handlersProvider,
		getters:          getters,
		senderFactory:    senderFactory,
		execChan:         out,
	}
	session.GetData(ret)
	out <- ret
	// } else {
	// 	out <- nil
	// }

}

//Action provides the context for the user command
type Action struct {
	session        ActionSessionInterface
	sessionFactory SessionFactory
	LastCommand    string
	getters        actionExecuterFactoryConfig

	handlersProvider ActionHandlersProvider
	senderFactory    senderFactory
	err              interface{}
	execChan         chan Executer
	passedCmd        string
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
	a.passedCmd = a.getters.cmdGetter()

	if _, ok = a.handlersProvider(a.passedCmd); ok == true && a.passedCmd != "" {
		a.LastCommand = a.passedCmd
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
func (a *Action) Args() Args {
	return a.getters.argsGetter()
}

//Error allow user to terminate ActionHandler through the Context()
func (a *Action) Error(e interface{}) {
	a.err = e
	a.LastCommand = ""
	panic(AbortedContextError{e})
}

//Session allow user to access the session inside ActionHandler through the Context()
func (a *Action) Session() ChatSession {
	if v, ok := a.session.(SessionInterface); ok {
		return v
	}
	return nil
}

//SourceMessage allow user to access the session inside ActionHandler through the Context()
func (a *Action) SourceMessage() BotMessageInterface {
	return a.getters.sourceMsgGetter()
}

//Output allow user to access the OutMsgFactoryInterface inside ActionHandler through the Context()
func (a *Action) Output() OutMsgFactoryInterface {
	return a.senderFactory(a.session)
}

//Finish allow user to access finish command processing inside ActionHandler through the Context()
func (a *Action) Finish() {
	a.LastCommand = ""
}

func (a *Action) Cmd() string {
	return a.passedCmd
}

func (a *Action) CreateSession(base SessionBase) error {
	session, err := a.sessionFactory(base)
	if err != nil {
		return err
	}
	if session.IsNew() {

		a.execChan <- &Action{
			session:          session,
			sessionFactory:   a.sessionFactory,
			handlersProvider: a.handlersProvider,
			getters: actionExecuterFactoryConfig{
				cmdGetter:       func() string { return "" },
				argsGetter:      func() Args { return args{[]arg{arg{session}}, ""} },
				sourceMsgGetter: func() (r BotMessageInterface) { return },
			},
			senderFactory: a.senderFactory,
			execChan:      a.execChan,
		}
	}
	return nil
}

type execHelper struct {
	a *Action
	f ActionHandler
}

func (helper execHelper) Id() int64 {
	return helper.a.Id()
}

func (helper execHelper) Execute() {
	helper.f(helper.a)
}

//ExecuteInSession allows to execute some function in the same goroutine as other action for given session.
//Can be used to exec commands for chat created from another chat
func (a *Action) ExecuteInSession(s ChatSession, f ActionHandler) {
	if session, ok := s.(ActionSessionInterface); ok {
		a.execChan <- execHelper{&Action{
			session:          session,
			getters:          a.getters,
			handlersProvider: a.handlersProvider,
			senderFactory:    a.senderFactory,
		}, f}
	}
}

//ActionContextInterface defines the context for ActionHandler
type ActionContextInterface interface {
	Cmd() string
	Args() Args
	Output() OutMsgFactoryInterface
	Error(interface{})
	Session() ChatSession
	SourceMessage() BotMessageInterface
	Finish()
	ExecuteInSession(s ChatSession, f ActionHandler)
	CreateSession(base SessionBase) error
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
