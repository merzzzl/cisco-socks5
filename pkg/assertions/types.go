package assertions

import "reflect"

func TypeOf[T any]() reflect.Type {
	var t T
	return reflect.TypeOf(t)
}

func As[T any](raw interface{}) (T, bool) {
	var zero T
	rv := reflect.ValueOf(raw)
	if rv.Type().AssignableTo(reflect.TypeOf(zero)) {
		return rv.Interface().(T), true
	}
	return zero, false
}
