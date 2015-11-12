package rebecca

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/waterlink/rebecca/driver"
	"github.com/waterlink/rebecca/driver/fake"
	"github.com/waterlink/rebecca/field"
)

func TestSaveCreates(t *testing.T) {
	driver.SetupDriver(fake.NewDriver())

	type Person struct {
		ModelMetadata `tablename:"people"`

		ID   int    `rebecca:"id" rebecca_primary:"true"`
		Name string `rebecca:"name"`
		Age  int    `rebecca:"age"`
	}

	expected := &Person{Name: "John Smith", Age: 31}
	if err := Save(expected); err != nil {
		t.Fatal(err)
	}

	actual := &Person{}
	if err := Get(actual, expected.ID); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %+v to equal %+v", actual, expected)
	}
}

func TestSaveUpdates(t *testing.T) {
	driver.SetupDriver(fake.NewDriver())

	type Person struct {
		ModelMetadata `tablename:"people"`

		ID   int    `rebecca:"id" rebecca_primary:"true"`
		Name string `rebecca:"name"`
		Age  int    `rebecca:"age"`
	}

	p := &Person{Name: "John Smith", Age: 31}
	if err := Save(p); err != nil {
		t.Fatal(err)
	}

	expected := &Person{}
	if err := Get(expected, p.ID); err != nil {
		t.Fatal(err)
	}

	expected.Name = "John Smith Jr"
	if err := Save(expected); err != nil {
		t.Fatal(err)
	}

	actual := &Person{}
	if err := Get(actual, p.ID); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %+v to equal %+v", actual, expected)
	}
}

func TestAll(t *testing.T) {
	driver.SetupDriver(fake.NewDriver())

	type Person struct {
		ModelMetadata `tablename:"people"`

		ID   int    `rebecca:"id" rebecca_primary:"true"`
		Name string `rebecca:"name"`
		Age  int    `rebecca:"age"`
	}

	p1 := &Person{Name: "John", Age: 37}
	p2 := &Person{Name: "Sarah", Age: 26}
	p3 := &Person{Name: "James", Age: 33}
	people := []*Person{p1, p2, p3}

	for _, p := range people {
		if err := Save(p); err != nil {
			t.Fatal(err)
		}
	}

	expected := []Person{*p1, *p2, *p3}
	actual := []Person{}
	if err := All(&actual); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v to equal to %+v", actual, expected)
	}
}

func TestWhere(t *testing.T) {
	d := fake.NewDriver()
	driver.SetupDriver(d)

	d.RegisterWhere("age < $1", func(record []field.Field, args ...interface{}) (bool, error) {
		for _, f := range record {
			if f.DriverName == "age" {
				return f.Value.(int) < args[0].(int), nil
			}
		}

		return false, fmt.Errorf("record %+v does not have age field", record)
	})

	d.RegisterWhere("age >= $1", func(record []field.Field, args ...interface{}) (bool, error) {
		for _, f := range record {
			if f.DriverName == "age" {
				return f.Value.(int) >= args[0].(int), nil
			}
		}

		return false, fmt.Errorf("record %+v does not have age field", record)
	})

	type Person struct {
		ModelMetadata `tablename:"people"`

		ID   int    `rebecca:"id" rebecca_primary:"true"`
		Name string `rebecca:"name"`
		Age  int    `rebecca:"age"`
	}

	p1 := &Person{Name: "John", Age: 9}
	p2 := &Person{Name: "Sarah", Age: 27}
	p3 := &Person{Name: "James", Age: 11}
	p4 := &Person{Name: "Monika", Age: 12}
	people := []*Person{p1, p2, p3, p4}

	for _, p := range people {
		if err := Save(p); err != nil {
			t.Fatal(err)
		}
	}

	expected := []Person{*p1, *p3}
	actual := []Person{}
	if err := Where(&actual, "age < $1", 12); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v to equal to %+v", actual, expected)
	}

	expected = []Person{*p2, *p4}
	actual = []Person{}
	if err := Where(&actual, "age >= $1", 12); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v to equal to %+v", actual, expected)
	}
}

func TestFirst(t *testing.T) {
	d := fake.NewDriver()
	driver.SetupDriver(d)

	d.RegisterWhere("age < $1", func(record []field.Field, args ...interface{}) (bool, error) {
		for _, f := range record {
			if f.DriverName == "age" {
				return f.Value.(int) < args[0].(int), nil
			}
		}

		return false, fmt.Errorf("record %+v does not have age field", record)
	})

	d.RegisterWhere("age >= $1", func(record []field.Field, args ...interface{}) (bool, error) {
		for _, f := range record {
			if f.DriverName == "age" {
				return f.Value.(int) >= args[0].(int), nil
			}
		}

		return false, fmt.Errorf("record %+v does not have age field", record)
	})

	type Person struct {
		ModelMetadata `tablename:"people"`

		ID   int    `rebecca:"id" rebecca_primary:"true"`
		Name string `rebecca:"name"`
		Age  int    `rebecca:"age"`
	}

	p1 := &Person{Name: "John", Age: 9}
	p2 := &Person{Name: "Sarah", Age: 27}
	p3 := &Person{Name: "James", Age: 11}
	p4 := &Person{Name: "Monika", Age: 12}
	people := []*Person{p1, p2, p3, p4}

	for _, p := range people {
		if err := Save(p); err != nil {
			t.Fatal(err)
		}
	}

	expected := p1
	actual := &Person{}
	if err := First(&actual, "age < $1", 12); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v to equal to %+v", actual, expected)
	}

	expected = p2
	actual = &Person{}
	if err := First(&actual, "age >= $1", 12); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v to equal to %+v", actual, expected)
	}
}

func TestRemove(t *testing.T) {
	d := fake.NewDriver()
	driver.SetupDriver(d)

	type Person struct {
		ModelMetadata `tablename:"people"`

		ID   int    `rebecca:"id" rebecca_primary:"true"`
		Name string `rebecca:"name"`
		Age  int    `rebecca:"age"`
	}

	p1 := &Person{Name: "John", Age: 9}
	p2 := &Person{Name: "Sarah", Age: 27}
	p3 := &Person{Name: "James", Age: 11}
	p4 := &Person{Name: "Monika", Age: 12}
	people := []*Person{p1, p2, p3, p4}

	for _, p := range people {
		if err := Save(p); err != nil {
			t.Fatal(err)
		}
	}

	if err := Remove(p2); err != nil {
		t.Fatal(err)
	}

	expected := []Person{*p1, *p3, *p4}
	actual := []Person{}
	if err := All(&actual); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v to equal to %+v", actual, expected)
	}
}
