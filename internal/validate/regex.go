package validate

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type vregex string

type validateRegex struct {
	path  string
	items []vregex
}

func (validateRegex) getIdentifier() string {
	return "Regex"
}

func (v *validateRegex) getPath() string {
	return v.path
}

func (v *validateRegex) validate(value any) error {
	for _, vrItem := range v.items {
		match, err := regexp.MatchString(string(vrItem), fmt.Sprintf("%v", value))

		if err != nil {
			return err
		}

		if !match {
			if len(v.path) == 0 {
				return fmt.Errorf("does not match (%v)", vrItem)
			}
			return fmt.Errorf("(%v) does not match (%v)", v.path, vrItem)
		}
	}
	return nil
}

func (v *validateRegex) validateItems(value any) error {
	vV := reflect.ValueOf(value)

	switch vV.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < vV.Len(); i++ {
			if err := v.validate(vV.Index(i).Interface()); err != nil {
				if strings.Contains(err.Error(), "regexp") {
					return err
				}
				if len(v.path) == 0 {
					return fmt.Errorf("does not match (%v)", v.items)
				}
				return fmt.Errorf("(%v) does not match (%v)", v.path, v.items)
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
