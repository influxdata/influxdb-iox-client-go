package influxdbiox_test

import (
	"context"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/influxdb-iox-client-go"
)

func ExampleClient_PrepareQuery() {
	config, _ := influxdbiox.ClientConfigFromJSONString("localhost:8082")
	client, _ := influxdbiox.NewClient(context.Background(), config)

	req, _ := client.PrepareQuery(context.Background(), "mydb", "select count(*) from t")
	reader, _ := req.Query(context.Background())
	for reader.Next() {
		record := reader.Record()
		for i, column := range record.Columns() {
			columnName := record.ColumnName(i)
			println(columnName)
			switch typedColumn := column.(type) {
			case *array.Timestamp:
				values := typedColumn.TimestampValues()
				for _, value := range values {
					var t time.Time = time.Unix(0, int64(value))
					println(t.String())
				}
			case *array.Int64:
				var values []int64 = typedColumn.Int64Values()
				println(values)
			default:
				// Unexpected types
			}
		}
	}
}
