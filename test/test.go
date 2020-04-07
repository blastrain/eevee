package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/types"
	"golang.org/x/tools/imports"
	"golang.org/x/xerrors"
)

type Generator struct {
	appName    string
	importList types.ImportList
	cfg        *config.Config
}

func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		appName:    cfg.ModulePath,
		importList: types.DefaultImportList(cfg.ModulePath, cfg.ContextImportPath()),
		cfg:        cfg,
	}
}

func (g *Generator) existsFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (g *Generator) writeTestData(path string, class *types.Class) error {
	opt := yaml.MarshalAnchor(func(anchor *ast.AnchorNode, value interface{}) error {
		if o, ok := value.(*types.TestObjectDecl); ok {
			anchor.Name.(*ast.StringNode).Value = o.Name
		}
		return nil
	})
	defaultTestData := class.TestData()
	if !g.existsFile(path) {
		var buf bytes.Buffer
		if err := yaml.NewEncoder(&buf, opt).Encode(defaultTestData); err != nil {
			return xerrors.Errorf("failed to marshal default test data %s: %w", buf.String(), err)
		}
		if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
			return xerrors.Errorf("cannot write file %s: %w", path, err)
		}
		return nil
	}
	testData, err := g.readTestData(path)
	if err != nil {
		return xerrors.Errorf("failed to read yaml for test data: %w", err)
	}
	testData.MergeDefault(class)
	var buf bytes.Buffer
	if err := yaml.NewEncoder(&buf, opt).Encode(testData); err != nil {
		return xerrors.Errorf("failed to marshal default test data %s: %w", buf.String(), err)
	}
	if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return xerrors.Errorf("cannot write file %s: %w", path, err)
	}
	return nil
}

func (g *Generator) readTestData(path string) (*types.TestData, error) {
	if !g.existsFile(path) {
		return nil, nil
	}
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, xerrors.Errorf("failed to read test data file %s: %w", path, err)
	}
	var data types.TestData
	if err := yaml.Unmarshal(source, &data); err != nil {
		return nil, xerrors.Errorf("cannot unmarshal from %s to test data: %w", string(source), err)
	}
	return &data, nil
}

func (g *Generator) modelFactoryCode(object *types.TestObject, class *types.Class) []code.Code {
	dict := code.Dict{}
	beforeBlocks := []code.Code{}
	extendBlocks := []code.Code{}
	mapValue := object.MergedMapValue()
	for _, member := range class.Members {
		value := mapValue[member.Name.SnakeName()]
		if member.Relation != nil {
			relation := member.Relation
			if relation.Custom || relation.All {
				continue
			}
			name := value.(string)[1:]
			subClass := member.Type.Class()
			var (
				rvalueName string
				rtype      string
			)
			rvalueName = member.Name.CamelName()
			if member.HasMany {
				rtype = subClass.Name.PluralCamelName()
			} else {
				rtype = subClass.Name.CamelName()
			}
			var fname string
			if name == "default" || name == "defaults" {
				if member.HasMany {
					fname = fmt.Sprintf("Default%s", subClass.Name.PluralCamelName())
				} else {
					fname = fmt.Sprintf("Default%s", subClass.Name.CamelName())
				}
			} else {
				fname = types.Name(name).CamelName()
			}
			beforeBlocks = append(beforeBlocks, code.Id(member.Name.CamelLowerName()).Op(":=").Id(fname).Call())
			extendBlocks = append(extendBlocks,
				code.Id("value").Dot(rvalueName).Op("=").Func().Params(
					code.Qual(g.importList.Package("context"), "Context"),
				).Params(
					code.Op("*").Qual(g.importList.Package("model"), rtype),
					code.Id("error"),
				).Block(
					code.Return(code.Id(member.Name.CamelLowerName()), code.Nil()),
				),
			)
		} else if value != nil {
			defaultLayout := "2006-01-02T15:04:05Z"
			if member.Type.Type.IsTime() {
				if _, ok := value.(string); ok {
					beforeBlocks = append(beforeBlocks,
						code.List(code.Id(member.Name.CamelLowerName()), code.Id("_")).Op(":=").
							Qual("time", "Parse").Call(code.Lit(defaultLayout), code.Lit(value)),
					)
				} else {
					beforeBlocks = append(beforeBlocks,
						code.Id(member.Name.CamelLowerName()).Op(":=").
							Qual("time", "Unix").Call(code.Lit(int(value.(uint64))), code.Lit(0)),
					)
				}
				if member.Type.IsPointer {
					dict[code.Id(member.Name.CamelName())] = code.Op("&").Id(member.Name.CamelLowerName())
				} else {
					dict[code.Id(member.Name.CamelName())] = code.Id(member.Name.CamelLowerName())
				}
			} else {
				dict[code.Id(member.Name.CamelName())] = member.Type.ValueToCode(value)
			}
		}
	}
	blocks := []code.Code{}
	blocks = append(blocks, beforeBlocks...)
	blocks = append(blocks, []code.Code{
		code.Id("value").Op(":=").Op("&").Qual(g.importList.Package("model"), class.Name.CamelName()).Values(code.Dict{
			code.Id(class.Name.CamelName()): code.Op("&").Qual(g.importList.Package("entity"), class.Name.CamelName()).Values(dict),
		}),
	}...)
	blocks = append(blocks, extendBlocks...)
	return blocks
}

