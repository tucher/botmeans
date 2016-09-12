package botmeans

import (
	"encoding/json"
	"reflect"
)

func serialize(current string, value interface{}) string {
	if current == "" {
		current = "{}"
	}
	if value == nil {
		return current
	}

	container := make(map[string]*json.RawMessage)
	json.Unmarshal([]byte(current), &container)
	d, _ := json.Marshal(value)
	rm := json.RawMessage(d)

	t := ""
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		t = reflect.TypeOf(value).Name()
	} else {
		t = reflect.Indirect(reflect.ValueOf(value)).Type().Name()
	}

	container[t] = &rm

	d, _ = json.Marshal(container)
	current = string(d)
	return current
}

func deserialize(current string, value interface{}) {
	if value == nil || reflect.TypeOf(value).Kind() != reflect.Ptr {
		return
	}
	t := reflect.Indirect(reflect.ValueOf(value)).Type().Name()
	container := make(map[string]*json.RawMessage)
	json.Unmarshal([]byte(current), &container)
	if v, ok := container[t]; ok {
		json.Unmarshal(*v, value)
	}
}
