package influxdbiox

import (
	"context"

	management "github.com/influxdata/influxdbiox/internal/management"
)

func (c *Client) ListDatabases(ctx context.Context) ([]string, error) {
	response, err := c.managementGRPCClient.ListDatabases(ctx, &management.ListDatabasesRequest{})
	if err != nil {
		return nil, err
	}
	return response.Names, nil
}

func (c *Client) CreateDatabase(ctx context.Context, databaseName string) error {
	request := &management.CreateDatabaseRequest{
		Rules: &management.DatabaseRules{
			Name: databaseName,
		},
	}
	_, err := c.managementGRPCClient.CreateDatabase(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetServerStatus(ctx context.Context) (string, error) {
	response, err := c.managementGRPCClient.GetServerStatus(ctx, &management.GetServerStatusRequest{})
	if err != nil {
		return "", err
	}
	if response.ServerStatus.Error != nil {
		return response.ServerStatus.Error.Message, nil
	}
	if response.ServerStatus.Initialized {
		return "ok", nil
	}
	return "?", nil
}
