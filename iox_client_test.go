package influxdbiox_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/apache/arrow/go/v10/arrow/array"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/influxdb-iox-client-go/v2"
)

// Return the environment value for env, or default to the provided fallback
// value.
func envOrDefault(env string, fallback string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		return fallback
	}
	return v
}

// Return the hostname of the test IOx instance.
func getTestHost() string {
	return envOrDefault("INFLUXDB_IOX_HOST", "localhost")
}

// Return the HTTP port for the test IOx instance.
func getTestHttpPort() string {
	return envOrDefault("INFLUXDB_IOX_HTTP_PORT", "8080")
}

// Return the gRPC port for the test IOx instance.
func getTestGRPCPort() string {
	return envOrDefault("INFLUXDB_IOX_GRPC_PORT", "8082")
}

// Initialises the IOx client with a randomly generated database name.
//
// Returns the client & per-client database name.
func openNewDatabase(ctx context.Context, t *testing.T) (*influxdbiox.Client, string) {
	databaseName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if testing.Verbose() {
		t.Logf("temporary database name: %q", databaseName)
	}

	host := getTestHost()
	grpcPort := getTestGRPCPort()

	config := influxdbiox.ClientConfig{
		Address:     fmt.Sprintf("%s:%s", host, grpcPort),
		Namespace:   databaseName,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	client, err := influxdbiox.NewClient(ctx, &config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Handshake(ctx))

	return client, databaseName
}

// Write some data to the specified table, within the specified database.
func writeDataset(ctx context.Context, t *testing.T, databaseName string, table string) *http.Response {
	writeURL, err := url.Parse(fmt.Sprintf("http://%s:%s/api/v2/write", getTestHost(), getTestHttpPort()))
	require.NoError(t, err)

	// Break the database name into an org/bucket pair.
	orgBucket := strings.SplitN(databaseName, "_", 2)
	require.Len(t, orgBucket, 2)

	queryValues := writeURL.Query()
	queryValues.Set("org", orgBucket[0])
	queryValues.Set("bucket", orgBucket[1])
	queryValues.Set("precision", "ns")
	writeURL.RawQuery = queryValues.Encode()

	e := new(lineprotocol.Encoder)
	e.SetLax(false)
	e.SetPrecision(lineprotocol.Nanosecond)

	baseTime := time.Date(2021, time.April, 15, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 10; i++ {
		e.StartLine(table)
		e.AddTag("foo", "bar")
		e.AddField("v", lineprotocol.MustNewValue(int64(i)))
		e.EndLine(baseTime.Add(time.Minute * time.Duration(i)))
	}
	require.NoError(t, e.Err())

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, writeURL.String(), bytes.NewReader(e.Bytes()))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, 2, response.StatusCode/100)

	return response
}

func TestClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	client, dbName := openNewDatabase(ctx, t)
	writeDataset(ctx, t, dbName, "t")

	req, err := client.PrepareQuery(ctx, "", "select count(*) from t;")
	require.NoError(t, err)

	reader, err := req.Query(ctx)
	require.NoError(t, err)
	t.Cleanup(reader.Release)

	for reader.Next() {
		record := reader.Record()
		t.Logf("%v", record.Column(0).(*array.Int64).Int64Values())
	}
}
