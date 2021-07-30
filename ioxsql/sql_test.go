package ioxsql_test

import (
	"context"
	"database/sql"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/influxdbiox"
	"github.com/influxdata/influxdbiox/ioxsql"
	_ "github.com/influxdata/influxdbiox/ioxsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openDB(t testing.TB) *sql.DB {
	config, err := influxdbiox.ClientConfigFromJSONString(os.Getenv("IOX_TEST_DATABASE"))
	require.NoError(t, err)
	connector := ioxsql.NewConnector(config)
	return sql.OpenDB(connector)
}

type preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

func prepareStmt(t *testing.T, p preparer, sql string) *sql.Stmt {
	stmt, err := p.Prepare(sql)
	require.NoError(t, err)
	return stmt
}

func closeStmt(t *testing.T, stmt *sql.Stmt) {
	err := stmt.Close()
	require.NoError(t, err)
}

func TestSQLOpen(t *testing.T) {
	db, err := sql.Open("influxdb-iox", os.Getenv("IOX_TEST_DATABASE"))
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestNormalLifeCycle(t *testing.T) {
	db := openDB(t)

	stmt := prepareStmt(t, db, "select 'foo', n from generate_series($1::int, $2::int) n")
	defer closeStmt(t, stmt)

	rows, err := stmt.Query(int32(1), int32(10))
	require.NoError(t, err)

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var s string
		var n int64
		err := rows.Scan(&s, &n)
		require.NoError(t, err)

		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}
		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}
	require.NoError(t, rows.Err())

	require.EqualValues(t, 10, rowCount)

	err = rows.Close()
	require.NoError(t, err)
}

func TestStmtExec(t *testing.T) {
	db := openDB(t)

	tx, err := db.Begin()
	require.NoError(t, err)

	createStmt := prepareStmt(t, tx, "create temporary table t(a varchar not null)")
	_, err = createStmt.Exec()
	require.NoError(t, err)
	closeStmt(t, createStmt)

	insertStmt := prepareStmt(t, tx, "insert into t values($1::text)")
	result, err := insertStmt.Exec("foo")
	require.NoError(t, err)

	n, err := result.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	closeStmt(t, insertStmt)
}

func TestQueryCloseRowsEarly(t *testing.T) {
	db := openDB(t)

	stmt := prepareStmt(t, db, "select 'foo', n from generate_series($1::int, $2::int) n")
	defer closeStmt(t, stmt)

	rows, err := stmt.Query(int32(1), int32(10))
	require.NoError(t, err)

	// Close rows immediately without having read them
	err = rows.Close()
	require.NoError(t, err)

	// Run the query again to ensure the connection and statement are still ok
	rows, err = stmt.Query(int32(1), int32(10))
	require.NoError(t, err)

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var s string
		var n int64
		err := rows.Scan(&s, &n)
		require.NoError(t, err)
		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}
		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}
	require.NoError(t, rows.Err())
	require.EqualValues(t, 10, rowCount)

	err = rows.Close()
	require.NoError(t, err)

}

func TestConnExec(t *testing.T) {
	db := openDB(t)

	_, err := db.Exec("create temporary table t(a varchar not null)")
	require.NoError(t, err)

	result, err := db.Exec("insert into t values('hey')")
	require.NoError(t, err)

	n, err := result.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
}

func TestArgsNotSupported(t *testing.T) {
	db := openDB(t)

	rows, err := db.Query("select v from t where k = $1", "arg")
	if !assert.Error(t, err, "query args not supported") {
		_ = rows.Close()
	}
}

func TestConnQuery(t *testing.T) {
	db := openDB(t)

	rows, err := db.Query("select 'foo', n from generate_series($1::int, $2::int) n", int32(1), int32(10))
	require.NoError(t, err)

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var s string
		var n int64
		err := rows.Scan(&s, &n)
		require.NoError(t, err)
		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}
		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}
	require.NoError(t, rows.Err())
	require.EqualValues(t, 10, rowCount)

	err = rows.Close()
	require.NoError(t, err)
}

func TestConnQueryNull(t *testing.T) {
	db := openDB(t)

	rows, err := db.Query("select $1::int", nil)
	require.NoError(t, err)

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var n sql.NullInt64
		err := rows.Scan(&n)
		require.NoError(t, err)
		if n.Valid != false {
			t.Errorf("Expected n to be null, but it was %v", n)
		}
	}
	require.NoError(t, rows.Err())
	require.EqualValues(t, 1, rowCount)

	err = rows.Close()
	require.NoError(t, err)
}

