package influxdbiox

import (
	"context"

	influxdbpbdataprotocol "github.com/influxdata/influxdb-pb-data-protocol/golang"
	"github.com/influxdata/influxdb-pb-data-protocol/golang/ipdpsugar"
	"google.golang.org/grpc"
)

func (c *Client) NewWriteBatch(database string) (*WriteBatch, error) {
	return newWriteBatch(c, database), nil
}

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

func (b *WriteBatch) WithCallOption(grpcCallOption grpc.CallOption) *WriteBatch {
	return &WriteBatch{
		client:          b.client,
		DatabaseBatch:   b.DatabaseBatch,
		grpcCallOptions: append(b.grpcCallOptions, grpcCallOption),
	}
}

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
