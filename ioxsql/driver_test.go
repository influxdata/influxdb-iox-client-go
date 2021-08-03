package ioxsql_test

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/apache/arrow/go/arrow"
	"github.com/influxdata/influxdb-iox-client-go"
	"github.com/influxdata/influxdb-iox-client-go/ioxsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func databaseAddress() string {
	address, found := os.LookupEnv("INFLUXDB_IOX_ADDRESS")
	if !found {
		address = "localhost:8082"
	}
	return address
}

func openNewDatabase(t *testing.T) (*sql.DB, *influxdbiox.Client) {
	address := databaseAddress()
	database := fmt.Sprintf("test-%d", time.Now().UnixNano())
	if testing.Verbose() {
		t.Logf("temporary database: %q", database)
	}
	dsn := fmt.Sprintf("%s/%s", address, database)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	config, err := influxdbiox.ClientConfigFromAddressString(dsn)
	require.NoError(t, err)
	client, err := influxdbiox.NewClient(ctx, config)
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	require.NoError(t, client.CreateDatabase(ctx, database))

	sqlDB, err := sql.Open(ioxsql.DriverName, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	return sqlDB, client
}

func writeDataset(t *testing.T, client *influxdbiox.Client) {
	batch, err := client.NewWriteBatch("")
	require.NoError(t, err)

	table, err := batch.Table("t")
	require.NoError(t, err)

	baseTime := time.Date(2021, time.April, 15, 0, 0, 0, 0, time.UTC)
	tags := map[string]string{
		"foo": "bar",
	}

	for i := 0; i < 10; i++ {
		ts := baseTime.Add(time.Minute * time.Duration(i))
		fields := map[string]interface{}{
			"v": int64(i),
		}
		require.NoError(t, table.AddLineProtocolPoint(ts, tags, fields))
	}

	require.NoError(t, batch.Write(context.Background()))
}

func prepareStmt(t *testing.T, db *sql.DB, query string) *sql.Stmt {
	stmt, err := db.Prepare(query)
	require.NoError(t, err)
	t.Cleanup(func() { stmt.Close() })
	return stmt
}

func queryStmt(t *testing.T, stmt *sql.Stmt, args ...interface{}) *sql.Rows {
	rows, err := stmt.Query(args...)
	require.NoError(t, err)
	t.Cleanup(func() { rows.Close() })
	return rows
}

func TestSQLOpen(t *testing.T) {
	db, err := sql.Open(ioxsql.DriverName, databaseAddress())
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestNormalLifeCycle(t *testing.T) {
	db, client := openNewDatabase(t)
	writeDataset(t, client)

	stmt := prepareStmt(t, db, "select foo, v from t ORDER BY v ASC")
	rows := queryStmt(t, stmt)

	rowCount := 0

	for rows.Next() {
		var s string
		var n int64
		require.NoError(t, rows.Scan(&s, &n))
		assert.Equal(t, "bar", s)
		assert.EqualValues(t, rowCount, n)
		rowCount++
	}
	require.NoError(t, rows.Err())

	assert.EqualValues(t, 10, rowCount)

	require.NoError(t, rows.Close())
	require.NoError(t, stmt.Close())
}

func TestTransactionsNotSupported(t *testing.T) {
	db, _ := openNewDatabase(t)

	_, err := db.Begin()
	require.EqualError(t, err, "transactions not supported")
}

func TestQueryCloseRowsEarly(t *testing.T) {
	db, client := openNewDatabase(t)
	writeDataset(t, client)

	stmt := prepareStmt(t, db, "select foo, v from t ORDER BY v ASC")
	rows := queryStmt(t, stmt)

	// Close rows immediately without having read them
	require.NoError(t, rows.Close())

	// Run the query again to ensure the connection and statement are still ok
	rows = queryStmt(t, stmt)

	rowCount := 0

	for rows.Next() {
		var s string
		var n int64
		require.NoError(t, rows.Scan(&s, &n))
		assert.Equal(t, "bar", s)
		assert.EqualValues(t, rowCount, n)
		rowCount++
	}
	require.NoError(t, rows.Err())

	assert.EqualValues(t, 10, rowCount)

	require.NoError(t, rows.Close())
	require.NoError(t, stmt.Close())
}

func TestExecNotSupported(t *testing.T) {
	db, _ := openNewDatabase(t)
	_, err := db.Exec("create table t(a varchar not null)")
	require.EqualError(t, err, "exec not implemented")
}

func TestArgsNotSupported(t *testing.T) {
	db, _ := openNewDatabase(t)

	_, err := db.Query("select v from t where k = $1", "arg")
	assert.EqualError(t, err, "query args not supported")

	_, err = db.Query("select v from t where k = ?", "arg")
	assert.EqualError(t, err, "query args not supported")
}

func TestConnQueryNull(t *testing.T) {
	t.Skip("IOx/CF/Arrow bug in null handling")

	db, client := openNewDatabase(t)
	wb, err := client.NewWriteBatch("")
	require.NoError(t, err)

	table, err := wb.Table("t")
	require.NoError(t, err)

	baseTime := time.Date(2021, time.April, 15, 0, 0, 0, 0, time.UTC)

	require.NoError(t, table.AddLineProtocolPoint(baseTime,
		map[string]string{"foo": "bar"},
		map[string]interface{}{"v": int64(0)},
	))
	require.NoError(t, table.AddLineProtocolPoint(baseTime.Add(time.Minute),
		nil,
		map[string]interface{}{"v": int64(1)},
	))

	require.NoError(t, wb.Write(context.Background()))

	row := db.QueryRow("select foo, v from t where v = 1")
	require.NoError(t, row.Err())

	var gotFoo sql.NullString
	var gotV sql.NullInt64
	require.NoError(t, row.Scan(&gotFoo, &gotV))

	assert.False(t, gotFoo.Valid)
	if assert.True(t, gotV.Valid) {
		assert.Equal(t, 1, gotV.Int64)
	}
}

func TestConnQueryConstantString(t *testing.T) {
	db, _ := openNewDatabase(t)

	var got string
	err := db.QueryRow(`select 'live beef'`).Scan(&got)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "live beef", got)
	}
}

