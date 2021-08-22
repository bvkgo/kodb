package internal

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
)

const StructTagName = "kodb"

var SupportedKinds = []reflect.Kind{
	reflect.Bool,
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.String,
}
var supportedKindsMap = map[reflect.Kind]struct{}{}

var SupportedTypes = []reflect.Type{
	reflect.TypeOf([]byte{}),
	reflect.TypeOf(net.IP{}),
}
var supportedTypesMap = map[reflect.Type]struct{}{}

func init() {
	for _, v := range SupportedKinds {
		supportedKindsMap[v] = struct{}{}
	}
	for _, v := range SupportedTypes {
		supportedTypesMap[v] = struct{}{}
	}
}

type IndexField struct {
	// Name holds the field name. TODO: How to represent nested field names in
	// canonical form?
	name string

	// Position indicates field index in the object. TODO: We should supported
	// nested field positions in future.
	position []int

	// stringer if not-nil holds the user-defined stringer for an index field.
	stringer func(reflect.Value) (string, error)
}

func NewIndexFields(sfield reflect.StructField) ([]*IndexField, error) {
	tag, ok := sfield.Tag.Lookup(StructTagName)
	if !ok {
		return nil, nil
	}
	indexed := false
	tags := strings.Split(tag, ",")
	for _, t := range tags {
		if t == "index" {
			indexed = true
		}
	}
	if !indexed {
		return nil, nil
	}
	// TODO: Add support to flatten struct members.
	if isStruct(sfield.Type) || isStructPtr(sfield.Type) {
		return nil, fmt.Errorf("could not flatten field %s: %w", sfield.Name, os.ErrInvalid)
	}
	ifield := &IndexField{
		name:     sfield.Name,
		position: append([]int{}, sfield.Index...),
	}
	return []*IndexField{ifield}, nil
}

func (f *IndexField) ToString(ovalue reflect.Value) (string, error) {
	if !isStructValue(ovalue) {
		return "", fmt.Errorf("input value must be a struct: %w", os.ErrInvalid)
	}
	fvalue := ovalue.FieldByIndex(f.position)
	if !fvalue.IsValid() {
		return "", fmt.Errorf("couldn't get index field value for %s: %w", f.name, os.ErrInvalid)
	}

	if f.stringer != nil {
		return f.stringer(ovalue)
	}

	if _, ok := supportedKindsMap[fvalue.Kind()]; ok {
		return toStringNative(fvalue.Interface()), nil
	}

	if _, ok := supportedTypesMap[fvalue.Type()]; ok {
		return toStringStandard(fvalue.Interface())
	}

	return FormatIndexFieldValue(fvalue)
}

func toStringNative(v interface{}) string {
	switch x := v.(type) {
	case bool:
		if x {
			return "true"
		} else {
			return "false"
		}
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case string:
		return x
	}
	return "unsupported-index-field-kind"
}

func toStringStandard(v interface{}) (string, error) {
	switch x := v.(type) {
	case []byte:
		return fmt.Sprintf("%x", x), nil
	case net.IP:
		return x.String(), nil
	}
	return "unsupported-index-field-type", os.ErrInvalid
}
