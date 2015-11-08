// Package pgdriver provides implementation of rebecca driver for postgres. It
// uses github.com/lib/pq and database/sql under the hood.
package pgdriver

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/lib/pq" // since this driver directly depends on it
	"github.com/waterlink/rebecca/context"
	"github.com/waterlink/rebecca/field"
)

// Driver implements rebecca.Driver interface
type Driver struct {
	db *sql.DB
}

// NewDriver is for constructing correct driver instance
func NewDriver(pgURL string) *Driver {
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		panic(fmt.Errorf("Unable to open connection to postgres database - %s", err))
	}

	return &Driver{db}
}

// Get is for fetching one record given its ID
func (d *Driver) Get(tablename string, fields []field.Field, ID field.Field) ([]field.Field, error) {
	names := fieldNames(fields)

	query := "SELECT %s FROM %s WHERE %s = $1 LIMIT 1"
	query = fmt.Sprintf(query, namesRepr(names), tablename, ID.DriverName)

	return d.readRow(fields, query, ID.Value)
}

// Create is for creating new record and updating its ID
func (d *Driver) Create(tablename string, fields []field.Field, ID *field.Field) error {
	names := fieldNamesWithoutID(fields, *ID)
	values := fieldValuesWithoutID(fields, *ID)

	query := "INSERT INTO %s (%s) VALUES (%s) RETURNING %s"
	query = fmt.Sprintf(query, tablename, namesRepr(names), valuesRepr(values, 0), ID.DriverName)

	idValue := reflect.New(ID.Ty)
	if err := d.db.QueryRow(query, values...).Scan(idValue.Interface()); err != nil {
		return fmt.Errorf("Unable to insert into %s - %s", tablename, err)
	}

	ID.Value = idValue.Elem().Interface()
	return nil
}

// Update is for updating existing record given its ID and fields to update
func (d *Driver) Update(tablename string, fields []field.Field, ID field.Field) error {
	names := fieldNamesWithoutID(fields, ID)
	values := fieldValuesWithoutID(fields, ID)

	query := "UPDATE %s SET (%s) = (%s) WHERE %s = $1"
	query = fmt.Sprintf(query, tablename, namesRepr(names), valuesRepr(values, 1), ID.DriverName)

	args := []interface{}{ID.Value}
	args = append(args, values...)

	if _, err := d.db.Exec(query, args...); err != nil {
		return fmt.Errorf("Unable to update record with primary key = %+v in table %s - %s", ID.Value, tablename, err)
	}

	return nil
}

// All is for fetching all records in current context
func (d *Driver) All(tablename string, fields []field.Field, ctx context.Context) ([][]field.Field, error) {
	names := fieldNames(fields)

	query := "SELECT %s FROM %s %s"
	query = fmt.Sprintf(query, namesRepr(names), tablename, contextFor(ctx))

	return d.readRows(fields, query)
}

// Where is for fetching specific records from current context given where query and arguments
func (d *Driver) Where(tablename string, fields []field.Field, ctx context.Context, where string, args ...interface{}) ([][]field.Field, error) {
	names := fieldNames(fields)

	query := "SELECT %s FROM %s WHERE %s %s"
	query = fmt.Sprintf(query, namesRepr(names), tablename, where, contextFor(ctx))
	return d.readRows(fields, query, args...)
}

// First is for fetching only first specific record from current context matching given where query and arguments
func (d *Driver) First(tablename string, fields []field.Field, ctx context.Context, where string, args ...interface{}) ([]field.Field, error) {
	firstCtx := ctx.SetLimit(1)
	names := fieldNames(fields)

	query := "SELECT %s FROM %s WHERE %s %s"
	query = fmt.Sprintf(query, namesRepr(names), tablename, where, contextFor(firstCtx))
	return d.readRow(fields, query, args...)
}

// Remove is for removing existing record given its ID
func (d *Driver) Remove(tablename string, ID field.Field) error {
	query := "DELETE FROM %s WHERE %s = $1"
	query = fmt.Sprintf(query, tablename, ID.DriverName)

	if _, err := d.db.Exec(query, ID.Value); err != nil {
		return fmt.Errorf("Unable to remove record with primary key = %+v in table %s - %s", ID.Value, tablename, err)
	}

	return nil
}

func (d *Driver) readRow(fields []field.Field, query string, args ...interface{}) ([]field.Field, error) {
	values := newValues(fields)
	if err := d.db.QueryRow(query, args...).Scan(scannableValues(values)...); err != nil {
		return nil, fmt.Errorf("Unable to scan row - query = %s - %s", query, err)
	}

	return recordFromValues(values, fields), nil
}

func (d *Driver) readRows(fields []field.Field, query string, args ...interface{}) ([][]field.Field, error) {
	rows, err := d.db.Query(query, args...)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	if err != nil {
		return nil, fmt.Errorf("Unable to execute query '%s' - %s", query, err)
	}

	result := [][]field.Field{}

	var resultErr error

	for rows.Next() {
		values := newValues(fields)
		if err := rows.Scan(scannableValues(values)...); err != nil {
			resultErr = fmt.Errorf("Unable to scan row - query = %s - %s", query, err)
			continue
		}

		result = append(result, recordFromValues(values, fields))
	}

	return result, resultErr
}

func fieldNames(fields []field.Field) []string {
	names := []string{}
	for _, f := range fields {
		names = append(names, f.DriverName)
	}
	return names
}

func fieldNamesWithoutID(fields []field.Field, ID field.Field) []string {
	names := []string{}

	for _, f := range fields {
		if f.DriverName != ID.DriverName {
			names = append(names, f.DriverName)
		}
	}

	return names
}

func fieldValuesWithoutID(fields []field.Field, ID field.Field) []interface{} {
	values := []interface{}{}

	for _, f := range fields {
		if f.DriverName != ID.DriverName {
			values = append(values, f.Value)
		}
	}

	return values
}

func newValues(fields []field.Field) []reflect.Value {
	values := []reflect.Value{}
	for _, f := range fields {
		values = append(values, reflect.New(f.Ty))
	}
	return values
}

func scannableValues(values []reflect.Value) []interface{} {
	interfaces := []interface{}{}
	for _, v := range values {
		interfaces = append(interfaces, v.Interface())
	}
	return interfaces
}

func namesRepr(names []string) string {
	return strings.Join(names, ", ")
}

func valuesRepr(values []interface{}, offset int) string {
	reprs := []string{}

	for i := range values {
		reprs = append(reprs, "$"+strconv.Itoa(i+offset+1))
	}

	return strings.Join(reprs, ", ")
}

func recordFromValues(values []reflect.Value, fields []field.Field) []field.Field {
	record := []field.Field{}
	for i, f := range fields {
		newField := f
		newField.Value = values[i].Elem().Interface()
		record = append(record, newField)
	}
	return record
}

func contextFor(ctx context.Context) string {
	queryCtx := ""

	if group := ctx.GetGroup(); group != "" {
		queryCtx = queryCtx + fmt.Sprintf(" GROUP BY %s", group)
	}

	if order := ctx.GetOrder(); order != "" {
		queryCtx = queryCtx + fmt.Sprintf(" ORDER BY %s", order)
	}

	if limit := ctx.GetLimit(); limit > 0 {
		queryCtx = queryCtx + fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip := ctx.GetSkip(); skip > 0 {
		queryCtx = queryCtx + fmt.Sprintf(" OFFSET %d", skip)
	}

	return queryCtx
}