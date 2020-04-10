package renderer

import (
	"fmt"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

type JSONRenderer struct{}

// appendIntCode generate like the following code
// ===============================================
// buf = strconv.AppendInt(buf, value, 10)
// or
// buf = strconv.AppendInt(buf, int64(value), 10)
// ===============================================
func (r *JSONRenderer) appendIntCode(h RendererHelper, member *types.Member) Code {
	typ := member.Type.Type
	appendInt := Id("buf").Op("=").Qual("strconv", "AppendInt")
	var value *Statement
	if member.Type.IsPointer {
		value = Op("*").Add(h.Field(member.Name.CamelName()))
	} else {
		value = h.Field(member.Name.CamelName())
	}
	if typ.Name != types.Int64Type.Name {
		value = Id("int64").Call(value)
	}
	if member.Type.IsPointer {
		return If(h.Field(member.Name.CamelName()).Op("==").Nil()).Block(
			Id("buf").Op("=").Append(Id("buf"), Lit("null").Op("...")),
		).Else().Block(
			appendInt.Call(Id("buf"), value, Lit(10)),
		)
	}
	return appendInt.Call(Id("buf"), value, Lit(10))
}

// appendUintCode generate like the following code
// ================================================
// buf = strconv.AppendUint(buf, value, 10)
// or
// buf = strconv.AppendUint(buf, uint64(value), 10)
// ================================================
func (r *JSONRenderer) appendUintCode(h RendererHelper, member *types.Member) Code {
	typ := member.Type.Type
	appendUint := Id("buf").Op("=").Qual("strconv", "AppendUint")
	var value *Statement
	if member.Type.IsPointer {
		value = Op("*").Add(h.Field(member.Name.CamelName()))
	} else {
		value = h.Field(member.Name.CamelName())
	}
	if typ.Name != types.Uint64Type.Name {
		value = Id("uint64").Call(value)
	}
	if member.Type.IsPointer {
		return If(h.Field(member.Name.CamelName()).Op("==").Nil()).Block(
			Id("buf").Op("=").Append(Id("buf"), Lit("null").Op("...")),
		).Else().Block(
			appendUint.Call(Id("buf"), value, Lit(10)),
		)
	}
	return appendUint.Call(Id("buf"), value, Lit(10))
}

// appendFloatCode generate like the following code
// ===========================================================
// buf = strconv.AppendFloat(buf, value, "E", -1, 64)
// or
// buf = strconv.AppendFloat(buf, float64(value), 'E', -1, 64)
// ===========================================================
func (r *JSONRenderer) appendFloatCode(h RendererHelper, member *types.Member) Code {
	typ := member.Type.Type
	appendFloat := Id("buf").Op("=").Qual("strconv", "AppendFloat")
	var value *Statement
	if member.Type.IsPointer {
		value = Op("*").Add(h.Field(member.Name.CamelName()))
	} else {
		value = h.Field(member.Name.CamelName())
	}
	if typ.Name != types.Float64Type.Name {
		value = Id("float64").Call(value)
	}
	if member.Type.IsPointer {
		return If(h.Field(member.Name.CamelName()).Op("==").Nil()).Block(
			Id("buf").Op("=").Append(Id("buf"), Lit("null").Op("...")),
		).Else().Block(
			appendFloat.Call(Id("buf"), value, LitRune('E'), Lit(-1), Lit(64)),
		)
	}
	return appendFloat.Call(Id("buf"), value, LitRune('E'), Lit(-1), Lit(64))
}

// appendStringCode generate like the following code
// ===========================================================
// buf = strconv.Append(buf, strconv.Quote(value)...)
// ===========================================================
func (r *JSONRenderer) appendStringCode(h RendererHelper, member *types.Member) Code {
	if member.Type.IsPointer {
		return If(h.Field(member.Name.CamelName()).Op("==").Nil()).Block(
			Id("buf").Op("=").Append(Id("buf"), Lit("null").Op("...")),
		).Else().Block(
			Id("buf").Op("=").Append(
				Id("buf"),
				Qual("strconv", "Quote").Call(Op("*").Add(h.Field(member.Name.CamelName()))).Op("..."),
			),
		)
	}
	return Id("buf").Op("=").Append(
		Id("buf"),
		Qual("strconv", "Quote").Call(h.Field(member.Name.CamelName())).Op("..."),
	)
}

// appendBytesCode generate like the following code
// ===========================================================
// buf = strconv.Append(buf, value)
// ===========================================================
func (r *JSONRenderer) appendBytesCode(h RendererHelper, member *types.Member) Code {
	if member.Type.IsPointer {
		return If(h.Field(member.Name.CamelName()).Op("==").Nil()).Block(
			Id("buf").Op("=").Append(Id("buf"), Lit("null").Op("...")),
		).Else().Block(
			Id("buf").Op("=").Append(Id("buf"), Op("*").Add(h.Field(member.Name.CamelName()))),
		)
	}
	return Id("buf").Op("=").Append(Id("buf"), h.Field(member.Name.CamelName()))
}

// appendBoolCode generate like the following code
// ===========================================================
// buf = strconv.AppendBool(buf, value)
// ===========================================================
func (r *JSONRenderer) appendBoolCode(h RendererHelper, member *types.Member) Code {
	if member.Type.IsPointer {
		return If(h.Field(member.Name.CamelName()).Op("==").Nil()).Block(
			Id("buf").Op("=").Append(Id("buf"), Lit("null").Op("...")),
		).Else().Block(
			Id("buf").Op("=").Qual("strconv", "AppendBool").Call(
				Id("buf"),
				Op("*").Add(h.Field(member.Name.CamelName())),
			),
		)
	}
	return Id("buf").Op("=").Qual("strconv", "AppendBool").Call(Id("buf"), h.Field(member.Name.CamelName()))
}

// appendTimeCode generate like the following code
// ===========================================================
// buf = strconv.AppendUint(buf, uint64(value), 10)
// ===========================================================
func (r *JSONRenderer) appendTimeCode(h RendererHelper, member *types.Member) Code {
	if member.Type.IsPointer {
		return If(h.Field(member.Name.CamelName()).Op("==").Nil()).Block(
			Id("buf").Op("=").Append(Id("buf"), Lit("null").Op("...")),
		).Else().Block(
			Id("buf").Op("=").Qual("strconv", "AppendUint").Call(
				Id("buf"),
				Id("uint64").Call(h.Field(member.Name.CamelName()).Dot("Unix").Call()),
				Lit(10),
			),
		)
	}
	return Id("buf").Op("=").Qual("strconv", "AppendUint").Call(
		Id("buf"),
		Id("uint64").Call(h.Field(member.Name.CamelName()).Dot("Unix").Call()),
		Lit(10),
	)
}

// appendRelationCode generate like the following code
// ===========================================================
// user, err := m.User(ctx)
// if err != nil {
//   return nil, xerrors.Errorf("cannot get User: %w", err)
// }
// userBytes, err := user.ToJSON(ctx)
// if err != nil {
//   return nil, xerrors.Errorf("cannot render to JSON: %w", err)
// }
// buf = append(buf, "\"user\":"...)
// buf = append(buf, userBytes...)
// ===========================================================
func (r *JSONRenderer) appendRelationCode(h RendererHelper, member *types.Member, withOption bool) []Code {
	bytesName := fmt.Sprintf("%sBytes", member.Name.CamelLowerName())
	renderMethod := "ToJSON"
	renderArgs := []Code{Id("ctx")}
	if withOption {
		renderMethod += "WithOption"
		renderArgs = append(renderArgs, Id("opt"))
	}
	block := []Code{
		List(Id(member.Name.CamelLowerName()), Err()).Op(":=").Add(h.MethodCall(member.Name.CamelName(), Id("ctx"))),
		If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, fmt.Sprintf("cannot get %s: %%w", member.Name.CamelName())))),
		List(Id(bytesName), Err()).Op(":=").Id(member.Name.CamelLowerName()).Dot(renderMethod).Call(renderArgs...),
		If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
	}
	if member.Render != nil && member.Render.IsInline {
		block = append(block, []Code{
			If(Op("!").Qual("bytes", "Equal").Call(Id(bytesName), Index().Byte().Parens(Lit("null")))).Block(
				Id("buf").Op("=").Append(Id("buf"), Id(bytesName).Index(Lit(1), Len(Id(bytesName)).Op("-").Lit(1)).Op("...")),
			),
		}...)
	} else {
		def := Id("buf").Op("=").Append(Id("buf"), Lit(`"`+member.RenderNameByProtocol("json")+`":`).Op("..."))
		block = append(block, def, Id("buf").Op("=").Append(Id("buf"), Id(bytesName).Op("...")))
	}
	return block
}

