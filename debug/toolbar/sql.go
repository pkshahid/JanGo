package toolbar

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"runtime"
	"strings"
	"time"
)

// DebugDriver wraps an existing sql.Driver to intercept queries.
type DebugDriver struct {
	Driver driver.Driver
}

func (d *DebugDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &DebugConn{Conn: conn}, nil
}

// Ensure DebugDriver implements driver.DriverContext
func (d *DebugDriver) OpenConnector(name string) (driver.Connector, error) {
	if driverCtx, ok := d.Driver.(driver.DriverContext); ok {
		connector, err := driverCtx.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &DebugConnector{Connector: connector, DriverObj: d}, nil
	}

	// Fallback connector
	return &FallbackConnector{driver: d, name: name}, nil
}

type FallbackConnector struct {
	driver driver.Driver
	name   string
}

func (c *FallbackConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.driver.Open(c.name)
}
func (c *FallbackConnector) Driver() driver.Driver {
	return c.driver
}

type DebugConnector struct {
	Connector driver.Connector
	DriverObj driver.Driver
}

func (c *DebugConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &DebugConn{Conn: conn}, nil
}

func (c *DebugConnector) Driver() driver.Driver {
	return c.DriverObj
}

// DebugConn wraps a driver.Conn
type DebugConn struct {
	driver.Conn
}

func (c *DebugConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &DebugStmt{Stmt: stmt, QueryStr: query}, nil
}

func (c *DebugConn) Close() error {
	return c.Conn.Close()
}

func (c *DebugConn) Begin() (driver.Tx, error) {
	return c.Conn.Begin()
}

func (c *DebugConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if connCtx, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err := connCtx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &DebugStmt{Stmt: stmt, QueryStr: query}, nil
	}
	return c.Prepare(query)
}

func (c *DebugConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()
	var res driver.Result
	var err error
	if execCtx, ok := c.Conn.(driver.ExecerContext); ok {
		res, err = execCtx.ExecContext(ctx, query, args)
	} else if execer, ok := c.Conn.(driver.Execer); ok {
		dargs, err2 := namedValueToValue(args)
		if err2 != nil {
			return nil, err2
		}
		res, err = execer.Exec(query, dargs)
	} else {
		return nil, driver.ErrSkip
	}
	duration := time.Since(start)
	recordQuery(ctx, query, args, duration, err)
	return res, err
}

func (c *DebugConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()
	var rows driver.Rows
	var err error
	if queryCtx, ok := c.Conn.(driver.QueryerContext); ok {
		rows, err = queryCtx.QueryContext(ctx, query, args)
	} else if queryer, ok := c.Conn.(driver.Queryer); ok {
		dargs, err2 := namedValueToValue(args)
		if err2 != nil {
			return nil, err2
		}
		rows, err = queryer.Query(query, dargs)
	} else {
		return nil, driver.ErrSkip
	}
	duration := time.Since(start)
	recordQuery(ctx, query, args, duration, err)
	return rows, err
}

// DebugStmt wraps a driver.Stmt
type DebugStmt struct {
	driver.Stmt
	QueryStr string
}

func (s *DebugStmt) Exec(args []driver.Value) (driver.Result, error) {
	start := time.Now()
	res, err := s.Stmt.Exec(args)
	duration := time.Since(start)
	recordQuery(context.Background(), s.QueryStr, args, duration, err)
	return res, err
}

func (s *DebugStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.Stmt.Query(args)
	duration := time.Since(start)
	recordQuery(context.Background(), s.QueryStr, args, duration, err)
	return rows, err
}

func (s *DebugStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()
	var res driver.Result
	var err error
	if stmtCtx, ok := s.Stmt.(driver.StmtExecContext); ok {
		res, err = stmtCtx.ExecContext(ctx, args)
	} else {
		dargs, err2 := namedValueToValue(args)
		if err2 != nil {
			return nil, err2
		}
		res, err = s.Exec(dargs)
	}
	duration := time.Since(start)
	recordQuery(ctx, s.QueryStr, args, duration, err)
	return res, err
}

func (s *DebugStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()
	var rows driver.Rows
	var err error
	if stmtCtx, ok := s.Stmt.(driver.StmtQueryContext); ok {
		rows, err = stmtCtx.QueryContext(ctx, args)
	} else {
		dargs, err2 := namedValueToValue(args)
		if err2 != nil {
			return nil, err2
		}
		rows, err = s.Query(dargs)
	}
	duration := time.Since(start)
	recordQuery(ctx, s.QueryStr, args, duration, err)
	return rows, err
}

func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			// Some drivers might not support names
			return nil, driver.ErrSkip
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}

// SQLQueryInfo holds data about a single intercepted query.
type SQLQueryInfo struct {
	Query     string
	Params    []any
	Duration  time.Duration
	Traceback string
	Duplicate bool
}

// recordQuery finds the toolbar in the context and adds the query info.
func recordQuery(ctx context.Context, query string, args any, duration time.Duration, err error) {
	tb := GetToolbarFromContext(ctx)
	if tb == nil {
		return
	}

	// Capture stack trace
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	// Simplify stack trace to hide our own driver logic
	var filteredStack []string
	lines := strings.Split(stack, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, "debug/toolbar") || strings.Contains(line, "database/sql") {
			if i+1 < len(lines) {
				i++ // skip the line number
			}
			continue
		}
		filteredStack = append(filteredStack, line)
	}

	info := SQLQueryInfo{
		Query:     query,
		Duration:  duration,
		Traceback: strings.Join(filteredStack, "\n"),
	}

	// Extract params
	if namedArgs, ok := args.([]driver.NamedValue); ok {
		for _, arg := range namedArgs {
			info.Params = append(info.Params, arg.Value)
		}
	} else if valArgs, ok := args.([]driver.Value); ok {
		for _, arg := range valArgs {
			info.Params = append(info.Params, arg)
		}
	}

	// We look for SQLPanel and inject it
	if p, ok := tb.Panels["SQL"]; ok {
		if sqlPanel, ok := p.(*SQLPanel); ok {
			sqlPanel.mu.Lock()

			// Detect duplicates (simple check based on exact query and params matching)
			// A true implementation might hash the query and params.
			// This is an O(N) check.
			for _, existing := range sqlPanel.Queries {
				if existing.Query == info.Query && len(existing.Params) == len(info.Params) {
					// Check params roughly
					dup := true
					for i, param := range info.Params {
						if existing.Params[i] != param {
							dup = false
							break
						}
					}
					if dup {
						info.Duplicate = true
						sqlPanel.Duplicates++
						break
					}
				}
			}

			sqlPanel.Queries = append(sqlPanel.Queries, info)
			sqlPanel.TotalTime += duration
			sqlPanel.mu.Unlock()
		}
	}
}

// RegisterDriver registers the wrapped driver with database/sql
func RegisterDriver(name string, drv driver.Driver) {
	sql.Register(name, &DebugDriver{Driver: drv})
}
