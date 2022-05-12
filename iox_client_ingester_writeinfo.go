package influxdbiox

import (
	"context"
	"errors"
	ingester "github.com/influxdata/influxdb-iox-client-go/internal/ingester"
	"net/http"
	"time"
)

const tokenWaitInterval = 500 * time.Millisecond

// Waits for the specified predicate to return true.
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

func WriteTokenFromHTTPResponse(response *http.Response) (string, error) {
	writeToken := response.Header.Get("X-IOx-Write-Token")
	if len(writeToken) == 0 {
		return "", errors.New("no write token found in HTTP response")
	}
	return writeToken, nil
}

func (c *Client) WaitForDurable(ctx context.Context, writeToken string) error {
	return c.waitForToken(ctx, writeToken, func(response *ingester.GetWriteInfoResponse) bool {
		for _, pi := range response.KafkaPartitionInfos {
			if pi.Status != ingester.KafkaPartitionStatus_KAFKA_PARTITION_STATUS_DURABLE {
				return false
			}
		}
		return true
	})
}

func (c *Client) WaitForReadable(ctx context.Context, writeToken string) error {
	return c.waitForToken(ctx, writeToken, func(response *ingester.GetWriteInfoResponse) bool {
		for _, pi := range response.KafkaPartitionInfos {
			if pi.Status != ingester.KafkaPartitionStatus_KAFKA_PARTITION_STATUS_READABLE {
				return false
			}
		}
		return true
	})
}

func (c *Client) WaitForPersisted(ctx context.Context, writeToken string) error {
	return c.waitForToken(ctx, writeToken, func(response *ingester.GetWriteInfoResponse) bool {
		for _, pi := range response.KafkaPartitionInfos {
			if pi.Status != ingester.KafkaPartitionStatus_KAFKA_PARTITION_STATUS_PERSISTED {
				return false
			}
		}
		return true
	})
}