func (r *JSONRenderer) appendClassCode(h RendererHelper, member *types.Member, withOption bool) []Code {
	bytesName := fmt.Sprintf("%sBytes", member.Name.CamelLowerName())
	renderMethod := "ToJSON"
	renderArgs := []Code{Id("ctx")}
	if withOption {
		renderMethod += "WithOption"
		renderArgs = append(renderArgs, Id("opt"))
	}
	codes := []Code{
		Id(member.Name.CamelLowerName()).Op(":=").Add(h.Field(member.Name.CamelName())),
		List(Id(bytesName), Err()).Op(":=").Id(member.Name.CamelLowerName()).Dot(renderMethod).Call(renderArgs...),
		If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
	}
	if member.Render != nil && member.Render.IsInline {
		codes = append(codes, []Code{
			If(Op("!").Qual("bytes", "Equal").Call(Id(bytesName), Index().Byte().Parens(Lit("null")))).Block(
				Id("buf").Op("=").Append(Id("buf"), Id(bytesName).Index(Lit(1), Len(Id(bytesName)).Op("-").Lit(1)).Op("...")),
			),
		}...)
	} else {
		def := Id("buf").Op("=").Append(Id("buf"), Lit(`"`+member.RenderNameByProtocol("json")+`":`).Op("..."))
		codes = append(codes, def, Id("buf").Op("=").Append(Id("buf"), Id(bytesName).Op("...")))
	}
	return codes
}

