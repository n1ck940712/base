package validate

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type ivalidation interface {
	getIdentifier() string
	getPath() string
	validate(value any) error
	validateItems(value any) error
}

type validate struct {
	source       any
	vValidations []ivalidation
}

func Compose(source any) *validate {
	return &validate{source, []ivalidation{}}
}

func (v *validate) Value(path string, items ...any) *validate {
	v.vValidations = append(v.vValidations, &validateValue{path, items})
	return v
}

func (v *validate) Regex(path string, items ...vregex) *validate {
	v.vValidations = append(v.vValidations, &validateRegex{path, items})
	return v
}

func (v *validate) Type(path string, items ...vtype) *validate {
	v.vValidations = append(v.vValidations, &validateType{path, items})
	return v
}

func (v *validate) Check() error {
	sfound := map[string]any{} //store found using path

	//validate validations
	if err := v.validateValidations(&sfound); err != nil {
		return err
	}
	return nil
}

func (v *validate) validateValidations(sfound *map[string]any) error {
	for _, vValidation := range v.vValidations {
		path := vValidation.getPath()
		toFind := v.source

		if len(path) > 0 && (*sfound)[path] != nil {
			toFind = (*sfound)[path] //use stored found
			canValidateItems := shouldValidateItems(&path)

			if canValidateItems {
				if err := vValidation.validateItems(toFind); err != nil {
					return err
				}
			} else {
				if err := vValidation.validate(toFind); err != nil {
					return err
				}
			}
		} else {
			pathLevels := strings.Split(path, ".")

			if err := evaluatePathLevels(true, vValidation, pathLevels, toFind, sfound); err != nil {
				return err
			}
		}
	}

	return nil
}

//isMain - true on main call
//for objects
func evaluatePathLevels(isMain bool, validation ivalidation, pathLevels []string, source any, sfound *map[string]any) error {
	fullPath := strings.Join(pathLevels, ".")
	toFind := source
	canValidateItems := false

	for vplIndex, vPathLevel := range pathLevels {
		if vPathLevel == "" || vPathLevel == "[]" {
			canValidateItems = shouldValidateItems(&vPathLevel)
			break
		}
		isArr, err := evaluateIsArr(validation, vplIndex, pathLevels, toFind, sfound)

		if err != nil {
			return err
		}
		if isArr {
			return nil
		}
		canValidateItems = shouldValidateItems(&vPathLevel)
		found, err := reflectFind(vPathLevel, toFind)

		if err != nil {
			if isMain {
				if strings.Contains(err.Error(), "find kind:") { //use find kind error
					return err
				}
				return fmt.Errorf("(%v) not found", strings.Join(pathLevels, "."))
			}
			return err
		}
		toFind = found
	}

	if isMain && len(fullPath) > 0 {
		(*sfound)[fullPath] = toFind //save found
	}

	if canValidateItems {
		if err := validation.validateItems(toFind); err != nil {
			return err
		}
	} else {
		if err := validation.validate(toFind); err != nil {
			return err
		}
	}
	return nil
}

//for array
func evaluateIsArr(validation ivalidation, currentIndex int, pathLevels []string, source any, sfound *map[string]any) (bool, error) {
	aV := reflect.ValueOf(source)

	switch aV.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < aV.Len(); i++ {
			if err := evaluatePathLevels(false, validation, pathLevels[currentIndex:], aV.Index(i).Interface(), sfound); err != nil {
				return true, err
			}
		}
		return true, nil
	default:
		return false, nil
	}
}

//find value with key
func reflectFind(key string, toFind any) (any, error) {
	var value reflect.Value

	if rValue, ok := toFind.(reflect.Value); ok {
		value = rValue
	} else {
		value = reflect.ValueOf(toFind)
	}

	switch rKind := value.Kind(); rKind {
	case reflect.String:
		return value.String(), nil
	case reflect.Int:
		return value.Int(), nil
	case reflect.Float64:
		return value.Float(), nil
	case reflect.Map:
		for _, v := range value.MapKeys() {
			if key == v.Interface() {
				return pointerToValueIfNeeded(value.MapIndex(v).Interface()), nil
			}
		}
	case reflect.Interface:
		return reflectFind(key, value.Elem())
	case reflect.Struct:
		if key == "" {
			return value.Interface(), nil
		}
		for i := 0; i < value.NumField(); i++ {

			if key == value.Type().Field(i).Name {
				return reflectFind("", pointerToValueIfNeeded(value.Field(i)))
			}
		}
	case reflect.Pointer:
		if key == "" {
			ptr := reflect.NewAt(value.Elem().Type(), value.UnsafePointer())

			return ptr.Interface(), nil
		}
		return reflectFind(key, value.Elem())
	default:
		return nil, fmt.Errorf("find kind:(%v) is not supported on validate.reflectFind for (%v)", rKind, key)
	}
	return nil, fmt.Errorf("%v not found", key)
}

//converts any or reflect value pointer to value
func pointerToValueIfNeeded(a any) any {
	var aV reflect.Value
	if rValue, ok := a.(reflect.Value); ok {
		aV = rValue
	} else {
		aV = reflect.ValueOf(a)
	}

	if aV.Kind() == reflect.Ptr {
		ptr := reflect.NewAt(aV.Elem().Type(), aV.UnsafePointer())

		return ptr.Elem().Interface()
	}
	return a
}

func shouldValidateItems(pathLevel *string) bool {
	plLen := len(*pathLevel)

	if plLen > 1 && (*pathLevel)[plLen-2:] == "[]" {
		*pathLevel = (*pathLevel)[:plLen-2]
		return true
	}

	return false
}

func isKindInt(kind reflect.Kind) bool {
	return kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 ||
		kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64
}

func isKindFloat(kind reflect.Kind) bool {
	return kind == reflect.Float32 ||
		kind == reflect.Float64
}

func isKindNumber(kind reflect.Kind) bool {
	return isKindFloat(kind) || isKindInt(kind)
}

func readable(a any) string {
	jBytes, _ := json.Marshal(a)

	return strings.ReplaceAll(strings.ReplaceAll(string(jBytes), `"`, `”`), ",", ", ")
}
