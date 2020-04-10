package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"go.knocknote.io/eevee"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/watcher"
	"golang.org/x/xerrors"
)

type Option struct {
	Init   InitCommand   `description:"create .eevee.yml for configuration"                              command:"init"`
	Run    RunCommand    `description:"generate files by referring to .eevee.yml"                        command:"run"`
	Plugin PluginCommand `description:"manage plugin"                                                    command:"plugin"`
	Watch  WatchCommand  `description:"watching changed files and regenerates (implemented by fsnotify)" command:"watch"`
	Serve  ServeCommand  `description:"serve dependency graph files"                                     command:"serve"`
}

type InitCommand struct {
	SchemaPath string `description:"schema file(or directory) path. try read 'sql' suffix file"          long:"schema" short:"s"`
	ClassPath  string `description:"generated class file(or directory) path. try read 'yml' suffix file" long:"class"  short:"c"`
	APIPath    string `description:"api definition file path. try read 'yml' suffix file"                long:"api"    short:"a"`
	GraphPath  string `description:"visualize relationships between tables"                              long:"graph"  short:"g"`
	OutputPath string `description:"specify an output directory of source code"                          long:"output" short:"o"`
}

type RunCommand struct {
	ConfigPath string `description:"config path" long:"config" short:"c"`
}

type PluginCommand struct {
	List    PluginListCommand    `description:"show installed plugins" command:"list"`
	Install PluginInstallCommand `description:"install plugin"         command:"install"`
	Remove  PluginRemoveCommand  `description:"remove plugin"          command:"remove"`
	Path    PluginPathCommand    `description:"show plugin directory"  command:"path"`
}

func pluginPath() string {
	return filepath.Join(eevee.InstalledDir(), "plugin")
}

type PluginListCommand struct {
}

func (cmd *PluginListCommand) Execute(args []string) error {
	pluginFile := filepath.Join(pluginPath(), "plugin.go")
	bytes, err := ioutil.ReadFile(pluginFile)
	if err != nil {
		return xerrors.Errorf("cannot read plugin.go %s: %w", pluginFile, err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", bytes, parser.ImportsOnly)
	if err != nil {
		return xerrors.Errorf("cannot parse to %s: %w", string(bytes), err)
	}
	plugins := []string{}
	for _, imported := range f.Imports {
		path := imported.Path.Value
		trimmedPath := path[1 : len(path)-1]
		pluginPrefix := "go.knocknote.io/eevee/plugin/"
		pluginName := trimmedPath[len(pluginPrefix):] // remove plugin prefix
		plugins = append(plugins, pluginName)
	}
	fmt.Printf("%s\n", strings.Join(plugins, "\n"))
	return nil
}

type PluginInstallCommand struct {
}

func (cmd *PluginInstallCommand) Execute(args []string) error {
	return nil
}

type PluginRemoveCommand struct {
}

func (cmd *PluginRemoveCommand) Execute(args []string) error {
	return nil
}

type PluginPathCommand struct {
}

func (cmd *PluginPathCommand) Execute(args []string) error {
	fmt.Printf("%s\n", pluginPath())
	return nil
}

type WatchCommand struct {
	ConfigPath string `description:"config path" long:"config" short:"c"`
}

type ServeCommand struct {
	ConfigPath string `description:"config path"        long:"config" short:"c"`
	Port       int    `description:"listen port number" long:"port"   short:"p"`
}

func (cmd *InitCommand) Execute(args []string) error {
	if config.ExistsConfig() {
		return xerrors.Errorf("already exists %s", config.ConfigFilePath)
	}
	modPath, err := eevee.ModulePath()
	if err != nil {
		return xerrors.Errorf("failed to read module name: %w", err)
	}
	if err := config.WriteConfig(&config.Config{
		ModulePath: modPath,
		SchemaPath: cmd.SchemaPath,
		ClassPath:  cmd.ClassPath,
		APIPath:    cmd.APIPath,
		GraphPath:  cmd.GraphPath,
		OutputPath: cmd.OutputPath,
	}); err != nil {
		return xerrors.Errorf("failed to write config: %w", err)
	}
	return nil
}

func (cmd *RunCommand) Execute(args []string) error {
	if !config.ExistsConfig() {
		return xerrors.Errorf("`eevee init` must be executed before `eevee run`")
	}
	cfg, err := config.ReadConfig()
	if err != nil {
		return xerrors.Errorf("failed to read %s: %w", config.ConfigFilePath, err)
	}
	if cfg.ModulePath == "" {
		return xerrors.Errorf("'module' value must be declared")
	}
	if err := eevee.Generate(cfg); err != nil {
		return xerrors.Errorf("failed to generate: %w", err)
	}
	return nil
}

func (cmd *WatchCommand) Execute(args []string) error {
	if !config.ExistsConfig() {
		return xerrors.Errorf("`eevee init` must be executed before `eevee run`")
	}
	cfg, err := config.ReadConfig()
	if err != nil {
		return xerrors.Errorf("failed to read %s: %w", config.ConfigFilePath, err)
	}
	if cfg.ModulePath == "" {
		return xerrors.Errorf("'module' value must be declared")
	}
	if err := watcher.New().Run(cfg); err != nil {
		return xerrors.Errorf("failed to watch: %w", err)
	}
	return nil
}

func (cmd *ServeCommand) Execute(args []string) error {
	if !config.ExistsConfig() {
		return xerrors.Errorf("`eevee init` must be executed before `eevee run`")
	}
	cfg, err := config.ReadConfig()
	if err != nil {
		return xerrors.Errorf("failed to read %s: %w", config.ConfigFilePath, err)
	}
	if cfg.ModulePath == "" {
		return xerrors.Errorf("'module' value must be declared")
	}
	graphPath := cfg.GraphPath
	fileServer := http.StripPrefix("/", http.FileServer(http.Dir(graphPath)))
	port := 3333
	if cmd.Port != 0 {
		port = cmd.Port
	}
	log.Printf("serving 127.0.0.1:%d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	}))
	return nil
}

var opts Option

//go:generate statik -src ../../static/resources -p static -dest ../.. -f -c '' -m

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.Parse()
}
