package config_test

import (
	"path/filepath"
	"testing"

	"go.knocknote.io/eevee/config"
)

func TestConfig(t *testing.T) {
	content := `
module: test
class:  config/class
api:    config/api
schema: schema
graph:  graph
document: docs
output: .
dao:
  name: infra
  default: db
  datastore:
    db:
      before-create: request-time
entity:
  name: entity
repository:
  name: repository
renderer:
  style: lower-camel
plural:
  - name: money
    one: moneys
context:
  import: app/context
`
	cfg, err := config.ConfigFromBytes([]byte(content))
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if cfg.OutputPathWithPackage("entity") != filepath.Join(".", "entity") {
		t.Fatalf("failed to get output path with package: %s", cfg.OutputPathWithPackage("entity"))
	}
}
