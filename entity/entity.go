package entity

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/config"
	. "go.knocknote.io/eevee/plugin/entity"
	"go.knocknote.io/eevee/types"
	"golang.org/x/xerrors"
)

type Generator struct {
	appName      string
	packageName  string
	receiverName string
	importList   types.ImportList
	cfg          *config.Config
}

func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		appName:      cfg.ModulePath,
		packageName:  cfg.EntityPackageName(),
		receiverName: "e",
		importList:   types.DefaultImportList(cfg.ModulePath, cfg.ContextImportPath()),
		cfg:          cfg,
	}
}

func (g *Generator) addMethod(f *File, mtd *types.Method) {
	f.Line()
	f.Add(mtd.Generate(g.importList))
}

func (g *Generator) sliceMethod(class *types.Class, member *types.Member) *types.Method {
	sliceName := class.Name.PluralCamelName()
	return &types.Method{
		Decl: &types.MethodDeclare{
			ReceiverName:         g.receiverName,
			ReceiverClassName:    sliceName,
			MethodName:           member.Name.PluralCamelName(),
			IsNotPointerReceiver: true,
			Return: types.ValueDeclares{
				{
					Type: types.TypeDeclareWithName(fmt.Sprintf("[]%#v", member.Type.CodePackage(g.packageName, g.importList))),
				},
			},
		},
		Body: []Code{
			Id("values").Op(":=").Make(Index().Add(member.Type.CodePackage(g.packageName, g.importList)), Lit(0), Len(Id("e"))),
			For(List(Id("_"), Id("value")).Op(":=").Range().Id("e")).Block(
				Id("values").Op("=").Append(Id("values"), Id("value").Dot(member.Name.CamelName())),
			),
			Return(Id("values")),
		},
	}
}

func (g *Generator) structCodes(class *types.Class) []Code {
	codes := []Code{}
	for _, member := range class.Members {
		if member.Extend {
			continue
		}
		tag := map[string]string{}
		for _, proto := range member.RenderProtocols() {
			tag[proto] = member.RenderNameByProtocol(proto)
		}
		field := Id(member.Name.CamelName()).Add(member.Type.CodePackage(g.packageName, g.importList))
		if len(tag) > 0 {
			field = field.Tag(tag)
		}
		codes = append(codes, field)
	}
	return codes
}

func (g *Generator) generate(class *types.Class, path string) ([]byte, error) {
	for _, pluginName := range g.cfg.EntityPlugins() {
		plg, ok := Plugin(pluginName)
		if !ok {
			continue
		}
		g.importList = plg.Imports(g.importList)
	}
	f := NewGeneratedFile(g.packageName, g.importList)
	entityName := class.Name.CamelName()
	sliceName := class.Name.PluralCamelName()
	AddStruct(f, entityName, g.structCodes(class))
	TypeDef(f, sliceName, Index().Op("*").Id(entityName))
	for _, member := range class.Members {
		if member.Extend {
			continue
		}
		g.addMethod(f, g.sliceMethod(class, member))
	}
	for _, pluginName := range g.cfg.EntityPlugins() {
		plg, ok := Plugin(pluginName)
		if !ok {
			continue
		}
		mtds := plg.AddMethods(&types.EntityMethodHelper{
			Class:        class,
			ReceiverName: g.receiverName,
			ImportList:   g.importList,
		})
		for _, mtd := range mtds {
			g.addMethod(f, mtd)
		}
	}
	return []byte(fmt.Sprintf("%#v", f)), nil
}

func (g *Generator) existsFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (g *Generator) writeFile(class *types.Class, basePath string, source []byte) error {
	path := filepath.Join(basePath, fmt.Sprintf("%s.go", class.Name.SnakeName()))
	if g.existsFile(path) {
		if err := os.Remove(path); err != nil {
			return xerrors.Errorf("failed to remove file %s: %w", path, err)
		}
	}
	if err := ioutil.WriteFile(path, source, 0444); err != nil {
		return xerrors.Errorf("cannot write file to %s: %w", path, err)
	}
	return nil
}

func (g *Generator) Generate(classes []*types.Class) error {
	path := g.cfg.OutputPathWithPackage(g.packageName)
	if err := os.MkdirAll(path, 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", path, err)
	}
	for _, class := range classes {
		source, err := g.generate(class, path)
		if err != nil {
			return xerrors.Errorf("cannot generate entity for %s: %w", class.Name.SnakeName(), err)
		}
		if err := g.writeFile(class, path, source); err != nil {
			return xerrors.Errorf("cannot write file to %s for %s: %w", path, class.Name.SnakeName(), err)
		}
	}
	return nil
}