func (r *JSONRenderer) appendMember(h RendererHelper, member *types.Member, withOption bool) []Code {
	def := Id("buf").Op("=").Append(Id("buf"), Lit(`"`+member.RenderNameByProtocol("json")+`":`).Op("..."))
	typ := member.Type.Type
	switch {
	case typ.IsInt():
		return []Code{def, r.appendIntCode(h, member)}
	case typ.IsUint():
		return []Code{def, r.appendUintCode(h, member)}
	case typ.IsFloat():
		return []Code{def, r.appendFloatCode(h, member)}
	case typ.IsString():
		return []Code{def, r.appendStringCode(h, member)}
	case typ.IsByte():
		return []Code{def, r.appendBytesCode(h, member)}
	case typ.IsBool():
		return []Code{def, r.appendBoolCode(h, member)}
	case typ.IsTime():
		return []Code{def, r.appendTimeCode(h, member)}
	case member.Relation != nil:
		return r.appendRelationCode(h, member, withOption)
	case member.Type.Class() != nil:
		return r.appendClassCode(h, member, withOption)
	}
	return []Code{}
}

func (r *JSONRenderer) Render(h RendererHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "ToJSON"
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
			Type: types.TypeDeclareWithName("[]byte"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	var assertTo Code
	if h.IsModelPackage() {
		assertTo = Id("BeforeRenderer")
	} else {
		assertTo = Qual(h.Package("model"), "BeforeRenderer")
	}
	codes := []Code{
		If(h.Receiver().Op("==").Nil()).Block(Return(Index().Byte().Parens(Lit("null")), Nil())),
		If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(assertTo)), Id("ok")).Block(
			If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
				Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
			),
		),
		Id("buf").Op(":=").Index().Byte().Values(),
	}
	codes = append(codes, Id("buf").Op("=").Append(Id("buf"), LitRune('{')))
	for idx, member := range h.GetClass().Members {
		if member.Render != nil && !member.Render.IsRender {
			continue
		}
		if idx != 0 {
			codes = append(codes, Id("buf").Op("=").Append(Id("buf"), LitRune(',')))
		}
		if member.Type.IsCustomPrimitiveType() {
			codes = append(codes,
				If(
					List(Id("marshaler"), Id("ok")).Op(":=").Add(
						Interface().Parens(h.Field(member.Name.CamelName())).Assert(Qual(h.Package("json"), "Marshaler")),
					),
					Id("ok"),
				).Block(
					List(Id("bytes"), Err()).Op(":=").Id("marshaler").Dot("MarshalJSON").Call(),
					If(Err().Op("!=").Nil()).Block(
						Return(Nil(), WrapError(h, "failed to MarshalJSON: %w")),
					),
					Id("buf").Op("=").Append(Id("buf"), Id("bytes").Op("...")),
				).Else().Block(r.appendMember(h, member, false)...),
			)
		} else {
			codes = append(codes, r.appendMember(h, member, false)...)
		}
	}
	codes = append(codes, Id("buf").Op("=").Append(Id("buf"), LitRune('}')))
	codes = append(codes, Return(Id("buf"), Nil()))
	return &types.Method{
		Decl: decl,
		Body: codes,
	}
}

