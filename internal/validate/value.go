package validate

import (
	"fmt"
	"reflect"
)

type ValueCallback func(value any) error

type validateValue struct {
	path  string
	items []any
}

func (validateValue) getIdentifier() string {
	return "Value"
}

func (v *validateValue) getPath() string {
	return v.path
}

func (v *validateValue) validate(value any) error {
	contains := false
	vCallbackErrors := map[int]error{}

	for vItemIndex, vVItem := range v.items {
		if vCallback, ok := vVItem.(ValueCallback); ok {
			if err := vCallback(value); err != nil {
				vCallbackErrors[vItemIndex] = err
			} else {
				contains = true
				break
			}
			continue
		}
		rVVItem := reflect.ValueOf(vVItem)
		rValue := reflect.ValueOf(value)

		if isKindNumber(rVVItem.Kind()) && isKindNumber(rValue.Kind()) {
			diff := 1.0
			rType := reflect.TypeOf(diff)
			cVVItem := rVVItem.Convert(rType).Float()
			cValue := rValue.Convert(rType).Float()
			if (cVVItem - cValue) == 0 {
				contains = true
				break
			}
		} else if reflect.DeepEqual(vVItem, value) {
			contains = true
			break
		}
	}

	if !contains {
		uVItems := []any{}

		for i, vItem := range v.items {
			if err := vCallbackErrors[i]; err != nil {
				return err
			} else {
				uVItems = append(uVItems, vItem)
			}
		}
		if len(v.path) == 0 {
			return fmt.Errorf("must contain (%s)", readable(uVItems))
		}
		return fmt.Errorf("(%v) must contain (%s)", v.path, readable(uVItems))
	}
	return nil
}

func (v *validateValue) validateItems(value any) error {
	vV := reflect.ValueOf(value)

	switch vV.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < vV.Len(); i++ {
			if err := v.validate(vV.Index(i).Interface()); err != nil {
				if len(v.path) == 0 {
					return fmt.Errorf("items must contain (%v)", readable(v.items))
				}
				return fmt.Errorf("(%v) items must contain (%v)", v.path, readable(v.items))
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
