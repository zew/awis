package recurse

import (
	"reflect"
	"testing"
)

type Pers struct {
	PersId   int
	LastName string
}

type Prod struct {
	ProdId      int
	PersId      int
	ProductName string `db:"product_name, size:200"`
}

type PersProd struct {
	Pers
	Prod
	// FooId int
}

var t1 = PersProd{
	Pers{17, "Heino"},
	Prod{32168, 18, "Cornflakes"},
}

func TestFieldsByName(txx *testing.T) {

	type tmp struct {
		// srcStruct interface{}
		key     string
		want    int
		classic bool
	}

	testCases := []tmp{
		{"PersId", 2, false},
		{"LastName", 1, true},
		{"ProdId", 1, true},
		{"ProductName", 1, true},
		{"product_name", 1, false},
		{"UNKNOWN", 0, false},
	}

	for _, fn := range testCases {
		fields, _ := FieldsByName(t1, fn.key)
		if len(fields) != fn.want {
			txx.Errorf("Fieldname -%v- got %v Fields - want %v\n", fn.key, len(fields), fn.want)
		}
		_, found := reflect.TypeOf(t1).FieldByName(fn.key)
		if found != fn.classic {
			txx.Errorf("Fieldname -%v- was classically found: %v ; expected: %v\n", fn.key, found, fn.classic)
		}
	}

	aSlice := []PersProd{t1, t1}

	for _, fn := range testCases {
		fields, _ := FieldsByName(aSlice, fn.key)
		if len(fields) != fn.want {
			txx.Errorf("Fieldname -%v- got %v Fields - want %v\n", fn.key, len(fields), fn.want)
		}
	}
	for _, fn := range testCases {
		fields, _ := FieldsByName(&aSlice, fn.key)
		if len(fields) != fn.want {
			txx.Errorf("Fieldname -%v- got %v Fields - want %v\n", fn.key, len(fields), fn.want)
		}
	}

}
