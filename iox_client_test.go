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

	"github.com/apache/arrow/go/v7/arrow/array"
	influxdbiox "github.com/influxdata/influxdb-iox-client-go"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/stretchr/testify/require"
)

func openNewDatabase(ctx context.Context, t *testing.T) (*influxdbiox.Client, string) {
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

	config := influxdbiox.ClientConfig{
		Address:     fmt.Sprintf("%s:%s", host, grpcPort),
		Database:    databaseName,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	client, err := influxdbiox.NewClient(ctx, &config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Handshake(ctx))

	writeURL, err := url.Parse(fmt.Sprintf("http://%s:%s/api/v2/write", host, httpPort))
	require.NoError(t, err)
	queryValues := writeURL.Query()
	orgBucket := strings.SplitN(databaseName, "_", 2)
	require.Len(t, orgBucket, 2)
	queryValues.Set("org", orgBucket[0])
	queryValues.Set("bucket", orgBucket[1])
	queryValues.Set("precision", "ns")
	writeURL.RawQuery = queryValues.Encode()

	return client, writeURL.String()
}

func writeDataset(ctx context.Context, t *testing.T, writeURL string) *http.Response {
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

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, writeURL, bytes.NewReader(e.Bytes()))
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

	client, writeURL := openNewDatabase(ctx, t)
	writeDataset(ctx, t, writeURL)

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
