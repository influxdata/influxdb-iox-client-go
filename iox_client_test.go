package influxdbiox

import (
	"context"
	"testing"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	config := ClientConfig{
		Address:               "localhost:8082",
		Database:              "mydb",
	}
	client, err := NewClient(ctx, &config)
	require.NoError(t, err)

	require.NoError(t, client.Ping(ctx))

	req, err := client.Prepare("select count(*) from t;")
	require.NoError(t, err)

	reader, err := req.Query(ctx)
	require.NoError(t, err)
	t.Cleanup(reader.Release)

	for reader.Next() {
		record := reader.Record()
		t.Logf("%v", record.Column(0).(*array.Uint64).Uint64Values())
	}
}
