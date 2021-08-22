package internal

import (
	"fmt"
	"os"
	"reflect"
	"sync"
)

// TODO: Replace these with sync.Map
var mapMutex sync.Mutex
var nameMap = map[string]*DataType{}
var typeMap = map[reflect.Type]*DataType{}

// Register adds a type name to a data type mapping to the key-value object
// database. This type name is associated with the serialized bytes of the
// objects of the data type and is also included in secondary indicies.
//
// This function may fail if indexed field values are not convertible to
// strings. TODO: Provide an API to let user-defined converters to indexed
// field values.
func Register(datatype string, object interface{}) error {
	mapMutex.Lock()
	defer mapMutex.Unlock()

	if _, ok := nameMap[datatype]; ok {
		return os.ErrExist
	}
	stype, ok := getStructType(object)
	if !ok {
		return fmt.Errorf("input object is not a struct or a pointer-to-struct type: %w", os.ErrInvalid)
	}
	if _, ok := typeMap[stype]; ok {
		return fmt.Errorf("type cannot be registered under multiple names: %w", os.ErrInvalid)
	}

	t, err := NewDataType(datatype, object)
	if err != nil {
		return err
	}

	nameMap[datatype] = t
	typeMap[stype] = t
	return nil
}

// GetDataType returns data type handler for the object. Data type handlers are
// created by the Register function.
func GetDataType(object interface{}) (*DataType, error) {
	stype, ok := getStructType(object)
	if !ok {
		return nil, fmt.Errorf("input object must be a struct or pointer-to-struct: %w", os.ErrInvalid)
	}

	mapMutex.Lock()
	defer mapMutex.Unlock()

	t, ok := typeMap[stype]
	if !ok {
		return nil, fmt.Errorf("object type %T is not registered: %w", object, os.ErrInvalid)
	}
	return t, nil
}
