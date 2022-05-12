package ioxsql

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/apache/arrow/go/v8/arrow"
	"github.com/apache/arrow/go/v8/arrow/array"
	"github.com/apache/arrow/go/v8/arrow/flight"
	influxdbiox "github.com/influxdata/influxdb-iox-client-go"
)

var (
	_ driver.Rows                           = (*rows)(nil)
	_ driver.RowsColumnTypeScanType         = (*rows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*rows)(nil)
	_ driver.RowsColumnTypeLength           = (*rows)(nil)
	_ driver.RowsColumnTypeNullable         = (*rows)(nil)
	_ driver.RowsColumnTypePrecisionScale   = (*rows)(nil)
)

type rows struct {
	flightReader *flight.Reader // stream of result sets
	fields       []arrow.Field
	record       arrow.Record // current result set
	rowI         int          // next row index for current result set
}

// queryRows constructs a new rows object by executing a query request
func queryRows(ctx context.Context, request *influxdbiox.QueryRequest, _argsReserved []interface{}) (*rows, error) {
	flightReader, err := request.Query(ctx) // n.b. this must be released
	if err != nil {
		return nil, err
	}

	return &rows{
		flightReader: flightReader,
		fields:       flightReader.Schema().Fields(),
	}, nil
}

// Close ensures that releasable pointers are released and set to nil.
func (r *rows) Close() error {
	if r.record != nil {
		r.record = nil
	}
	if r.flightReader != nil {
		r.flightReader.Release()
		r.flightReader = nil
	}
	return nil
}

func (r *rows) Columns() []string {
	if r.flightReader == nil {
		return nil
	}

	columns := make([]string, len(r.fields))
	for i, field := range r.fields {
		columns[i] = field.Name
	}
	return columns
}

func (r *rows) Next(dest []driver.Value) error {
	for r.record == nil || r.rowI >= int(r.record.NumRows()) {
		if nextRecord, err := r.flightReader.Read(); err == io.EOF {
			r.record = nil
			_ = r.Close()
			return io.EOF
		} else if err != nil {
			_ = r.Close()
			return err
		} else {
			r.record = nextRecord
			r.rowI = 0
		}
	}

	for i := 0; i < int(r.record.NumCols()); i++ {
		col := r.record.Column(i)
		value, err := driverValueFromArrowColumn(col, r.rowI)
		if err != nil {
			_ = r.Close()
			return err
		}
		dest[i] = value
	}

	r.rowI++
	return nil
}

func driverValueFromArrowColumn(column arrow.Array, row int) (driver.Value, error) {
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

func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	if index >= len(r.fields) {
		return nil
	}
	switch r.fields[index].Type.ID() {
	case arrow.TIMESTAMP:
		return reflect.TypeOf(time.Time{})
	case arrow.FLOAT32:
		return reflect.TypeOf(float32(0))
	case arrow.DECIMAL, arrow.FLOAT64:
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
	if index >= len(r.fields) {
		return ""
	}
	return r.fields[index].Type.ID().String()
}

func (r *rows) ColumnTypeLength(index int) (length int64, ok bool) {
	if index >= len(r.fields) {
		return 0, false
	}
	switch r.fields[index].Type.ID() {
	case arrow.TIMESTAMP, arrow.FLOAT64, arrow.UINT64, arrow.INT64, arrow.BOOL:
		return 0, false
	case arrow.STRING, arrow.BINARY:
		return math.MaxInt64, true
	default:
		return 0, false
	}
}

func (r *rows) ColumnTypeNullable(index int) (nullable, ok bool) {
	if index >= len(r.fields) {
		return false, false
	}
	return r.fields[index].Nullable, true
}

func (r *rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return 0, 0, false
}
