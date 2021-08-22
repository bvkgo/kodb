package internal

import (
	"encoding/gob"
	"fmt"
	"strings"
)

type Value struct {
	// Data holds serialized read-only bytes of the object.
	Data string

	// Type name for the object.
	Type string

	// ObjectKey holds the object-key for the value.
	ObjectKey ObjectKey

	// References keeps track of all index keys that can refer to this object.
	// any value. Index keys are in stored lexical, ascending order.
	IndexKeys []IndexKey
}

func NewValue(okey ObjectKey, object interface{}, datatype *DataType) (*Value, error) {
	s, err := datatype.Marshal(object)
	if err != nil {
		return nil, err
	}
	ikMap, err := datatype.IndexKeyMap(object)
	if err != nil {
		return nil, err
	}
	var iks []IndexKey
	for _, ik := range ikMap {
		x, err := ik.WithObjectKey(okey)
		if err != nil {
			return nil, err
		}
		iks = append(iks, x)
	}
	SortIndexKeys(iks)
	v := &Value{
		Data:      s,
		Type:      datatype.name,
		ObjectKey: okey,
		IndexKeys: iks,
	}
	return v, nil
}

func NewStringValue(okey ObjectKey, value string) *Value {
	return &Value{Data: value, Type: "string", ObjectKey: okey}
}

func ParseValue(s string) (*Value, error) {
	v := new(Value)
	if err := gob.NewDecoder(strings.NewReader(s)).Decode(v); err != nil {
		return nil, err
	}
	if _, err := ParseObjectKey(string(v.ObjectKey)); err != nil {
		return nil, fmt.Errorf("no object key reference: %w", err)
	}
	for _, x := range v.IndexKeys {
		if _, err := ParseIndexKey(string(x)); err != nil {
			return nil, fmt.Errorf("bad index keys:%w", err)
		}
	}
	return v, nil
}

// TODO: Allow for json encoding for easier inspection.
func (v *Value) String() string {
	var sb strings.Builder
	if err := gob.NewEncoder(&sb).Encode(v); err != nil {
		panic("unexpected gob encode failure")
	}
	return sb.String()
}

func (v *Value) HasAllIndexKeys(iks []IndexKey) bool {
	for _, ik := range iks {
		i := SearchIndexKeys(v.IndexKeys, ik)
		if i < len(v.IndexKeys) && v.IndexKeys[i] == ik {
			continue
		}
		return false
	}
	return true
}