func TestConnQueryRowByteSlice(t *testing.T) {
	db := openDB(t)

	expected := []byte{222, 173, 190, 239}
	var actual []byte

	err := db.QueryRow(`select E'\\xdeadbeef'::bytea`).Scan(&actual)
	require.NoError(t, err)
	require.EqualValues(t, expected, actual)
}

func TestConnQueryFailure(t *testing.T) {
	db := openDB(t)

	_, err := db.Query("select 'foo")
	require.Error(t, err)
}

func TestConnSimpleSlicePassThrough(t *testing.T) {
	db := openDB(t)

	var n int64
	err := db.QueryRow("select cardinality($1::text[])", []string{"a", "b", "c"}).Scan(&n)
	require.NoError(t, err)
	assert.EqualValues(t, 3, n)
}

// Test type that pgx would handle natively in binary, but since it is not a
// database/sql native type should be passed through as a string
func TestConnQueryRowPgxBinary(t *testing.T) {
	db := openDB(t)

	sql := "select $1::int4[]"
	expected := "{1,2,3}"
	var actual string

	err := db.QueryRow(sql, expected).Scan(&actual)
	require.NoError(t, err)
	require.EqualValues(t, expected, actual)
}

func TestConnQueryRowUnknownType(t *testing.T) {
	db := openDB(t)

	sql := "select $1::point"
	expected := "(1,2)"
	var actual string

	err := db.QueryRow(sql, expected).Scan(&actual)
	require.NoError(t, err)
	require.EqualValues(t, expected, actual)
}

func TestExecNotImplemented(t *testing.T) {

}

func TestTransactionNotImplemented(t *testing.T) {

}

func TestConnRaw(t *testing.T) {
	db := openDB(t)

	conn, err := db.Conn(context.Background())
	require.NoError(t, err)

	err = conn.Raw(func(driverConn interface{}) error {
		client := driverConn.(*ioxsql.Connection).Client()
		return client.Ping(context.Background())
	})
	require.NoError(t, err)
}

func TestConnPingContextSuccess(t *testing.T) {
	db := openDB(t)

	err := db.PingContext(context.Background())
	require.NoError(t, err)
}

func TestConnPrepareContextSuccess(t *testing.T) {
	db := openDB(t)

	stmt, err := db.PrepareContext(context.Background(), "select now()")
	require.NoError(t, err)
	err = stmt.Close()
	require.NoError(t, err)
}

func TestConnQueryContextSuccess(t *testing.T) {
	db := openDB(t)

	rows, err := db.QueryContext(context.Background(), "select * from generate_series(1,10) n")
	require.NoError(t, err)

	for rows.Next() {
		var n int64
		err := rows.Scan(&n)
		require.NoError(t, err)
	}
	require.NoError(t, rows.Err())
}

func TestConnQueryContextFailureRetry(t *testing.T) {
	db := openDB(t)

	// We get a connection, immediately close it, and then get it back;
	// DB.Conn along with Conn.ResetSession does the retry for us.
	{
		conn, err := db.Conn(context.Background())
		require.NoError(t, err)
		err = conn.Raw(func(driverConn interface{}) error {
			client := driverConn.(*ioxsql.Connection).Client()
			return client.Close()
		})
		require.NoError(t, err)
	}
	conn, err := db.Conn(context.Background())
	require.NoError(t, err)

	_, err = conn.QueryContext(context.Background(), "select 1")
	require.NoError(t, err)
}

func TestRowsColumnTypeDatabaseTypeName(t *testing.T) {
	db := openDB(t)

	rows, err := db.Query("select 42::bigint")
	require.NoError(t, err)

	columnTypes, err := rows.ColumnTypes()
	require.NoError(t, err)
	require.Len(t, columnTypes, 1)

	if columnTypes[0].DatabaseTypeName() != "INT8" {
		t.Errorf("columnTypes[0].DatabaseTypeName() => %v, want %v", columnTypes[0].DatabaseTypeName(), "INT8")
	}

	err = rows.Close()
	require.NoError(t, err)
}

func TestStmtExecContextSuccess(t *testing.T) {
	db := openDB(t)

	_, err := db.Exec("create temporary table t(id int primary key)")
	require.NoError(t, err)

	stmt, err := db.Prepare("insert into t(id) values ($1::int4)")
	require.NoError(t, err)
	defer stmt.Close()

	_, err = stmt.ExecContext(context.Background(), 42)
	require.NoError(t, err)
}

func TestStmtExecContextCancel(t *testing.T) {
	db := openDB(t)

	require.NoError(t, db.QueryRow("select 1").Err())

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	require.Error(t, db.QueryRowContext(ctx, "select 1").Err())
	// TODO check error type
}

