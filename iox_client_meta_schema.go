package influxdbiox

import (
	"context"
	"errors"

	schema "github.com/influxdata/influxdb-iox-client-go/internal/schema"
)

// ColumnType defines the column data types IOx can represent.
type ColumnType int32

const (
	// ColumnTypeUnknown is an invalid column type.
	ColumnTypeUnknown ColumnType = 0
	// ColumnType_I64 is an int64.
	ColumnType_I64 ColumnType = 1
	// ColumnType_U64 is an uint64.
	ColumnType_U64 ColumnType = 2
	// ColumnType_F64 is an float64.
	ColumnType_F64 ColumnType = 3
	// ColumnType_BOOL is a bool.
	ColumnType_BOOL ColumnType = 4
	// ColumnType_STRING is a string.
	ColumnType_STRING ColumnType = 5
	// ColumnType_TIME is a timestamp.
	ColumnType_TIME ColumnType = 6
	// ColumnType_TAG is a tag value.
	ColumnType_TAG ColumnType = 7
)

func (c ColumnType) String() string {
	switch c {
	case ColumnType_I64:
		return "int64"
	case ColumnType_U64:
		return "uint64"
	case ColumnType_F64:
		return "float64"
	case ColumnType_BOOL:
		return "bool"
	case ColumnType_STRING:
		return "string"
	case ColumnType_TIME:
		return "timestamp"
	case ColumnType_TAG:
		return "tag"
	default:
		return "unknown"
	}
}

// Return a map of column name to data types for the specified table in
// namespace.
func (c *Client) GetSchema(ctx context.Context, namespace string, table string) (map[string]ColumnType, error) {
	client := schema.NewSchemaServiceClient(c.grpcClient)
	resp, err := client.GetSchema(ctx, &schema.GetSchemaRequest{
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	// Extract the (possibly nil) map of column name -> data type for the
	// requested table.
	cols := resp.GetSchema().GetTables()[table].GetColumns()
	if cols == nil {
		return nil, errors.New("table not found")
	}

	// Iterate over all the columns for table, mapping the proto data type to a
	// package const.
	ret := make(map[string]ColumnType, len(cols))
	for colName, col := range cols {
		// Attempt to map the proto type to the package const.
		//
		// This can fail if the server sends a data type identifier this client
		// does not know - this would indicate a client/server version mismatch.
		colType, err := mapProtoColumnType(col.GetColumnType())
		if err != nil {
			return nil, err
		}

		ret[colName] = colType
	}

	return ret, nil
}

// Convert the given proto data type const into an exported typed const.
func mapProtoColumnType(v schema.ColumnSchema_ColumnType) (ColumnType, error) {
	switch v {
	case schema.ColumnSchema_COLUMN_TYPE_I64:
		return ColumnType_I64, nil
	case schema.ColumnSchema_COLUMN_TYPE_U64:
		return ColumnType_U64, nil
	case schema.ColumnSchema_COLUMN_TYPE_F64:
		return ColumnType_F64, nil
	case schema.ColumnSchema_COLUMN_TYPE_BOOL:
		return ColumnType_BOOL, nil
	case schema.ColumnSchema_COLUMN_TYPE_STRING:
		return ColumnType_STRING, nil
	case schema.ColumnSchema_COLUMN_TYPE_TIME:
		return ColumnType_TIME, nil
	case schema.ColumnSchema_COLUMN_TYPE_TAG:
		return ColumnType_TAG, nil
	default:
		return 0, errors.New("unknown column data type in response")
	}
}
