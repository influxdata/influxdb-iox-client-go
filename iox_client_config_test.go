package influxdbiox_test

import (
	"fmt"
	"testing"

	"github.com/influxdata/influxdb-iox-client-go"
	"github.com/stretchr/testify/assert"
)

func ExampleClientConfig_ToJSONString() {
	config := &influxdbiox.ClientConfig{
		Address: "localhost:8082",
	}
	s, _ := config.ToJSONString()
	println(s)
}

func ExampleClientConfigFromJSONString() {
	// Multiline, indented JSON is accepted.
	dsn := `{
"address": "localhost:8082",
"tls_ca":  "..."
}`
	config, _ := influxdbiox.ClientConfigFromJSONString(dsn)
	println(config)

	config, _ = influxdbiox.ClientConfigFromJSONString(`{"address":"localhost:8082","tls_ca":"..."}`)
	println(config)
}

func ExampleClientConfigFromAddressString() {
	config, _ := influxdbiox.ClientConfigFromAddressString("localhost:8082")
	println(config)

	config, _ = influxdbiox.ClientConfigFromAddressString("localhost:8082/mydb")
	println(config)
}

func TestClientConfigFromAddressString(t *testing.T) {
	tests := []struct {
		s            string
		expectConfig *influxdbiox.ClientConfig
		expectError  bool
	}{{
		s:            "",
		expectConfig: nil,
		expectError:  true,
	}, {
		s:            "localhost:8082",
		expectConfig: &influxdbiox.ClientConfig{Address: "localhost:8082"},
		expectError:  false,
	}, {
		s:            "localhost:8082/mydb",
		expectConfig: &influxdbiox.ClientConfig{Address: "localhost:8082", Database: "mydb"},
		expectError:  false,
	}, {
		s:            "localhost",
		expectConfig: nil,
		expectError:  true,
	}, {
		s:            "localhost/mydb",
		expectConfig: nil,
		expectError:  true,
	}}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			gotConfig, gotErr := influxdbiox.ClientConfigFromAddressString(test.s)
			if test.expectError {
				assert.Nil(t, gotConfig)
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, test.expectConfig, gotConfig)
				assert.NoError(t, gotErr)
			}
		})
	}
}
