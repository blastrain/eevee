package renderer

import (
	"fmt"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

type MapRenderer struct{}

func (r *MapRenderer) appendRelationCode(h RendererHelper, member *types.Member, withOption bool) []Code {
	mapName := fmt.Sprintf("%sMap", member.Name.CamelLowerName())
	renderMethod := "ToMap"
	renderArgs := []Code{Id("ctx")}
	if withOption {
		renderMethod += "WithOption"
		renderArgs = append(renderArgs, Id("opt"))
	}
	block := []Code{
		List(Id(member.Name.CamelLowerName()), Err()).Op(":=").Add(h.MethodCall(member.Name.CamelName(), Id("ctx"))),
		If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, fmt.Sprintf("cannot get %s: %%w", member.Name.CamelName())))),
		List(Id(mapName), Err()).Op(":=").Id(member.Name.CamelLowerName()).Dot(renderMethod).Call(renderArgs...),
		If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to map: %w"))),
	}
	if member.Render != nil && member.Render.IsInline {
		block = append(block, []Code{
			If(Id(mapName).Op("!=").Nil()).Block(
				For(List(Id("k"), Id("v")).Op(":=").Range().Id(mapName)).Block(
					Id("value").Index(Id("k")).Op("=").Id("v"),
				),
			),
		}...)
	} else {
		block = append(block,
			Id("value").Index(Lit(member.RenderNameByProtocol("json"))).Op("=").Id(mapName),
		)
	}
	return block
}

func (r *MapRenderer) appendClassCode(h RendererHelper, member *types.Member, withOption bool) []Code {
	mapName := fmt.Sprintf("%sMap", member.Name.CamelLowerName())
	renderMethod := "ToMap"
	renderArgs := []Code{Id("ctx")}
	if withOption {

		renderMethod += "WithOption"
		renderArgs = append(renderArgs, Id("opt"))
	}
	codes := []Code{
		Id(member.Name.CamelLowerName()).Op(":=").Add(h.Field(member.Name.CamelName())),
		List(Id(mapName), Err()).Op(":=").Id(member.Name.CamelLowerName()).Dot(renderMethod).Call(renderArgs...),
		If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to map: %w"))),
	}
	if member.Render != nil && member.Render.IsInline {
		codes = append(codes, []Code{
			If(Id(mapName).Op("!=").Nil()).Block(
				For(List(Id("k"), Id("v")).Op(":=").Range().Id(mapName)).Block(
					Id("value").Index(Id("k")).Op("=").Id("v"),
				),
			),
		}...)
	} else {
		codes = append(codes,
			Id("value").Index(Lit(member.RenderNameByProtocol("json"))).Op("=").Id(mapName),
		)
	}
	return codes
}

func (r *MapRenderer) appendMember(h RendererHelper, member *types.Member, withOption bool) []Code {
	target := Id("value").Index(Lit(member.RenderNameByProtocol("json"))).Op("=")
	if member.Relation != nil {
		return r.appendRelationCode(h, member, withOption)
	} else if member.Type.Class() != nil {
		return r.appendClassCode(h, member, withOption)
	}
	return []Code{target.Add(h.Field(member.Name.CamelName()))}
}

func (r *MapRenderer) Render(h RendererHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "ToMap"
	decl.Args = types.ValueDeclares{
		{
			Name: "ctx",
			Type: &types.TypeDeclare{
				Type: &types.Type{
					PackageName: "context",
					Name:        "Context",
				},
			},
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName("map[string]interface{}"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	codes := []Code{
		If(h.Receiver().Op("==").Nil()).Block(Return(Nil(), Nil())),
		If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(Id("BeforeRenderer"))), Id("ok")).Block(
			If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
				Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
			),
		),
		Id("value").Op(":=").Map(String()).Interface().Values(),
	}
	for _, member := range h.GetClass().Members {
		if member.Render != nil && !member.Render.IsRender {
			continue
		}
		codes = append(codes, r.appendMember(h, member, false)...)
	}
	codes = append(codes, Return(Id("value"), Nil()))
	return &types.Method{
		Decl: decl,
		Body: codes,
	}
}

