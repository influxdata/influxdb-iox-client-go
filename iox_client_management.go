package influxdbiox

import (
	"context"
	"time"

	management "github.com/influxdata/influxdb-iox-client-go/internal/management"
)

func (c *Client) ListDatabases(ctx context.Context) ([]string, error) {
	response, err := c.managementGRPCClient.ListDatabases(ctx, &management.ListDatabasesRequest{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, rules := range response.Rules {
		names = append(names, rules.Name)
	}
	return names, nil
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

func (c *Client) Delete(ctx context.Context, databaseName, tableName string, startTime, stopTime time.Time, predicate string) error {
	_, err := c.managementGRPCClient.Delete(ctx, &management.DeleteRequest{
		DbName:    databaseName,
		TableName: tableName,
		StartTime: startTime.UTC().Format(time.RFC3339),
		StopTime:  stopTime.UTC().Format(time.RFC3339),
		Predicate: predicate,
	})
	if err != nil {
		return err
	}
	return nil
}