func TestStmtQueryContextSuccess(t *testing.T) {
	db := openDB(t)

	stmt, err := db.Prepare("select * from generate_series(1,$1::int4) n")
	require.NoError(t, err)
	defer stmt.Close()

	rows, err := stmt.QueryContext(context.Background(), 5)
	require.NoError(t, err)

	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
	}

	if rows.Err() != nil {
		t.Error(rows.Err())
	}
}

func TestRowsColumnTypes(t *testing.T) {
	db := openDB(t)

	columnTypesTests := []struct {
		Name     string
		TypeName string
		Length   struct {
			Len int64
			OK  bool
		}
		DecimalSize struct {
			Precision int64
			Scale     int64
			OK        bool
		}
		ScanType reflect.Type
	}{
		{
			Name:     "a",
			TypeName: "INT8",
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			DecimalSize: struct {
				Precision int64
				Scale     int64
				OK        bool
			}{
				Precision: 0,
				Scale:     0,
				OK:        false,
			},
			ScanType: reflect.TypeOf(int64(0)),
		}, {
			Name:     "bar",
			TypeName: "TEXT",
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: math.MaxInt64,
				OK:  true,
			},
			DecimalSize: struct {
				Precision int64
				Scale     int64
				OK        bool
			}{
				Precision: 0,
				Scale:     0,
				OK:        false,
			},
			ScanType: reflect.TypeOf(""),
		}, {
			Name:     "dec",
			TypeName: "NUMERIC",
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			DecimalSize: struct {
				Precision int64
				Scale     int64
				OK        bool
			}{
				Precision: 9,
				Scale:     2,
				OK:        true,
			},
			ScanType: reflect.TypeOf(float64(0)),
		}, {
			Name:     "d",
			TypeName: "1266",
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			DecimalSize: struct {
				Precision int64
				Scale     int64
				OK        bool
			}{
				Precision: 0,
				Scale:     0,
				OK:        false,
			},
			ScanType: reflect.TypeOf(""),
		},
	}

	rows, err := db.Query("SELECT 1::bigint AS a, text 'bar' AS bar, 1.28::numeric(9, 2) AS dec, '12:00:00'::timetz as d")
	require.NoError(t, err)

	columns, err := rows.ColumnTypes()
	require.NoError(t, err)
	assert.Len(t, columns, 4)

	for i, tt := range columnTypesTests {
		c := columns[i]
		if c.Name() != tt.Name {
			t.Errorf("(%d) got: %s, want: %s", i, c.Name(), tt.Name)
		}
		if c.DatabaseTypeName() != tt.TypeName {
			t.Errorf("(%d) got: %s, want: %s", i, c.DatabaseTypeName(), tt.TypeName)
		}
		l, ok := c.Length()
		if l != tt.Length.Len {
			t.Errorf("(%d) got: %d, want: %d", i, l, tt.Length.Len)
		}
		if ok != tt.Length.OK {
			t.Errorf("(%d) got: %t, want: %t", i, ok, tt.Length.OK)
		}
		p, s, ok := c.DecimalSize()
		if p != tt.DecimalSize.Precision {
			t.Errorf("(%d) got: %d, want: %d", i, p, tt.DecimalSize.Precision)
		}
		if s != tt.DecimalSize.Scale {
			t.Errorf("(%d) got: %d, want: %d", i, s, tt.DecimalSize.Scale)
		}
		if ok != tt.DecimalSize.OK {
			t.Errorf("(%d) got: %t, want: %t", i, ok, tt.DecimalSize.OK)
		}
		if c.ScanType() != tt.ScanType {
			t.Errorf("(%d) got: %v, want: %v", i, c.ScanType(), tt.ScanType)
		}
	}
}

func TestQueryLifeCycle(t *testing.T) {
	db := openDB(t)

	rows, err := db.Query("SELECT 'foo', n FROM generate_series($1::int, $2::int) n WHERE 3 = $3", 1, 10, 3)
	require.NoError(t, err)

	rowCount := int64(0)

	for rows.Next() {
		rowCount++
		var (
			s string
			n int64
		)

		err := rows.Scan(&s, &n)
		require.NoError(t, err)

		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}

		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}
	require.NoError(t, rows.Err())

	err = rows.Close()
	require.NoError(t, err)

	rows, err = db.Query("select 1 where false")
	require.NoError(t, err)

	rowCount = int64(0)

	for rows.Next() {
		rowCount++
	}
	require.NoError(t, rows.Err())
	require.EqualValues(t, 0, rowCount)

	err = rows.Close()
	require.NoError(t, err)
}
