package internal

import (
	"os"
	"reflect"
	"sync"
)

var userDefinedTypesMapMutex sync.Mutex
var userDefinedTypesMap = map[reflect.Type]func(reflect.Value) (string, error){}

func RegisterIndexFieldType(ftype reflect.Type, stringer func(reflect.Value) (string, error)) error {
	userDefinedTypesMapMutex.Lock()
	defer userDefinedTypesMapMutex.Unlock()

	if _, ok := userDefinedTypesMap[ftype]; ok {
		return os.ErrExist
	}
	userDefinedTypesMap[ftype] = stringer
	return nil
}

func FormatIndexFieldValue(v reflect.Value) (string, error) {
	userDefinedTypesMapMutex.Lock()
	stringer, ok := userDefinedTypesMap[v.Type()]
	userDefinedTypesMapMutex.Unlock()

	if !ok {
		return "", os.ErrNotExist
	}
	return stringer(v)
}
