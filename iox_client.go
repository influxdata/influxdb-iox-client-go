package influxdbiox

import (
	"context"

	"github.com/apache/arrow/go/arrow/flight"
	influxdbpbdataprotocol "github.com/influxdata/influxdb-pb-data-protocol/golang"
	management "github.com/influxdata/influxdbiox/internal/management"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Client is the primary handle to interact with InfluxDB/IOx.
type Client struct {
	config               *ClientConfig
	grpcClient           *grpc.ClientConn
	managementGRPCClient management.ManagementServiceClient
	flightClient         flight.FlightServiceClient
	writeGRPCClient      influxdbpbdataprotocol.WriteServiceClient
}

// NewClient instantiates a connection with the InfluxDB/IOx gRPC services.
//
// The gRPC client does not establish a connection here, unless
// ClientConfig.GRPCClient has been configured with dialer option grpc.WithBlock.
// For use of the context.Context object in this function, see grpc.DialContext.
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	grpcClient, err := config.GetGRPCClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{
		config:     config,
		grpcClient: grpcClient,
	}, nil
}

// GetState gets the state of the wrapped gRPC client.
func (c *Client) GetState() connectivity.State {
	return c.grpcClient.GetState()
}

// Close closes the instance of Client.
func (c *Client) Close() error {
	return c.grpcClient.Close()
}
