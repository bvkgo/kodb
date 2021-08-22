# An API for Key-Value Object Database

A Key-Value object database is similar to a Key-Value database, but also allows
storing objects instead of raw bytes. This allows us to "index* objects using
their fields and struct-tags automatically.

API in this module compliments the key-value database API defined in
[github.com/bvkgo/kv](https://pkg.go.dev/github.com//bvkgo/kv) module.
Function names in this API do not conflicts with the key-value database API, so
that a database can implement both APIs simultaneously.

## Type-Safety

To ensure type-safety, API is designed to let user pass a pointer to the object
when retrieving the object from database. Database also keeps track of the a
type name for the object for debugging purposes.

## Keys are Absolute Paths

Keys must be non-empty, absolute and clean paths. This restriction makes it
easier to disambiguate the user-keys internally.

## Object keys and Index keys

Indexing use special keys internally to identify the objects in the
index. These keys are hidden from the user level api.

Backend keyspace is partitioned into object keyspace and index keyspace, with
`/ob/` and `/ix/` key prefixes respectively. This detail would be useful to
know if backend key-value store is scanned independently.

## Indexing with StructTags

Member fields of the objects can be tagged for index with the help of
struct-tags on the fields. Indexed fields identify all objects with the same
non-zero value for the field.

For example, the following data type definition marks the `Age` as an indexed
field:

```go
type User struct {
  Name string

  Age int `kodb:"index"`
}
```

For the above type definition, all user objects stored in the database are also
indexed by their age. Database can be queried for all users with certain age
(eg: 10) using the `FindByIndex` api.

When the age of an user object is modified (eg: say 10 to 11) then database
automatically *moves* the object in the index appropriately (i.e., removes the
user id from index 10 and adds the user id to index 11).

Note that, indexing ignores zero valued fields, so it is not possible to find
all objects with the zero value for a indexed field. With the above example, it
is not possible to find all users with zero age.

## Index Consistency

Data object and it's references from the index should be kept in-sync. Since
these are separate operations, we may face (network) errors in between. This
package guarantees the following:

    If an object with indexed fields is present in the database, it MUST be
    found through the index.

### Object Deletion

When an object is deleted using `Delete` api, index keys are removed *after*
the object is removed. This could leave dangling index keys, so index based
lookups verify that target object is still valid for the index key.

### Object Updates

When an object is updated using `Set` or `Store` apis, new index keys inserted
*before* the object is updated and old index keys are removed *after* the
object is updated. Failures in-between could leave stale index keys, so index
based lookups verify that target object is still valid for a index key.
