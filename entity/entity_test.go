package entity_test

import (
	"path/filepath"
	"testing"

	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/entity"
	_ "go.knocknote.io/eevee/plugin"
	"go.knocknote.io/eevee/class"
)

func TestGenerate(t *testing.T) {
	cfg := &config.Config{
		OutputPath: filepath.Join("testdata"),
		ClassPath:  filepath.Join("testdata", "class"),
	}
	classes, err := class.NewReader().ClassByConfig(cfg)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if err := entity.NewGenerator(cfg).Generate(classes); err != nil {
		t.Fatalf("%+v", err)
	}
}
