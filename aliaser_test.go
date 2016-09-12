package botmeans

import (
	"bytes"
	"io"
	"testing"
)

func TestAliaser(t *testing.T) {
	sources := []io.Reader{
		bytes.NewBuffer([]byte(`
{
    "ReplyKeyboard": {
        "ru": [
            [{"Text": "Команда в тексте", "Command": "/cmd1", "Args": "bla bla"}, {"Text": "Ещё команда в тексте", "Command": "/cmd2"}]
        ],
        "en": [
            [{"Text": "Command in text", "Command": "/cmd3"}, {"Text": "Another command in text", "Command": "/fffuuu", "Args": "9.87 qwe 10.5"}]
        ]
    }
}
            `)),
		bytes.NewBuffer([]byte(`buf`)),
	}
	aliaser := AliaserFromTemplates(sources)

	if cmd, args, ok := aliaser("ffuuu"); ok == true {
		t.Log(cmd)
		t.Log(args)
		t.Fail()
	}
	if cmd, args, ok := aliaser("Команда в тексте"); ok == true {
		if cmd != "/cmd1" {
			t.Error()
		}
		if len(args) != 2 {
			t.Error()
		} else {
			if a, _ := args[0].String(); a != "bla" {
				t.Error()
			}
			if a, _ := args[1].String(); a != "bla" {
				t.Error()
			}
		}
	} else {
		t.Error(cmd, args, ok)
	}
	if cmd, args, ok := aliaser("Ещё команда в тексте"); ok == true {
		if cmd != "/cmd2" {
			t.Error()
		}
		if len(args) > 0 {
			t.Error()
		}
	} else {
		t.Error()
	}
	if cmd, args, ok := aliaser("Command in text"); ok == true {
		if cmd != "/cmd3" {
			t.Error()
		}
		if len(args) > 0 {
			t.Error()
		}
	} else {
		t.Error()
	}
	if cmd, args, ok := aliaser("Another command in text"); ok == true {
		if cmd != "/fffuuu" {
			t.Error()
		}
		if len(args) != 3 {
			t.Error()
		} else {
			if a, _ := args[0].Float(); a != 9.87 {
				t.Error()
			}
			if a, _ := args[1].String(); a != "qwe" {
				t.Error()
			}
			if a, _ := args[2].Float(); a != 10.5 {
				t.Error()
			}
		}
	} else {
		t.Error(cmd, args, ok)
	}

}
