package influxdbiox

import (
	"context"

	"github.com/apache/arrow/go/arrow/flight"
	management "github.com/influxdata/influxdb-iox-client-go/internal/management"
	influxdbpbdataprotocol "github.com/influxdata/influxdb-pb-data-protocol/golang"
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
	c := &Client{
		config: config,
	}
	if err := c.Reconnect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// Reconnect closes the gRPC connection, if open, and creates a new connection.
func (c *Client) Reconnect(ctx context.Context) error {
	if c.grpcClient != nil {
		_ = c.grpcClient.Close()
	}

	grpcClient, err := c.config.GetGRPCClient(ctx)
	if err != nil {
		return err
	}
	c.grpcClient = grpcClient
	c.managementGRPCClient = management.NewManagementServiceClient(grpcClient)
	c.flightClient = flight.NewFlightServiceClient(grpcClient)
	c.writeGRPCClient = influxdbpbdataprotocol.NewWriteServiceClient(grpcClient)

	return nil
}

// GetState gets the state of the wrapped gRPC client.
func (c *Client) GetState() connectivity.State {
	return c.grpcClient.GetState()
}

// Close closes the instance of Client.
func (c *Client) Close() error {
	return c.grpcClient.Close()
}
