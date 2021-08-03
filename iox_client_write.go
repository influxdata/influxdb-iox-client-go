package influxdbiox

import (
	"context"

	influxdbpbdataprotocol "github.com/influxdata/influxdb-pb-data-protocol/golang"
	"github.com/influxdata/influxdb-pb-data-protocol/golang/ipdpsugar"
	"google.golang.org/grpc"
)

// NewWriteBatch constructs a new WriteBatch.
//
// If database is "" then the configured default is used.
func (c *Client) NewWriteBatch(database string) (*WriteBatch, error) {
	if database == "" {
		database = c.config.Database
	}
	return newWriteBatch(c, database), nil
}

// WriteBatch wraps ipdpsugar.DatabaseBatch.
//
// When the batch is complete, call the Write method to persist the batch.
type WriteBatch struct {
	client *Client
	*ipdpsugar.DatabaseBatch
	grpcCallOptions []grpc.CallOption
}

func newWriteBatch(client *Client, database string) *WriteBatch {
	return &WriteBatch{
		client:        client,
		DatabaseBatch: ipdpsugar.NewDatabaseBatch(database),
	}
}

// WithCallOption adds a grpc.CallOption to this WriteBatch, which will
// be used when the Write method is called.
func (b *WriteBatch) WithCallOption(grpcCallOption grpc.CallOption) *WriteBatch {
	return &WriteBatch{
		client:          b.client,
		DatabaseBatch:   b.DatabaseBatch,
		grpcCallOptions: append(b.grpcCallOptions, grpcCallOption),
	}
}

// Write writes this WriteBatch to the database.
func (b *WriteBatch) Write(ctx context.Context) error {
	pb, err := b.ToProto()
	if err != nil {
		return err
	}
	req := &influxdbpbdataprotocol.WriteRequest{
		DatabaseBatch: pb,
	}
	_, err = b.client.writeGRPCClient.Write(ctx, req, b.grpcCallOptions...)
	if err != nil {
		return err
	}
	return nil
}
