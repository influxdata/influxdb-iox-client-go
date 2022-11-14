package influxdbiox_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	influxdbiox "github.com/influxdata/influxdb-iox-client-go"
)

func ExampleClient_GetSchema() {
	config, _ := influxdbiox.ClientConfigFromAddressString("localhost:8082")
	client, _ := influxdbiox.NewClient(context.Background(), config)

	table := "my_measurement"
	req, _ := client.GetSchema(context.Background(), "mydb", table)

	fmt.Printf("Columns for table %q:\n", table)
	for name, dataType := range req {
		fmt.Printf("%-15s: %s\n", name, dataType)
	}
}

func TestClient_GetSchema(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	client, dbName := openNewDatabase(ctx, t)
	writeDataset(ctx, t, dbName, "bananas")

	// Fetch the schema for the table/namespace pair that was just created by
	// the writeDataset call.
	schema, err := client.GetSchema(ctx, dbName, "bananas")
	require.NoError(t, err)
	require.Equal(t, influxdbiox.ColumnType_TIME, schema["time"])
	require.Equal(t, influxdbiox.ColumnType_TAG, schema["foo"])
	require.Equal(t, influxdbiox.ColumnType_I64, schema["v"])
	require.Equal(t, influxdbiox.ColumnTypeUnknown, schema["does_not_exist"])
}

func TestClient_GetSchema_no_table(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	// Create a randomly-named namespace
	client, dbName := openNewDatabase(ctx, t)

	// Force creation of the namespace by writing to it.
	writeDataset(ctx, t, dbName, "a_test_table")

	// But ask for a table that was never created.
	_, err := client.GetSchema(ctx, dbName, "bananas")
	require.ErrorContains(t, err, "table not found")
}

func TestClient_GetSchema_no_namespace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	// Create a client
	client, _ := openNewDatabase(ctx, t)

	// But ask for some other namespace that (shouldn't) exist.
	_, err := client.GetSchema(ctx, "platanos", "bananas")
	require.ErrorContains(t, err, "namespace platanos not found")
}
