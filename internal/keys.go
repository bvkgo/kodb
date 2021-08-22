package internal

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	ObjectKeyspace = "ob"
	IndexKeyspace  = "ix"
)

// ObjectKey holds the user specified key with the ObjectKeyspace prefix. For
// example, user key "config/cluster" would be represented as:
//
//    /ob/config/cluster
//
type ObjectKey string

// IndexKey holds an object key with IndexKeyspace, a object type, field name
// and field value prefix. For example, object key /ob/a/b/c of data type User
// with indexed field Phone would be represented as below:
//
//     /ix/User/Phone/888-000-1234/ob/a/b/c
//
type IndexKey string

func NewObjectKey(key string) (ObjectKey, error) {
	if !path.IsAbs(key) {
		return "", fmt.Errorf("key must be an absolute path: %w", os.ErrInvalid)
	}
	if s := path.Clean(key); s != key {
		return "", fmt.Errorf("key must be a clean path: %w", os.ErrInvalid)
	}
	return ObjectKey(path.Join("/", ObjectKeyspace, key)), nil
}

func ParseObjectKey(s string) (ObjectKey, error) {
	if !strings.HasPrefix(s, "/"+ObjectKeyspace) {
		return "", fmt.Errorf("not an object key: %w", os.ErrInvalid)
	}
	return ObjectKey(s), nil
}

func (ok ObjectKey) String() string {
	return string(ok)
}

func (ok ObjectKey) UserKey() string {
	return string(ok[len(ObjectKeyspace)+1:])
}

func NewIndexKey(okey ObjectKey, typeName, fieldName, fieldValue string) (IndexKey, error) {
	if len(okey) == 0 {
		return "", fmt.Errorf("object key cannot be empty: %w", os.ErrInvalid)
	}
	if len(typeName) == 0 || len(fieldName) == 0 || len(fieldValue) == 0 {
		return "", fmt.Errorf("type name/field name/field value can't be empty: %w", os.ErrInvalid)
	}
	s := path.Join("/", IndexKeyspace, url.PathEscape(typeName), url.PathEscape(fieldName), url.PathEscape(fieldValue), string(okey))
	return IndexKey(s), nil
}

func ParseIndexKey(s string) (IndexKey, error) {
	keyspacePos := indexRuneN(s, '/', 1)
	typeNamePos := indexRuneN(s, '/', 2)
	fieldNamePos := indexRuneN(s, '/', 3)
	fieldValuePos := indexRuneN(s, '/', 4)
	objectKeyPos := indexRuneN(s, '/', 5)
	if keyspacePos != 0 || typeNamePos == -1 || fieldNamePos == -1 || fieldValuePos == -1 || objectKeyPos == -1 {
		return "", fmt.Errorf("index key format is invalid: %w", os.ErrInvalid)
	}
	keyspace := s[keyspacePos+1 : typeNamePos]
	typeName := s[typeNamePos+1 : fieldNamePos]
	fieldName := s[fieldNamePos+1 : fieldValuePos]
	fieldValue := s[fieldValuePos+1 : objectKeyPos]
	objectKey := s[objectKeyPos:]
	if len(keyspace) == 0 || len(typeName) == 0 || len(fieldName) == 0 || len(fieldValue) == 0 || len(objectKey) == 0 {
		return "", fmt.Errorf("index key format is illegal: %w", os.ErrInvalid)
	}
	if _, err := url.PathUnescape(typeName); err != nil {
		return "", err
	}
	if _, err := url.PathUnescape(fieldName); err != nil {
		return "", err
	}
	if _, err := url.PathUnescape(fieldValue); err != nil {
		return "", err
	}
	if _, err := ParseObjectKey(objectKey); err != nil {
		return "", err
	}
	return IndexKey(s), nil
}

func (ik IndexKey) String() string {
	return string(ik)
}

func (ik IndexKey) GetTypeName() (string, error) {
	s := string(ik)
	p := indexRuneN(s, '/', 2)
	q := indexRuneN(s, '/', 3)
	if p != -1 && q != -1 && p+1 < q {
		return url.PathUnescape(s[p+1 : q])
	}
	return "", fmt.Errorf("index key has no type name: %w", os.ErrInvalid)
}

func (ik IndexKey) GetFieldName() (string, error) {
	s := string(ik)
	p := indexRuneN(s, '/', 3)
	q := indexRuneN(s, '/', 4)
	if p != -1 && q != -1 && p+1 < q {
		return url.PathUnescape(s[p+1 : q])
	}
	return "", fmt.Errorf("index key has no field name: %w", os.ErrInvalid)
}

func (ik IndexKey) GetFieldValue() (string, error) {
	s := string(ik)
	p := indexRuneN(s, '/', 4)
	q := indexRuneN(s, '/', 5)
	if p != -1 && q != -1 && p+1 < q {
		return url.PathUnescape(s[p+1 : q])
	}
	return "", fmt.Errorf("index key has no field value: %w", os.ErrInvalid)
}

func (ik IndexKey) GetObjectKey() (ObjectKey, error) {
	s := string(ik)
	p := indexRuneN(s, '/', 5)
	if p != -1 && p+1 < len(ik) {
		return ObjectKey(s[p:]), nil
	}
	return "", fmt.Errorf("index key has no object key part: %w", os.ErrInvalid)
}

func (ik IndexKey) WithObjectKey(okey ObjectKey) (IndexKey, error) {
	s := string(ik)
	p := indexRuneN(s, '/', 5)
	if p != -1 {
		return IndexKey(s[:p] + string(okey)), nil
	}
	return "", fmt.Errorf("index key has no object key part: %w", os.ErrInvalid)
}

func (ik IndexKey) IndexKeyRange() ([2]string, error) {
	s := string(ik)
	p := indexRuneN(s, '/', 5)
	if p == -1 {
		return [2]string{}, fmt.Errorf("invalid index key: %w", os.ErrInvalid)
	}
	begin := s[:p+1]
	end := s[:p] + string([]byte{'/' + 1})
	return [2]string{begin, end}, nil
}

func SortIndexKeys(iks []IndexKey) {
	sort.Slice(iks, func(i, j int) bool { return iks[i] < iks[j] })
}

func SearchIndexKeys(iks []IndexKey, ik IndexKey) int {
	return sort.Search(len(iks), func(i int) bool {
		return iks[i] >= ik
	})
}
