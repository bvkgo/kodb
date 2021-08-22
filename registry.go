package kodb

import (
	"reflect"

	"github.com/bvkgo/kodb/internal"
)

// RegisterDataType adds a new object type and it's type name to the database.
func RegisterDataType(name string, otype reflect.Type) error {
	return internal.Register(name, reflect.New(otype).Interface())
}

// RegisterIndexStringer adds a string converter for an index field type.
func RegisterIndexStringer(ftype reflect.Type, stringer func(reflect.Value) (string, error)) error {
	return internal.RegisterIndexFieldType(ftype, stringer)
}
