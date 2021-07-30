package ioxsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/influxdata/influxdbiox"
	"google.golang.org/grpc/connectivity"
)

var (
	_ driver.Driver        = (*Driver)(nil)
	_ driver.DriverContext = (*Driver)(nil)
)

func init() {
	sql.Register("influxdb-iox", thisDriver)
}

type Driver struct{}

var thisDriver = &Driver{}

func (d *Driver) Open(dataSourceName string) (driver.Conn, error) {
	connector, err := d.OpenConnector(dataSourceName)
	if err != nil {
		return nil, err
	}
	return connector.Connect(context.Background())
}

func (_ *Driver) OpenConnector(dataSourceName string) (driver.Connector, error) {
	config, err := influxdbiox.ClientConfigFromJSONString(dataSourceName)
	if err != nil {
		return nil, err
	}
	return NewConnector(config), nil
}

var _ driver.Connector = (*Connector)(nil)

type Connector struct {
	config *influxdbiox.ClientConfig
}

func NewConnector(config *influxdbiox.ClientConfig) *Connector {
	return &Connector{
		config: config,
	}
}

func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	client, err := influxdbiox.NewClient(ctx, c.config)
	if err != nil {
		return nil, err
	}

	return newConnection(client), nil
}

func (c *Connector) Driver() driver.Driver {
	return thisDriver
}

var (
	_ driver.Conn            = (*Connection)(nil)
	_ driver.Pinger          = (*Connection)(nil)
	_ driver.SessionResetter = (*Connection)(nil)
	_ driver.Validator       = (*Connection)(nil)
)

type Connection struct {
	client *influxdbiox.Client
}

func newConnection(client *influxdbiox.Client) *Connection {
	return &Connection{
		client: client,
	}
}

// Client returns the instance of *influxdbiox.Client backing this Connection.
// This is useful for sql.Conn.Raw():
//  conn, err := db.Conn(context.Background())
//  err = conn.Raw(func(driverConn interface{}) error {
//    // This client object has type *influxdbiox.Client
//    client := driverConn.(*ioxsql.Connection).Client()
//    ...
//    return nil
//  })
func (c *Connection) Client() *influxdbiox.Client {
	return c.client
}

func (c *Connection) Prepare(query string) (driver.Stmt, error) {
	request, err := c.client.Prepare(query)
	if err != nil {
		return nil, err
	}
	return newStatement(request), nil
}

func (c *Connection) Close() error {
	return c.client.Close()
}

func (c *Connection) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported")
}

func (c *Connection) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}

func (c *Connection) ResetSession(_ context.Context) error {
	if c.IsValid() {
		return nil
	}
	return driver.ErrBadConn
}

func (c *Connection) IsValid() bool {
	return c.client.GetState() != connectivity.Shutdown
}
