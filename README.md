# InfluxDB IOx Client for Go

Package `influxdbiox` is the official Go client for InfluxDB/IOx.

InfluxDB/IOx uses Arrow Flight gRPC for queries, and InfluxDB Protobuf Data Protocol for writes.
This client makes it easy to use those interfaces.

Take a look at the godoc for usage.

## SQL

Package [`ioxsql`](ioxsql) contains an implementation of the `database/sql` driver interface.
