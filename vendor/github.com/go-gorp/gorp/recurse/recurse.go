package recurse

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
)

var logl = log.New(os.Stdout, "", log.Lshortfile)

func init() {
	logl.SetOutput(ioutil.Discard)
}

func scalarField(k reflect.Kind) bool {

	if k == reflect.Bool ||
		k == reflect.Int ||
		k == reflect.Int8 ||
		k == reflect.Int16 ||
		k == reflect.Int32 ||
		k == reflect.Int64 ||
		k == reflect.Uint ||
		k == reflect.Uint8 ||
		k == reflect.Uint16 ||
		k == reflect.Uint32 ||
		k == reflect.Uint64 ||
		// k == reflect.Uintptr ||
		k == reflect.Float32 ||
		k == reflect.Float64 ||
		k == reflect.Complex64 ||
		k == reflect.Complex128 ||
		// k == reflect.Array ||
		// k == reflect.Chan ||
		// k == reflect.Func ||
		k == reflect.Interface ||
		// k == reflect.Map ||
		// k == reflect.Ptr ||
		// k == reflect.Slice ||
		k == reflect.String ||
		// k == reflect.Struct ||
		// k == reflect.UnsafePointer ||
		false {
		return true
	}
	return false
}

func appendMultiFieldIndexChain(chain [][]int, appendix []int) [][]int {
	// Unbelievable:
	// Unless we do this *hard copy* of appendix
	// the appended slice keeps changing
	detachedNewSlice := make([]int, len(appendix))
	copy(detachedNewSlice, appendix)

	chain = append(chain, detachedNewSlice)
	return chain
}

// Returns all fields of a struct with the given name.
//
func FieldsByName(i interface{}, name string) ([]reflect.StructField, [][]int) {

	name = strings.ToLower(name)

	fds := []reflect.StructField{}  // fields
	multiFieldIdxChain := [][]int{} // indexes to delve into master struct

	curIdxQueue := []int{}

	var recF func(i interface{})
	recF = func(i interface{}) {

		v := reflect.ValueOf(i) // value of interface
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		logl.Printf("    %v subfields for %-14v - %+v \n", v.NumField(), reflect.TypeOf(v.Interface()).Name(), v.Interface())

		for i := 0; i < v.NumField(); i++ {

			fieldName := strings.ToLower(v.Type().Field(i).Name)

			if !v.Field(i).CanInterface() {
				mayBeUppper := v.Type().Field(i).Name
				if fieldName[:1] == mayBeUppper[:1] {
					logl.Printf("\t\t\tSkipping unexported field %v", fieldName)
					continue
				} else {
					logl.Printf("\t\tCannot get interface for %v. Should never happen.", fieldName)
				}
			}

			kind := v.Field(i).Kind()
			anon := v.Type().Field(i).Anonymous
			namedStruct := false
			if !anon && kind == reflect.Struct {
				namedStruct = true
			}

			added := len(v.Type().Field(i).Index)
			curIdxQueue = append(curIdxQueue, v.Type().Field(i).Index...)

			// found struct variable
			if fieldName == name && scalarField(kind) {
				fds = append(fds, v.Type().Field(i))
				multiFieldIdxChain = appendMultiFieldIndexChain(multiFieldIdxChain, curIdxQueue)
			} else {
				// dont do it twice
				// found struct variable with alias
				tg := v.Type().Field(i).Tag.Get("db")
				cArguments := strings.Split(tg, ",")
				alternFieldName := strings.ToLower(strings.TrimSpace(cArguments[0]))
				if alternFieldName == name && scalarField(kind) {
					fds = append(fds, v.Type().Field(i))
					multiFieldIdxChain = appendMultiFieldIndexChain(multiFieldIdxChain, curIdxQueue)
				} else {
					if namedStruct && fieldName == name && isScannable(v.Type().Field(i)) {
						// found struct with custom scan method
						fds = append(fds, v.Type().Field(i))
						multiFieldIdxChain = appendMultiFieldIndexChain(multiFieldIdxChain, curIdxQueue)
					}
				}
			}

			logl.Printf("\t\t%-14v %-8v %-8v %v\n", fieldName, anon, namedStruct, kind)
			if fieldName == name && isScannable(v.Type().Field(i)) {
				logl.Printf("\t\t with exception for struct with custom sql.Scanner, i.e. NullTime")
			}

			// recurse deeper
			if anon && kind == reflect.Struct {
				// if kind == reflect.Struct {
				realVal := v.Field(i).Interface()
				recF(realVal)
			}

			curIdxQueue = curIdxQueue[:len(curIdxQueue)-added]

		}
	}

	//
	//
	v := reflect.ValueOf(i)

	// dereference pointers to interface()
	doDoubleElemBelow := false
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		doDoubleElemBelow = true
	}

	// having a slice, we need its element type
	if v.Kind() == reflect.Slice {
		sliceType := reflect.TypeOf(i).Elem()
		if doDoubleElemBelow {
			sliceType = sliceType.Elem()
		}

		if sliceType.Kind() == reflect.Ptr {
			sliceType = sliceType.Elem() // this is to remove the second asterisk in case of  *[]*type
		}

		v = reflect.New(sliceType)
		i = v.Interface()
		v = reflect.ValueOf(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		logl.Printf("\tbase type of slice is %q of kind %q\n", sliceType.Name(), v.Kind())
	}

	//
	//
	// Double dereferencing
	// var p3 *Person
	// SelectOne(&p3, "select * from ...
	if v.Kind() == reflect.Ptr {
		logl.Printf("    it is still a pointer")
		// i = v.Interface()
		// v = reflect.ValueOf(i)
		if v.IsNil() {
			logl.Printf("    %+v is NIL ", v)
			t1, _ := ToStructType(i)
			i = reflect.New(t1).Interface()
			v = reflect.ValueOf(i)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			logl.Printf("    t1 %+v - i %v %T - v %v", t1, i, i, v)
		}
	}

	logl.Printf("i is of type %T and of kind %-12q  - %+v\n", i, v.Kind(), i)

	if v.Kind() == reflect.Struct {
		recF(i) // recurse into struct
	}

	if len(multiFieldIdxChain) > 1 {
		logl.Printf("chain of indexes for several fields: %v ", fds[0].Name)
		logl.Printf("index paths: %v ", multiFieldIdxChain)
	}

	return fds, multiFieldIdxChain

}

// To struct type
func ToStructType(i interface{}) (reflect.Type, error) {
	t := reflect.TypeOf(i)
	// If a Pointer to a type, follow
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("gorp: cannot SELECT into this type: %v", reflect.TypeOf(i))
	}
	return t, nil
}

func isScannable(sf reflect.StructField) (scannable bool) {

	typ := sf.Type

	if false {
		// I failed to create a type *sql.Scanner
		compareType := reflect.TypeOf(new(sql.Scanner)).Elem()
		scannable = typ.Implements(compareType)
	}

	sfInstance := reflect.New(typ).Interface()
	_, scannable = interface{}(sfInstance).(sql.Scanner)

	if scannable {
		logl.Printf("Type %q implements sql.Scanner", typ)
	}

	return scannable

}
