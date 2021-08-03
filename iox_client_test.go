package influxdbiox_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/influxdb-iox-client-go"
	"github.com/stretchr/testify/require"
)

func databaseAddress() string {
	address, found := os.LookupEnv("INFLUXDB_IOX_ADDRESS")
	if !found {
		address = "localhost:8082"
	}
	return address
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

func TestClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	database := fmt.Sprintf("test-%d", time.Now().UnixNano())
	if testing.Verbose() {
		t.Logf("temporary database: %q", database)
	}

	config := influxdbiox.ClientConfig{
		Address:  databaseAddress(),
		Database: database,
	}
	client, err := influxdbiox.NewClient(ctx, &config)
	require.NoError(t, err)

	require.NoError(t, client.Handshake(ctx))

	require.NoError(t, client.CreateDatabase(ctx, database))
	writeDataset(t, client)

	req, err := client.PrepareQuery(ctx, "", "select count(*) from t;")
	require.NoError(t, err)

	reader, err := req.Query(ctx)
	require.NoError(t, err)
	t.Cleanup(reader.Release)

	for reader.Next() {
		record := reader.Record()
		t.Logf("%v", record.Column(0).(*array.Uint64).Uint64Values())
	}
}