func TestConnQueryConstantByteSlice(t *testing.T) {
	// This might be implemented in DataFusion later, at which time, this test will fail
	db, _ := openNewDatabase(t)

	// expected := []byte{222, 173, 190, 239}
	// var actual []byte

	_, err := db.Query(`select X'deadbeef'`) // .Scan(&actual)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), `Unsupported ast node Value(HexStringLiteral("deadbeef")) in sqltorel`)
	}

	// require.NoError(t, err)
	// require.EqualValues(t, expected, actual)
}

func TestConnQueryFailure(t *testing.T) {
	db, _ := openNewDatabase(t)

	_, err := db.Query("select 'foo")
	require.Error(t, err)
}

func TestConnQueryRowUnsupportedType(t *testing.T) {
	db, _ := openNewDatabase(t)

	query := "select 1::decimal"

	row := db.QueryRow(query)
	if assert.Error(t, row.Err()) {
		assert.Contains(t, row.Err().Error(), "Unsupported SQL type Decimal")
	}
}

func TestConnRaw(t *testing.T) {
	db, _ := openNewDatabase(t)

	conn, err := db.Conn(context.Background())
	require.NoError(t, err)

	err = conn.Raw(func(driverConn interface{}) error {
		client := driverConn.(*ioxsql.Connection).Client()
		return client.Handshake(context.Background())
	})
	require.NoError(t, err)
}

func TestConnPingContextSuccess(t *testing.T) {
	db, _ := openNewDatabase(t)

	require.NoError(t, db.PingContext(context.Background()))
}

func TestConnPrepareContextSuccess(t *testing.T) {
	db, _ := openNewDatabase(t)

	stmt, err := db.PrepareContext(context.Background(), "select now()")
	assert.NoError(t, err)
	assert.NoError(t, stmt.Close())
}

func TestConnQueryContextSuccess(t *testing.T) {
	db, client := openNewDatabase(t)
	writeDataset(t, client)

	rows, err := db.QueryContext(context.Background(), "select foo, v from t ORDER BY v ASC")
	require.NoError(t, err)

	for rows.Next() {
		var foo string
		var n int64
		require.NoError(t, rows.Scan(&foo, &n))
	}
	require.NoError(t, rows.Err())
}

