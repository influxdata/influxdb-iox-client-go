package influxdbiox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/apache/arrow/go/v7/arrow/flight"
	"github.com/apache/arrow/go/v7/arrow/ipc"
	"github.com/apache/arrow/go/v7/arrow/memory"
	"google.golang.org/grpc"
)

// Handshake the InfluxDB/IOx service, possibly (re-)connecting to the gRPC
// service in the process.
func (c *Client) Handshake(ctx context.Context) error {
	response, err := c.flightClient.Handshake(ctx)
	if err != nil {
		return err
	}
	payload := make([]byte, 16)
	if _, err = rand.Read(payload); err != nil {
		return err
	}
	if err = response.Send(&flight.HandshakeRequest{Payload: payload}); err != nil {
		return err
	}
	resp, err := response.Recv()
	if err != nil {
		return err
	}
	if !bytes.Equal(resp.Payload, payload) {
		return errors.New("handshake payload response does not match request")
	}
	return nil
}

type ticketReadInfo struct {
	DatabaseName string `json:"database_name"`
	SQLQuery     string `json:"sql_query"`
}

// PrepareQuery prepares a query request.
//
// If database is "" then the configured default is used.
func (c *Client) PrepareQuery(ctx context.Context, database, query string) (*QueryRequest, error) {
	if database == "" {
		database = c.config.Database
	}
	return newRequest(c, database, query), nil
}

// QueryRequest represents a prepared query.
type QueryRequest struct {
	client          *Client
	database        string
	query           string
	grpcCallOptions []grpc.CallOption
	allocator       memory.Allocator
}

func newRequest(client *Client, database, query string) *QueryRequest {
	return &QueryRequest{
		client:    client,
		database:  database,
		query:     query,
		allocator: memory.DefaultAllocator,
	}
}

// WithCallOption adds a grpc.CallOption to be included when the gRPC service
// is called.
func (r *QueryRequest) WithCallOption(grpcCallOption grpc.CallOption) *QueryRequest {
	return &QueryRequest{
		client:          r.client,
		database:        r.database,
		query:           r.query,
		grpcCallOptions: append(r.grpcCallOptions, grpcCallOption),
		allocator:       r.allocator,
	}
}

// WithAllocator provides an Arrow allocator the that flight.Reader will use to
// account for memory allocated for record batches pulled off the wire.
func (r *QueryRequest) WithAllocator(alloc memory.Allocator) *QueryRequest {
	return &QueryRequest{
		client:          r.client,
		database:        r.database,
		query:           r.query,
		grpcCallOptions: r.grpcCallOptions,
		allocator:       alloc,
	}
}

// Query sends a query via the Flight RPC DoGet.
//
// The returned *flight.Reader must be released when the caller is done with it.
//
//	reader, err := request.Query(ctx)
//	defer reader.Release()
//	...
func (r *QueryRequest) Query(ctx context.Context, args ...interface{}) (*flight.Reader, error) {
	if len(args) > 0 {
		return nil, errors.New("query arguments are not supported")
	}
	ticket, err := json.Marshal(ticketReadInfo{
		DatabaseName: r.database,
		SQLQuery:     r.query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Arrow DoGet ticket: %w", err)
	}
	doGetClient, err := r.client.flightClient.DoGet(ctx, &flight.Ticket{Ticket: ticket}, r.grpcCallOptions...)
	if err != nil {
		return nil, fmt.Errorf("arrow Flight DoGet request failed: %w", err)
	}
	flightReader, err := flight.NewRecordReader(doGetClient, ipc.WithAllocator(r.allocator))
	if err != nil {
		return nil, fmt.Errorf("failed to create Flight record reader: %w", err)
	}
	return flightReader, nil
}
