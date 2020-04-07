package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/goccy/go-yaml"
	"go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/renderer"
	_ "go.knocknote.io/eevee/static"
	"go.knocknote.io/eevee/types"
	"github.com/rakyll/statik/fs"
	"golang.org/x/xerrors"
)

type Generator struct {
	appName      string
	receiverName string
	importList   types.ImportList
	cfg          *config.Config
}

func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		appName:      cfg.ModulePath,
		receiverName: "r",
		importList:   types.ImportList{},
		cfg:          cfg,
	}
}

func (g *Generator) existsFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (g *Generator) readAPI(path string) ([]*types.API, error) {
	if !g.existsFile(path) {
		return nil, nil
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, xerrors.Errorf("cannot read api file: %w", err)
	}
	var api []*types.API
	if err := yaml.Unmarshal(bytes, &api); err != nil {
		return nil, xerrors.Errorf("cannot unmarshal from %s to api: %w", string(bytes), err)
	}
	return api, nil
}

func (g *Generator) helper(class *types.Class) *types.APIResponseHelper {
	return &types.APIResponseHelper{
		Class:        class,
		ReceiverName: g.receiverName,
		ImportList:   g.importList,
	}
}

func (g Generator) addMethod(f *code.File, mtd *types.Method) {
	f.Line()
	f.Add(mtd.Generate(g.importList))
}

func (g *Generator) loadTemplate(filename string) (*template.Template, error) {
	statikFS, err := fs.New()
	if err != nil {
		return nil, xerrors.Errorf("failed to create statik fs: %w", err)
	}
	file, err := statikFS.Open(filename)
	if err != nil {
		return nil, xerrors.Errorf("failed to open %s: %w", filename, err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, xerrors.Errorf("failed to read from %s: %w", filename, err)
	}
	tmpl, err := template.New("").Parse(string(bytes))
	if err != nil {
		return nil, xerrors.Errorf("failed to parse text template for %s: %w", filename, err)
	}
	return tmpl, nil
}

func (g *Generator) writeAPIIndex(api []*types.API) error {
	indexTmpl, err := g.loadTemplate("/index.tmpl")
	if err != nil {
		return xerrors.Errorf("failed to load index.tmpl: %w", err)
	}
	var buf bytes.Buffer
	if err := indexTmpl.Execute(&buf, api); err != nil {
		return xerrors.Errorf("failed to execute template: %w", err)
	}
	if err := os.MkdirAll("docs", 0755); err != nil {
		return xerrors.Errorf("cannot create directory to docs: %w", err)
	}
	path := filepath.Join("docs", "index.md")
	if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return xerrors.Errorf("cannot write file %s: %w", path, err)
	}
	return nil
}

func (g *Generator) generateDocument(api []*types.API, classMap *map[string]*types.Class) error {
	if err := g.writeAPIIndex(api); err != nil {
		return xerrors.Errorf("failed to write index.md for api: %w", err)
	}
	docTmpl, err := g.loadTemplate("/doc.tmpl")
	if err != nil {
		return xerrors.Errorf("failed to load doc.tmpl: %w", err)
	}
	for _, subAPI := range api {
		var buf bytes.Buffer
		if err := docTmpl.Execute(&buf, subAPI); err != nil {
			return xerrors.Errorf("failed to execute template: %w", err)
		}
		path := filepath.Join("docs", fmt.Sprintf("%s.md", subAPI.Name.SnakeName()))
		if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
			return xerrors.Errorf("cannot write file %s: %w", path, err)
		}
	}
	return nil
}

