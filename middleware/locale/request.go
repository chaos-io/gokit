package locale

import (
	"reflect"
	"strings"
)

func ResolveFromBase(req any) (string, bool) {
	if req == nil {
		return "", false
	}

	value := reflect.ValueOf(req)
	if !value.IsValid() {
		return "", false
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "", false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return "", false
	}

	baseField := value.FieldByName("Base")
	if !baseField.IsValid() {
		return "", false
	}
	if baseField.Kind() == reflect.Pointer {
		if baseField.IsNil() {
			return "", false
		}
		baseField = baseField.Elem()
	}
	if baseField.Kind() != reflect.Struct {
		return "", false
	}

	localeField := baseField.FieldByName("Locale")
	if !localeField.IsValid() || localeField.Kind() != reflect.String {
		return "", false
	}

	locale := strings.TrimSpace(localeField.String())
	return locale, len(locale) > 0
}
