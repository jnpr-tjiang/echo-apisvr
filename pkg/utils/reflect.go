package utils

import "reflect"

// TypeOf returns object type
func TypeOf(myvar interface{}) (res string) {
	t := reflect.TypeOf(myvar)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return res + t.Name()
}
