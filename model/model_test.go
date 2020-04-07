package model_test

import (
	"path/filepath"
	"testing"

	"go.knocknote.io/eevee/class"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/model"
	_ "go.knocknote.io/eevee/plugin"
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
	if err := model.NewGenerator(cfg).Generate(classes); err != nil {
		t.Fatalf("%+v", err)
	}
}
