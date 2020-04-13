package eevee

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"go.knocknote.io/eevee/api"
	"go.knocknote.io/eevee/class"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/dao"
	"go.knocknote.io/eevee/entity"
	"go.knocknote.io/eevee/graph"
	"go.knocknote.io/eevee/model"
	_ "go.knocknote.io/eevee/plugin"
	"go.knocknote.io/eevee/repository"
	"go.knocknote.io/eevee/schema"
	"go.knocknote.io/eevee/test"
	"go.knocknote.io/eevee/types"
	"golang.org/x/xerrors"
)

func InstalledDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func ModulePath() (string, error) {
	if !existsGoMod() {
		return "", nil
	}
	content, err := ioutil.ReadFile(GoModPath)
	if err != nil {
		return "", xerrors.Errorf("failed to read go.mod: %w", err)
	}
	return parseModulePath(content), nil
}

func GenerateByClass(cfg *config.Config, class *types.Class) error {
	if err := GenerateByClasses(cfg, []*types.Class{class}); err != nil {
		return xerrors.Errorf("failed to generate by classes: %w", err)
	}
	return nil
}

func GenerateByClasses(cfg *config.Config, classes []*types.Class) error {
	if err := dao.NewGenerator(cfg).Generate(classes); err != nil {
		return xerrors.Errorf("failed to generate dao package: %w", err)
	}
	if err := repository.NewGenerator(cfg).Generate(classes); err != nil {
		return xerrors.Errorf("failed to generate repository package: %w", err)
	}
	if err := entity.NewGenerator(cfg).Generate(classes); err != nil {
		return xerrors.Errorf("failed to generate entity package: %w", err)
	}
	if err := model.NewGenerator(cfg).Generate(classes); err != nil {
		return xerrors.Errorf("failed to generate model package: %w", err)
	}
	if err := test.NewGenerator(cfg).Generate(classes); err != nil {
		return xerrors.Errorf("failed to generate testdata: %w", err)
	}
	if err := api.NewGenerator(cfg).Generate(classes); err != nil {
		return xerrors.Errorf("failed to generate api: %w", err)
	}
	if cfg.GraphPath != "" {
		if err := graph.Generate(cfg, classes); err != nil {
			return xerrors.Errorf("failed to generate graph: %w", err)
		}
	}
	return nil
}

func Generate(cfg *config.Config) error {
	schemata, err := getSchemata(cfg)
	if err != nil {
		return xerrors.Errorf("failed to get schemata: %w", err)
	}
	writer, err := class.NewWriter(cfg.ClassPath)
	if err != nil {
		return xerrors.Errorf("failed to initialize relation writer by %s: %w", cfg.ClassPath, err)
	}
	for _, schema := range schemata {
		class := schema.ToClass()
		class.DataStore = cfg.DataStore()
		if err := writer.Write(cfg, class); err != nil {
			return xerrors.Errorf("failed to write by schema: %w", err)
		}
	}
	reader := class.NewReader()
	classes, err := reader.ClassByConfig(cfg)
	if err != nil {
		return xerrors.Errorf("failed to read relation file from %s: %w", cfg.ClassPath, err)
	}
	if err := GenerateByClasses(cfg, classes); err != nil {
		return xerrors.Errorf("failed to generate by class: %w", err)
	}
	return nil
}

func getSchemata(cfg *config.Config) ([]*schema.Schema, error) {
	reader := schema.NewReader()
	schemata, err := reader.SchemaFromPath(cfg.SchemaPath)
	if err != nil {
		return nil, xerrors.Errorf("failed to read schema files from %s: %w", cfg.SchemaPath, err)
	}
	return schemata, nil
}

const GoModPath = "go.mod"

func existsGoMod() bool {
	_, err := os.Stat(GoModPath)
	return err == nil
}

var (
	slashSlash = []byte("//")
	moduleStr  = []byte("module")
)

// this original source is cmd/go/internal/modfile/read.go
func parseModulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}
