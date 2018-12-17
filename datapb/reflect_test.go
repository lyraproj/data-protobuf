package datapb

import (
	"fmt"
	"reflect"
)

func ExampleToData() {
	d, _ := ToData(reflect.ValueOf(map[string]int{`t`: 3}))
	fmt.Println(d)
	// Output:
	// hash_value:<entries:<key:<string_value:"t" > value:<integer_value:3 > > >
}

func ExampleToData_fromReflect() {
	d, _ := ToData(reflect.ValueOf(map[string]reflect.Value{`t`: reflect.ValueOf(3)}))
	fmt.Println(d)
	// Output:
	// hash_value:<entries:<key:<string_value:"t" > value:<integer_value:3 > > >
}

func ExampleFromData() {
	d, _ := ToData(reflect.ValueOf(map[string]int{`t`: 3}))

	v, _ := FromData(d)
	fmt.Printf("%T '%v'\n", v.Interface(), v.Interface())
	// Output:
	// map[string]int64 'map[t:3]'
}

func ExampleFromData_intkeys() {
	d, _ := ToData(reflect.ValueOf(map[int]int{1: 3}))

	v, _ := FromData(d)
	fmt.Printf("%T '%v'\n", v.Interface(), v.Interface())
	// Output:
	// map[int64]int64 'map[1:3]'
}
