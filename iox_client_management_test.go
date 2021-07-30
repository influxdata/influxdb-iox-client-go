package influxdbiox_test

import (
	"context"

	"github.com/influxdata/influxdbiox"
)

func ExampleClient_CreateDatabase() {
	config, err := influxdbiox.ClientConfigFromJSONString("localhost:8082")
	client, err := influxdbiox.NewClient(context.Background(), config)

	err := client.CreateDatabase(context.Background(), "mydb")
}

func ExampleClient_ListDatabases() {
	config, err := influxdbiox.ClientConfigFromJSONString("localhost:8082")
	client, err := influxdbiox.NewClient(context.Background(), config)

	databases, err := client.ListDatabases(context.Background())
	for _, database := range databases {
		println(database)
	}
}
