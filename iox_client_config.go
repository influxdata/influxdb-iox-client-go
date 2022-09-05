package influxdbiox

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ClientConfig contains all the options used to establish a connection.
type ClientConfig struct {
	// Address string as host:port
	Address string `json:"address"`
	// Default database; optional unless using sql.Open
	Database string `json:"database,omitempty"`

	// Filename containing PEM encoded certificate for root certificate authority
	// to use when verifying server certificates.
	TLSCA string `json:"tls_ca,omitempty"`
	// Filename of certificate to present to service. TODO say more here
	TLSCert string `json:"tls_cert,omitempty"`
	TLSKey  string `json:"tls_key,omitempty"`
	// Do not verify the server's certificate chain and host name
	TLSInsecureSkipVerify bool `json:"tls_insecure_skip_verify,omitempty"`
	// Used to verify the server's hostname on the returned certificates
	// unless TLSInsecureSkipVerify is true
	TLSServerName string `json:"tls_server_name,omitempty"`

	// DialOptions are passed to grpc.DialContext when a new gRPC connection
	// is created.
	DialOptions []grpc.DialOption `json:"-"`

	// Use this TLS config, instead of allowing this library to generate one
	// from fields named with prefix "TLS".
	TLSConfig *tls.Config `json:"-"`
}

// ToJSONString converts this instance of *ClientConfig to a JSON string,
// which can be used as an argument for sql.Open().
//
// Example output:
//
//	{"address":"localhost:8082","database":"mydb"}
//
// To customize the way the JSON string is constructed, call json.Marshal
// with a *ClientConfig.
func (dc *ClientConfig) ToJSONString() (string, error) {
	b := bytes.NewBuffer(nil)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(dc); err != nil {
		return "", err
	}
	return b.String(), nil
}

// ClientConfigFromJSONString constructs an instance of *ClientConfig from a JSON string.
//
// See ConfigClient for a description of all fields.
// Example:
//
//	{
//	  "address": "localhost:8082",
//	  "tls_cert": "...",
//	  "tls_key": "..."
//	}
func ClientConfigFromJSONString(s string) (*ClientConfig, error) {
	var dc ClientConfig
	if err := json.Unmarshal([]byte(s), &dc); err != nil {
		return nil, fmt.Errorf("failed to parse client config from JSON string: %w", err)
	}
	if _, err := dc.getTLSConfig(); err != nil {
		return nil, fmt.Errorf("TLS config parse failed: %w", err)
	}
	return &dc, nil
}

// ClientConfigFromAddressString constructs an instance of *ClientConfig from an address string.
//
// Example, IPv4:
//
//	localhost:8082
//
// Example, IPv6:
//
//	[::1]:8082
//
// To specify a default database, as required by ioxsql (the database/sql driver),
// append a slash to the address.
//
// Example:
//
//	localhost:8082/mydb
func ClientConfigFromAddressString(s string) (*ClientConfig, error) {
	var address, database string
	if index := strings.IndexRune(s, '/'); index >= 0 {
		address = s[:index]
		database = s[index+1:]
	} else {
		address = s
	}

	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client config from address string: %w", err)
	}
	return &ClientConfig{
		Address:  address,
		Database: database,
	}, nil
}

// newGRPCClient returns a *grpc.ClientConn based on the config, or returns
// the instance already set as ClientConfig.GRPCClient.
func (dc *ClientConfig) newGRPCClient(ctx context.Context) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if tlsConfig, err := dc.getTLSConfig(); err != nil {
		return nil, err
	} else if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}
	dialOptions := append([]grpc.DialOption{grpc.WithTransportCredentials(creds)}, dc.DialOptions...)

	grpcClient, err := grpc.DialContext(ctx, dc.Address, dialOptions...)
	if err != nil {
		return nil, err
	}

	return grpcClient, nil
}

func (dc *ClientConfig) getTLSConfig() (*tls.Config, error) {
	if dc.TLSConfig != nil {
		return dc.TLSConfig, nil
	}

	if dc.TLSCA == "" && dc.TLSKey == "" && dc.TLSCert == "" && !dc.TLSInsecureSkipVerify && dc.TLSServerName == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: dc.TLSInsecureSkipVerify,
		Renegotiation:      tls.RenegotiateNever,
	}

	if dc.TLSCA != "" {
		pool := x509.NewCertPool()
		pem, err := ioutil.ReadFile(dc.TLSCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read root certificate file %q: %w", dc.TLSCA, err)
		}
		if ok := pool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to parse PEM certificate in root certificate file %q: %w", dc.TLSCA, err)
		}
		tlsConfig.RootCAs = pool
	}

	if dc.TLSCert != "" && dc.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(dc.TLSCert, dc.TLSKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if dc.TLSServerName != "" {
		tlsConfig.ServerName = dc.TLSServerName
	}
	dc.TLSConfig = tlsConfig

	return tlsConfig, nil
}
