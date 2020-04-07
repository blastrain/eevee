package repository_test

import (
	"path/filepath"
	"testing"

	"go.knocknote.io/eevee/class"
	"go.knocknote.io/eevee/config"
	_ "go.knocknote.io/eevee/plugin"
	"go.knocknote.io/eevee/repository"
)

func TestGenerate(t *testing.T) {
	cfg := &config.Config{
		ClassPath:  filepath.Join("testdata", "class"),
		OutputPath: filepath.Join("testdata"),
	}
	classes, err := class.NewReader().ClassByConfig(cfg)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if err := repository.NewGenerator(cfg).Generate(classes); err != nil {
		t.Fatalf("%+v", err)
	}
}
