package eevee

import (
	"fmt"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/plugin/entity"
	"go.knocknote.io/eevee/types"
)

type RapidashEntityHandler struct {
}

func init() {
	entity.Register("rapidash", &RapidashEntityHandler{})
}

func (*RapidashEntityHandler) Imports(pkgs types.ImportList) types.ImportList {
	for _, decl := range []*types.ImportDeclare{
		{
			Path: "go.knocknote.io/rapidash",
			Name: "rapidash",
		},
		{
			Path: "golang.org/x/xerrors",
			Name: "xerrors",
		},
	} {
		pkgs[decl.Name] = decl
	}
	return pkgs
}

func (h *RapidashEntityHandler) Struct(helper *types.EntityMethodHelper) *types.Method {
	decl := helper.CreateMethodDeclare()
	decl.MethodName = "Struct"
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: helper.ImportList.Package("rapidash"),
				Name:        "Struct",
			},
			IsPointer: true,
		},
	})
	class := helper.Class
	stmt := Qual(helper.Package("rapidash"), "NewStruct").Call(Lit(class.Name.PluralSnakeName()))
	fields := []Code{stmt}
	for _, member := range class.Members {
		if member.Extend {
			continue
		}
		fields = append(fields, Line().Id(fmt.Sprintf("Field%s", member.CamelType())).Call(Lit(member.Name.SnakeName())))
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Return(Id("").Custom(Options{Separator: "."}, fields...)),
		},
	}
}

func (h *RapidashEntityHandler) EncodeRapidash(helper *types.EntityMethodHelper) *types.Method {
	decl := helper.CreateMethodDeclare()
	decl.MethodName = "EncodeRapidash"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "enc",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: helper.Package("rapidash"),
				Name:        "Encoder",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	codes := []Code{}
	for _, member := range helper.Class.Members {
		if member.Extend {
			continue
		}
		if member.Name.SnakeName() == "id" {
			codes = append(codes,
				If(helper.Field("ID").Op("!=").Lit(0)).Block(
					Id("enc").Dot("Uint64").Call(Lit("id"), helper.Field("ID")),
				),
			)
			continue
		}
		codes = append(codes,
			Id("enc").Dot(member.CamelType()).Call(
				Lit(member.Name.SnakeName()),
				helper.Field(member.Name.CamelName()),
			),
		)
	}
	codes = append(codes, []Code{
		If(
			Err().Op(":=").Id("enc").Dot("Error").Call(),
			Err().Op("!=").Nil(),
		).Block(
			Return(
				Qual(helper.Package("xerrors"), "Errorf").Call(Lit("failed to encode: %w"), Err()),
			),
		),
		Return(Nil()),
	}...)
	return &types.Method{
		Decl: decl,
		Body: codes,
	}
}

func (h *RapidashEntityHandler) EncodeRapidashPlural(helper *types.EntityMethodHelper) *types.Method {
	decl := helper.CreatePluralMethodDeclare()
	decl.MethodName = "EncodeRapidash"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "enc",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: helper.Package("rapidash"),
				Name:        "Encoder",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			For(List(Id("_"), Id("v")).Op(":=").Range().Op("*").Id(helper.ReceiverName)).Block(
				If(
					Err().Op(":=").Id("v").Dot("EncodeRapidash").Call(Id("enc").Dot("New").Call()),
					Err().Op("!=").Nil(),
				).Block(
					Return(Qual(helper.Package("xerrors"), "Errorf").Call(Lit("failed to encode: %w"), Err())),
				),
			),
			Return(Nil()),
		},
	}
}

func (h *RapidashEntityHandler) DecodeRapidash(helper *types.EntityMethodHelper) *types.Method {
	decl := helper.CreateMethodDeclare()
	decl.MethodName = "DecodeRapidash"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "dec",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: helper.Package("rapidash"),
				Name:        "Decoder",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	codes := []Code{}
	for _, member := range helper.Class.Members {
		if member.Extend {
			continue
		}
		codes = append(codes,
			helper.Field(member.Name.CamelName()).
				Op("=").
				Id("dec").Dot(member.CamelType()).Call(Lit(member.Name.SnakeName())),
		)
	}
	codes = append(codes, []Code{
		If(
			Err().Op(":=").Id("dec").Dot("Error").Call(),
			Err().Op("!=").Nil(),
		).Block(
			Return(
				Qual(helper.Package("xerrors"), "Errorf").Call(Lit("failed to decode: %w"), Err()),
			),
		),
		Return(Nil()),
	}...)
	return &types.Method{
		Decl: decl,
		Body: codes,
	}
}

func (h *RapidashEntityHandler) DecodeRapidashPlural(helper *types.EntityMethodHelper) *types.Method {
	decl := helper.CreatePluralMethodDeclare()
	decl.MethodName = "DecodeRapidash"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "dec",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: helper.Package("rapidash"),
				Name:        "Decoder",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.ErrorType),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Id("decLen").Op(":=").Id("dec").Dot("Len").Call(),
			Id("values").Op(":=").Make(Id(helper.Class.Name.PluralCamelName()), Id("decLen")),
			For(
				Id("i").Op(":=").Lit(0),
				Id("i").Op("<").Id("decLen"),
				Id("i").Op("++"),
			).Block(
				Var().Id("v").Id(helper.Class.Name.CamelName()),
				If(
					Err().Op(":=").Id("v").Dot("DecodeRapidash").Call(Id("dec").Dot("At").Call(Id("i"))),
					Err().Op("!=").Nil(),
				).Block(
					Return(Qual(helper.Package("xerrors"), "Errorf").Call(Lit("failed to decode: %w"), Err())),
				),
				Id("values").Index(Id("i")).Op("=").Op("&").Id("v"),
			),
			Op("*").Id(helper.ReceiverName).Op("=").Id("values"),
			Return(Nil()),
		},
	}
}

func (h *RapidashEntityHandler) AddMethods(helper *types.EntityMethodHelper) types.Methods {
	return types.Methods{
		h.Struct(helper),
		h.EncodeRapidash(helper),
		h.EncodeRapidashPlural(helper),
		h.DecodeRapidash(helper),
		h.DecodeRapidashPlural(helper),
	}
}