func (g *Generator) generateCastCode(param *types.RequestParam) []code.Code {
	toStr := code.Qual(g.importList.Package("fmt"), "Sprint").Call(code.Id("v"))
	switch param.Type {
	case "int", "int8", "int16", "int32", "int64":
		return []code.Code{
			code.List(code.Id("i"), code.Id("_")).Op(":=").Qual(
				g.importList.Package("strconv"),
				"ParseInt",
			).Call(toStr, code.Lit(10), code.Lit(64)),
			code.Id("req").Dot(param.Name.CamelName()).Op("=").Id(param.Type).Call(code.Id("i")),
		}
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return []code.Code{
			code.List(code.Id("i"), code.Id("_")).Op(":=").Qual(
				g.importList.Package("strconv"),
				"ParseUint",
			).Call(toStr, code.Lit(10), code.Lit(64)),
			code.Id("req").Dot(param.Name.CamelName()).Op("=").Id(param.Type).Call(code.Id("i")),
		}
	case "float32", "float64":
		return []code.Code{
			code.List(code.Id("f"), code.Id("_")).Op(":=").Qual(
				g.importList.Package("strconv"),
				"ParseFloat",
			).Call(toStr, code.Lit(64)),
			code.Id("req").Dot(param.Name.CamelName()).Op("=").Id(param.Type).Call(code.Id("f")),
		}
	case "[]byte":
		return []code.Code{
			code.Id("req").Dot(param.Name.CamelName()).Op("=").Id(param.Type).Call(toStr),
		}
	case "string":
		return []code.Code{
			code.Id("req").Dot(param.Name.CamelName()).Op("=").Add(toStr),
		}
	case "bool":
		return []code.Code{
			code.List(code.Id("b"), code.Id("_")).Op(":=").Qual(
				g.importList.Package("strconv"),
				"ParseBool",
			).Call(toStr),
			code.Id("req").Dot(param.Name.CamelName()).Op("=").Id("b"),
		}
	}
	return []code.Code{}
}

