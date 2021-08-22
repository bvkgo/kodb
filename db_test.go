package kodb

import (
	"context"
	"errors"
	"os"
	"path"
	"testing"

	"github.com/bvkgo/kodb/internal"
	"github.com/bvkgo/kv"
	"github.com/bvkgo/kvmemdb"
)

func TestIndexing(t *testing.T) {
	ctx := context.Background()

	type ExampleType struct {
		Name string
		Age  int `kodb:"index"`
	}

	findUsersByAge := func(tx *Tx, age int) []string {
		var it Iter
		if err := tx.FindByIndex(ctx, ExampleType{Age: age}, &it); err != nil {
			t.Fatal(err)
		}
		var matched []string
		var user ExampleType
		for err := it.LoadNext(ctx, nil /* key */, &user); err == nil; err = it.LoadNext(ctx, nil /* key */, &user) {
			matched = append(matched, user.Name)
		}
		return matched
	}

	if err := internal.Register("TestIndexing.ExampleType", ExampleType{}); err != nil {
		if !errors.Is(err, os.ErrExist) {
			t.Fatal(err)
		}
	}

	var kvdb kvmemdb.DB
	newTx := func(context.Context) (kv.Transaction, error) { return kvdb.NewTx(), nil }
	newIt := func(context.Context) (kv.Iterator, error) { return new(kvmemdb.Iter), nil }
	db := New(newTx, newIt)

	alex := &ExampleType{Name: "alex", Age: 10}
	ben := &ExampleType{Name: "ben", Age: 20}
	carter := &ExampleType{Name: "carter", Age: 30}
	dave := &ExampleType{Name: "dave", Age: 10}
	ethan := &ExampleType{Name: "ethan", Age: 20}
	users := []*ExampleType{alex, ben, carter, dave, ethan}

	{
		t1, err := db.NewTx(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, u := range users {
			if err := t1.Store(ctx, path.Join("/users", u.Name), u); err != nil {
				t.Fatal(err)
			}
		}
		if err := t1.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Check that users are as expected.
	{
		t2, err := db.NewTx(ctx)
		if err != nil {
			t.Fatal(err)
		}

		at10 := findUsersByAge(t2, 10)
		if len(at10) != 2 || !(internal.SliceContainsTrimFold(at10, "alex") && internal.SliceContainsTrimFold(at10, "dave")) {
			t.Fatalf("alex and dave must be in 10y index")
		}
		at20 := findUsersByAge(t2, 20)
		if len(at20) != 2 || !(internal.SliceContainsTrimFold(at20, "ben") && internal.SliceContainsTrimFold(at20, "ethan")) {
			t.Fatalf("ben and ethan must be in 20y index")
		}

		if err := t2.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Updating an user should update the index, but rollback should undo it.
	{
		t3, err := db.NewTx(ctx)
		if err != nil {
			t.Fatal(err)
		}
		alex.Age = 20
		if err := t3.Store(ctx, path.Join("/users", alex.Name), alex); err != nil {
			t.Fatal(err)
		}

		at10 := findUsersByAge(t3, 10)
		if len(at10) != 1 || !internal.SliceContainsTrimFold(at10, "dave") {
			t.Fatalf("only dave must be in 10y index")
		}
		at20 := findUsersByAge(t3, 20)
		if len(at20) != 3 || !(internal.SliceContainsTrimFold(at20, "alex") && internal.SliceContainsTrimFold(at20, "ben") && internal.SliceContainsTrimFold(at20, "ethan")) {
			t.Fatalf("alex, ben and ethan must be in 20y index")
		}

		if err := t3.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Check that users are as expected.
	{
		t4, err := db.NewTx(ctx)
		if err != nil {
			t.Fatal(err)
		}

		at10 := findUsersByAge(t4, 10)
		if len(at10) != 2 || !(internal.SliceContainsTrimFold(at10, "alex") && internal.SliceContainsTrimFold(at10, "dave")) {
			t.Fatalf("alex and dave must be in 10y index")
		}
		at20 := findUsersByAge(t4, 20)
		if len(at20) != 2 || !(internal.SliceContainsTrimFold(at20, "ben") && internal.SliceContainsTrimFold(at20, "ethan")) {
			t.Fatalf("ben and ethan must be in 20y index")
		}

		if err := t4.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
}
