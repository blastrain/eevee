package model

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/dao"
	"go.knocknote.io/eevee/renderer"
	"go.knocknote.io/eevee/types"
	"golang.org/x/tools/imports"
	"golang.org/x/xerrors"
)

type Generator struct {
	packageName  string
	receiverName string
	daoPath      string
	importList   types.ImportList
	cfg          *config.Config
}

func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		packageName:  cfg.ModelPackageName(),
		receiverName: "m",
		importList:   types.DefaultImportList(cfg.ModulePath, cfg.ContextImportPath()),
		cfg:          cfg,
	}
}

func (g *Generator) helper(class *types.Class) *types.ModelMethodHelper {
	return &types.ModelMethodHelper{
		Class:        class,
		ReceiverName: g.receiverName,
		ImportList:   g.importList,
	}
}

func (g *Generator) addMethod(f *code.File, mtd *types.Method) {
	if mtd == nil {
		return
	}
	f.Line()
	f.Add(mtd.Generate(g.importList))
}

func (g *Generator) addRendererMethod(f *code.File, class *types.Class, r renderer.Renderer) {
	g.addMethod(f, r.Render(g.helper(class)))
	g.addMethod(f, r.RenderWithOption(g.helper(class)))
	g.addMethod(f, r.RenderCollection(g.helper(class)))
	g.addMethod(f, r.RenderCollectionWithOption(g.helper(class)))
	g.addMethod(f, r.Marshaler(g.helper(class)))
	g.addMethod(f, r.MarshalerContext(g.helper(class)))
	g.addMethod(f, r.MarshalerCollection(g.helper(class)))
	g.addMethod(f, r.MarshalerCollectionContext(g.helper(class)))
	g.addMethod(f, r.Unmarshaler(g.helper(class)))
	g.addMethod(f, r.UnmarshalerCollection(g.helper(class)))
}

