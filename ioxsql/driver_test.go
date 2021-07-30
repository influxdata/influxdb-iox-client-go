package ioxsql_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/influxdata/influxdbiox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriver_connect(t *testing.T) {
	source := &influxdbiox.ClientConfig{
		Address:  "localhost:8082",
		Database: "mydb",
	}
	dsn, err := source.ToJSONString()
	require.NoError(t, err)

	db, err := sql.Open("influxdb-iox", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	require.NoError(t, db.PingContext(context.Background()))
	rows, err := db.QueryContext(context.Background(), "select * from t;")
	require.NoError(t, err)

	for rows.Next() {
		println("row")
		ct, err := rows.ColumnTypes()
		assert.NoError(t, err)
		for _, c := range ct {
			nullable, hasNullable := c.Nullable()
			length, hasLength := c.Length()
			fmt.Printf("%+v %+v %+v %+v %+v %+v %+v\n",
				c.Name(), c.DatabaseTypeName(), nullable, hasNullable, c.ScanType(), length, hasLength)
		}

		// fmt.Printf("%+v\n", ))
	}

	if rows.NextResultSet() {
		for rows.Next() {
			println("row2")
		}
	}
}
