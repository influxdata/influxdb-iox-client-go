// Package ioxsql is the compatibility layer from influxdbiox to database/sql.
//
// A database/sql connection can be established through sql.Open.
// Two data source name formats are supported.
//
// The complete influxdbiox.ClientConfig struct is supported as JSON:
//
//	dsn := `{
//	  "address": "localhost:8082",
//	  "tls_cert": "...",
//	  "tls_key": "..."
//	}`
//	db, err := sql.Open("influxdb-iox", dsn)
//
// The influxdbiox.ClientConfig struct serializes to JSON with method ToJSONString:
//
//	config := &influxdbiox.ClientConfig{
//	  Address:               "localhost:8082",
//	  TLSInsecureSkipVerify: true,
//	}
//	dsn, err = config.ToJSONString()
//	db, err := sql.Open("influxdb-iox", dsn)
//
// The host:port address format is simpler to type, but only sets the address field:
//
//	db, err := sql.Open("influxdb-iox", "localhost:8082")
//
// Or a influxdbiox.ClientConfig can be used directly.
//
//	config := &influxdbiox.ClientConfig{
//	  Address:               "localhost:8082",
//	  TLSInsecureSkipVerify: true,
//	}
//	db := sql.OpenDB(ioxsql.NewConnector(config))
package ioxsql