func (g *Generator) generate(class *types.Class, path string) ([]byte, error) {
	f := code.NewFile(g.packageName)
	f.HeaderComment(code.GeneratedMarker)
	for _, importDeclare := range g.importList {
		f.ImportName(importDeclare.Path, importDeclare.Name)
	}
	daoGenerator := dao.NewGenerator(g.cfg)
	daoPackageDecl, err := daoGenerator.PackageDeclare(class, g.daoPath)
	if err != nil {
		return nil, xerrors.Errorf("cannot create package declaration for dao(%s): %w", class.Name.SnakeName(), err)
	}
	interfaceBody := []code.Code{}
	for _, method := range daoPackageDecl.Methods {
		if !strings.HasPrefix(method.MethodName, "Find") {
			continue
		}
		decl := &types.MethodDeclare{
			MethodName: method.MethodName,
			Args:       method.Args,
			Return:     types.ValueDeclares{},
		}
		for _, retDecl := range method.Return {
			name := retDecl.Type.Type.Name
			if name == class.Name.CamelName() || name == class.Name.PluralCamelName() {
				decl.Return = append(decl.Return, &types.ValueDeclare{
					Type: &types.TypeDeclare{
						Type: &types.Type{
							Name: name,
						},
						IsPointer: true,
					},
				})
			} else {
				decl.Return = append(decl.Return, &types.ValueDeclare{
					Type: types.TypeDeclareWithType(&types.Type{
						Name: name,
					}),
				})
			}
		}
		interfaceBody = append(interfaceBody, decl.Interface(g.importList))
	}
	f.Add(code.GoType().Id(fmt.Sprintf("%sFinder", class.Name.CamelName())).Interface(interfaceBody...))
	structFields := []code.Code{code.Op("*").Qual(g.importList.Package("entity"), class.Name.CamelName())}
	structFields = append(structFields,
		code.Id(fmt.Sprintf("%sDAO", class.Name.CamelLowerName())).Qual(g.importList.Package("dao"), class.Name.CamelName()),
	)
	for _, member := range class.ExtendMembers() {
		structFields = append(structFields, code.Id(member.Name.CamelName()).Add(member.Type.Code(g.importList)))
	}
	for _, member := range class.RelationMembers() {
		if member.Relation.Custom {
			continue
		}
		var rtype string
		if member.HasMany || member.Relation.All {
			rtype = member.Type.Class().Name.PluralCamelName()
		} else {
			rtype = member.Type.Class().Name.CamelName()
		}
		structFields = append(structFields,
			code.Id(member.Name.CamelName()).Func().Params(code.Qual(g.importList.Package("context"), "Context")).
				Params(code.Op("*").Id(rtype), code.Id("error")),
		)
	}
	structFields = append(structFields, code.Id("isAlreadyCreated").Bool())
	structFields = append(structFields, code.Id("savedValue").Qual(g.importList.Package("entity"), class.Name.CamelName()))
	structFields = append(structFields, code.Id("conv").Id("ModelConverter"))
	f.Line()
	f.Add(code.GoType().Id(class.Name.CamelName()).Struct(structFields...))

	collectionStructFields := []code.Code{code.Id("values").Index().Op("*").Id(class.Name.CamelName())}
	for _, value := range g.helper(class).CollectionProperties() {
		collectionStructFields = append(collectionStructFields, value.Code(g.importList))
	}
	f.Line()
	f.Add(code.GoType().Id(class.Name.PluralCamelName()).Struct(collectionStructFields...))
	f.Line()
	f.Add(code.GoType().Id(fmt.Sprintf("%sCollection", class.Name.PluralCamelName())).Index().Op("*").Id(class.Name.PluralCamelName()))

	f.Line()
	f.Add(g.Constructor(g.helper(class)))
	f.Line()
	f.Add(g.CollectionConstructor(g.helper(class)))
	for _, mtdFn := range g.Methods() {
		g.addMethod(f, mtdFn(g.helper(class)))
	}
	g.addRendererMethod(f, class, &renderer.JSONRenderer{})
	g.addRendererMethod(f, class, &renderer.MapRenderer{})
	g.addMethod(f, g.SetConverter(g.helper(class)))
	if !class.ReadOnly {
		g.addMethod(f, g.Create(g.helper(class)))
		g.addMethod(f, g.Update(g.helper(class)))
		if class.MemberByName("id") != nil {
			g.addMethod(f, g.Delete(g.helper(class)))
		}
		g.addMethod(f, g.SetAlreadyCreated(g.helper(class)))
		g.addMethod(f, g.SetSavedValue(g.helper(class)))
		g.addMethod(f, g.Save(g.helper(class)))
		g.addMethod(f, g.CreateForCollection(g.helper(class)))
		g.addMethod(f, g.UpdateForCollection(g.helper(class)))
		g.addMethod(f, g.SaveForCollection(g.helper(class)))
	}
	for _, member := range class.Members {
		if member.Relation == nil && !member.Extend {
			g.addMethod(f, g.Unique(g.helper(class), member))
			g.addMethod(f, g.GroupBy(g.helper(class), member))
			g.addMethod(f, g.Collection(g.helper(class), member))
		}
		if member.Relation != nil && !member.Relation.All {
			if !member.Relation.Custom {
				g.addMethod(f, g.findBy(g.helper(class), member))
			}
			g.addMethod(f, g.Collection(g.helper(class), member))
		} else if member.Relation != nil && member.Relation.All {
			g.addMethod(f, g.findAll(g.helper(class), member))
		}
	}
	primaryKey := class.PrimaryKey()
	definedKeyPair := map[string]struct{}{}
	if primaryKey != nil {
		definedKeyPair[primaryKey.Name.SnakeName()] = struct{}{}
		g.addMethod(f, g.FirstBy(g.helper(class), types.Members{primaryKey}))
		g.addMethod(f, g.FilterBy(g.helper(class), types.Members{primaryKey}))
	}
	for _, uniqueKey := range class.UniqueKeys() {
		if _, exists := definedKeyPair[uniqueKey.JoinedName()]; !exists {
			g.addMethod(f, g.FirstBy(g.helper(class), uniqueKey))
			g.addMethod(f, g.FilterBy(g.helper(class), uniqueKey))
			definedKeyPair[uniqueKey.JoinedName()] = struct{}{}
		}
		if len(uniqueKey) < 2 {
			continue
		}
		uniqueKey = uniqueKey[:len(uniqueKey)-1]
		for i := len(uniqueKey); i > 0; i-- {
			if _, exists := definedKeyPair[uniqueKey.JoinedName()]; !exists {
				g.addMethod(f, g.FirstBy(g.helper(class), uniqueKey))
				g.addMethod(f, g.FilterBy(g.helper(class), uniqueKey))
				definedKeyPair[uniqueKey.JoinedName()] = struct{}{}
			}
			uniqueKey = uniqueKey[:len(uniqueKey)-1]
		}
	}
	for _, key := range class.Keys() {
		if _, exists := definedKeyPair[key.JoinedName()]; !exists {
			g.addMethod(f, g.FirstBy(g.helper(class), key))
			g.addMethod(f, g.FilterBy(g.helper(class), key))
			definedKeyPair[key.JoinedName()] = struct{}{}
		}
		if len(key) < 2 {
			continue
		}
		key = key[:len(key)-1]
		for i := len(key); i > 0; i-- {
			if _, exists := definedKeyPair[key.JoinedName()]; !exists {
				g.addMethod(f, g.FirstBy(g.helper(class), key))
				g.addMethod(f, g.FilterBy(g.helper(class), key))
				definedKeyPair[key.JoinedName()] = struct{}{}
			}
			key = key[:len(key)-1]
		}
	}
	bytes := []byte(fmt.Sprintf("%#v", f))
	source, err := imports.Process("", bytes, nil)
	if err != nil {
		return nil, xerrors.Errorf("cannot format by goimport: %w", err)
	}
	return source, nil
}

