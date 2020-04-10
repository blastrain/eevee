package model

import (
	. "go.knocknote.io/eevee/code"
)

func (g *Generator) generateRenderOption(f *File) {
	f.Line()
	f.Add(GoType().Id("BeforeRenderer").Interface(
		Id("BeforeRender").Params(Qual(g.importList.Package("context"), "Context")).Id("error"),
	))
	f.Line()
	f.Add(GoType().Id("RenderOption").Struct(
		Id("Name").String(),
		Id("IsIncludeAll").Bool(),
		Id("onlyNames").Map(String()).Struct(),
		Id("exceptNames").Map(String()).Struct(),
		Id("includes").Map(String()).Op("*").Id("RenderOption"),
	))
	f.Line()
	f.Add(Func().Params(Id("ro").Op("*").Id("RenderOption")).Id("Exists").Params(Id("name").String()).Bool().Block(
		If(Len(Id("ro").Dot("onlyNames")).Op(">").Lit(0)).Block(
			If(
				List(Id("_"), Id("exists")).Op(":=").Id("ro").Dot("onlyNames").Index(Id("name")),
				Id("exists"),
			).Block(
				Return(True()),
			),
			Return(False()),
		),
		If(Len(Id("ro").Dot("exceptNames")).Op(">").Lit(0)).Block(
			If(
				List(Id("_"), Id("exists")).Op(":=").Id("ro").Dot("exceptNames").Index(Id("name")),
				Id("exists"),
			).Block(
				Return(False()),
			),
			Return(True()),
		),
		Return(True()),
	))
	f.Line()
	f.Add(Func().Params(Id("ro").Op("*").Id("RenderOption")).Id("IncludeOption").Params(Id("name").String()).Op("*").Id("RenderOption").Block(
		If(Id("ro").Dot("Name").Op("==").Id("name")).Block(Return(Id("ro"))),
		Return(Id("ro").Dot("includes").Index(Id("name"))),
	))
	f.Line()
	f.Add(GoType().Id("RenderOptionBuilder").Struct(
		Id("onlyNames").Map(String()).Struct(),
		Id("exceptNames").Map(String()).Struct(),
		Id("includes").Map(String()).Op("*").Id("RenderOption"),
		Id("isIncludeAll").Bool(),
	))
	f.Line()
	f.Add(Func().Id("NewRenderOptionBuilder").Params().Op("*").Id("RenderOptionBuilder").Block(
		Return(Op("&").Id("RenderOptionBuilder").Values(Dict{
			Id("onlyNames"):   Map(String()).Struct().Values(),
			Id("exceptNames"): Map(String()).Struct().Values(),
			Id("includes"):    Map(String()).Op("*").Id("RenderOption").Values(),
		})),
	))
	f.Line()
	f.Add(Func().Params(Id("b").Op("*").Id("RenderOptionBuilder")).Id("Only").Params(Id("names").Op("...").String()).Op("*").Id("RenderOptionBuilder").Block(
		For(List(Id("_"), Id("name")).Op(":=").Range().Id("names")).Block(
			Id("b").Dot("onlyNames").Index(Id("name")).Op("=").Struct().Values(),
		),
		Return(Id("b")),
	))
	f.Line()
	f.Add(Func().Params(Id("b").Op("*").Id("RenderOptionBuilder")).Id("Except").Params(Id("names").Op("...").String()).Op("*").Id("RenderOptionBuilder").Block(
		For(List(Id("_"), Id("name")).Op(":=").Range().Id("names")).Block(
			Id("b").Dot("exceptNames").Index(Id("name")).Op("=").Struct().Values(),
		),
		Return(Id("b")),
	))
	f.Line()
	f.Add(Func().Params(Id("b").Op("*").Id("RenderOptionBuilder")).Id("Include").Params(Id("name").String()).Op("*").Id("RenderOptionBuilder").Block(
		Id("b").Dot("includes").Index(Id("name")).Op("=").Op("&").Id("RenderOption").Values(Dict{Id("Name"): Id("name")}),
		Return(Id("b")),
	))
	f.Line()
	f.Add(Func().Params(Id("b").Op("*").Id("RenderOptionBuilder")).Id("IncludeWithCallback").Params(
		Id("name").String(),
		Id("callback").Func().Params(Op("*").Id("RenderOptionBuilder")),
	).Op("*").Id("RenderOptionBuilder").Block(
		Id("builder").Op(":=").Id("NewRenderOptionBuilder").Call(),
		Id("callback").Call(Id("builder")),
		Id("opt").Op(":=").Id("builder").Dot("Build").Call(),
		Id("opt").Dot("Name").Op("=").Id("name"),
		Id("b").Dot("includes").Index(Id("name")).Op("=").Id("opt"),
		Return(Id("b")),
	))
	f.Line()
	f.Add(Func().Params(Id("b").Op("*").Id("RenderOptionBuilder")).Id("IncludeAll").Params().Op("*").Id("RenderOptionBuilder").Block(
		Id("b").Dot("isIncludeAll").Op("=").True(),
		Return(Id("b")),
	))
	f.Line()
	f.Add(Func().Params(Id("b").Op("*").Id("RenderOptionBuilder")).Id("Build").Params().Op("*").Id("RenderOption").Block(
		Return(Op("&").Id("RenderOption").Values(Dict{
			Id("onlyNames"):    Id("b").Dot("onlyNames"),
			Id("exceptNames"):  Id("b").Dot("exceptNames"),
			Id("includes"):     Id("b").Dot("includes"),
			Id("IsIncludeAll"): Id("b").Dot("isIncludeAll"),
		})),
	))
}
