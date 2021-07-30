package ioxsql

import (
	"database/sql/driver"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/flight"
)

var (
	_ driver.Rows                           = (*rows)(nil)
	_ driver.RowsNextResultSet              = (*rows)(nil)
	_ driver.RowsColumnTypeScanType         = (*rows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*rows)(nil)
	_ driver.RowsColumnTypeLength           = (*rows)(nil)
	_ driver.RowsColumnTypeNullable         = (*rows)(nil)
	_ driver.RowsColumnTypePrecisionScale   = (*rows)(nil)
)

type rows struct {
	flightReader *flight.Reader // multiple result sets
	record       array.Record   // current result set
	rowI         int            // next row index
}

func newRows(flightReader *flight.Reader) *rows {
	return &rows{
		flightReader: flightReader,
	}
}

func (r *rows) Columns() []string {
	columns := make([]string, r.record.NumCols())
	for i := int64(0); i < r.record.NumCols(); i++ {
		columns[i] = r.record.ColumnName(int(i))
	}
	return columns
}

func (r *rows) Close() error {
	r.flightReader.Release()
	return nil
}

func (r *rows) Next(dest []driver.Value) error {
	if r.record == nil {
		return io.EOF
	}
	if r.rowI >= int(r.record.NumRows()) {
		r.record.Release()
		r.record = nil
		return io.EOF
	}

	for i := int64(0); i < r.record.NumCols(); i++ {
		col := r.record.Column(int(i))
		value, err := driverValueFromArrowColumn(col, r.rowI)
		if err != nil {
			return err
		}
		dest[i] = value
	}
	r.rowI++
	return nil
}

func driverValueFromArrowColumn(column array.Interface, row int) (driver.Value, error) {
	if column.IsNull(row) {
		return nil, nil
	}
	switch typedColumn := column.(type) {
	case *array.Timestamp:
		return time.Unix(0, int64(typedColumn.Value(row))), nil
	case *array.Float64:
		return typedColumn.Value(row), nil
	case *array.Uint64:
		return typedColumn.Value(row), nil
	case *array.Int64:
		return typedColumn.Value(row), nil
	case *array.String:
		return typedColumn.Value(row), nil
	case *array.Binary:
		return typedColumn.Value(row), nil
	case *array.Boolean:
		return typedColumn.Value(row), nil
	default:
		return nil, fmt.Errorf("unsupported arrow type %q", column.DataType().Name())
	}
}

func (r *rows) HasNextResultSet() bool {
	return r.flightReader.Next()
}

func (r *rows) NextResultSet() error {
	record, err := r.flightReader.Read()
	if err != nil {
		return err
	}
	r.record = record
	r.rowI = 0

	for _, column := range r.record.Columns() {
		column.Data()
	}
	return nil
}

func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	if r.record == nil || index >= int(r.record.NumCols()) {
		return nil
	}
	switch r.record.Column(index).DataType().ID() {
	case arrow.TIMESTAMP:
		return reflect.TypeOf(time.Time{})
	case arrow.FLOAT64:
		return reflect.TypeOf(float64(0))
	case arrow.UINT64:
		return reflect.TypeOf(uint64(0))
	case arrow.INT64:
		return reflect.TypeOf(int64(0))
	case arrow.STRING:
		return reflect.TypeOf("")
	case arrow.BINARY:
		return reflect.TypeOf([]byte(nil))
	case arrow.BOOL:
		return reflect.TypeOf(true)
	default:
		return nil
	}
}

func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	if r.record == nil || index >= int(r.record.NumCols()) {
		return ""
	}
	return r.record.Column(index).DataType().Name()
}

func (r *rows) ColumnTypeLength(index int) (length int64, ok bool) {
	if r.record == nil || index >= int(r.record.NumCols()) {
		return 0, false
	}
	switch r.record.Column(index).DataType().ID() {
	case arrow.TIMESTAMP, arrow.FLOAT64, arrow.UINT64, arrow.INT64, arrow.BOOL:
		return 0, false
	case arrow.STRING, arrow.BINARY:
		return math.MaxInt64, true
	default:
		return 0, false
	}
}

func (r *rows) ColumnTypeNullable(index int) (nullable, ok bool) {
	if r.record == nil || index >= len(r.record.Schema().Fields()) {
		return false, false
	}
	return r.record.Schema().Field(index).Nullable, true
}

func (r *rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return 0, 0, false
}
