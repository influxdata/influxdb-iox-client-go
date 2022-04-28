package ioxsql_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/apache/arrow/go/v7/arrow"
	influxdbiox "github.com/influxdata/influxdb-iox-client-go"
	"github.com/influxdata/influxdb-iox-client-go/ioxsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func openNewDatabase(ctx context.Context, t *testing.T) (*sql.DB, *influxdbiox.Client, string) {
	databaseName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if testing.Verbose() {
		t.Logf("temporary database name: %q", databaseName)
	}

	host, found := os.LookupEnv("INFLUXDB_IOX_HOST")
	if !found {
		host = "localhost"
	}
	grpcPort, found := os.LookupEnv("INFLUXDB_IOX_GRPC_PORT")
	if !found {
		grpcPort = "8082"
	}
	httpPort, found := os.LookupEnv("INFLUXDB_IOX_HTTP_PORT")
	if !found {
		httpPort = "8080"
	}

	dsn := fmt.Sprintf("%s:%s/%s", host, grpcPort, databaseName)
	config, err := influxdbiox.ClientConfigFromAddressString(dsn)
	require.NoError(t, err)
	config.DialOptions = append(config.DialOptions, grpc.WithBlock())

	client, err := influxdbiox.NewClient(ctx, config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Handshake(ctx))

	sqlDB, err := sql.Open(ioxsql.DriverName, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	writeURL, err := url.Parse(fmt.Sprintf("http://%s:%s/api/v2/write", host, httpPort))
	require.NoError(t, err)
	queryValues := writeURL.Query()
	orgBucket := strings.SplitN(databaseName, "_", 2)
	require.Len(t, orgBucket, 2)
	queryValues.Set("org", orgBucket[0])
	queryValues.Set("bucket", orgBucket[1])
	queryValues.Set("precision", "ns")
	writeURL.RawQuery = queryValues.Encode()

	return sqlDB, client, writeURL.String()
}

func writeDataset(t *testing.T, writeURL string) {
	e := new(lineprotocol.Encoder)
	e.SetLax(false)
	e.SetPrecision(lineprotocol.Nanosecond)

	baseTime := time.Date(2021, time.April, 15, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 10; i++ {
		e.StartLine("t")
		e.AddTag("foo", "bar")
		e.AddField("v", lineprotocol.MustNewValue(int64(i)))
		e.EndLine(baseTime.Add(time.Minute * time.Duration(i)))
	}
	require.NoError(t, e.Err())

	resp, err := http.Post(writeURL, "text/plain; charset=utf-8", bytes.NewReader(e.Bytes()))
	require.NoError(t, err)
	require.Equal(t, 2, resp.StatusCode/100)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	openNewDatabase(ctx, t)
}

func TestNormalLifeCycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	db, _, writeURL := openNewDatabase(ctx, t)
	writeDataset(t, writeURL)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	_, err := db.Begin()
	require.EqualError(t, err, "transactions not supported")
}

func TestQueryCloseRowsEarly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, writeURL := openNewDatabase(ctx, t)
	writeDataset(t, writeURL)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)
	_, err := db.Exec("create table t(a varchar not null)")
	require.EqualError(t, err, "exec not implemented")
}

func TestArgsNotSupported(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	_, err := db.Query("select v from t where k = $1", "arg")
	assert.EqualError(t, err, "query args not supported")

	_, err = db.Query("select v from t where k = ?", "arg")
	assert.EqualError(t, err, "query args not supported")
}

func TestConnQueryNull(t *testing.T) {
	t.Skip("IOx/CF/Arrow bug in null handling")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, writeURL := openNewDatabase(ctx, t)
	writeDataset(t, writeURL)

	row := db.QueryRow("select foo, v from t where v = 1")
	require.NoError(t, row.Err())

	var gotFoo sql.NullString
	var gotV sql.NullInt64
	require.NoError(t, row.Scan(&gotFoo, &gotV))

	assert.False(t, gotFoo.Valid)
	if assert.True(t, gotV.Valid) {
		assert.EqualValues(t, 1, gotV.Int64)
	}
}

func TestConnQueryConstantString(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	var got string
	err := db.QueryRow(`select 'live beef'`).Scan(&got)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "live beef", got)
	}
}

func TestConnQueryConstantByteSlice(t *testing.T) {
	// This might be implemented in DataFusion later, at which time, this test will fail
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	_, err := db.Query("select 'foo")
	require.Error(t, err)
}

func TestConnQueryRowUnsupportedType(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	query := "select 1::UUID"

	row := db.QueryRow(query)
	if assert.Error(t, row.Err()) {
		assert.Contains(t, row.Err().Error(), "Unsupported SQL type Uuid")
	}
}

func TestConnRaw(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	conn, err := db.Conn(ctx)
	require.NoError(t, err)

	err = conn.Raw(func(driverConn interface{}) error {
		client := driverConn.(*ioxsql.Connection).Client()
		return client.Handshake(ctx)
	})
	require.NoError(t, err)
}

func TestConnPingContextSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	require.NoError(t, db.PingContext(ctx))
}

func TestConnPrepareContextSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	stmt, err := db.PrepareContext(ctx, "select now()")
	assert.NoError(t, err)
	assert.NoError(t, stmt.Close())
}

func TestConnQueryContextSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, writeURL := openNewDatabase(ctx, t)
	writeDataset(t, writeURL)

	rows, err := db.QueryContext(ctx, "select foo, v from t ORDER BY v ASC")
	require.NoError(t, err)

	for rows.Next() {
		var foo string
		var n int64
		require.NoError(t, rows.Scan(&foo, &n))
	}
	require.NoError(t, rows.Err())
}

func TestConnQueryContextFailureRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	{
		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		err = conn.Raw(func(driverConn interface{}) error {
			client := driverConn.(*ioxsql.Connection).Client()
			return client.Close()
		})
		require.NoError(t, err)
	}

	_, err := db.QueryContext(ctx, "select 1")
	require.NoError(t, err)
}

func TestRowsColumnTypeDatabaseTypeName(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	rows, err := db.Query("select 42::bigint as v")
	require.NoError(t, err)

	columnTypes, err := rows.ColumnTypes()
	require.NoError(t, err)
	require.Len(t, columnTypes, 1)

	assert.Equal(t, arrow.INT64.String(), columnTypes[0].DatabaseTypeName())
	require.NoError(t, rows.Close())
}

func TestStmtQueryContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	stmt, err := db.PrepareContext(ctx, "select 1")
	require.NoError(t, err)

	ctx2, cancel2 := context.WithTimeout(ctx, 0)
	defer cancel2()
	_, err = stmt.QueryContext(ctx2)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestStmtQueryContextSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

	stmt, err := db.PrepareContext(ctx, "select 1")
	require.NoError(t, err)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, _ := openNewDatabase(ctx, t)

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
			TypeName: arrow.DECIMAL.String(),
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

	rows, err := db.Query("SELECT 1::bigint AS a, varchar 'bar' AS bar, 1.28::DECIMAL(10,0) AS dec, '2004-10-19 10:23:54'::timestamp as d")
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, _, writeURL := openNewDatabase(ctx, t)
	writeDataset(t, writeURL)

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
