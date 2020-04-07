package class

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/goccy/go-yaml"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/types"
	"golang.org/x/xerrors"
)

type Reader struct {
}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) existsFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (r *Reader) ReadClass(cfg *config.Config, path string) (*types.Class, error) {
	if !r.existsFile(path) {
		return nil, nil
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, xerrors.Errorf("cannot read class file: %w", err)
	}
	var class types.Class
	if err := yaml.Unmarshal(bytes, &class); err != nil {
		return nil, xerrors.Errorf("cannot unmarshal from %s to class: %w", string(bytes), err)
	}
	if err := class.Validate(); err != nil {
		return nil, xerrors.Errorf("invalid class %s: %w", class.Name.SnakeName(), err)
	}
	if class.DataStore == "" {
		class.DataStore = cfg.DataStore()
	}
	if cfg.Renderer != nil {
		for _, member := range class.Members {
			if member.Render != nil {
				if !member.Render.IsRender {
					continue
				}
				if member.Render.DefaultName != "" {
					continue
				}
			}
			if member.Render == nil {
				member.Render = &types.Render{IsRender: true}
			}
			switch cfg.Renderer.Style {
			case config.RenderStyleLowerCamel:
				member.Render.DefaultName = member.Name.CamelLowerName()
			case config.RenderStyleUpperCamel:
				member.Render.DefaultName = member.Name.CamelName()
			case config.RenderStyleLowerSnake:
				member.Render.DefaultName = member.Name.SnakeName()
			}
		}
	}
	return &class, nil
}

func (r *Reader) ClassByConfig(cfg *config.Config) ([]*types.Class, error) {
	path := cfg.ClassPath
	if filepath.Ext(path) == "yml" {
		class, err := r.ReadClass(cfg, path)
		if err != nil {
			return nil, xerrors.Errorf("cannot read class file %s: %w", path, err)
		}
		return []*types.Class{class}, nil
	}

	var classes []*types.Class
	classMap := map[string]*types.Class{}
	ymlFilePattern := regexp.MustCompile(`\.yml$`)
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !ymlFilePattern.MatchString(path) {
			return nil
		}
		class, err := r.ReadClass(cfg, path)
		if err != nil {
			return xerrors.Errorf("cannot read class file %s: %w", path, err)
		}
		if class != nil {
			classes = append(classes, class)
		}
		class.SetClassMap(&classMap)
		classMap[class.Name.SnakeName()] = class
		return nil
	}); err != nil {
		return nil, xerrors.Errorf("interrupt walk in %s: %w", path, err)
	}
	for _, class := range classes {
		if err := class.ResolveTypeReference(); err != nil {
			return nil, xerrors.Errorf("cannot resolve type reference: %w", err)
		}
	}
	return classes, nil
}

type ClassMarshaler interface {
	FileName() string
	MarshalClass() ([]byte, error)
}

type Writer struct {
	path   string
	reader *Reader
}

func NewWriter(path string) (*Writer, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, xerrors.Errorf("cannot mkdir for %s: %w", path, err)
	}
	return &Writer{
		path:   path,
		reader: NewReader(),
	}, nil
}

func (w *Writer) Write(cfg *config.Config, schema *types.Class) error {
	path := filepath.Join(w.path, fmt.Sprintf("%s.yml", schema.Name.SnakeName()))
	class := schema
	if w.reader.existsFile(path) {
		definedClass, err := w.reader.ReadClass(cfg, path)
		if err != nil {
			return xerrors.Errorf("cannot read class file %s: %w", path, err)
		}
		definedClass.Merge(schema)
		class = definedClass
	}
	source, err := yaml.Marshal(class)
	if err != nil {
		return xerrors.Errorf("cannot marshal class: %w", err)
	}
	if err := ioutil.WriteFile(path, source, 0644); err != nil {
		return xerrors.Errorf("cannot write file to %s: %w", path, err)
	}
	return nil
}
