package influxdbiox_test

import (
	"context"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/influxdbiox"
)

func ExampleClient_PrepareQuery() {
	config, err := influxdbiox.ClientConfigFromJSONString("localhost:8082")
	client, err := influxdbiox.NewClient(context.Background(), config)

	req, err := client.PrepareQuery("mydb", "select count(*) from t")
	reader, err := req.Query(context.Background())
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
				}
			case *array.Int64:
				var values []int64 = typedColumn.Int64Values()
			default:
				// Unexpected types
			}
		}
	}
}
