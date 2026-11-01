package tests

import (
	"strings"

	"github.com/goravel/framework/testing"

	"goravel/app/facades"
	"goravel/bootstrap"
)

func init() {
	bootstrap.Boot()
	// config/database.go switches to the test database inside `go test`; make sure it did before any test runs.
	connection := facades.Config().GetString("database.default")
	name := facades.Config().GetString("database.connections." + connection + ".database")
	if !strings.HasSuffix(name, "_test") {
		panic("tests are connected to " + name + ", not a *_test database")
	}
}

type TestCase struct {
	testing.TestCase
}