func (r *JSONRenderer) RenderWithOption(h RendererHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "ToJSONWithOption"
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
			Type: types.TypeDeclareWithName("[]byte"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	members := h.GetClass().Members
	isMultipleMembers := len(members) > 1
	var assertTo Code
	if h.IsModelPackage() {
		assertTo = Id("BeforeRenderer")
	} else {
		assertTo = Qual(h.Package("model"), "BeforeRenderer")
	}
	codes := []Code{
		If(h.Receiver().Op("==").Nil()).Block(Return(Index().Byte().Parens(Lit("null")), Nil())),
		If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(assertTo)), Id("ok")).Block(
			If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
				Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
			),
		),
		Id("buf").Op(":=").Index().Byte().Values(),
	}
	if isMultipleMembers {
		codes = append(codes, Id("isWritten").Op(":=").False())
	}
	codes = append(codes, Id("buf").Op("=").Append(Id("buf"), LitRune('{')))
	for idx, member := range members {
		if member.Render != nil && !member.Render.IsRender {
			continue
		}
		blocks := []Code{}
		if idx != 0 {
			blocks = append(blocks, If(Id("isWritten")).Block(Id("buf").Op("=").Append(Id("buf"), LitRune(','))))
		}
		if member.Relation != nil || member.Type.Class() != nil {
			ifBlocks := append(blocks, r.appendMember(h, member, false)...)
			if isMultipleMembers {
				ifBlocks = append(ifBlocks, Id("isWritten").Op("=").True())
			}
			elseIfBlocks := append(blocks, r.appendMember(h, member, true)...)
			if isMultipleMembers {
				elseIfBlocks = append(elseIfBlocks, Id("isWritten").Op("=").True())
			}
			codes = append(codes, []Code{
				If(Id("option").Dot("IsIncludeAll")).Block(ifBlocks...).Else().If(
					Id("opt").Op(":=").Id("option").Dot("IncludeOption").Call(Lit(member.Name.SnakeName())),
					Id("opt").Op("!=").Nil(),
				).Block(elseIfBlocks...),
			}...)
		} else {
			if member.Type.IsCustomPrimitiveType() {
				blocks = append(blocks,
					If(
						List(Id("marshaler"), Id("ok")).Op(":=").Add(
							Interface().Parens(h.Field(member.Name.CamelName())).Assert(Qual(h.Package("json"), "Marshaler")),
						),
						Id("ok"),
					).Block(
						List(Id("bytes"), Err()).Op(":=").Id("marshaler").Dot("MarshalJSON").Call(),
						If(Err().Op("!=").Nil()).Block(
							Return(Nil(), WrapError(h, "failed to MarshalJSON: %w")),
						),
						Id("buf").Op("=").Append(Id("buf"), Id("bytes").Op("...")),
					).Else().Block(r.appendMember(h, member, false)...),
				)
			} else {
				blocks = append(blocks, r.appendMember(h, member, false)...)
			}
			if isMultipleMembers {
				blocks = append(blocks, Id("isWritten").Op("=").True())
			}
			codes = append(codes, If(Id("option").Dot("Exists").Call(Lit(member.Name.SnakeName()))).Block(blocks...))
		}
	}
	codes = append(codes, Id("buf").Op("=").Append(Id("buf"), LitRune('}')))
	codes = append(codes, Return(Id("buf"), Nil()))
	return &types.Method{
		Decl: decl,
		Body: codes,
	}
}

