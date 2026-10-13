package graph

import (
	"strings"
	"testing"
)

// TestServedSchemaNamesAreBeanFree guards the item vocabulary of the API
// contract: no type, field, argument, input field, enum value or directive of
// the served schema may carry the legacy noun in any letter case.
func TestServedSchemaNamesAreBeanFree(t *testing.T) {
	schema := NewExecutableSchema(Config{}).Schema()

	check := func(kind, name string) {
		t.Helper()
		if strings.Contains(strings.ToLower(name), "bean") {
			t.Errorf("%s name %q contains the legacy noun", kind, name)
		}
	}

	for name, def := range schema.Types {
		check("type", name)
		for _, f := range def.Fields {
			check("field "+name+".", f.Name)
			for _, a := range f.Arguments {
				check("argument "+name+"."+f.Name+"(", a.Name)
			}
		}
		for _, v := range def.EnumValues {
			check("enum value "+name+".", v.Name)
		}
	}
	for name, d := range schema.Directives {
		check("directive", name)
		for _, a := range d.Arguments {
			check("directive argument @"+name+"(", a.Name)
		}
	}
}
