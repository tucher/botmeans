### botmeans ###

Telegram bot framework

[![GoDoc](https://godoc.org/github.com/tucher/botmeans?status.svg)](https://godoc.org/github.com/tucher/botmeans)
[![Go Report Card](https://goreportcard.com/badge/github.com/tucher/botmeans)](https://goreportcard.com/report/github.com/tucher/botmeans)
[![Build Status](https://travis-ci.org/tucher/botmeans.svg?branch=master)](https://travis-ci.org/tucher/botmeans)
[![codebeat badge](https://codebeat.co/badges/29cd8e22-6895-4a25-9556-c6579e4684a8)](https://codebeat.co/projects/github-com-tucher-botmeans)

Центральная идея использования — создаёшь бота с конфигом телеграммных параметров и постгресовских функцией `botmeans.New`, а потом запускаешь работу с помощью:

```go
func (ui *MeansBot) Run(handlersProvider ActionHandlersProvider)
```

`ActionHandlersProvider` - это функция, которая принимает айдишник команды и возвращает функцию обработки команды.

*Функция обработки* — это функция, которая принимает ActionContextInterface и делает собственно какую то безнес логику.

`ActionContextInterface` – это типа как context в том же *gin*. То есть из него есть доступ практически ко всем фичам:

```go
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
```

`Cmd` — это имя команды, `Args` — аргументы, `Output()` — позволяет отправлять сообщения разных видов, включая рендеринг из шаблонов, создание клавиатур, и т.д. `Session` — доступ к сессии, там же можно сохранять произвольные данные. `SourceMessage` — можно изнать всякое о том сообщении, которое привело к исполнению команды. `ExecuteInSession` — это чтобы залезть в чужую сессию, например, отправить другому пользователю что-то. `CreateSession` — чтобы вручную добавить сессию, например, чтоб бот мог увидеть юзера, который ещё не слал сообщений.

# Test

`MEANS_DB_USERNAME=postgres MEANS_DBNAME=meansbot go test`
