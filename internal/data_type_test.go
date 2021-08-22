package internal

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

type ExampleType struct {
	Bool bool `kodb:"index"`

	Int   int   `kodb:"index"`
	Int8  int8  `kodb:"index"`
	Int16 int16 `kodb:"index"`
	Int32 int32 `kodb:"index"`
	Int64 int64 `kodb:"index"`

	Uint   uint   `kodb:"index"`
	Uint8  uint8  `kodb:"index"`
	Uint16 uint16 `kodb:"index"`
	Uint32 uint32 `kodb:"index"`
	Uint64 uint64 `kodb:"index"`

	String string `kodb:"index"`

	Bytes []byte `kodb:"index"`
	IP    net.IP `kodb:"index"`

	// Time  time.Time `kodb:"index"` // TODO Needs more work and also not supported by quick.Value
}

func TestDataType(t *testing.T) {
	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))

	if err := Register("ExampleType", ExampleType{}); err != nil {
		if !errors.Is(err, os.ErrExist) {
			t.Fatal(err)
		}
	}

	datatype, err := GetDataType(new(ExampleType))
	if err != nil {
		t.Fatal(err)
	}

	rv, ok := quick.Value(reflect.TypeOf(ExampleType{}), random)
	if !ok {
		t.Fatalf("could not create aa randomized value")
	}

	if s, err := datatype.Marshal(rv.Interface()); err != nil {
		t.Fatal(err)
	} else if err := datatype.Unmarshal(s, new(ExampleType)); err != nil {
		t.Fatal(err)
	} else if _, err := datatype.Clone(rv.Interface()); err != nil {
		t.Fatal(err)
	}

	if _, err := datatype.IndexKeyMap(rv.Interface()); err != nil {
		t.Fatal(err)
	}
}
