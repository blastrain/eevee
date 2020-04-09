package config

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"go.knocknote.io/eevee/plural"
	"go.knocknote.io/eevee/types"
	"golang.org/x/xerrors"
)

const (
	ConfigFilePath   = ".eevee.yml"
	DefaultDataStore = "db"
)

type Plugin struct {
	Name    string `yaml:"-"`
	Repo    string `yaml:"repo"`
	Version string `yaml:"version"`
}

type PrimitiveType struct {
	Name         string      `yaml:"name"`
	PackageName  string      `yaml:"package_name"`
	Import       string      `yaml:"import"`
	DefaultValue interface{} `yaml:"default"`
	As           string      `yaml:"as"`
}

type Config struct {
	ModulePath   string             `yaml:"module"`
	ClassPath    string             `yaml:"class,omitempty"`
	APIPath      string             `yaml:"api,omitempty"`
	SchemaPath   string             `yaml:"schema,omitempty"`
	GraphPath    string             `yaml:"graph,omitempty"`
	DocumentPath string             `yaml:"document,omitempty"`
	OutputPath   string             `yaml:"output,omitempty"`
	Plugins      map[string]*Plugin `yaml:"plugins,omitempty"`
	DAO          *DAO               `yaml:"dao,omitempty"`
	Entity       *Entity            `yaml:"entity,omitempty"`
	Model        *Model             `yaml:"model,omitempty"`
	Repository   *Repository        `yaml:"repository,omitempty"`
	Renderer     *Renderer          `yaml:"renderer,omitempty"`
	Plural       []*Plural          `yaml:"plural,omitempty"`
	Context      *Context           `yaml:"context,omitempty"`
	Types        []*PrimitiveType   `yaml:"primitive_types,omitempty"`
}

func (cfg *Config) OutputPathWithPackage(pkg string) string {
	return filepath.Join(cfg.OutputPath, pkg)
}

func (cfg *Config) TestDataPath() string {
	return filepath.Join(cfg.OutputPath, "testdata", "seeds")
}

func (cfg *Config) ContextImportPath() string {
	if cfg.Context == nil {
		return ""
	}
	return cfg.Context.Import
}

func (cfg *Config) DAOPackageName() string {
	defaultName := "dao"
	if cfg.DAO == nil {
		return defaultName
	}
	if cfg.DAO.Name != "" {
		return cfg.DAO.Name
	}
	return defaultName
}

func (cfg *Config) EntityPackageName() string {
	defaultName := "entity"
	if cfg.Entity == nil {
		return defaultName
	}
	if cfg.Entity.Name != "" {
		return cfg.Entity.Name
	}
	return defaultName
}

func (cfg *Config) ModelPackageName() string {
	defaultName := "model"
	if cfg.Model == nil {
		return defaultName
	}
	if cfg.Model.Name != "" {
		return cfg.Model.Name
	}
	return defaultName
}

func (cfg *Config) RepositoryPackageName() string {
	defaultName := "repository"
	if cfg.Repository == nil {
		return defaultName
	}
	if cfg.Repository.Name != "" {
		return cfg.Repository.Name
	}
	return defaultName
}

func (cfg *Config) RequestPackageName() string {
	defaultName := "request"
	return defaultName
}

func (cfg *Config) ResponsePackageName() string {
	defaultName := "response"
	return defaultName
}

func (cfg *Config) DataStore() string {
	if cfg.DAO != nil && cfg.DAO.Default != "" {
		return cfg.DAO.Default
	}
	return DefaultDataStore
}

func (cfg *Config) EntityPlugins() []string {
	if cfg.Entity == nil {
		return nil
	}
	return cfg.Entity.Plugins
}

type DataStore struct {
	Hooks map[string]interface{}
}

func (s *DataStore) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v map[string]interface{}
	if err := unmarshal(&v); err != nil {
		return xerrors.Errorf("cannot unmarshal datastore: %w", err)
	}
	s.Hooks = v
	return nil
}

type DAO struct {
	Name      string                `yaml:"name,omitempty"`
	Default   string                `yaml:"default,omitempty"`
	DataStore map[string]*DataStore `yaml:"datastore,omitempty"`
}

type Entity struct {
	Name    string   `yaml:"name,omitempty"`
	Plugins []string `yaml:"plugins,omitempty"`
}

type Model struct {
	Name string `yaml:"name,omitempty"`
}

type Repository struct {
	Name string `yaml:"name,omitempty"`
}

type RenderStyle string

const (
	RenderStyleLowerCamel RenderStyle = "lower-camel"
	RenderStyleUpperCamel RenderStyle = "upper-camel"
	RenderStyleLowerSnake RenderStyle = "lower-snake"
)

type Renderer struct {
	Style RenderStyle `yaml:"style,omitempty"`
}

type Plural struct {
	Name string `yaml:"name"`
	One  string `yaml:"one"`
}

type Context struct {
	Import string `yaml:"import"`
}

func ExistsConfig() bool {
	_, err := os.Stat(ConfigFilePath)
	return err == nil
}

func WriteConfig(cfg *Config) error {
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return xerrors.Errorf("failed to marshal YAML(%s): %w", string(bytes), err)
	}
	if err := ioutil.WriteFile(ConfigFilePath, bytes, 0644); err != nil {
		return xerrors.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func ConfigFromBytes(bytes []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, xerrors.Errorf("cannot unmarshal from %s to config: %w", string(bytes), err)
	}
	for name, plugin := range cfg.Plugins {
		plugin.Name = name
	}
	for _, p := range cfg.Plural {
		plural.Register(p.One, p.Name)
	}
	for _, typ := range cfg.Types {
		types.PrimitiveTypes[typ.Name] = &types.Type{
			Name:         typ.Name,
			PackageName:  typ.PackageName,
			ImportPath:   typ.Import,
			As:           typ.As,
			DefaultValue: typ.DefaultValue,
			IsPrimitive:  true,
		}
	}
	return &cfg, nil
}

func ConfigFromReader(r io.Reader) (*Config, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, xerrors.Errorf("cannot read file %s: %w", ConfigFilePath, err)
	}
	cfg, err := ConfigFromBytes(bytes)
	if err != nil {
		return nil, xerrors.Errorf("failed to get config from bytes: %w", err)
	}
	return cfg, nil
}

func ConfigFromPath(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, xerrors.Errorf("cannot read file %s: %w", ConfigFilePath, err)
	}
	cfg, err := ConfigFromBytes(bytes)
	if err != nil {
		return nil, xerrors.Errorf("failed to get config from bytes: %w", err)
	}
	return cfg, nil
}

func ReadConfig() (*Config, error) {
	cfg, err := ConfigFromPath(ConfigFilePath)
	if err != nil {
		return nil, xerrors.Errorf("failed to get config from path: %w", err)
	}
	return cfg, nil
}
