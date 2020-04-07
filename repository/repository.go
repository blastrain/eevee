package repository

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/dao"
	"go.knocknote.io/eevee/types"
	"golang.org/x/tools/imports"
	"golang.org/x/xerrors"
)

type Generator struct {
	packageName    string
	receiverName   string
	daoPath        string
	importList     types.ImportList
	constructorMap map[*types.Class]*types.ConstructorDeclare
	cfg            *config.Config
}

func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		packageName:    cfg.RepositoryPackageName(),
		receiverName:   "r",
		importList:     types.DefaultImportList(cfg.ModulePath, cfg.ContextImportPath()),
		constructorMap: map[*types.Class]*types.ConstructorDeclare{},
		cfg:            cfg,
	}
}

func (g *Generator) helper(class *types.Class) *types.RepositoryMethodHelper {
	return &types.RepositoryMethodHelper{
		AppName:      g.cfg.ModulePath,
		Class:        class,
		ReceiverName: g.receiverName,
		ImportList:   g.importList,
	}
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
	subClassConstructorMap := map[*types.Class]*types.ConstructorDeclare{}
	for _, member := range class.RelationMembers() {
		relation := member.Relation
		if relation.Custom {
			continue
		}
		subClass := member.Type.Class()
		if subClass == nil {
			continue
		}
		pkgDecl, err := daoGenerator.PackageDeclare(subClass, g.daoPath)
		if err != nil {
			return nil, xerrors.Errorf("cannot create package declaration for dao(%s): %w", subClass.Name.SnakeName(), err)
		}
		subClassConstructorMap[subClass] = pkgDecl.Constructor
	}
	toModelMethod := g.ToModel(g.helper(class))
	toModelsMethod := g.ToModels(g.helper(class))
	interfaceBody := []code.Code{
		toModelMethod.Decl.Interface(g.importList),
		toModelsMethod.Decl.Interface(g.importList),
	}
	createMethods := []*types.Method{}
	findMethods := []*types.Method{}
	updateByMethods := []*types.Method{}
	deleteByMethods := []*types.Method{}
	otherMethods := []*types.Method{}
	methodNameMap := map[string]struct{}{}
	for _, method := range daoPackageDecl.Methods {
		if !strings.HasPrefix(method.MethodName, "Find") {
			continue
		}
		methodNameMap[method.MethodName] = struct{}{}
		findMethods = append(findMethods, g.FindBy(g.helper(class), method))
	}
	if !class.ReadOnly {
		createMethods = append(createMethods, g.Create(g.helper(class)))
		createMethods = append(createMethods, g.Creates(g.helper(class)))
		methodNameMap["Create"] = struct{}{}
		methodNameMap["Creates"] = struct{}{}
		for _, method := range daoPackageDecl.Methods {
			if !strings.HasPrefix(method.MethodName, "UpdateBy") {
				continue
			}
			methodNameMap[method.MethodName] = struct{}{}
			updateByMethods = append(updateByMethods, g.UpdateBy(g.helper(class), method))
		}
		for _, method := range daoPackageDecl.Methods {
			if !strings.HasPrefix(method.MethodName, "DeleteBy") {
				continue
			}
			methodNameMap[method.MethodName] = struct{}{}
			deleteByMethods = append(deleteByMethods, g.DeleteBy(g.helper(class), method))
		}
	}
	for _, method := range daoPackageDecl.Methods {
		if _, exists := methodNameMap[method.MethodName]; exists {
			continue
		}
		otherMethods = append(otherMethods, g.Other(g.helper(class), method))
	}
	for _, mtd := range createMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	for _, mtd := range findMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	for _, mtd := range updateByMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	for _, mtd := range deleteByMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	for _, mtd := range otherMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	f.Add(code.Type().Id(class.Name.CamelName()).Interface(interfaceBody...))
	structFields := []code.Code{
		code.Id(fmt.Sprintf("%sDAO", class.Name.CamelLowerName())).Qual(g.importList.Package("dao"), class.Name.CamelName()),
		code.Id("repo").Id("Repository"),
	}
	subClassNames := []string{}
	subClassNameMap := map[string]*types.Class{}
	for subClass := range subClassConstructorMap {
		subClassName := subClass.Name.CamelLowerName()
		subClassNames = append(subClassNames, subClassName)
		subClassNameMap[subClassName] = subClass
	}
	sort.Strings(subClassNames)
	for _, subClassName := range subClassNames {
		subClass := subClassNameMap[subClassName]
		structFields = append(structFields, code.Id(subClass.Name.CamelLowerName()).Id(subClass.Name.CamelName()))
	}
	f.Line()
	f.Add(
		code.Type().Id(fmt.Sprintf("%sImpl", class.Name.CamelName())).Struct(structFields...),
	)
	f.Line()
	constructor, block := g.Constructor(g.helper(class), daoPackageDecl.Constructor, subClassConstructorMap)
	g.constructorMap[constructor.Class] = constructor
	f.Add(block)
	f.Line()
	f.Add(toModelMethod.Generate(g.importList))
	f.Line()
	f.Add(toModelsMethod.Generate(g.importList))
	for _, mtd := range createMethods {
		f.Line()
		f.Add(mtd.Generate(g.importList))
	}
	for _, mtd := range findMethods {
		f.Line()
		f.Add(mtd.Generate(g.importList))
	}
	for _, mtd := range updateByMethods {
		f.Line()
		f.Add(mtd.Generate(g.importList))
	}
	for _, mtd := range deleteByMethods {
		f.Line()
		f.Add(mtd.Generate(g.importList))
	}
	for _, mtd := range otherMethods {
		f.Line()
		f.Add(mtd.Generate(g.importList))
	}
	f.Line()
	f.Add(g.createCollection(g.helper(class)).Generate(g.importList))
	f.Line()
	f.Add(g.create(g.helper(class)).Generate(g.importList))
	bytes := []byte(fmt.Sprintf("%#v", f))
	source, err := imports.Process("", bytes, nil)
	if err != nil {
		return nil, xerrors.Errorf("cannot format by goimport %s: %w", string(bytes), err)
	}
	return source, nil
}

