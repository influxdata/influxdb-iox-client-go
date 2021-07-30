package influxdbiox_test

import "github.com/influxdata/influxdbiox"

func ExampleClientConfig_ToJSONString() {
	config := &influxdbiox.ClientConfig{
		Address: "localhost:8082",
	}
	s, err := config.ToJSONString()
	println(s)
}

func ExampleClientConfigFromJSONString() {
	// Multiline, indented JSON is accepted.
	dsn := `{
"address": "localhost:8082",
"tls_ca":  "..."
}`
	config, err := influxdbiox.ClientConfigFromJSONString(dsn)

	config, err = influxdbiox.ClientConfigFromJSONString(`{"address":"localhost:8082","tls_ca":"..."}`)
}

func ExampleClientConfigFromAddressString() {
	config, err := influxdbiox.ClientConfigFromAddressString("localhost:8082")
}
