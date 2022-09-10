package influxdbiox_test

import (
	"context"
	"net/http"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/apache/arrow/go/v10/arrow/array"
	"github.com/influxdata/influxdb-iox-client-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteTokenFromHTTPResponse(t *testing.T) {
	response := &http.Response{
		Header: map[string][]string{
			textproto.CanonicalMIMEHeaderKey("X-IOx-Write-Token"): {"foo"},
			textproto.CanonicalMIMEHeaderKey("Something-Else"):    {"bar"},
		},
	}
	writeToken, err := influxdbiox.WriteTokenFromHTTPResponse(response)
	require.NoError(t, err)
	assert.Equal(t, "foo", writeToken)

	response = &http.Response{
		Header: map[string][]string{
			textproto.CanonicalMIMEHeaderKey("Something-Else"): {"bar"},
		},
	}
	writeToken, err = influxdbiox.WriteTokenFromHTTPResponse(response)
	require.Empty(t, writeToken)
	assert.Contains(t, strings.ToLower(err.Error()), "no write token found")

	response = &http.Response{
		Header: nil,
	}
	writeToken, err = influxdbiox.WriteTokenFromHTTPResponse(response)
	require.Empty(t, writeToken)
	assert.Contains(t, strings.ToLower(err.Error()), "no write token found")
}

func TestClient_WaitForReadable(t *testing.T) {
	/*
		In this test, WaitForDurable should most often result in zero results found,
		since durability does not always imply readability. There is no other way to
		verify that written data is durable, so we don't test it.

		Similarly, WaitForPersisted should always result in 10 results found, since
		persisted implies readable. There is no other way to test that the written
		data is persisted, so we don't test it.
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	client, dbName := openNewDatabase(ctx, t)
	response := writeDataset(ctx, t, dbName, "bananas")

	writeToken, err := influxdbiox.WriteTokenFromHTTPResponse(response)
	require.NoError(t, err)

	err = client.WaitForReadable(ctx, writeToken)
	require.NoError(t, err)

	queryRequest, err := client.PrepareQuery(ctx, "", "select count(*) from bananas;")
	require.NoError(t, err)

	reader, err := queryRequest.Query(ctx)
	require.NoError(t, err)
	t.Cleanup(reader.Release)

	require.True(t, reader.Next())
	record := reader.Record()
	assert.Equal(t, []int64{10}, record.Column(0).(*array.Int64).Int64Values())
	require.False(t, reader.Next())
}