func (g *Generator) generateExpect(class *types.Class, f *code.File, mtd *types.Method) {
	structFields := []code.Code{
		code.Id("expect").Op("*").Id(fmt.Sprintf("%sExpect", class.Name.CamelName())),
		code.Id("isOutOfOrder").Bool(),
		code.Id("isAnyTimes").Bool(),
		code.Id("requiredTimes").Int(),
		code.Id("calledTimes").Int(),
		code.Id("actions").Index().Func().Params(mtd.Decl.Args.Code(g.importList)...),
	}
	structFields = append(structFields, mtd.Decl.Args.Code(g.importList)...)
	structFields = append(structFields, mtd.Decl.Return.Code(g.importList)...)
	f.Line()
	f.Add(code.Type().Id(fmt.Sprintf("%s%sExpect", class.Name.CamelName(), mtd.Decl.MethodName)).Struct(structFields...))
	f.Line()
	f.Add(g.expectReturn(g.helper(class), mtd).Generate(g.importList))
	f.Line()
	f.Add(g.expectDo(g.helper(class), mtd).Generate(g.importList))
	f.Line()
	f.Add(g.expectOutOfOrder(g.helper(class), mtd).Generate(g.importList))
	f.Line()
	f.Add(g.expectAnyTimes(g.helper(class), mtd).Generate(g.importList))
	f.Line()
	f.Add(g.expectTimes(g.helper(class), mtd).Generate(g.importList))
	f.Line()
	f.Add(mtd.Generate(g.importList))
	f.Line()
	f.Add(g.expectMethod(g.helper(class), mtd).Generate(g.importList))
}