func TestConnQueryContextFailureRetry(t *testing.T) {
	db, _ := openNewDatabase(t)

	{
		conn, err := db.Conn(context.Background())
		require.NoError(t, err)
		err = conn.Raw(func(driverConn interface{}) error {
			client := driverConn.(*ioxsql.Connection).Client()
			return client.Close()
		})
		require.NoError(t, err)
	}

	_, err := db.QueryContext(context.Background(), "select 1")
	require.NoError(t, err)
}

func TestRowsColumnTypeDatabaseTypeName(t *testing.T) {
	db, _ := openNewDatabase(t)

	rows, err := db.Query("select 42::bigint as v")
	require.NoError(t, err)

	columnTypes, err := rows.ColumnTypes()
	require.NoError(t, err)
	require.Len(t, columnTypes, 1)

	assert.Equal(t, arrow.INT64.String(), columnTypes[0].DatabaseTypeName())
	require.NoError(t, rows.Close())
}

func TestStmtQueryContextCancel(t *testing.T) {
	db, _ := openNewDatabase(t)

	stmt, err := db.PrepareContext(context.Background(), "select 1")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	_, err = stmt.QueryContext(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestStmtQueryContextSuccess(t *testing.T) {
	db, _ := openNewDatabase(t)

	stmt, err := db.PrepareContext(context.Background(), "select 1")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx)
	if assert.NoError(t, err) && assert.True(t, rows.Next()) {
		var n int64
		require.NoError(t, rows.Scan(&n))
		require.False(t, rows.Next())
	}
	if assert.NoError(t, rows.Err()) {
		assert.NoError(t, rows.Close())
	}
}

func TestRowsColumnTypes(t *testing.T) {
	db, _ := openNewDatabase(t)

	columnTypesTests := []struct {
		Name     string
		TypeName string
		Length   struct {
			Len int64
			OK  bool
		}
		ScanType reflect.Type
	}{
		{
			Name:     "a",
			TypeName: arrow.INT64.String(),
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			ScanType: reflect.TypeOf(int64(0)),
		}, {
			Name:     "bar",
			TypeName: arrow.STRING.String(),
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: math.MaxInt64,
				OK:  true,
			},
			ScanType: reflect.TypeOf(""),
		}, {
			Name:     "dec",
			TypeName: arrow.FLOAT64.String(),
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			ScanType: reflect.TypeOf(float64(0)),
		}, {
			Name:     "d",
			TypeName: arrow.TIMESTAMP.String(),
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			ScanType: reflect.TypeOf(time.Time{}),
		},
	}

	rows, err := db.Query("SELECT 1::bigint AS a, varchar 'bar' AS bar, 1.28::float AS dec, '12:00:00'::timestamp as d")
	require.NoError(t, err)

	columns, err := rows.ColumnTypes()
	require.NoError(t, err)
	assert.Len(t, columns, 4)

	for i, tt := range columnTypesTests {
		c := columns[i]
		assert.Equal(t, tt.Name, c.Name())
		assert.Equal(t, tt.TypeName, c.DatabaseTypeName())
		l, ok := c.Length()
		if assert.Equal(t, tt.Length.OK, ok) && ok {
			assert.Equal(t, tt.Length.Len, l)
		}
		if c.ScanType() != tt.ScanType {
			t.Errorf("(%d) got: %v, want: %v", i, c.ScanType(), tt.ScanType)
		}
	}
}

func TestQueryLifeCycle(t *testing.T) {
	db, client := openNewDatabase(t)
	writeDataset(t, client)

	rows, err := db.Query("SELECT foo, v FROM t WHERE 3 = 3 ORDER BY v ASC")
	require.NoError(t, err)

	rowCount := int64(0)

	for rows.Next() {
		var (
			s string
			n int64
		)

		err := rows.Scan(&s, &n)
		require.NoError(t, err)

		assert.Equal(t, "bar", s)
		assert.Equal(t, rowCount, n)
		rowCount++
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