func (r *MapRenderer) RenderWithOption(h RendererHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "ToMapWithOption"
	var renderOptionTypeName string
	if h.IsModelPackage() {
		renderOptionTypeName = "*RenderOption"
	} else {
		renderOptionTypeName = "*model.RenderOption"
	}
	decl.Args = types.ValueDeclares{
		{
			Name: "ctx",
			Type: &types.TypeDeclare{
				Type: &types.Type{
					PackageName: "context",
					Name:        "Context",
				},
			},
		},
		{
			Name: "option",
			Type: types.TypeDeclareWithName(renderOptionTypeName),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName("map[string]interface{}"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	members := h.GetClass().Members
	codes := []Code{
		If(h.Receiver().Op("==").Nil()).Block(Return(Nil(), Nil())),
		If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(Id("BeforeRenderer"))), Id("ok")).Block(
			If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
				Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
			),
		),
		Id("value").Op(":=").Map(String()).Interface().Values(),
	}
	for _, member := range members {
		if member.Render != nil && !member.Render.IsRender {
			continue
		}
		blocks := []Code{}
		if member.Relation != nil || member.Type.Class() != nil {
			ifBlocks := append(blocks, r.appendMember(h, member, false)...)
			elseIfBlocks := append(blocks, r.appendMember(h, member, true)...)
			codes = append(codes, []Code{
				If(Id("option").Dot("IsIncludeAll")).Block(ifBlocks...).Else().If(
					Id("opt").Op(":=").Id("option").Dot("IncludeOption").Call(Lit(member.Name.SnakeName())),
					Id("opt").Op("!=").Nil(),
				).Block(elseIfBlocks...),
			}...)
		} else {
			blocks = append(blocks, r.appendMember(h, member, true)...)
			codes = append(codes, If(Id("option").Dot("Exists").Call(Lit(member.Name.SnakeName()))).Block(blocks...))
		}
	}
	codes = append(codes, Return(Id("value"), Nil()))
	return &types.Method{
		Decl: decl,
		Body: codes,
	}
}

func (r *MapRenderer) RenderCollection(h RendererHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "ToMap"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithName("[]map[string]interface{}"),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil(), Nil())),
			If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(Id("BeforeRenderer"))), Id("ok")).Block(
				If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
					Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
				),
			),
			Id("value").Op(":=").Index().Map(String()).Interface().Values(),
			For(List(Id("_"), Id("v")).Op(":=").Range().Add(h.Field("values"))).Block(
				List(Id("mapValue"), Err()).Op(":=").Id("v").Dot("ToMap").Call(Id("ctx")),
				If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to map: %w"))),
				Id("value").Op("=").Append(Id("value"), Id("mapValue")),
			),
			Return(Id("value"), Nil()),
		},
	}
}

func (r *MapRenderer) RenderCollectionWithOption(h RendererHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "ToMapWithOption"
	decl.Args = types.ValueDeclares{
		{
			Name: "ctx",
			Type: &types.TypeDeclare{
				Type: &types.Type{
					PackageName: "context",
					Name:        "Context",
				},
			},
		},
		{
			Name: "option",
			Type: types.TypeDeclareWithName("*RenderOption"),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName("[]map[string]interface{}"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil(), Nil())),
			If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(Id("BeforeRenderer"))), Id("ok")).Block(
				If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
					Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
				),
			),
			Id("value").Op(":=").Index().Map(String()).Interface().Values(),
			For(List(Id("_"), Id("v")).Op(":=").Range().Add(h.Field("values"))).Block(
				List(Id("mapValue"), Err()).Op(":=").Id("v").Dot("ToMapWithOption").Call(Id("ctx"), Id("option")),
				If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to map: %w"))),
				Id("value").Op("=").Append(Id("value"), Id("mapValue")),
			),
			Return(Id("value"), Nil()),
		},
	}
}

func (*MapRenderer) Marshaler(h RendererHelper) *types.Method                  { return nil }
func (*MapRenderer) MarshalerContext(h RendererHelper) *types.Method           { return nil }
func (*MapRenderer) MarshalerCollection(h RendererHelper) *types.Method        { return nil }
func (*MapRenderer) MarshalerCollectionContext(h RendererHelper) *types.Method { return nil }
func (*MapRenderer) Unmarshaler(h RendererHelper) *types.Method                { return nil }
func (*MapRenderer) UnmarshalerCollection(h RendererHelper) *types.Method      { return nil }
