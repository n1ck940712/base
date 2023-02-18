package validate

import (
	"fmt"
	"reflect"
)

type vtype string

const (
	String vtype = "string"
	Int    vtype = "int"
	Float  vtype = "float"
	Number vtype = "number"
	Map    vtype = "map"
	Array  vtype = "array"
	Struct vtype = "struct"
	Nil    vtype = "nil"
)

type validateType struct {
	path  string
	items []vtype
}

func (validateType) getIdentifier() string {
	return "Type"
}

func (v *validateType) getPath() string {
	return v.path
}

func (v *validateType) validate(value any) error {
	notFound := true
	vtKind := reflect.ValueOf(value).Kind()

	for _, vtItem := range v.items {
		switch vtItem {
		case String:
			if vtKind == reflect.String {
				notFound = false
			}
		case Int:
			if isKindInt(vtKind) {
				notFound = false
			}
		case Float:
			if isKindFloat(vtKind) {
				notFound = false
			}
		case Number:
			if isKindNumber(vtKind) {
				notFound = false
			}
		case Map:
			if vtKind == reflect.Map {
				notFound = false
			}
		case Array:
			if vtKind == reflect.Array ||
				vtKind == reflect.Slice {
				notFound = false
			}
		case Struct:
			if vtKind == reflect.Struct {
				notFound = false
			}
		case Nil:
			if vtKind == reflect.Invalid {
				notFound = false
			}
		}
	}
	if notFound {
		if len(v.path) == 0 {
			if vtKind == reflect.Invalid {
				return fmt.Errorf("must be type (%v) but (%v)", v.items, "null")
			}
			return fmt.Errorf("must be type (%v) but (%v)", v.items, vtKind)
		}
		if vtKind == reflect.Invalid {
			return fmt.Errorf("(%v) must be type (%v) but (%v)", v.path, v.items, "null")
		}
		return fmt.Errorf("(%v) must be type (%v) but (%v)", v.path, v.items, vtKind)
	}
	return nil
}

func (v *validateType) validateItems(value any) error {
	vV := reflect.ValueOf(value)

	switch vV.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < vV.Len(); i++ {
			if err := v.validate(vV.Index(i).Interface()); err != nil {
				if len(v.path) == 0 {
					return fmt.Errorf("items must be type (%v)", v.items)
				}
				return fmt.Errorf("(%v) items must be type (%v)", v.path, v.items)
			}
		}
		return nil
	default:
		if len(v.path) == 0 {
			return fmt.Errorf("%v: (%v) is (%v) not an array/slice", v.getIdentifier(), value, vV.Kind())
		}
		return fmt.Errorf("%v: (%v) is (%v) not an array/slice for (%v)", v.getIdentifier(), value, vV.Kind(), v.path)
	}
}
