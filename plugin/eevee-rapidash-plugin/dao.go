package eevee

import (
	"fmt"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/plugin/dao"
	"go.knocknote.io/eevee/types"
)

type RapidashDataStore struct{}

func init() {
	dao.RegisterDataStore("rapidash", &RapidashDataStore{})
}

func (*RapidashDataStore) Imports(pkgs types.ImportList) types.ImportList {
	for _, decl := range []*types.ImportDeclare{
		{
			Path: "fmt",
			Name: "fmt",
		},
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

func (*RapidashDataStore) StructFields(class *types.Class, fields types.StructFieldList) types.StructFieldList {
	values := types.ValueDeclares{
		{
			Name: "tx",
			Type: &types.TypeDeclare{
				Type: &types.Type{
					PackageName: "rapidash",
					Name:        "Tx",
				},
				IsPointer: true,
			},
		},
	}
	for _, value := range values {
		fields[value.Name] = value
	}
	return fields
}

func (*RapidashDataStore) ConstructorDeclare(d *types.ConstructorDeclare) error {
	d.Args = append(d.Args, &types.ValueDeclare{
		Name: "tx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: d.Package("rapidash"),
				Name:        "Tx",
			},
			IsPointer: true,
		},
	})
	return nil
}

func (*RapidashDataStore) Constructor(p *types.ConstructorParam) []Code {
	return []Code{
		Return(Op("&").Id(p.ImplName).Values(Dict{
			Id("tx"): Id("tx"),
		})),
	}
}

func (*RapidashDataStore) Create(p *types.CreateParam) []Code {
	var idMember *types.Member
	for _, member := range p.Class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		if member.Name.SnakeName() == "id" {
			idMember = member
		}
	}
	return []Code{
		List(Id("id"), Err()).Op(":=").Add(p.Field("tx").Dot("CreateByTableContext").Call(
			Id("ctx"),
			Lit(p.Class.Name.PluralSnakeName()),
			Id("value"),
		)),
		If(Err().Op("!=").Nil()).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Create: %w"), Err())),
		),
		Id("value").Dot("ID").Op("=").Add(idMember.Type.Code(p.ImportList)).Call(Id("id")), // TODO: need to AUTO_INCREMENT column only
		Return(Nil()),
	}
}

func (*RapidashDataStore) Update(p *types.UpdateParam) []Code {
	updateMap := Dict{}
	for _, member := range p.Class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		if member.Name.SnakeName() == "id" {
			continue
		}
		updateMap[Lit(member.Name.SnakeName())] = Id("value").Dot(member.Name.CamelName())
	}
	return []Code{
		Id("updateMap").Op(":=").Map(String()).Interface().Values(updateMap),
		Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
			Lit(p.Class.Name.PluralSnakeName()),
		).Dot("Eq").Call(Lit("id"), Id("value").Dot("ID")),
		If(Err().Op(":=").Add(p.Field("tx").Dot("UpdateByQueryBuilderContext").Call(
			Id("ctx"),
			Id("builder"),
			Id("updateMap"),
		)), Err().Op("!=").Nil()).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Update: %w"), Err())),
		),
		Return(Nil()),
	}
}

func (*RapidashDataStore) Delete(p *types.DeleteParam) []Code {
	return []Code{
		Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
			Lit(p.Class.Name.PluralSnakeName()),
		).Dot("Eq").Call(Lit("id"), Id("value").Dot("ID")),
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("DeleteByQueryBuilderContext").Call(Id("ctx"), Id("builder"))),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Delete: %w"), Err())),
		),
		Return(Nil()),
	}
}

func (*RapidashDataStore) FindAll(p *types.FindParam) []Code {
	return []Code{
		Id("values").Op(":=").Qual(p.Package("entity"), p.Class.Name.PluralCamelName()).Block(),
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("FindAllByTable").Call(Lit(p.Class.Name.PluralSnakeName()), Op("&").Id("values"))),
			Err().Op("!=").Nil(),
		).Block(
			Return(List(Id("values"), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to FindAll %w"), Err()))),
		),
		Return(List(Id("values"), Nil())),
	}
}

