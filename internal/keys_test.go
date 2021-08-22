package internal

import (
	"fmt"
	"testing"
)

func TestObjectKey(t *testing.T) {
	okey, err := NewObjectKey("/foo")
	if err != nil {
		t.Fatal(err)
	}
	if s := okey.String(); s != "/ob/foo" {
		t.Fatalf("want /ob/foo got %s", s)
	}
	if k := okey.UserKey(); k != "/foo" {
		t.Fatalf("want /foo got %s", k)
	}
	if _, err := ParseObjectKey(okey.String()); err != nil {
		t.Fatal(err)
	}
}

func TestIndexKey(t *testing.T) {
	okey, err := NewObjectKey("/x")
	if err != nil {
		t.Fatal(err)
	}
	ikey, err := NewIndexKey(okey, "X", "Field", "Value")
	if err != nil {
		t.Fatal(err)
	}
	if s := ikey.String(); s != "/ix/X/Field/Value/ob/x" {
		t.Fatalf("index key %s is in unexpected format", s)
	}
	if _, err := ParseIndexKey(ikey.String()); err != nil {
		t.Fatal(err)
	}
	if v, err := ikey.GetTypeName(); err != nil {
		t.Fatal(err)
	} else if v != "X" {
		t.Fatalf("want X got %s", v)
	}
	if v, err := ikey.GetFieldName(); err != nil {
		t.Fatal(err)
	} else if v != "Field" {
		t.Fatalf("want Field got %s", v)
	}
	if v, err := ikey.GetFieldValue(); err != nil {
		t.Fatal(err)
	} else if v != "Value" {
		t.Fatalf("want Value got %s", v)
	}
	if v, err := ikey.GetObjectKey(); err != nil {
		t.Fatal(err)
	} else if v != okey {
		t.Fatalf("want %s got %s", okey, v)
	}
	okey2, err := NewObjectKey("/y")
	if err != nil {
		t.Fatal(err)
	}
	ikey, err = ikey.WithObjectKey(okey2)
	if err != nil {
		t.Fatal(err)
	}
	if s := ikey.String(); s != "/ix/X/Field/Value/ob/y" {
		t.Fatalf("object key must now be different")
	}
	if _, err := ParseIndexKey(ikey.String()); err != nil {
		t.Fatal(err)
	}
	if r, err := ikey.IndexKeyRange(); err != nil {
		t.Fatal(err)
	} else if r[0] != "/ix/X/Field/Value/" {
		t.Fatalf("index key range start key %q is unexpected", r[0])
	} else if r[1] != "/ix/X/Field/Value"+fmt.Sprintf("%c", '/'+1) {
		t.Fatalf("index key range end key %q is unexpected", r[1])
	}
}

func TestIndexKeyEscapes(t *testing.T) {
	okey, err := NewObjectKey("/a/b/c")
	if err != nil {
		t.Fatal(err)
	}
	ikey, err := NewIndexKey(okey, "Type/Name", "Field/Name", "Field/Value")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseIndexKey(ikey.String()); err != nil {
		t.Fatal(err)
	}
	if v, err := ikey.GetTypeName(); err != nil {
		t.Fatal(err)
	} else if v != "Type/Name" {
		t.Fatalf("want X got %s", v)
	}
	if v, err := ikey.GetFieldName(); err != nil {
		t.Fatal(err)
	} else if v != "Field/Name" {
		t.Fatalf("want Field got %s", v)
	}
	if v, err := ikey.GetFieldValue(); err != nil {
		t.Fatal(err)
	} else if v != "Field/Value" {
		t.Fatalf("want Value got %s", v)
	}
	if v, err := ikey.GetObjectKey(); err != nil {
		t.Fatal(err)
	} else if v != okey {
		t.Fatalf("want %s got %s", okey, v)
	}
}
