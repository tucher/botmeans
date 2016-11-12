package botmeans

import (
	"encoding/json"
	"io"
	"io/ioutil"
	// "log"
	"os"
	"path"
	// "strconv"
	"strings"
)

//AliaserFromTemplates creates aliaser from given set of templates
func AliaserFromTemplates(sourceList []io.Reader) CommandAliaser {

	ret := make(map[string]retStruct)

	for _, reader := range sourceList {
		if c, err := ioutil.ReadAll(reader); err == nil {

			c = []byte(strings.Replace(string(c), "\n", "", -1))

			template := MessageTemplate{}

			err := json.Unmarshal(c, &template)
			if err != nil {
				continue
			}
			handleTemplate(template, &ret)

		}

	}

	return func(text string) (cmd string, args Args, ook bool) {
		if r, ok := ret[text]; ok {
			cmd = r.Cmd
			args = r.Args
			ook = true
		}
		return
	}
}

//AliaserFromTemplateDir reads the given dir and calls AliaserFromTemplates
func AliaserFromTemplateDir(p string) CommandAliaser {
	files, _ := ioutil.ReadDir(p)

	readers := []io.Reader{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if f, err := os.Open(path.Join(p, file.Name())); err == nil {
			readers = append(readers, f)
		}

	}
	return AliaserFromTemplates(readers)
}

type retStruct struct {
	Cmd  string
	Args Args
}

func handleRow(row []MessageButton, ret *map[string]retStruct) {
	for _, button := range row {
		cmd := button.Command
		text := button.Args
		arguments := []arg{}

		splitted := []string{}
		for _, a := range strings.Split(text, " ") {
			trimmed := strings.TrimSpace(a)
			if len(trimmed) > 0 {
				splitted = append(splitted, trimmed)
			}
		}

		for _, str := range splitted {
			arguments = append(arguments, arg{str})
		}
		(*ret)[button.Text] = retStruct{cmd, args{arguments, text}}
	}
}

func handleTemplate(template MessageTemplate, ret *map[string]retStruct) {

	for _, keyboard := range template.ReplyKeyboard {
		for _, row := range keyboard {
			handleRow(row, ret)
		}
	}
}