func (g *Generator) generateMock(class *types.Class, path string) ([]byte, error) {
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
	createMethod := g.CreateMock(g.helper(class))
	createsMethod := g.CreatesMock(g.helper(class))
	toModelMethod := g.ToModelMock(g.helper(class))
	toModelsMethod := g.ToModelsMock(g.helper(class))
	methodNameMap := map[string]struct{}{
		"Create":  struct{}{},
		"Creates": struct{}{},
	}
	findMethods := []*types.Method{}
	for _, method := range daoPackageDecl.Methods {
		if !strings.HasPrefix(method.MethodName, "Find") {
			continue
		}
		methodNameMap[method.MethodName] = struct{}{}
		findMethods = append(findMethods, g.FindByMock(g.helper(class), method))
	}
	updateByMethods := []*types.Method{}
	for _, method := range daoPackageDecl.Methods {
		if !strings.HasPrefix(method.MethodName, "UpdateBy") {
			continue
		}
		methodNameMap[method.MethodName] = struct{}{}
		updateByMethods = append(updateByMethods, g.UpdateByMock(g.helper(class), method))
	}
	deleteByMethods := []*types.Method{}
	for _, method := range daoPackageDecl.Methods {
		if !strings.HasPrefix(method.MethodName, "DeleteBy") {
			continue
		}
		methodNameMap[method.MethodName] = struct{}{}
		deleteByMethods = append(deleteByMethods, g.DeleteByMock(g.helper(class), method))
	}
	otherMethods := []*types.Method{}
	for _, method := range daoPackageDecl.Methods {
		if _, exists := methodNameMap[method.MethodName]; exists {
			continue
		}
		otherMethods = append(otherMethods, g.OtherMock(g.helper(class), method))
	}
	interfaceBody := []code.Code{
		createMethod.Decl.Interface(g.importList),
		createsMethod.Decl.Interface(g.importList),
		toModelMethod.Decl.Interface(g.importList),
		toModelsMethod.Decl.Interface(g.importList),
	}
	for _, mtd := range findMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	for _, mtd := range updateByMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	for _, mtd := range deleteByMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	for _, mtd := range otherMethods {
		interfaceBody = append(interfaceBody, mtd.Decl.Interface(g.importList))
	}
	f.Line()
	f.Add(
		code.Type().Id(fmt.Sprintf("%sMock", class.Name.CamelName())).Struct(
			code.Id("expect").Op("*").Id(fmt.Sprintf("%sExpect", class.Name.CamelName())),
		),
	)
	expectMethod := g.EXPECT(g.helper(class))
	f.Line()
	f.Add(expectMethod.Generate(g.importList))
	f.Line()
	_, block := g.ConstructorMock(g.helper(class))
	f.Add(block)
	methods := []*types.Method{
		toModelMethod,
		toModelsMethod,
		createMethod,
		createsMethod,
	}
	expectFields := []code.Code{}
	expectValues := code.Dict{}
	methods = append(methods, findMethods...)
	methods = append(methods, updateByMethods...)
	methods = append(methods, deleteByMethods...)
	methods = append(methods, otherMethods...)
	for _, mtd := range methods {
		expectFields = append(expectFields,
			code.Id(types.Name(mtd.Decl.MethodName).CamelLowerName()).
				Index().Op("*").Id(fmt.Sprintf("%s%sExpect", class.Name.CamelName(), mtd.Decl.MethodName)),
		)
		expectValues[code.Id(types.Name(mtd.Decl.MethodName).CamelLowerName())] =
			code.Index().Op("*").Id(fmt.Sprintf("%s%sExpect", class.Name.CamelName(), mtd.Decl.MethodName)).Values()
		g.generateExpect(class, f, mtd)
	}
	f.Line()
	f.Add(code.Type().Id(fmt.Sprintf("%sExpect", class.Name.CamelName())).Struct(expectFields...))
	f.Line()
	f.Add(
		code.Func().Id(fmt.Sprintf("New%sExpect", class.Name.CamelName())).Params().
			Op("*").Id(fmt.Sprintf("%sExpect", class.Name.CamelName())).Block(
			code.Return(code.Op("&").Id(fmt.Sprintf("%sExpect", class.Name.CamelName())).Values(expectValues)),
		),
	)
	bytes := []byte(fmt.Sprintf("%#v", f))
	source, err := imports.Process("", bytes, nil)
	if err != nil {
		return nil, xerrors.Errorf("cannot format by goimport: %s %w", string(bytes), err)
	}
	return source, nil
}

