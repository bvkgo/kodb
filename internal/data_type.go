package internal

import (
	"fmt"
	"os"
	"reflect"
)

type DataType struct {
	// gotype holds the reflect.Type of reflect.Struct kind for the data type.
	gotype reflect.Type

	// name holds the user chosen name for the data type, which may not match
	// the Golang data type name.
	name string

	// indexFields holds metadata for struct fields that should be indexed
	// automatically.
	indexFields []*IndexField

	cloner func(interface{}) (interface{}, error)

	marshaler func(interface{}) (string, error)

	unmarshaler func(string, interface{}) error
}

func NewDataType(name string, sample interface{}) (*DataType, error) {
	stype, ok := getStructType(sample)
	if !ok {
		return nil, fmt.Errorf("input object must be a struct or pointer to struct: %w", os.ErrInvalid)
	}

	var indexFields []*IndexField
	for i := 0; i < stype.NumField(); i++ {
		sfield := stype.Field(i)
		ifields, err := NewIndexFields(sfield)
		if err != nil {
			return nil, fmt.Errorf("couldn't determine index fields: %w", err)
		}
		indexFields = append(indexFields, ifields...)
	}
	t := &DataType{
		gotype:      stype,
		name:        name,
		indexFields: indexFields,
		marshaler:   gobMarshalString,
		unmarshaler: gobUnmarshalString,
	}
	return t, nil
}

func (t *DataType) goodValue(ob interface{}) (reflect.Value, bool) {
	ovalue, ok := getStructValue(ob)
	if !ok {
		return reflect.Value{}, false
	}
	if ovalue.Type() != t.gotype {
		return reflect.Value{}, false
	}
	if !ovalue.IsValid() {
		return reflect.Value{}, false
	}
	return ovalue, true
}

func (t *DataType) IndexKeyMap(ob interface{}) (map[string]IndexKey, error) {
	ovalue, ok := t.goodValue(ob)
	if !ok {
		return nil, fmt.Errorf("input object is not a struct or pointer to struct of %s type: %w", t.name, os.ErrInvalid)
	}
	okey, err := NewObjectKey("/x")
	if err != nil {
		return nil, err
	}
	ikMap := make(map[string]IndexKey)
	for _, ifield := range t.indexFields {
		fstring, err := ifield.ToString(ovalue)
		if err != nil {
			return nil, err
		}
		// Don't index fields that translate to empty strings.
		if len(fstring) == 0 {
			return nil, nil
		}
		ik, err := NewIndexKey(okey, t.name, ifield.name, fstring)
		if err != nil {
			return nil, err
		}
		ikMap[ifield.name] = ik
	}
	return ikMap, nil
}

func (t *DataType) Marshal(ob interface{}) (string, error) {
	if _, ok := t.goodValue(ob); !ok {
		return "", fmt.Errorf("input object is not a struct or pointer to struct of %s type: %w", t.name, os.ErrInvalid)
	}
	return t.marshaler(ob)
}

func (t *DataType) Unmarshal(s string, ob interface{}) error {
	if _, ok := t.goodValue(ob); !ok {
		return fmt.Errorf("input object is not a struct or pointer to struct of %s type: %w", t.name, os.ErrInvalid)
	}
	return t.unmarshaler(s, ob)
}

func (t *DataType) Clone(ob interface{}) (interface{}, error) {
	if _, ok := t.goodValue(ob); !ok {
		return nil, fmt.Errorf("input object is not a struct or pointer to struct of %s type: %w", t.name, os.ErrInvalid)
	}
	if t.cloner != nil {
		return t.cloner(ob)
	}
	s, err := t.Marshal(ob)
	if err != nil {
		return nil, err
	}
	tmp := reflect.New(t.gotype).Interface()
	if err := t.Unmarshal(s, tmp); err != nil {
		return nil, err
	}
	return tmp, nil
}
