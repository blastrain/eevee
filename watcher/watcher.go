package watcher

import (
	"log"
	"path/filepath"

	eevee "go.knocknote.io/eevee"
	"go.knocknote.io/eevee/class"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/schema"
	"go.knocknote.io/eevee/test"
	"go.knocknote.io/eevee/types"
	"golang.org/x/xerrors"
	"gopkg.in/fsnotify.v1"
)

type Watcher struct {
	schemaWatcher   *fsnotify.Watcher
	classWatcher    *fsnotify.Watcher
	testDataWatcher *fsnotify.Watcher
	watchingCh      chan bool
	classMap        map[string]*types.Class
}

func (w *Watcher) validateFilename(path string) bool {
	filename := filepath.Base(path)
	if len(filename) == 0 {
		return false
	}
	prefix := string(filename[0])
	if prefix == "#" || prefix == "." {
		return false
	}
	return true
}

func (w *Watcher) validateSchema(path string) bool {
	ext := filepath.Ext(path)
	if ext != ".sql" {
		return false
	}
	return w.validateFilename(path)
}

func (w *Watcher) validateClass(path string) bool {
	ext := filepath.Ext(path)
	if ext != ".yml" {
		return false
	}
	return w.validateFilename(path)
}

func (w *Watcher) validateTestData(path string) bool {
	ext := filepath.Ext(path)
	if ext != ".yml" {
		return false
	}
	return w.validateFilename(path)
}

func (w *Watcher) recoverRuntimeError() {
	if err := recover(); err != nil {
		log.Printf("%+v", err)
	}
}

func (w *Watcher) generateClassBySchemaPath(cfg *config.Config, path string) error {
	writer, err := class.NewWriter(cfg.ClassPath)
	if err != nil {
		return xerrors.Errorf("failed to create class writer from %s: %w", cfg.ClassPath)
	}
	reader := schema.NewReader()
	schemata, err := reader.SchemaFromPath(path)
	if err != nil {
		return xerrors.Errorf("failed to read schema files from %s: %w", path, err)
	}
	for _, schema := range schemata {
		if err := writer.Write(cfg, schema.ToClass()); err != nil {
			return xerrors.Errorf("failed to write by schema: %w", err)
		}
	}
	return nil
}

func (w *Watcher) generateByClassPath(cfg *config.Config, path string) error {
	class, err := class.NewReader().ReadClass(cfg, path)
	if err != nil {
		return xerrors.Errorf("failed to read class file %s: %w", path, err)
	}
	if err := eevee.GenerateByClass(cfg, class); err != nil {
		return xerrors.Errorf("failed to generate by class %s: %w", class.Name.SnakeName(), err)
	}
	return nil
}

func (w *Watcher) classByName(classes []*types.Class, name string) *types.Class {
	for _, class := range classes {
		if class.Name.SnakeName() == name {
			return class
		}
	}
	return nil
}

func (w *Watcher) generateByTestDataPath(cfg *config.Config, path string) error {
	reader := class.NewReader()
	classes, err := reader.ClassByConfig(cfg)
	if err != nil {
		return xerrors.Errorf("failed to read relation file from %s: %w", cfg.ClassPath, err)
	}
	fileName := filepath.Base(path)
	className := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	class := w.classByName(classes, className)
	if err := test.NewGenerator(cfg).GenerateMock(class); err != nil {
		return xerrors.Errorf("failed to generate testdata: %w", err)
	}
	return nil
}

func (w *Watcher) updateSchema(cfg *config.Config, path string) error {
	if !w.validateSchema(path) {
		return nil
	}
	log.Printf("modified: %s", path)
	if err := w.generateClassBySchemaPath(cfg, path); err != nil {
		return xerrors.Errorf("failed to generate class file by schema path %s: %w", path, err)
	}
	return nil
}

func (w *Watcher) createSchema(cfg *config.Config, path string) error {
	if !w.validateSchema(path) {
		return nil
	}
	log.Printf("created: %s", path)
	if err := w.generateClassBySchemaPath(cfg, path); err != nil {
		return xerrors.Errorf("failed to generate class file by schema path %s: %w", path, err)
	}
	return nil
}

func (w *Watcher) updateClass(cfg *config.Config, path string) error {
	if !w.validateClass(path) {
		return nil
	}
	log.Printf("modified: %s\n", path)
	if err := w.generateByClassPath(cfg, path); err != nil {
		return xerrors.Errorf("failed to generate by class path %s: %w", path, err)
	}
	return nil
}