func (g *Generator) generateRepositoryClass(path string, classes []*types.Class) error {
	allArgs := types.ValueDeclares{}
	for _, constructor := range g.constructorMap {
		for _, arg := range constructor.Args {
			allArgs = append(allArgs, arg)
		}
	}
	valueNameMap := map[string]struct{}{}
	methodArgs := []code.Code{}
	for _, arg := range allArgs {
		if _, exists := valueNameMap[arg.Name]; exists {
			continue
		}
		methodArgs = append(methodArgs, arg.Code(g.importList))
		valueNameMap[arg.Name] = struct{}{}
	}
	variables := []code.Code{code.Id("repo").Op("*").Id("RepositoryImpl")}
	properties := code.Dict{}
	fields := []code.Code{}
	interfaceFields := []code.Code{}
	for _, class := range classes {
		constructor := g.constructorMap[class]
		className := class.Name.CamelName()
		classLowerName := class.Name.CamelLowerName()
		variables = append(variables, code.Id(classLowerName).Op("*").Id(fmt.Sprintf("%sImpl", className)))
		args := []code.Code{}
		for _, arg := range constructor.Args {
			args = append(args, code.Id(arg.Name))
		}
		interfaceFields = append(interfaceFields, code.Id(className).Params().Id(className))
		interfaceFields = append(interfaceFields,
			code.Id(fmt.Sprintf("To%s", className)).
				Params(code.Op("*").Qual(g.importList.Package("entity"), className)).
				Op("*").Qual(g.importList.Package("model"), className),
		)
		fields = append(fields, code.Id(classLowerName).Func().Params().Id(className))
		properties[code.Id(classLowerName)] = code.Func().Params().Id(className).Block(
			code.If(code.Id(classLowerName).Op("!=").Nil()).Block(code.Return(code.Id(classLowerName))),
			code.Id(classLowerName).Op("=").Id(constructor.MethodName).Call(args...),
			code.Id(classLowerName).Dot("repo").Op("=").Id("repo"),
			code.Return(code.Id(classLowerName)),
		)
	}
	f := code.NewFile(g.packageName)
	f.HeaderComment(code.GeneratedMarker)
	for _, importDeclare := range g.importList {
		f.ImportName(importDeclare.Path, importDeclare.Name)
	}
	f.Add(code.Type().Id("Repository").Interface(interfaceFields...))
	f.Line()
	f.Add(code.Type().Id("RepositoryImpl").Struct(fields...))
	f.Line()
	for _, class := range classes {
		f.Add(code.Func().Params(code.Id("r").Op("*").Id("RepositoryImpl")).Id(class.Name.CamelName()).Params().Id(class.Name.CamelName()).Block(
			code.Return(code.Id("r").Dot(class.Name.CamelLowerName()).Call()),
		))
		f.Line()
	}
	f.Add(code.Func().Id("New").Params(methodArgs...).Op("*").Id("RepositoryImpl").Block(
		code.Var().Defs(variables...),
		code.Id("repo").Op("=").Op("&").Id("RepositoryImpl").Values(properties),
		code.Return(code.Id("repo")),
	))
	for _, class := range classes {
		f.Line()
		method := &types.Method{
			Decl: &types.MethodDeclare{
				Class:             class,
				ReceiverName:      "r",
				ReceiverClassName: "RepositoryImpl",
				ImportList:        g.importList,
				MethodName:        fmt.Sprintf("To%s", class.Name.CamelName()),
				Args: types.ValueDeclares{
					{
						Name: "value",
						Type: g.helper(class).EntityClassType(),
					},
				},
				Return: types.ValueDeclares{
					{
						Type: g.helper(class).ModelClassType(),
					},
				},
			},
			Body: []code.Code{
				code.Return(code.Id("r").Dot(class.Name.CamelName()).Call().Dot("ToModel").Call(code.Id("value"))),
			},
		}
		f.Add(method.Generate(g.importList))
	}
	source := []byte(fmt.Sprintf("%#v", f))
	repositoryGoPath := filepath.Join(path, "repository.go")
	if g.existsFile(repositoryGoPath) {
		if err := os.Remove(repositoryGoPath); err != nil {
			return xerrors.Errorf("failed to remove file %s: %w", repositoryGoPath, err)
		}
	}
	if err := ioutil.WriteFile(repositoryGoPath, source, 0444); err != nil {
		return xerrors.Errorf("cannot write file to %s: %w", path, err)
	}
	return nil
}

