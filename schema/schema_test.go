package schema

import (
	"testing"
)

func TestSchema(t *testing.T) {
	reader := NewReader()
	_, err := reader.SchemaFromPath("testdata/schema")
	if err != nil {
		t.Fatalf("%+v", err)
	}
}
