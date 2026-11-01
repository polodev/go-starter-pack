package config

import (
	"strings"

	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/support"
	mysqlfacades "github.com/goravel/mysql/facades"
	postgresfacades "github.com/goravel/postgres/facades"

	"goravel/app/facades"
)

// databaseName is DB_DATABASE, except inside `go test`: then DB_TEST_DATABASE (default <DB_DATABASE>_test).
// Tests refresh their database, so a test run refuses any name that does not end in _test. Goravel marks a test
// run before any config is read, which is why this check lives here and not in the tests.
func databaseName() string {
	config := facades.Config()
	name, _ := config.Env("DB_DATABASE", "").(string)
	if support.RuntimeMode != support.RuntimeTest {
		return name
	}
	testName, _ := config.Env("DB_TEST_DATABASE", "").(string)
	if testName == "" && name != "" {
		testName = name + "_test"
	}
	if !strings.HasSuffix(testName, "_test") {
		panic("tests only run against a database whose name ends in _test; set DB_TEST_DATABASE in .env")
	}
	return testName
}

func init() {
	config := facades.Config()
	config.Add("database", map[string]any{
		// Default database connection name
		"default": config.Env("DB_CONNECTION"),
		// Database connections
		"connections": map[string]any{
			"postgres": map[string]any{
				"host":     config.Env("DB_HOST"),
				"port":     config.Env("DB_PORT"),
				"database": databaseName(),
				"username": config.Env("DB_USERNAME"),
				"password": config.Env("DB_PASSWORD"),
				"sslmode":  "disable",
				"singular": false,
				"prefix":   "",
				"schema":   config.Env("DB_SCHEMA", "public"),
				"via": func() (driver.Driver, error) {
					return postgresfacades.Postgres("postgres")
				},
			},
			"mysql": map[string]any{
				"host":     config.Env("DB_HOST"),
				"port":     config.Env("DB_PORT"),
				"database": databaseName(),
				"username": config.Env("DB_USERNAME"),
				"password": config.Env("DB_PASSWORD"),
				"charset":  "utf8mb4",
				"prefix":   "",
				"singular": false,
				"via": func() (driver.Driver, error) {
					return mysqlfacades.Mysql("mysql")
				},
			},
		},
		// Pool configuration
		"pool": map[string]any{
			// Sets the maximum number of connections in the idle
			// connection pool.
			//
			// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
			// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
			//
			// If n <= 0, no idle connections are retained.
			"max_idle_conns": 10,
			// Sets the maximum number of open connections to the database.
			//
			// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
			// MaxIdleConns, then MaxIdleConns will be reduced to match the new
			// MaxOpenConns limit.
			//
			// If n <= 0, then there is no limit on the number of open connections.
			"max_open_conns": 100,
			// Sets the maximum amount of time a connection may be idle.
			//
			// Expired connections may be closed lazily before reuse.
			//
			// If d <= 0, connections are not closed due to a connection's idle time.
			// Unit: Second
			"conn_max_idletime": 3600,
			// Sets the maximum amount of time a connection may be reused.
			//
			// Expired connections may be closed lazily before reuse.
			//
			// If d <= 0, connections are not closed due to a connection's age.
			// Unit: Second
			"conn_max_lifetime": 3600,
		},
		// Sets the threshold for slow queries in milliseconds, the slow query will be logged.
		// Unit: Millisecond
		"slow_threshold": 200,
		// Migration Repository Table
		//
		// This table keeps track of all the migrations that have already run for
		// your application. Using this information, we can determine which of
		// the migrations on disk haven't actually been run in the database.
		"migrations": map[string]any{
			"table": "migrations",
		},
	})
}
