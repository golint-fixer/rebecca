# rebecca [![Build Status](https://travis-ci.org/waterlink/rebecca.svg?branch=master)](https://travis-ci.org/waterlink/rebecca)

Simple database convenience wrapper for Go language.

## Documentation

Docs are available on http://godoc.org/github.com/waterlink/rebecca

## Installation

```bash
go get -u github.com/waterlink/rebecca
```

## Usage

```go
import "github.com/waterlink/rebecca"
```

It is recommended to import it as a `bec` shortcut to save some keystrokes:

```go
import bec "github.com/waterlink/rebecca"
```

### Designing record

```go
type Person struct {
        rebecca.ModelMetadata `tablename:"people"`

        ID   int    `rebecca:"id" rebecca_primary:"true"`
        Name string `rebecca:"name"`
        Age  int    `rebecca:"age"`
}
```

### Enabling specific driver

```go
import "github.com/waterlink/rebecca/pgdriver"
```

And in code:

```go
rebecca.SetupDriver(pgdriver.NewDriver("postgres://user:pass@host:port/database?sslmode=sslmode"))
```

The same can be done with empty import:

```go
import _ "github.com/waterlink/rebecca/pgdriver/auto"
```

Empty import will fetch connection options from respective environment
variables:

- `REBECCA_PG_USER`
- `REBECCA_PG_PASS`
- `REBECCA_PG_HOST`
- `REBECCA_PG_PORT`
- `REBECCA_PG_DATABASE`
- `REBECCA_PG_SSLMODE`

or, if present: `REBECCA_PG_URL`

To find out each driver's respective environment variables, see its
documentation.

### List of supported drivers

- `github.com/waterlink/rebecca/pgdriver` - driver for postgresql.
- TODO:
  - `github.com/waterlink/rebecca/cassdriver` - driver for cassandra.
  - `github.com/waterlink/rebecca/mysqldriver` - driver for mysql.
  - `github.com/waterlink/rebecca/mongodriver` - driver for mongodb.

### Fetching the record

```go
var p Person
if err := rebecca.Get(ID, &p); err != nil {
        // handle error here
}

// use &p at this point as a found model instance
```

### Saving record

```go
// creates new record
p := &Person{Name: "John Smith", Age: 31}
if err := rebecca.Save(p); err != nil {
        // handle error here
}

// updates the record
p := &Person{}
if err := rebecca.Get(ID, p); err != nil {
        // handle error here
}

p.Age++
if err := rebecca.Save(p); err != nil {
        // handle error here
}
```

### Fetching all records

```go
people := []Person{}
if err := rebecca.All(&people); err != nil {
        // handle error here
}

// people slice will contain found records
```

### Fetching specific records

```go
kids := []Person{}
if err := rebecca.Where(&kids, "age < $1", 12); err != nil {
        // handle error here
}

// kids slice will contain found records
```

### Fetching only first record

```go
kid := &Person{}
if err := rebecca.First(kid, "age < $1", 12); err != nil {
        // handle error here
}

// kid will contain found first record
```

### Removing record

```go
// Given p is *Person:
if err := rebecca.Remove(p); err != nil {
        // handle error here
}
```

### Fetching count for something

First lets define the view for this purpose:

```go
type PeopleCount struct {
          rebecca.ModelMetadata `tablename:"people"`

          Count int `rebecca:"count(id)"`
}
```

And then, lets query for this count:

```go
kidsCount := &PeopleCount{}
if err := rebecca.First(kidsCount, "age < $1", 12); err != nil {
        // handle error here
}

// Now you can use `kidsCount.Count`
```

### Using order, limit and skip

For example, to fetch second 300 records ordered by age.

For that purpose use `rebecca.Context` struct, documentation on which and all
available options can be found here:
[rebecca.Context](https://godoc.org/github.com/waterlink/rebecca#Context)

```go
kidsBatch := []Person{}
ctx := &rebecca.Context{Order: "age ASC", Limit: 300, Skip: 300}
if err := ctx.Where(&kidsBatch, "age < $1", 12); err != nil {
        // handle error here
}

// Now you can use kidsBatch as a second batch of 300 records ordered by age.
```

Don't confuse `rebecca.Context` with this interface:
[rebecca/context](https://godoc.org/github.com/waterlink/rebecca/context). This
interface is internal and used only by drivers. `rebecca.Context` implements
it. This interface is required to avoid circular dependencies.

### Using aggregation

First, lets define our view for aggregation results:

```go
type PeopleByAge struct {
        rebecca.ModelMetadata `tablename:"people"`

        Age   int `rebecca:"age"`
        Count int `rebecca:"count(distinct(id))"`
}
```

Next, lets query this using the `Context`:

```go
peopleByAge := []PeopleByAge{}
ctx := &rebecca.Context{Group: "age"}
if err := ctx.All(&peopleByAge); err != nil {
        // handle error here
}

// Now peopleByAge represents slice of age => count relationship.
```

### Using transactions

First, obtain `rebecca.Transaction` object by doing:

```go
tx, err := rebecca.Begin()
if err != nil {
        // handle error here
}
defer tx.Rollback()
```

Next, use `tx` as if it was `rebecca` normally:

```go
p := &Person{Name: "John Smith", Age: 29}
if err := tx.Save(p); err != nil {
        // handle error here
}

// use tx.Save, tx.Get, tx.Where, tx.All, tx.First as you would normally do
// with `rebecca`
```

And finally, commit the transaction:

```go
if err := tx.Commit(); err != nil {
        // handle error, if it is impossible to commit the transaction
}
```

Or rollback, if you need to:

```go
tx.Rollback()
```

Don't worry about doing `defer tx.Rollback()` in your functions.
`tx.Rollback()`, when done after commit, is a noop.

### Using Context with Transactions

To use context with transaction, you just need to create context using your
transaction instance:

```go
ctx := tx.Context(&rebecca.Context{Order: "age DESC"})
people := []Person{}
if err := ctx.Where(&people, "age < $1", 23); err != nil {
        // handle error here
}
```

## Development

After cloning and `cd`-ing into this repo, run `go get ./...` to get all
dependencies going.

Next make sure current codebase is in green state with `go test ./...`.

It is encouraged to use TDD, i.e.: first write/change a test for your change,
then make it green by making necessary modifications to the source.

Make sure your editor runs `goimports` tool on save.

When you are done, make sure you have run whole test suite, `golint ./...` and
`go vet ./...`.

## Contributing

1. Fork it (https://github.com/waterlink/rebecca)
1. Clone it (`git clone git@github.com:my-username/rebecca.git`)
1. `cd` to it (`cd rebecca`)
1. Create your feature branch (`git checkout -b my-new-feature`)
1. Commit your changes (`git commit -am 'Add some new feature'`)
1. Push the branch to your fork (`git push -u origin my-new-feature`)
1. Create a new Pull Request on Github

## Contributors

- [waterlink](https://github.com/waterlink) - Oleksii Fedorov, author,
  maintainer
