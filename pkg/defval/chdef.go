package defval

import "reflect"

// DWR - Default Value Replacement - Проверяет значение переданной переменной val на значение по-умолчанию
// В случае если val имеет значение по-умлочанию в ответе возвращается subsVal
// Исползует рефлексию
func DVR[T any](val T, subsVal T) T {

	lval := reflect.ValueOf(val)
	zero := reflect.Zero(lval.Type())

	if reflect.DeepEqual(lval.Interface(), zero.Interface()) {
		return subsVal
	}

	return val
}