func (r *JSONRenderer) RenderCollection(h RendererHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "ToJSON"
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
		Type: types.TypeDeclareWithName("[]byte"),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Index().Byte().Parens(Lit("null")), Nil())),
			If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(Id("BeforeRenderer"))), Id("ok")).Block(
				If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
					Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
				),
			),
			Id("buf").Op(":=").Index().Byte().Values(),
			Id("buf").Op("=").Append(Id("buf"), LitRune('[')),
			For(List(Id("idx"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Id("idx").Op("!=").Lit(0)).Block(
					Id("buf").Op("=").Append(Id("buf"), LitRune(',')),
				),
				List(Id("bytes"), Err()).Op(":=").Id("value").Dot("ToJSON").Call(Id("ctx")),
				If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
				Id("buf").Op("=").Append(Id("buf"), Id("bytes").Op("...")),
			),
			Id("buf").Op("=").Append(Id("buf"), LitRune(']')),
			Return(Id("buf"), Nil()),
		},
	}
}

func (r *JSONRenderer) RenderCollectionWithOption(h RendererHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "ToJSONWithOption"
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
			Type: types.TypeDeclareWithName("[]byte"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Index().Byte().Parens(Lit("null")), Nil())),
			If(List(Id("r"), Id("ok")).Op(":=").Add(Interface().Parens(h.Receiver()).Assert(Id("BeforeRenderer"))), Id("ok")).Block(
				If(Err().Op(":=").Id("r").Dot("BeforeRender").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
					Return(Nil(), WrapError(h, "failed to BeforeRender: %w")),
				),
			),
			Id("buf").Op(":=").Index().Byte().Values(),
			Id("buf").Op("=").Append(Id("buf"), LitRune('[')),
			For(List(Id("idx"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Id("idx").Op("!=").Lit(0)).Block(
					Id("buf").Op("=").Append(Id("buf"), LitRune(',')),
				),
				List(Id("bytes"), Err()).Op(":=").Id("value").Dot("ToJSONWithOption").Call(Id("ctx"), Id("option")),
				If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
				Id("buf").Op("=").Append(Id("buf"), Id("bytes").Op("...")),
			),
			Id("buf").Op("=").Append(Id("buf"), LitRune(']')),
			Return(Id("buf"), Nil()),
		},
	}
}

// Marshaler generate the following code
// ===============================================
// func (m *User) MarshalJSON() ([]byte, error) {
//    bytes, err := m.ToJSON(context.Background())
// 	  if err != nil {
//        return nil, xerrors.Errorf("cannot render to JSON: %w", err)
//    }
//    return bytes, nil
// }
// ===============================================
func (r *JSONRenderer) Marshaler(h RendererHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "MarshalJSON"
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName("[]byte"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			List(Id("bytes"), Err()).Op(":=").Add(
				h.MethodCall("ToJSON", Qual(h.Package("context"), "Background").Call()),
			),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
			Return(Id("bytes"), Nil()),
		},
	}
}

// MarshalerContext generate the following code
// ===============================================
// func (m *User) MarshalJSONContext(ctx context.Context) ([]byte, error) {
//    bytes, err := m.ToJSON(ctx)
//    if err != nil {
//        return nil, xerrors.Errorf("cannot render to JSON: %w", err)
//    }
//    return bytes, nil
// }
// ===============================================
func (r *JSONRenderer) MarshalerContext(h RendererHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "MarshalJSONContext"
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
			Type: types.TypeDeclareWithName("[]byte"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			List(Id("bytes"), Err()).Op(":=").Add(h.MethodCall("ToJSON", Id("ctx"))),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
			Return(Id("bytes"), Nil()),
		},
	}
}

