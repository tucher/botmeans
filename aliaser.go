package botmeans

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strconv"
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
				//ui.Logger.Println("Error in template ", file.Name())
				continue
			}
			handleTemplate(template, ret)

		}

	}
	return func(text string) (cmd string, args []ArgInterface, ook bool) {
		if r, ok := ret[text]; ok {
			cmd = r.Cmd
			args = r.Args
			ook = true
		}
		return
	}
}

//AliaserFromTemplateDir reads the given dir and calls AliaserFromTemplates
func AliaserFromTemplateDir(path string) CommandAliaser {
	files, _ := ioutil.ReadDir(path)

	readers := []io.Reader{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if f, err := os.Open(path + file.Name()); err == nil {
			readers = append(readers, f)
		}

	}
	return AliaserFromTemplates(readers)
}

type retStruct struct {
	Cmd  string
	Args []ArgInterface
}

func handleRow(row []MessageButton, ret map[string]retStruct) {
	for _, button := range row {
		cmd := button.Command
		text := button.Args
		Args := []ArgInterface{}

		splitted := []string{}
		for _, a := range strings.Split(text, " ") {
			trimmed := strings.TrimSpace(a)
			if len(trimmed) > 0 {
				splitted = append(splitted, trimmed)
			}
		}

		for _, str := range splitted {
			if val, ok := strconv.ParseFloat(str, 64); ok == nil {
				Args = append(Args, Arg{val})
			} else {
				Args = append(Args, Arg{str})
			}
		}
		ret[button.Text] = struct {
			Cmd  string
			Args []ArgInterface
		}{cmd, Args}
	}
}

func handleTemplate(template MessageTemplate, ret map[string]retStruct) {

	for _, keyboard := range template.ReplyKeyboard {
		for _, row := range keyboard {
			handleRow(row, ret)
		}
	}
}