func (g *Generator) generateRepositoryMockClass(path string, classes []*types.Class) error {
	variables := []code.Code{code.Id("repo").Op("*").Id("RepositoryMock")}
	properties := code.Dict{}
	fields := []code.Code{}
	for _, class := range classes {
		constructor := g.constructorMap[class]
		className := class.Name.CamelName()
		classLowerName := class.Name.CamelLowerName()
		variables = append(variables, code.Id(classLowerName).Op("*").Id(fmt.Sprintf("%sMock", className)))
		fields = append(fields, code.Id(classLowerName).Func().Params().Op("*").Id(fmt.Sprintf("%sMock", className)))
		properties[code.Id(classLowerName)] = code.Func().Params().Op("*").Id(fmt.Sprintf("%sMock", className)).Block(
			code.If(code.Id(classLowerName).Op("!=").Nil()).Block(code.Return(code.Id(classLowerName))),
			code.Id(classLowerName).Op("=").Id(fmt.Sprintf("%sMock", constructor.MethodName)).Call(),
			code.Return(code.Id(classLowerName)),
		)
	}
	f := code.NewFile(g.packageName)
	f.HeaderComment(code.GeneratedMarker)
	for _, importDeclare := range g.importList {
		f.ImportName(importDeclare.Path, importDeclare.Name)
	}
	f.ImportName(fmt.Sprintf("%s/repository", g.cfg.ModulePath), "repository")
	f.Add(code.Type().Id("RepositoryMock").Struct(fields...))
	f.Line()
	for _, class := range classes {
		f.Add(code.Func().Params(code.Id("r").Op("*").Id("RepositoryMock")).Id(class.Name.CamelName()).Params().
			Qual(g.importList.Package("repository"), class.Name.CamelName()).Block(
			code.Return(code.Id("r").Dot(class.Name.CamelLowerName()).Call()),
		))
		f.Add(code.Func().Params(code.Id("r").Op("*").Id("RepositoryMock")).Id(fmt.Sprintf("%sMock", class.Name.CamelName())).Params().
			Op("*").Id(fmt.Sprintf("%sMock", class.Name.CamelName())).Block(
			code.Return(code.Id("r").Dot(class.Name.CamelLowerName()).Call()),
		))
		f.Line()
	}
	f.Add(code.Func().Id("NewMock").Params().Op("*").Id("RepositoryMock").Block(
		code.Var().Defs(variables...),
		code.Id("repo").Op("=").Op("&").Id("RepositoryMock").Values(properties),
		code.Return(code.Id("repo")),
	))
	for _, class := range classes {
		f.Line()
		method := &types.Method{
			Decl: &types.MethodDeclare{
				Class:             class,
				ReceiverName:      "r",
				ReceiverClassName: "RepositoryMock",
				ImportList:        g.importList,
				MethodName:        fmt.Sprintf("To%s", class.Name.CamelName()),
				Args: types.ValueDeclares{
					{
						Name: "value",
						Type: g.helper(class).EntityClassType(),
					},
				},
				Return: types.ValueDeclares{
					{
						Type: g.helper(class).ModelClassType(),
					},
				},
			},
			Body: []code.Code{
				code.Return(code.Id("r").Dot(class.Name.CamelName()).Call().Dot("ToModel").Call(code.Id("value"))),
			},
		}
		f.Add(method.Generate(g.importList))
	}
	source := []byte(fmt.Sprintf("%#v", f))
	repositoryGoPath := filepath.Join(path, "repository.go")
	if g.existsFile(repositoryGoPath) {
		if err := os.Remove(repositoryGoPath); err != nil {
			return xerrors.Errorf("failed to remove file %s: %w", repositoryGoPath, err)
		}
	}
	if err := ioutil.WriteFile(repositoryGoPath, source, 0444); err != nil {
		return xerrors.Errorf("cannot write file to %s: %w", path, err)
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

	mockPath := filepath.Join(g.cfg.OutputPath, "mock", g.packageName)
	if err := os.MkdirAll(mockPath, 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", mockPath, err)
	}
	for _, class := range classes {
		source, err := g.generate(class, path)
		if err != nil {
			return xerrors.Errorf("cannot generate repository for %s: %w", class.Name.SnakeName(), err)
		}
		if err := g.writeFile(class, path, source); err != nil {
			return xerrors.Errorf("cannot write file for %s: %w", class.Name.SnakeName(), err)
		}
		mockSource, err := g.generateMock(class, mockPath)
		if err != nil {
			return xerrors.Errorf("cannot generate repository for %s: %w", class.Name.SnakeName(), err)
		}
		if err := g.writeFile(class, mockPath, mockSource); err != nil {
			return xerrors.Errorf("cannot write file for %s: %w", class.Name.SnakeName(), err)
		}
	}
	g.generateRepositoryClass(path, classes)
	g.generateRepositoryMockClass(mockPath, classes)
	return nil
}