func (g *Generator) generateRequest(path string, api []*types.API, classMap *map[string]*types.Class) error {
	f := code.NewFile("request")
	f.HeaderComment(code.GeneratedMarker)
	for _, decl := range []*types.ImportDeclare{
		{
			Path: "golang.org/x/xerrors",
			Name: "xerrors",
		},
		{
			Path: "net/http",
			Name: "http",
		},
		{
			Path: "encoding/json",
			Name: "json",
		},
	} {
		g.importList[decl.Name] = decl
	}
	for _, importDeclare := range g.importList {
		f.ImportName(importDeclare.Path, importDeclare.Name)
	}
	for _, subAPI := range api {
		if subAPI.Request == nil {
			continue
		}
		builderFields := []code.Code{}
		requestFields := []code.Code{}
		buildBlock := []code.Code{
			code.Id("req").Op(":=").Op("&").Id(subAPI.Name.CamelName()).Values(),
		}
		if subAPI.Request.Params.HasInBodyParam() {
			buildBlock = append(buildBlock,
				code.Id("dec").Op(":=").Qual(g.importList.Package("json"), "NewDecoder").Call(
					code.Id("r").Dot("Body"),
				),
				code.Var().Id("body").Map(code.String()).Interface(),
				code.If(
					code.Err().Op(":=").Id("dec").Dot("Decode").Call(code.Op("&").Id("body")),
					code.Err().Op("!=").Nil(),
				).Block(
					code.Return(code.Nil(), code.WrapError(g.helper(nil), "failed to decode: %w")),
				),
			)
		}
		if subAPI.Request.Params.HasInQueryParam() {
			buildBlock = append(buildBlock,
				code.If(code.Err().Op(":=").Id("r").Dot("ParseForm").Call()).Block(
					code.Return(code.Nil(), code.WrapError(g.helper(nil), "failed to parse form: %w")),
				),
			)
		}
		for _, param := range subAPI.Request.Params {
			requestFields = append(requestFields,
				code.Id(param.Name.CamelName()).Id(param.Type),
			)
			switch param.In {
			case types.InHeader:
				buildBlock = append(buildBlock,
					code.Id("req").Dot(param.Name.CamelName()).Op("=").
						Id("r").Dot("Header").Dot("Get").Call(
						code.Lit(fmt.Sprintf("X-%s", param.Name.CamelName())),
					),
				)
			case types.InPath:
				builderFields = append(builderFields,
					code.Id(param.Name.CamelLowerName()).Op("*").Id(param.Type),
				)
				if param.Required {
					buildBlock = append(buildBlock,
						code.If(
							code.Id("b").Dot(param.Name.CamelLowerName()).Op("==").Nil(),
						).Block(
							code.Return(
								code.Nil(),
								code.Qual(g.importList.Package("xerrors"), "New").Call(
									code.Lit(fmt.Sprintf("%s is required. but doesn't assigned to builder", param.Name.CamelLowerName())),
								),
							),
						),
					)
				}
				buildBlock = append(buildBlock,
					code.If(
						code.Id("b").Dot(param.Name.CamelLowerName()).Op("!=").Nil(),
					).Block(
						code.Id("req").Dot(param.Name.CamelName()).Op("=").
							Op("*").Id("b").Dot(param.Name.CamelLowerName()),
					),
				)
			case types.InQuery:
				buildBlock = append(buildBlock,
					code.Id("req").Dot(param.Name.CamelName()).Op("=").
						Id("r").Dot("FormValue").Call(code.Lit(param.RenderName())),
				)
			case types.InBody:
				buildBlock = append(buildBlock,
					code.If(
						code.List(
							code.Id("v"),
							code.Id("exists"),
						).Op(":=").Id("body").Index(code.Lit(param.RenderName())),
						code.Id("exists"),
					).Block(
						g.generateCastCode(param)...,
					),
				)
			}
		}
		buildBlock = append(buildBlock, code.Return(code.Id("req"), code.Nil()))
		f.Add(code.Type().Id(subAPI.Name.CamelName()).Struct(requestFields...))
		f.Line()
		f.Add(code.Type().Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())).Struct(builderFields...))
		f.Line()
		for _, param := range subAPI.Request.Params {
			if param.In != types.InPath {
				continue
			}
			f.Add(
				code.Func().Params(
					code.Id("b").Op("*").Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())),
				).Id(fmt.Sprintf("Set%s", param.Name.CamelName())).Params(code.Id("a").Id(param.Type)).Params(
					code.Op("*").Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())),
				).Block([]code.Code{
					code.Id("b").Dot(param.Name.CamelLowerName()).Op("=").Op("&").Id("a"),
					code.Return(code.Id("b")),
				}...),
			)
			f.Line()
		}
		f.Add(
			code.Func().Params(
				code.Id("b").Op("*").Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())),
			).Id("Build").Params(code.Id("r").Op("*").Qual(g.importList.Package("http"), "Request")).Params(
				code.Op("*").Id(subAPI.Name.CamelName()), code.Id("error"),
			).Block(buildBlock...),
		)
		source := []byte(fmt.Sprintf("%#v", f))
		apiPath := filepath.Join(path, fmt.Sprintf("%s.go", subAPI.Name.SnakeName()))
		if g.existsFile(apiPath) {
			if err := os.Remove(apiPath); err != nil {
				return xerrors.Errorf("failed to remove file %s: %w", apiPath, err)
			}
		}
		if err := ioutil.WriteFile(apiPath, source, 0444); err != nil {
			return xerrors.Errorf("cannot write file %s: %w", apiPath, err)
		}
	}
	return nil
}