func (g *Generator) generateByTestObject(name string, object *types.TestObject, class *types.Class) code.Code {
	blocks := g.modelFactoryCode(object, class)
	var fname string
	if name == "default" {
		fname = fmt.Sprintf("Default%s", class.Name.CamelName())
	} else {
		fname = types.Name(name).CamelName()
	}
	blocks = append(blocks, code.Return(code.Id("value")))
	return code.Func().Id(fname).Params().Params(
		code.Op("*").Qual(g.importList.Package("model"), class.Name.CamelName()),
	).Block(blocks...)
}

func (g *Generator) generate(f *code.File, class *types.Class, testData *types.TestData) {
	singleKeys := []string{}
	collectionKeys := []string{}
	for k := range testData.Single {
		singleKeys = append(singleKeys, k)
	}
	for k := range testData.Collection {
		collectionKeys = append(collectionKeys, k)
	}
	sort.Strings(singleKeys)
	sort.Strings(collectionKeys)
	for _, k := range singleKeys {
		f.Add(g.generateByTestObject(k, testData.Single[k], class))
		f.Line()
	}
	for _, k := range collectionKeys {
		f.Add(g.generateByTestObjects(k, testData.Collection[k], class))
		f.Line()
	}
}

func (g *Generator) generateByTestObjects(name string, objects []*types.TestObject, class *types.Class) code.Code {
	blocks := []code.Code{
		code.Id("values").Op(":=").Op("&").Qual(g.importList.Package("model"), class.Name.PluralCamelName()).Values(),
	}
	for _, object := range objects {
		subBlocks := g.modelFactoryCode(object, class)
		subBlocks = append(subBlocks, code.Id("values").Dot("Add").Call(code.Id("value")))
		blocks = append(blocks, code.Block(subBlocks...))
	}
	blocks = append(blocks, code.Return(code.Id("values")))
	var fname string
	if name == "defaults" {
		fname = fmt.Sprintf("Default%s", class.Name.PluralCamelName())
	} else {
		fname = types.Name(name).CamelName()
	}
	return code.Func().Id(fname).Params().Params(code.Op("*").Qual(g.importList.Package("model"), class.Name.PluralCamelName())).Block(blocks...)
}

func (g *Generator) writeFile(class *types.Class, basePath string, source []byte) error {
	path := filepath.Join(basePath, fmt.Sprintf("%s.go", class.Name.SnakeName()))
	if g.existsFile(path) {
		if err := os.Remove(path); err != nil {
			return xerrors.Errorf("failed to remove file %s: %w", path, err)
		}
	}
	if err := ioutil.WriteFile(path, source, 0444); err != nil {
		return xerrors.Errorf("cannot write file %s: %w", path, err)
	}
	return nil
}

func (g *Generator) currentTestData(path string, class *types.Class) (*types.TestData, error) {
	seedPath := filepath.Join(path, fmt.Sprintf("%s.yml", class.Name.SnakeName()))
	testData, err := g.readTestData(seedPath)
	if err != nil {
		return nil, xerrors.Errorf("failed to write yaml for test data: %w", err)
	}
	return testData, nil
}

func (g *Generator) mockModelPath() string {
	return filepath.Join(g.cfg.OutputPath, "mock", "model", "factory")
}

func (g *Generator) GenerateMock(class *types.Class) error {
	path := g.cfg.TestDataPath()
	testData, err := g.currentTestData(path, class)
	if err != nil {
		return xerrors.Errorf("failed to get current testdata: %w", err)
	}
	f := code.NewFile("factory")
	f.HeaderComment(code.GeneratedMarker)
	for _, decl := range []*types.ImportDeclare{
		{
			Path: fmt.Sprintf("%s/entity", g.appName),
			Name: "entity",
		},
		{
			Path: fmt.Sprintf("%s/model", g.appName),
			Name: "model",
		},
	} {
		g.importList[decl.Name] = decl
	}
	for _, importDeclare := range g.importList {
		f.ImportName(importDeclare.Path, importDeclare.Name)
	}
	g.generate(f, class, testData)
	bytes := []byte(fmt.Sprintf("%#v", f))
	source, err := imports.Process("", bytes, nil)
	if err != nil {
		return xerrors.Errorf("cannot format by goimport %s: %w", string(bytes), err)
	}
	if err := g.writeFile(class, g.mockModelPath(), source); err != nil {
		return xerrors.Errorf("cannot write file for %s: %w", class.Name.SnakeName(), err)
	}
	return nil
}

func (g *Generator) Generate(classes []*types.Class) error {
	testDataPath := g.cfg.TestDataPath()
	if err := os.MkdirAll(testDataPath, 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", testDataPath, err)
	}
	if err := os.MkdirAll(g.mockModelPath(), 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", g.mockModelPath(), err)
	}
	for _, class := range classes {
		seedPath := filepath.Join(testDataPath, fmt.Sprintf("%s.yml", class.Name.SnakeName()))
		if err := g.writeTestData(seedPath, class); err != nil {
			return xerrors.Errorf("failed to write testdata: %w", err)
		}
	}
	for _, class := range classes {
		if err := g.GenerateMock(class); err != nil {
			return xerrors.Errorf("failed to generate mock/model: %w", err)
		}
	}
	return nil
}
