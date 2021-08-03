package ioxsql

import (
	"context"
	"database/sql/driver"
	"errors"

	"github.com/influxdata/influxdb-iox-client-go"
)

var (
	_ driver.Stmt             = (*statement)(nil)
	_ driver.StmtExecContext  = (*statement)(nil)
	_ driver.StmtQueryContext = (*statement)(nil)
)

type statement struct {
	request *influxdbiox.QueryRequest
}

func newStatement(request *influxdbiox.QueryRequest) *statement {
	return &statement{
		request: request,
	}
}

func (s *statement) Close() error {
	return nil
}

func (s *statement) NumInput() int {
	return -1
}

func (s *statement) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("exec not implemented")
}

func (s *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("exec not implemented")
}

func (s *statement) Query(args []driver.Value) (driver.Rows, error) {
	if len(args) > 0 {
		return nil, errors.New("query args not supported")
	}
	return queryRows(context.Background(), s.request, nil)
}

func (s *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if len(args) > 0 {
		return nil, errors.New("query args not supported")
	}
	return queryRows(ctx, s.request, nil)
}