func (g *Generator) generateResponse(path string, api []*types.API, classMap *map[string]*types.Class) error {
	renderer := &renderer.JSONRenderer{}
	f := code.NewFile("response")
	f.HeaderComment(code.GeneratedMarker)
	for _, decl := range []*types.ImportDeclare{
		{
			Path: "golang.org/x/xerrors",
			Name: "xerrors",
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
	subClassMap := map[string]*types.Class{}
	for _, subAPI := range api {
		if subAPI.Response == nil {
			continue
		}
		for _, subtype := range subAPI.Response.SubTypes {
			class := subtype.Class
			subClassMap[class.Name.SnakeName()] = &class
		}
	}
	f.Line()
	for _, subAPI := range api {
		if subAPI.Response == nil {
			continue
		}
		subAPI.Response.ResolveClassReference(classMap, &subClassMap)
		for _, subtype := range subAPI.Response.SubTypes {
			class := &subtype.Class
			class.SetClassMap(classMap)
			class.ResolveTypeReference()
			structFields := []code.Code{}
			for _, member := range class.Members {
				typeName := code.Id(member.Type.Name())
				if member.Type.Class() != nil {
					if member.IsCollectionType() {
						typeName = code.Op("*").Qual(g.importList.Package("model"), member.Type.Class().Name.PluralCamelName())
					} else {
						typeName = code.Op("*").Qual(g.importList.Package("model"), member.Type.Class().Name.CamelName())
					}
				}
				structFields = append(structFields,
					code.Id(member.Name.CamelName()).Add(typeName),
				)
			}
			f.Add(code.Type().Id(class.Name.CamelName()).Struct(structFields...))
			g.addMethod(f, renderer.Render(g.helper(class)))
			g.addMethod(f, renderer.RenderWithOption(g.helper(class)))
		}
		builderCodes := subAPI.Response.BuilderCode(g.helper(nil))
		builderFields := []code.Code{}
		responseFields := []code.Code{}
		memberTypeMap := map[string]code.Code{}
		for _, member := range subAPI.Response.Type.Members {
			typeName := code.Id(member.Type.Name())
			if member.Type.Class() != nil {
				if member.Type.IsSchemaClass() {
					if member.IsCollectionType() {
						typeName = code.Op("*").Qual(g.importList.Package("model"), member.Type.Class().Name.PluralCamelName())
					} else {
						typeName = code.Op("*").Qual(g.importList.Package("model"), member.Type.Class().Name.CamelName())
					}
				} else {
					if member.IsCollectionType() {
						typeName = code.Op("*").Id(member.Type.Class().Name.PluralCamelName())
					} else {
						typeName = code.Op("*").Id(member.Type.Class().Name.CamelName())
					}
				}
			}
			memberTypeMap[member.Name.SnakeName()] = typeName
			builderFields = append(builderFields,
				code.Id(member.Name.CamelLowerName()).Add(typeName),
			)
			responseFields = append(responseFields,
				code.Id(member.Name.CamelName()).Add(typeName),
			)
		}
		responseFields = append(responseFields, code.Id("renderedBytes").Index().Byte())
		subAPI.Response.Type.Name = subAPI.Name
		f.Add(code.Type().Id(subAPI.Name.CamelName()).Struct(responseFields...))
		g.addMethod(f, renderer.Render(g.helper(&subAPI.Response.Type.Class)))
		g.addMethod(f, renderer.RenderWithOption(g.helper(&subAPI.Response.Type.Class)))
		f.Line()
		f.Add(
			code.Func().Params(
				code.Id("r").Op("*").Id(subAPI.Name.CamelName()),
			).Id("MarshalJSON").Params().Params(
				code.Index().Byte(), code.Id("error"),
			).Block([]code.Code{
				code.If(code.Id("r").Dot("renderedBytes").Op("==").Nil()).Block(
					code.Return(code.Nil(), code.Qual(g.importList.Package("xerrors"), "New").Call(code.Lit("response object must be created by builder"))),
				),
				code.Return(code.Id("r").Dot("renderedBytes"), code.Nil()),
			}...),
		)
		f.Add(code.Type().Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())).Struct(builderFields...))
		buildBlocks := []code.Code{
			code.Var().Id("res").Id(subAPI.Name.CamelName()),
		}
		for _, member := range subAPI.Response.Type.Members {
			f.Line()
			f.Add(
				code.Func().Params(
					code.Id("b").Op("*").Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())),
				).Id(fmt.Sprintf("Set%s", member.Name.CamelName())).Params(
					code.Id("value").Add(memberTypeMap[member.Name.SnakeName()]),
				).Params(
					code.Op("*").Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())),
				).Block(
					code.Id("b").Dot(member.Name.CamelLowerName()).Op("=").Id("value"),
					code.Return(code.Id("b")),
				),
			)
			buildBlocks = append(buildBlocks, code.Id("res").Dot(member.Name.CamelName()).Op("=").Id("b").Dot(member.Name.CamelLowerName()))
		}
		buildBlocks = append(buildBlocks, builderCodes...)
		buildBlocks = append(buildBlocks, []code.Code{
			code.List(code.Id("bytes"), code.Err()).Op(":=").Id("res").Dot("ToJSONWithOption").Call(code.Id("ctx"), code.Id("optBuilder").Dot("Build").Call()),
			code.If(code.Err().Op("!=").Nil()).Block(code.Return(code.Nil(), code.Qual(g.importList.Package("xerrors"), "Errorf").Call(code.Lit("failed to render json: %w"), code.Err()))),
			code.Id("res").Dot("renderedBytes").Op("=").Id("bytes"),
			code.Return(code.Op("&").Id("res"), code.Nil()),
		}...)
		f.Line()
		f.Add(
			code.Func().Params(
				code.Id("b").Op("*").Id(fmt.Sprintf("%sBuilder", subAPI.Name.CamelName())),
			).Id("Build").Params(
				code.Id("ctx").Qual(g.importList.Package("context"), "Context"),
			).Params(
				code.Op("*").Id(subAPI.Name.CamelName()), code.Id("error"),
			).Block(buildBlocks...),
		)
		source := []byte(fmt.Sprintf("%#v", f))
		apiPath := filepath.Join(path, fmt.Sprintf("%s.go", subAPI.Name.SnakeName()))
		if g.existsFile(apiPath) {
			if err := os.Remove(apiPath); err != nil {
				return xerrors.Errorf("failed to remove file %s: %w", apiPath, err)
			}
		}
		if err := ioutil.WriteFile(apiPath, source, 0444); err != nil {
			return xerrors.Errorf("cannot write file %s: %w", apiPath, err)
		}
	}
	return nil
}

