package influxdbiox

import (
	"context"
	"errors"
	"net/http"
	"time"

	ingester "github.com/influxdata/influxdb-iox-client-go/internal/ingester"
)

const tokenWaitInterval = 500 * time.Millisecond

// Blocks until the specified predicate is true.
func (c *Client) waitForToken(ctx context.Context, writeToken string, predicate func(*ingester.GetWriteInfoResponse) bool) error {
	request := &ingester.GetWriteInfoRequest{
		WriteToken: writeToken,
	}
	for {
		response, err := c.ingesterWriteInfoClient.GetWriteInfo(ctx, request)
		if err != nil {
			return err
		}
		if predicate(response) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(tokenWaitInterval):
			continue
		}
	}
}

// WriteTokenFromHTTPResponse fetches the IOx write token, if available, from
// the http.Response object
func WriteTokenFromHTTPResponse(response *http.Response) (string, error) {
	writeToken := response.Header.Get("X-IOx-Write-Token")
	if len(writeToken) == 0 {
		return "", errors.New("no write token found in HTTP response")
	}
	return writeToken, nil
}

// WaitForDurable blocks until the write associated with writeToken is durable,
// meaning that the data has been safely stored in a write-ahead log.
func (c *Client) WaitForDurable(ctx context.Context, writeToken string) error {
	return c.waitForToken(ctx, writeToken, func(response *ingester.GetWriteInfoResponse) bool {
		for _, pi := range response.ShardInfos {
			if !((pi.Status == ingester.ShardStatus_SHARD_STATUS_DURABLE) || (pi.Status == ingester.ShardStatus_SHARD_STATUS_READABLE) || (pi.Status == ingester.ShardStatus_SHARD_STATUS_PERSISTED)) {
				return false
			}
		}
		return true
	})
}

// WaitForReadable blocks until the write associated with writeToken is readable,
// meaning that the data can be queried.
func (c *Client) WaitForReadable(ctx context.Context, writeToken string) error {
	return c.waitForToken(ctx, writeToken, func(response *ingester.GetWriteInfoResponse) bool {
		for _, pi := range response.ShardInfos {
			if !((pi.Status == ingester.ShardStatus_SHARD_STATUS_READABLE) || (pi.Status == ingester.ShardStatus_SHARD_STATUS_PERSISTED)) {
				return false
			}
		}
		return true
	})
}

// WaitForPersisted blocks until the write associated with writeToken is persisted,
// meaning that the data has been batched, sorted, compacted, and persisted to disk
// or object storage.
func (c *Client) WaitForPersisted(ctx context.Context, writeToken string) error {
	return c.waitForToken(ctx, writeToken, func(response *ingester.GetWriteInfoResponse) bool {
		for _, pi := range response.ShardInfos {
			if pi.Status != ingester.ShardStatus_SHARD_STATUS_PERSISTED {
				return false
			}
		}
		return true
	})
}