// MarshalerCollection generate the following code
// ===============================================
// func (m *Users) MarshalJSON() ([]byte, error) {
//    bytes, err := m.ToJSON(context.Background())
// 	  if err != nil {
//        return nil, xerrors.Errorf("cannot render to JSON: %w", err)
//    }
//    return bytes, nil
// }
// ===============================================
func (r *JSONRenderer) MarshalerCollection(h RendererHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "MarshalJSON"
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName("[]byte"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			List(Id("bytes"), Err()).Op(":=").Add(
				h.MethodCall("ToJSON", Qual(h.Package("context"), "Background").Call()),
			),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
			Return(Id("bytes"), Nil()),
		},
	}
}

// MarshalerCollectionContext generate the following code
// ===============================================
// func (m *Users) MarshalJSONContext(ctx context.Context) ([]byte, error) {
//    bytes, err := m.ToJSON(ctx)
// 	  if err != nil {
//        return nil, xerrors.Errorf("cannot render to JSON: %w", err)
//    }
//    return bytes, nil
// }
// ===============================================
func (r *JSONRenderer) MarshalerCollectionContext(h RendererHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "MarshalJSONContext"
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
			Type: types.TypeDeclareWithName("[]byte"),
		},
		{
			Type: types.TypeDeclareWithType(types.ErrorType),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			List(Id("bytes"), Err()).Op(":=").Add(h.MethodCall("ToJSON", Id("ctx"))),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), WrapError(h, "cannot render to JSON: %w"))),
			Return(Id("bytes"), Nil()),
		},
	}
}

// Unmarshaler generate the following code
// ===============================================
// func (m *User) UnmarshalJSON(bytes []byte) error {
//   var user struct {
//     *entity.User
//   }
//   if err := json.Unmarshal(bytes, &user); err != nil {
//     return errors.Trace(err)
//   }
//   m.User = &user
//   return nil
// }
// ===============================================
func (r *JSONRenderer) Unmarshaler(h RendererHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "UnmarshalJSON"
	decl.Args = types.ValueDeclares{
		{
			Name: "bytes",
			Type: types.TypeDeclareWithName("[]byte"),
		},
	}
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	structFields := []Code{}
	class := h.GetClass()
	structFields = append(structFields, Op("*").Id(fmt.Sprintf("entity.%s", class.Name.CamelName())))
	/*
		for _, member := range class.Members {
			if member.Render != nil && !member.Render.IsRender {
				continue
			}
			if member.Extend {
				continue
			}
			tag := map[string]string{
				"json": member.RenderNameByProtocol("json"),
			}
			structFields = append(structFields,
				Id(member.Name.CamelName()).Id(member.Type.FormatName(h.GetImportList())).Tag(tag),
			)
		}
	*/
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Var().Id("value").Struct(structFields...),
			If(Err().Op(":=").Qual(h.Package("json"), "Unmarshal").Call(Id("bytes"), Op("&").Id("value")), Err().Op("!=").Nil()).Block(
				Return(WrapError(h, "failed to unmarshal: %w")),
			),
			h.Field(class.Name.CamelName()).Op("=").Id("value").Dot(class.Name.CamelName()),
			Return(Nil()),
		},
	}
}

// UnmarshalerCollection generate the following code
// ===============================================
// func (m *Users) UnmarshalJSON(bytes []byte) error {
//   var values []*User
//   if err := json.Unmarshal(bytes, &values); err != nil {
//     return errors.Trace(err)
//   }
//   m.values = values
//   return nil
// }
// ===============================================
func (r *JSONRenderer) UnmarshalerCollection(h RendererHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "UnmarshalJSON"
	decl.Args = types.ValueDeclares{
		{
			Name: "bytes",
			Type: types.TypeDeclareWithName("[]byte"),
		},
	}
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Var().Id("values").Index().Op("*").Id(h.GetClass().Name.CamelName()),
			If(Err().Op(":=").Qual(h.Package("json"), "Unmarshal").Call(Id("bytes"), Op("&").Id("values")), Err().Op("!=").Nil()).Block(
				Return(WrapError(h, "failed to unmarshal: %w")),
			),
			h.Field("values").Op("=").Id("values"),
			Return(Nil()),
		},
	}
}