func (g *Generator) generate(path string, classMap *map[string]*types.Class) error {
	api, err := g.readAPI(path)
	if err != nil {
		return xerrors.Errorf("cannot read api file %s: %w", path, err)
	}
	if err := g.generateRequest(g.cfg.RequestPackageName(), api, classMap); err != nil {
		return xerrors.Errorf("cannot generate request: %w", err)
	}
	if err := g.generateResponse(g.cfg.ResponsePackageName(), api, classMap); err != nil {
		return xerrors.Errorf("cannot generate response: %w", err)
	}
	if err := g.generateDocument(api, classMap); err != nil {
		return xerrors.Errorf("failed to generate api document: %w", err)
	}
	return nil
}

func (g *Generator) Generate(classes []*types.Class) error {
	path := g.cfg.APIPath
	if path == "" {
		return nil
	}
	requestPkgName := g.cfg.RequestPackageName()
	responsePkgName := g.cfg.ResponsePackageName()
	if err := os.MkdirAll(requestPkgName, 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", requestPkgName, err)
	}
	if err := os.MkdirAll(responsePkgName, 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", responsePkgName, err)
	}
	classMap := map[string]*types.Class{}
	for _, class := range classes {
		classMap[class.Name.SnakeName()] = class
	}
	if filepath.Ext(path) == "yml" {
		if err := g.generate(path, &classMap); err != nil {
			return xerrors.Errorf("failed to generate: %w", err)
		}
		return nil
	}
	ymlFilePattern := regexp.MustCompile(`\.yml$`)
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !ymlFilePattern.MatchString(path) {
			return nil
		}
		if err := g.generate(path, &classMap); err != nil {
			return xerrors.Errorf("failed to generate: %w", err)
		}
		return nil
	}); err != nil {
		return xerrors.Errorf("interrupt walk in %s: %w", path, err)
	}
	return nil
}