func (g *Generator) generateModelClass(path string, classes []*types.Class) error {
	f := code.NewFile(g.packageName)
	f.HeaderComment(code.GeneratedMarker)
	for _, importDeclare := range g.importList {
		f.ImportName(importDeclare.Path, importDeclare.Name)
	}
	interfaceBody := []code.Code{}
	for _, class := range classes {
		decl := &types.MethodDeclare{
			Class:      class,
			ImportList: g.importList,
			MethodName: fmt.Sprintf("To%s", class.Name.CamelName()),
			Args: types.ValueDeclares{
				{
					Name: "value",
					Type: &types.TypeDeclare{
						Type: &types.Type{
							PackageName: g.importList.Package("entity"),
							Name:        class.Name.CamelName(),
						},
						IsPointer: true,
					},
				},
			},
			Return: types.ValueDeclares{
				{
					Type: &types.TypeDeclare{
						Type: &types.Type{
							Name: class.Name.CamelName(),
						},
						IsPointer: true,
					},
				},
			},
		}
		interfaceBody = append(interfaceBody, decl.Interface(g.importList))
	}
	f.Add(code.GoType().Id("ModelConverter").Interface(interfaceBody...))
	g.generateRenderOption(f)
	source := []byte(fmt.Sprintf("%#v", f))
	modelGoPath := filepath.Join(path, "model.go")
	if g.existsFile(modelGoPath) {
		if err := os.Remove(modelGoPath); err != nil {
			return xerrors.Errorf("failed to remove file %s: %w", modelGoPath, err)
		}
	}
	if err := ioutil.WriteFile(modelGoPath, source, 0444); err != nil {
		return xerrors.Errorf("cannot write file %s: %w", path, err)
	}
	return nil
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
		return xerrors.Errorf("cannot write file %s: %w", path, err)
	}
	return nil
}

func (g *Generator) Generate(classes []*types.Class) error {
	path := g.cfg.OutputPathWithPackage(g.packageName)
	if err := os.MkdirAll(path, 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", path, err)
	}
	splittedPaths := strings.Split(path, string(filepath.Separator))
	daoPaths := splittedPaths[:len(splittedPaths)-1]
	daoPaths = append(daoPaths, "dao")
	daoPath := filepath.Join(daoPaths...)
	g.daoPath = daoPath
	for _, class := range classes {
		source, err := g.generate(class, path)
		if err != nil {
			return xerrors.Errorf("cannot generate model for %s: %w", class.Name.SnakeName(), err)
		}
		if err := g.writeFile(class, path, source); err != nil {
			return xerrors.Errorf("cannot write file to %s for %s: %w", path, class.Name.SnakeName(), err)
		}
	}
	if err := g.generateModelClass(path, classes); err != nil {
		return xerrors.Errorf("cannot generate model.go: %w", err)
	}
	return nil
}
