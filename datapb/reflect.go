package datapb

import (
	"fmt"
	"reflect"
)

// ToData converts reflect.Value to a datapb.Data
func ToData(v reflect.Value) (*Data, error) {
	if !v.IsValid() {
		return &Data{Kind: &Data_UndefValue{}}, nil
	}

	switch v.Kind() {
	case reflect.Bool:
		return &Data{Kind: &Data_BooleanValue{v.Bool()}}, nil
	case reflect.Float32, reflect.Float64:
		return &Data{Kind: &Data_FloatValue{v.Float()}}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Data{Kind: &Data_IntegerValue{v.Int()}}, nil
	case reflect.String:
		return &Data{Kind: &Data_StringValue{v.String()}}, nil
	case reflect.Slice, reflect.Array:
		cnt := v.Len()
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			return &Data{Kind: &Data_BinaryValue{v.Bytes()}}, nil
		}
		els := make([]*Data, cnt)
		for i := 0; i < cnt; i++ {
			elem, err := ToData(v.Index(i))
			if err != nil {
				return nil, err
			}
			els[i] = elem
		}
		return &Data{Kind: &Data_ArrayValue{&DataArray{els}}}, nil
	case reflect.Map:
		cnt := v.Len()
		els := make([]*DataEntry, cnt)
		for i, k := range v.MapKeys() {
			key, err := ToData(k)
			if err != nil {
				return nil, err
			}
			val, err := ToData(v.MapIndex(k))
			if err != nil {
				return nil, err
			}
			els[i] = &DataEntry{key, val}
		}
		return &Data{Kind: &Data_HashValue{&DataHash{els}}}, nil
	case reflect.Struct:
		if v.Type().String() == `reflect.Value` {
			return ToData(v.Interface().(reflect.Value))
		}
	case reflect.Interface:
		if v.Type() == interfaceType && v.Interface() == nil {
			// The interface{} nil value represents a generic nil
			return &Data{Kind: &Data_UndefValue{}}, nil
		}
		return ToData(v.Elem())
	}
	return nil, fmt.Errorf(`unable to convert a value of kind '%s' and type '%s' to Data`, v.Kind(), v.Type().Name())
}

var interfaceType = reflect.TypeOf([]interface{}{}).Elem()
var GenericNilValue = reflect.Zero(interfaceType)
var InvalidValue = reflect.ValueOf(nil)

// FromData converts a datapb.Data to a reflect.Value
func FromData(v *Data) (reflect.Value, error) {
	if v.Kind == nil {
		return GenericNilValue, nil
	}

	switch v.Kind.(type) {
	case *Data_BooleanValue:
		return reflect.ValueOf(v.GetBooleanValue()), nil
	case *Data_FloatValue:
		return reflect.ValueOf(v.GetFloatValue()), nil
	case *Data_IntegerValue:
		return reflect.ValueOf(v.GetIntegerValue()), nil
	case *Data_StringValue:
		return reflect.ValueOf(v.GetStringValue()), nil
	case *Data_UndefValue:
		return GenericNilValue, nil
	case *Data_BinaryValue:
		return reflect.ValueOf(v.GetBinaryValue()), nil
	case *Data_ArrayValue:
		av := v.GetArrayValue().GetValues()
		vals := make([]reflect.Value, len(av))
		var et reflect.Type = nil
		for i, elem := range av {
			rv, err := FromData(elem)
			if err != nil {
				return InvalidValue, err
			}
			rt := rv.Type()
			if et == nil {
				et = rt
			} else if et != rt {
				et = nil
			}
			vals[i] = rv
		}
		if et == nil {
			et = interfaceType
		}
		return reflect.Append(reflect.MakeSlice(reflect.SliceOf(et), 0, len(vals)), vals...), nil
	case *Data_HashValue:
		av := v.GetHashValue().Entries
		vals := make([]reflect.Value, len(av))
		keys := make([]reflect.Value, len(av))
		var kType reflect.Type = nil
		var vType reflect.Type = nil
		for i, elem := range av {
			rv, err := FromData(elem.Key)
			if err != nil {
				return InvalidValue, err
			}
			keys[i] = rv
			rt := rv.Type()
			if kType == nil {
				kType = rt
			} else if kType != rt {
				kType = nil
			}

			rv, err = FromData(elem.Value)
			if err != nil {
				return InvalidValue, err
			}
			vals[i] = rv
			rt = rv.Type()
			if vType == nil {
				vType = rt
			} else if vType != rt {
				vType = nil
			}
		}
		if kType == nil {
			kType = interfaceType
		}
		if vType == nil {
			vType = interfaceType
		}
		hash := reflect.MakeMapWithSize(reflect.MapOf(kType, vType), len(vals))
		for i, k := range keys {
			hash.SetMapIndex(k, vals[i])
		}
		return hash, nil
	default:
		return InvalidValue, fmt.Errorf(`unable to convert a value of type '%T' to reflect.Value`, v.Kind)
	}
}