func (*RapidashDataStore) Count(p *types.CountParam) []Code {
	return []Code{
		Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
			Lit(p.Class.Name.PluralSnakeName()),
		),
		List(Id("count"), Err()).Op(":=").Add(p.Field("tx").Dot("CountByQueryBuilderContext").Call(Id("ctx"), Id("builder"))),
		If(Err().Op("!=").Nil()).Block(
			Return(List(Lit(0), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Count: %w"), Err()))),
		),
		Return(Int64().Call(Id("count")), Nil()),
	}
}

func (*RapidashDataStore) FindBy(p *types.FindParam) []Code {
	builder := Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
		Lit(p.Class.Name.PluralSnakeName()),
	)
	for idx, member := range p.Args.Members {
		builder = builder.Dot("Eq").Call(Lit(member.Name.SnakeName()), Id(fmt.Sprintf("a%d", idx)))
	}
	if p.IsSingleReturnValue {
		return []Code{
			builder,
			Var().Id("value").Qual(p.Package("entity"), p.Class.Name.CamelName()),
			If(
				Err().Op(":=").Add(p.Field("tx").Dot("FindByQueryBuilderContext").Call(
					Id("ctx"), Id("builder"), Op("&").Id("value"),
				)),
				Err().Op("!=").Nil(),
			).Block(
				Return(Nil(), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Find: %w"), Err())),
			),
			Return(Op("&").Id("value"), Nil()),
		}
	}
	return []Code{
		builder,
		Var().Id("values").Qual(p.Package("entity"), p.Class.Name.PluralCamelName()),
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("FindByQueryBuilderContext").Call(
				Id("ctx"), Id("builder"), Op("&").Id("values"),
			)),
			Err().Op("!=").Nil(),
		).Block(
			Return(Nil(), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Find: %w"), Err())),
		),
		Return(Id("values"), Nil()),
	}
}

func (*RapidashDataStore) FindByPlural(p *types.FindParam) []Code {
	builder := Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
		Lit(p.Class.Name.PluralSnakeName()),
	)
	for idx, member := range p.Args.Members {
		builder = builder.Dot("In").Call(Lit(member.Name.SnakeName()), Id(fmt.Sprintf("a%d", idx)))
	}
	return []Code{
		builder,
		Id("values").Op(":=").Qual(p.Package("entity"), p.Class.Name.PluralCamelName()).Values(),
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("FindByQueryBuilderContext").Call(
				Id("ctx"), Id("builder"), Op("&").Id("values"),
			)),
			Err().Op("!=").Nil(),
		).Block(
			Return(Nil(), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Find: %w"), Err())),
		),
		Return(Id("values"), Nil()),
	}
}

func (*RapidashDataStore) UpdateBy(p *types.UpdateParam) []Code {
	builder := Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
		Lit(p.Class.Name.PluralSnakeName()),
	)
	for idx, member := range p.Args.Members {
		builder = builder.Dot("Eq").Call(Lit(member.Name.SnakeName()), Id(fmt.Sprintf("a%d", idx)))
	}
	return []Code{
		builder,
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("UpdateByQueryBuilderContext").Call(
				Id("ctx"), Id("builder"), Id("updateMap"),
			)),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Update: %w"), Err())),
		),
		Return(Nil()),
	}
}

func (*RapidashDataStore) UpdateByPlural(p *types.UpdateParam) []Code {
	builder := Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
		Lit(p.Class.Name.PluralSnakeName()),
	)
	for idx, member := range p.Args.Members {
		builder = builder.Dot("In").Call(Lit(member.Name.SnakeName()), Id(fmt.Sprintf("a%d", idx)))
	}
	return []Code{
		builder,
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("UpdateByQueryBuilderContext").Call(
				Id("ctx"), Id("builder"), Id("updateMap"),
			)),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Update: %w"), Err())),
		),
		Return(Nil()),
	}
}

func (*RapidashDataStore) DeleteBy(p *types.DeleteParam) []Code {
	builder := Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
		Lit(p.Class.Name.PluralSnakeName()),
	)
	for idx, member := range p.Args.Members {
		builder = builder.Dot("Eq").Call(Lit(member.Name.SnakeName()), Id(fmt.Sprintf("a%d", idx)))
	}
	return []Code{
		builder,
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("DeleteByQueryBuilderContext").Call(
				Id("ctx"), Id("builder"),
			)),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Delete: %w"), Err())),
		),
		Return(Nil()),
	}
}

func (*RapidashDataStore) DeleteByPlural(p *types.DeleteParam) []Code {
	builder := Id("builder").Op(":=").Qual(p.Package("rapidash"), "NewQueryBuilder").Call(
		Lit(p.Class.Name.PluralSnakeName()),
	)
	for idx, member := range p.Args.Members {
		builder = builder.Dot("In").Call(Lit(member.Name.SnakeName()), Id(fmt.Sprintf("a%d", idx)))
	}
	return []Code{
		builder,
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("DeleteByQueryBuilderContext").Call(
				Id("ctx"), Id("builder"),
			)),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failed to Delete: %w"), Err())),
		),
		Return(Nil()),
	}
}