func (w *Watcher) createClass(cfg *config.Config, path string) error {
	if !w.validateClass(path) {
		return nil
	}
	log.Printf("created: %s\n", path)
	if err := w.generateByClassPath(cfg, path); err != nil {
		return xerrors.Errorf("failed to generate by class path %s: %w", path, err)
	}
	return nil
}

func (w *Watcher) updateTestData(cfg *config.Config, path string) error {
	if !w.validateTestData(path) {
		return nil
	}
	log.Printf("modified: %s\n", path)
	if err := w.generateByTestDataPath(cfg, path); err != nil {
		return xerrors.Errorf("failed to generate by testdata path %s: %w", path, err)
	}
	return nil
}

func (w *Watcher) createTestData(cfg *config.Config, path string) error {
	if !w.validateTestData(path) {
		return nil
	}
	log.Printf("created: %s\n", path)
	if err := w.generateByTestDataPath(cfg, path); err != nil {
		return xerrors.Errorf("failed to generate by testdata path %s: %w", path, err)
	}
	return nil
}

func (w *Watcher) watchSchema(cfg *config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return xerrors.Errorf("failed to create fsnotify instance: %w", err)
	}
	if err := watcher.Add(cfg.SchemaPath); err != nil {
		return xerrors.Errorf("failed to add path %s: %w", cfg.SchemaPath, err)
	}
	go func() {
		defer w.recoverRuntimeError()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := w.updateSchema(cfg, event.Name); err != nil {
						log.Printf("%+v", err)
					}
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if err := w.createSchema(cfg, event.Name); err != nil {
						log.Printf("%+v", err)
					}
				}
			case err := <-watcher.Errors:
				log.Printf("%+v", err)
			}
		}
	}()
	w.schemaWatcher = watcher
	return nil
}

func (w *Watcher) watchClass(cfg *config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return xerrors.Errorf("failed to create fsnotify instance: %w", err)
	}
	if err := watcher.Add(cfg.ClassPath); err != nil {
		return xerrors.Errorf("failed to add path %s: %w", cfg.ClassPath, err)
	}
	go func() {
		defer w.recoverRuntimeError()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := w.updateClass(cfg, event.Name); err != nil {
						log.Printf("%+v", err)
					}
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if err := w.createClass(cfg, event.Name); err != nil {
						log.Printf("%+v", err)
					}
				}
			case err := <-watcher.Errors:
				log.Printf("%+v", err)
			}
		}
	}()
	w.classWatcher = watcher
	return nil
}

func (w *Watcher) watchTestData(cfg *config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return xerrors.Errorf("failed to create fsnotify instance: %w", err)
	}
	if err := watcher.Add(cfg.TestDataPath()); err != nil {
		return xerrors.Errorf("failed to add path %s: %w", cfg.TestDataPath(), err)
	}
	go func() {
		defer w.recoverRuntimeError()
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := w.updateTestData(cfg, event.Name); err != nil {
						log.Printf("%+v", err)
					}
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if err := w.createTestData(cfg, event.Name); err != nil {
						log.Printf("%+v", err)
					}
				}
			case err := <-watcher.Errors:
				log.Printf("%+v", err)
			}
		}
	}()
	w.testDataWatcher = watcher
	return nil
}

func (w *Watcher) Run(cfg *config.Config) error {
	if err := w.watchSchema(cfg); err != nil {
		return xerrors.Errorf("failed to watch schema: %w", err)
	}
	if err := w.watchClass(cfg); err != nil {
		return xerrors.Errorf("failed to watch class: %w", err)
	}
	if err := w.watchTestData(cfg); err != nil {
		return xerrors.Errorf("failed to watch testdata: %w", err)
	}
	log.Println("watching...")
	<-w.watchingCh
	return nil
}

func (w *Watcher) Stop() {
	w.watchingCh <- true
}

func (w *Watcher) Close() {
	if w.schemaWatcher != nil {
		w.schemaWatcher.Close()
	}
	if w.classWatcher != nil {
		w.classWatcher.Close()
	}
	if w.testDataWatcher != nil {
		w.testDataWatcher.Close()
	}
}

func New() *Watcher {
	return &Watcher{watchingCh: make(chan bool)}
}
