package kodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bvkgo/kodb/internal"
	"github.com/bvkgo/kv"
)

type NewTx = func(context.Context) (kv.Transaction, error)
type NewIt = func(context.Context) (kv.Iterator, error)

type DB struct {
	newTx NewTx
	newIt NewIt
}

type Tx struct {
	db *DB
	tx kv.Transaction
}

type Iter struct {
	tx   *Tx
	next int
	keys []internal.ObjectKey
	refs [][]internal.IndexKey
}

// New creates a key-object database out of a key-value database.
func New(ntx NewTx, nit NewIt) *DB {
	return &DB{newTx: ntx, newIt: nit}
}

// NewTx creates a new transaction.
func (d *DB) NewTx(ctx context.Context) (*Tx, error) {
	tx, err := d.newTx(ctx)
	if err != nil {
		return nil, err
	}
	t := &Tx{
		db: d,
		tx: tx,
	}
	return t, nil
}

// Commit commits all changes made by the transaction.
func (t *Tx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback drops all changes made by the transaction.
func (t *Tx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// Get returns the value stored at the given key. If an object was previously
// stored at the key, it serialized bytes will be returned.
func (t *Tx) Get(ctx context.Context, key string) (string, error) {
	okey, err := internal.NewObjectKey(key)
	if err != nil {
		return "", err
	}
	s, err := t.tx.Get(ctx, okey.String())
	if err != nil {
		return "", err
	}
	v, err := internal.ParseValue(s)
	if err != nil {
		return "", err
	}
	return v.Data, nil
}

// Set updates the data stored at the given key to the value.
func (t *Tx) Set(ctx context.Context, key, value string) error {
	okey, err := internal.NewObjectKey(key)
	if err != nil {
		return err
	}
	v := internal.NewStringValue(okey, value)
	// NOTE: We don't bother to erase index keys referring to previous value
	// cause stale index key references are checked when dereferenced.
	return t.tx.Set(ctx, okey.String(), v.String())
}

// Delete removes data or object stored at the given key.
func (t *Tx) Delete(ctx context.Context, key string) error {
	okey, err := internal.NewObjectKey(key)
	if err != nil {
		return err
	}
	s, err := t.tx.Get(ctx, okey.String())
	if err != nil {
		return err
	}
	v, err := internal.ParseValue(s)
	if err != nil {
		return err
	}
	if err := t.tx.Delete(ctx, okey.String()); err != nil {
		return err
	}
	for _, k := range v.IndexKeys {
		if err := t.tx.Delete(ctx, k.String()); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				panic("index consistency bug")
			}
			return err
		}
	}
	return nil
}

// Load reads and unmarshals the value stored at the given key into the object
// pointer.
func (t *Tx) Load(ctx context.Context, key string, ob interface{}) error {
	okey, err := internal.NewObjectKey(key)
	if err != nil {
		return err
	}
	datatype, err := internal.GetDataType(ob)
	if err != nil {
		return err
	}
	s, err := t.tx.Get(ctx, okey.String())
	if err != nil {
		return err
	}
	v, err := internal.ParseValue(s)
	if err != nil {
		return err
	}
	if err := datatype.Unmarshal(v.Data, ob); err != nil {
		return err
	}
	return nil
}

// Store saves the input object at the given key. Index is updated to reflect
// the new indexed field values if any.
func (t *Tx) Store(ctx context.Context, key string, ob interface{}) error {
	okey, err := internal.NewObjectKey(key)
	if err != nil {
		return err
	}
	datatype, err := internal.GetDataType(ob)
	if err != nil {
		return err
	}
	s, err := t.tx.Get(ctx, okey.String())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	old := new(internal.Value)
	if len(s) > 0 {
		v, err := internal.ParseValue(s)
		if err != nil {
			return fmt.Errorf("could not find indexed keys for old instance: %w", err)
		}
		old = v
	}
	cur, err := internal.NewValue(okey, ob, datatype)
	if err != nil {
		return err
	}
	// We must first add new index keys followed by insert/replace the object and
	// only then should remove the stale index keys. A failure may leave
	// stale/wrong index keys, but it is handled when read through the
	// iterator. See the indexing guarantees in the README file.
	deletions, additions := internal.DiffIndexKeys(old.IndexKeys, cur.IndexKeys)
	for _, i := range additions {
		if err := t.tx.Set(ctx, i.String(), ""); err != nil {
			return err
		}
		log.Printf("adding %s with index key %s", key, i)
	}
	if err := t.tx.Set(ctx, okey.String(), cur.String()); err != nil {
		return err
	}
	for _, d := range deletions {
		if err := t.tx.Delete(ctx, d.String()); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				panic("index consistency bug")
			}
			return err
		}
	}
	return nil
}

