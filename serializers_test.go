package botmeans

import (
	"testing"
)

func TestSerializers(t *testing.T) {
	current := ""
	type Struct1 struct {
		Foo int
		Moo string
	}
	type Struct2 struct {
		Loo []int
		Zoo []string
	}
	in := Struct1{12, "42"}

	current = serialize(current, in)
	current = serialize(current, Struct2{[]int{12, 42}, []string{"42", "54"}})

	out := Struct1{}
	deserialize(current, &out)
	if in != out {
		t.Fail()
	}

	out2 := Struct2{}
	deserialize(current, &out2)

	in = Struct1{13, "43"}
	current = serialize(current, in)

	out = Struct1{}
	deserialize(current, &out)

	if in != out {
		t.Fail()
	}

}
