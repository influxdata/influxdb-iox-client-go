> [!WARNING]
> This module is no longer being maintained. Please use the official module at [github.com/InfluxCommunity/influxdb3-go](https://github.com/InfluxCommunity/influxdb3-go).

# InfluxDB IOx Client for Go

Package `influxdbiox` is a Go client for InfluxDB/IOx.

InfluxDB/IOx uses Arrow Flight gRPC for queries.
This client makes it easy to use that interface.

Take a look at the godoc for usage.

## SQL

Package [`ioxsql`](ioxsql) contains an implementation of the `database/sql` driver interface.

## Tests

This project does not run tests as part of CI.
Most tests depend on a running instance of InfluxDB/IOx, and each creates its own database.
To start an in-memory instance, from the [InfluxDB/IOx repository](https://github.com/influxdata/influxdb_iox/) root:
```console
$ cargo build
$ ./target/debug/influxdb_iox run all-in-one
```

Then run the tests like any golang test:
```console
$ go test ./...
```
