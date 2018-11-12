package datapb

import (
	"reflect"
	"fmt"
)

func ExampleToData() {
	d, _ := ToData(reflect.ValueOf(map[string]reflect.Value{`t`: reflect.ValueOf(3)}))
	fmt.Printf("'%v'\n", d)

	d, _ = ToData(reflect.ValueOf(map[string]int{`t`: 3}))
	fmt.Printf("'%v'\n", d)
	// Output:
	// 'hash_value:<entries:<key:"t" value:<integer_value:3 > > > '
	// 'hash_value:<entries:<key:"t" value:<integer_value:3 > > > '
}