// FindByIndex scans the database index for objects with indexed field values
// matching the input object.
func (t *Tx) FindByIndex(ctx context.Context, part interface{}, iterator Iterator) error {
	iter, ok := iterator.(*Iter)
	if !ok {
		return os.ErrInvalid
	}

	datatype, err := internal.GetDataType(part)
	if err != nil {
		return err
	}
	ikMap, err := datatype.IndexKeyMap(part)
	if err != nil {
		return err
	}

	var refs []string
	for _, ik := range ikMap {
		r, err := ik.IndexKeyRange()
		if err != nil {
			return err
		}
		it, err := t.db.newIt(ctx)
		if err != nil {
			return err
		}
		if err := t.tx.Ascend(ctx, r[0], r[1], it); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			continue
		}

		for k, _, err := it.GetNext(ctx); true; k, _, err = it.GetNext(ctx) {
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return err
				}
				break
			}
			refs = append(refs, k)
		}
	}

	// Dedup the object keys out of index keys.
	refMap := map[internal.ObjectKey][]internal.IndexKey{}
	for _, ref := range refs {
		ik, err := internal.ParseIndexKey(ref)
		if err != nil {
			return fmt.Errorf("unexpected index key failure: %w", err)
		}
		ok, err := ik.GetObjectKey()
		if err != nil {
			return fmt.Errorf("index key with invalid object key: %w", err)
		}
		refMap[ok] = append(refMap[ok], ik)
	}

	oks := make([]internal.ObjectKey, 0, len(refMap))
	iks := make([][]internal.IndexKey, 0, len(refMap))
	for k, v := range refMap {
		oks = append(oks, k)
		iks = append(iks, v)
	}

	iter.tx = t
	iter.next = 0
	iter.keys = oks
	iter.refs = iks
	return nil
}

// GetNext returns the value at the iterator in the serialized form.
func (it *Iter) GetNext(ctx context.Context) (string, string, error) {
	for ; it.next < len(it.keys); it.next++ {
		k := it.keys[it.next]
		s, err := it.tx.tx.Get(ctx, k.String())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", "", err
		}
		v, err := internal.ParseValue(s)
		if err != nil {
			return "", "", err
		}
		// There can be index key with stale object key references due to errors
		// (when using txes after a non-nil error). So, validate with the target
		// object metadata.  See the indexing guarantees in README file.
		if v.ObjectKey != k {
			continue
		}
		if len(it.refs) > 0 {
			if !v.HasAllIndexKeys(it.refs[it.next]) {
				continue
			}
		}
		it.next++
		return k.UserKey(), v.Data, nil
	}
	return "", "", os.ErrNotExist
}

// LoadNext reads current value at the iterator and also advances the iterator
// to the next object.
func (it *Iter) LoadNext(ctx context.Context, key *string, ob interface{}) error {
	datatype, err := internal.GetDataType(ob)
	if err != nil {
		return err
	}

	for ; it.next < len(it.keys); it.next++ {
		k := it.keys[it.next]
		s, err := it.tx.tx.Get(ctx, k.String())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
		v, err := internal.ParseValue(s)
		if err != nil {
			return err
		}
		// There can be index key with stale object key references due to errors
		// (when using txes after a non-nil error). So, validate with the target
		// object metadata.  See the indexing guarantees in README file.
		if v.ObjectKey != k {
			continue
		}
		if len(it.refs) > 0 {
			if !v.HasAllIndexKeys(it.refs[it.next]) {
				continue
			}
		}
		if err := datatype.Unmarshal(v.Data, ob); err != nil {
			return err
		}
		if key != nil {
			*key = k.UserKey()
		}
		it.next++
		return nil
	}
	return os.ErrNotExist
}
