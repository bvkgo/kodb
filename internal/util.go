package internal

import (
	"encoding/gob"
	"reflect"
	"strings"
)

// DiffIndexKeys compares two set of index-keys of an object and returns
// necessary deletions and additions to the secondary index.
func DiffIndexKeys(old, cur []IndexKey) (deletions []IndexKey, additions []IndexKey) {
	for _, k := range cur {
		if i := SearchIndexKeys(old, k); i < len(old) && old[i] == k {
			continue
		}
		additions = append(additions, k)
	}
	for _, k := range old {
		if i := SearchIndexKeys(cur, k); i < len(cur) && cur[i] == k {
			continue
		}
		deletions = append(deletions, k)
	}
	return
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func isStructValue(v reflect.Value) bool {
	return v.Kind() == reflect.Struct
}

func isStructPtr(t reflect.Type) bool {
	if k := t.Kind(); k == reflect.Ptr {
		return t.Elem().Kind() == reflect.Struct
	}
	return false
}

func getStructType(ob interface{}) (reflect.Type, bool) {
	otype := reflect.TypeOf(ob)
	if isStruct(otype) {
		return otype, true
	}
	if isStructPtr(otype) {
		return otype.Elem(), true
	}
	return nil, false
}

func getStructValue(ob interface{}) (reflect.Value, bool) {
	otype := reflect.TypeOf(ob)
	ovalue := reflect.ValueOf(ob)
	if isStruct(otype) {
		return ovalue, true
	}
	if isStructPtr(otype) {
		if ovalue.IsNil() {
			return ovalue, true
		}
		return ovalue.Elem(), true
	}
	return ovalue, false
}

func SliceContainsTrimFold(vs []string, k string) bool {
	for _, v := range vs {
		if strings.EqualFold(strings.TrimSpace(v), k) {
			return true
		}
	}
	return false
}

func gobMarshalString(v interface{}) (string, error) {
	var sb strings.Builder
	if err := gob.NewEncoder(&sb).Encode(v); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func gobUnmarshalString(s string, v interface{}) error {
	if err := gob.NewDecoder(strings.NewReader(s)).Decode(v); err != nil {
		return err
	}
	return nil
}

func indexRuneN(s string, r rune, n int) int {
	offset := 0
	for i := 0; i < n; i++ {
		if p := strings.IndexRune(s[offset:], r); p == -1 {
			return p
		} else {
			offset += p + 1
		}
	}
	return offset - 1
}
