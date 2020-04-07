package dao

import (
	"fmt"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

type DBDataStore struct {
}

func init() {
	RegisterDataStore("db", &DBDataStore{})
}

func (*DBDataStore) Imports(pkgs types.ImportList) types.ImportList {
	for _, decl := range []*types.ImportDeclare{
		{
			Path: "fmt",
			Name: "fmt",
		},
		{
			Path: "database/sql",
			Name: "sql",
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

func (*DBDataStore) StructFields(class *types.Class, fields types.StructFieldList) types.StructFieldList {
	values := types.ValueDeclares{
		{
			Name: "tx",
			Type: &types.TypeDeclare{
				Type: &types.Type{
					PackageName: "sql",
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

func (*DBDataStore) ConstructorDeclare(d *types.ConstructorDeclare) error {
	d.Args = append(d.Args, &types.ValueDeclare{
		Name: "tx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: d.Package("sql"),
				Name:        "Tx",
			},
			IsPointer: true,
		},
	})
	return nil
}

func (*DBDataStore) Constructor(p *types.ConstructorParam) []Code {
	return []Code{
		Return(Op("&").Id(p.ImplName).Values(Dict{
			Id("tx"): Id("tx"),
		})),
	}
}

func (*DBDataStore) Create(p *types.CreateParam) []Code {
	values := []*Statement{}
	var idMember *types.Member
	for _, member := range p.Class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		values = append(values, p.Args.Value().Dot(member.Name.CamelName()))
		if member.Name.SnakeName() == "id" {
			idMember = member
		}
	}
	args := []Code{p.Args.Context(), Id("query")}
	for _, value := range values {
		args = append(args, value)
	}
	return []Code{
		Id("query").Op(":=").Lit(p.SQL.Query),
		List(Id("result"), Err()).Op(":=").Add(p.Field("tx").Dot("ExecContext").Call(args...)),
		If(
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err"))),
		),
		List(Id("id"), Err()).Op(":=").Id("result").Dot("LastInsertId").Call(),
		If(Err().Op("!=").Nil()).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("cannot get LastInsertId: %w"), Id("err"))),
		),
		Id("value").Dot("ID").Op("=").Add(idMember.Type.Code(p.ImportList)).Call(Id("id")), // TODO: need to AUTO_INCREMENT column only
		Return(Nil()),
	}
}

func (*DBDataStore) Update(p *types.UpdateParam) []Code {
	values := []Code{}
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
		} else {
			values = append(values, Id("value").Dot(member.Name.CamelName()))
		}
	}
	values = append(values, Id("value").Dot(idMember.Name.CamelName()))
	return []Code{
		Id("args").Op(":=").Index().Interface().Values(values...),
		Id("query").Op(":=").Lit(p.SQL.Query),
		If(
			List(Id("_"), Err()).Op(":=").Add(
				p.Field("tx").Dot("ExecContext").Call(p.Args.Context(), Id("query"), Id("args").Op("...")),
			),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err"))),
		),
		Return(Nil()),
	}
}

func (*DBDataStore) Delete(p *types.DeleteParam) []Code {
	return []Code{
		Id("query").Op(":=").Lit(p.SQL.Query),
		If(
			List(Id("_"), Err()).Op(":=").Add(
				p.Field("tx").Dot("ExecContext").Call(p.Args.Context(), Id("query"), Id("value").Dot("ID")),
			),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err"))),
		),
		Return(Nil()),
	}
}

func (*DBDataStore) FindAll(p *types.FindParam) []Code {
	return []Code{
		Id("values").Op(":=").Qual(p.Package("entity"), p.Class.Name.PluralCamelName()).Block(),
		Id("query").Op(":=").Lit(p.SQL.Query),
		List(Id("rows"), Err()).Op(":=").Add(p.Field("tx").Dot("QueryContext").Call(p.Args.Context(), Id("query"))),
		If(Err().Op("!=").Nil()).Block(
			Return(List(Id("values"), Qual(p.Package("xerrors"), "Errorf").Call(Lit("falure query %s: %w"), Id("query"), Id("err")))),
		),
		Defer().Func().Call().Block(
			If(
				Err().Op(":=").Id("rows").Dot("Close").Call(),
				Err().Op("!=").Nil(),
			).Block(
				Id("e").Op("=").Qual(p.Package("xerrors"), "Errorf").Call(Lit("cannot close rows: %w"), Err()),
			),
		).Call(),
		For(Id("rows").Dot("Next").Call()).Block(
			Var().Id("value").Qual(p.Package("entity"), p.Class.Name.CamelName()),
			If(
				Err().Op(":=").Id("rows").Dot("Scan").Call(p.SQL.ScanValues...),
				Err().Op("!=").Nil(),
			).Block(
				Return(List(Id("values"), Qual(p.Package("xerrors"), "Errorf").Call(Lit("cannot scan value: %w"), Id("err")))),
			),
			Id("values").Op("=").Append(Id("values"), Op("&").Id("value")),
		),
		Return(List(Id("values"), Nil())),
	}
}

func (*DBDataStore) Count(p *types.CountParam) []Code {
	return []Code{
		Var().Id("value").Id("int64"),
		Id("query").Op(":=").Lit(p.SQL.Query),
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("QueryRowContext").Call(p.Args.Context(), Id("query"))).Dot("Scan").Call(Op("&").Id("value")),
			Err().Op("!=").Nil(),
		).Block(
			If(Err().Op("==").Qual(p.Package("sql"), "ErrNoRows")).Block(
				Return(Lit(0), Nil()),
			),
			Return(List(Lit(0), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err")))),
		),
		Return(Id("value"), Nil()),
	}
}

func (s *DBDataStore) findForSingleReturnValue(p *types.FindParam, query string, queryArgs []Code, scanValues []Code) []Code {
	return []Code{
		Var().Id("value").Qual(p.Package("entity"), p.Class.Name.CamelName()),
		Id("query").Op(":=").Lit(query),
		If(
			Err().Op(":=").Add(p.Field("tx").Dot("QueryRowContext").Call(queryArgs...)).Dot("Scan").Call(scanValues...),
			Err().Op("!=").Nil(),
		).Block(
			If(Err().Op("==").Qual(p.Package("sql"), "ErrNoRows")).Block(
				Return(Nil(), Nil()),
			),
			Return(List(Nil(), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err")))),
		),
		Return(List(Op("&").Id("value"), Nil())),
	}
}

func (s *DBDataStore) findForSliceReturnValue(p *types.FindParam, query string, queryArgs []Code, scanValues []Code) []Code {
	return []Code{
		Id("values").Op(":=").Qual(p.Package("entity"), p.Class.Name.PluralCamelName()).Block(),
		Id("query").Op(":=").Lit(query),
		List(Id("rows"), Err()).Op(":=").Add(p.Field("tx").Dot("QueryContext").Call(queryArgs...)),
		If(Err().Op("!=").Nil()).Block(
			Return(List(Id("values"), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err")))),
		),
		Defer().Func().Call().Block(
			If(
				Err().Op(":=").Id("rows").Dot("Close").Call(),
				Err().Op("!=").Nil(),
			).Block(
				Id("e").Op("=").Qual(p.Package("xerrors"), "Errorf").Call(Lit("cannot close rows: %w"), Err()),
			),
		).Call(),
		For(Id("rows").Dot("Next").Call()).Block(
			Var().Id("value").Qual(p.Package("entity"), p.Class.Name.CamelName()),
			If(
				Err().Op(":=").Id("rows").Dot("Scan").Call(scanValues...),
				Err().Op("!=").Nil(),
			).Block(
				Return(List(Id("values"), Qual(p.Package("xerrors"), "Errorf").Call(Lit("cannot scan value: %w"), Id("err")))),
			),
			Id("values").Op("=").Append(Id("values"), Op("&").Id("value")),
		),
		Return(List(Id("values"), Nil())),
	}
}

func (s *DBDataStore) FindWithQueryArgs(p *types.FindParam, queryArgs []Code) []Code {
	if p.IsSingleReturnValue {
		return s.findForSingleReturnValue(p, p.SQL.Query, queryArgs, p.SQL.ScanValues)
	}
	return s.findForSliceReturnValue(p, p.SQL.Query, queryArgs, p.SQL.ScanValues)
}

func (s *DBDataStore) FindBy(p *types.FindParam) []Code {
	queryArgs := []Code{Code(p.Args.Context()), Id("query")}
	queryArgs = append(queryArgs, p.SQL.Args...)
	return s.FindWithQueryArgs(p, queryArgs)
}

func (*DBDataStore) FindByPlural(p *types.FindParam) []Code {
	return []Code{
		Id("values").Op(":=").Qual(p.Package("entity"), p.Class.Name.PluralCamelName()).Block(),
		Id("query").Op(":=").Lit(p.SQL.Query),
		Id("args").Op(":=").Index().Interface().Values(),
		Id("placeholders").Op(":=").Make(Index().String(), Lit(0), Len(Id("a0"))),
		For(
			List(Id("_"), Id("v")).Op(":=").Range().Id("a0"),
		).Block(
			Id("args").Op("=").Append(Id("args"), Id("v")),
			Id("placeholders").Op("=").Append(Id("placeholders"), Lit("?")),
		),
		Id("selectQuery").Op(":=").Qual(p.Package("fmt"), "Sprintf").Call(Id("query"), Qual(p.Package("strings"), "Join").Call(Id("placeholders"), Lit(", "))),
		List(Id("rows"), Err()).Op(":=").Add(p.Field("tx").Dot("QueryContext").Call(p.Args.Context(), Id("selectQuery"), Id("args").Op("..."))),
		If(Err().Op("!=").Nil()).Block(
			Return(List(Id("values"), Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err")))),
		),
		Defer().Func().Call().Block(
			If(
				Err().Op(":=").Id("rows").Dot("Close").Call(),
				Err().Op("!=").Nil(),
			).Block(
				Id("e").Op("=").Qual(p.Package("xerrors"), "Errorf").Call(Lit("cannot close rows: %w"), Err()),
			),
		).Call(),
		For(Id("rows").Dot("Next").Call()).Block(
			Var().Id("value").Qual(p.Package("entity"), p.Class.Name.CamelName()),
			If(
				Err().Op(":=").Id("rows").Dot("Scan").Call(p.SQL.ScanValues...),
				Err().Op("!=").Nil(),
			).Block(
				Return(List(Id("values"), Qual(p.Package("xerrors"), "Errorf").Call(Lit("cannot scan value: %w"), Id("err")))),
			),
			Id("values").Op("=").Append(Id("values"), Op("&").Id("value")),
		),
		Return(List(Id("values"), Nil())),
	}
}

func (*DBDataStore) UpdateWithAppendStmts(p *types.UpdateParam, appendStmts []Code) []Code {
	codes := []Code{
		Id("columns").Op(":=").Index().String().Values(),
		Id("args").Op(":=").Index().Interface().Values(),
		For(
			List(Id("column"), Id("v")).Op(":=").Range().Id("updateMap"),
		).Block(
			Id("columns").Op("=").Append(Id("columns"), Qual(p.Package("fmt"), "Sprintf").Call(Lit("`%s` = ?"), Id("column"))),
			Id("args").Op("=").Append(Id("args"), Id("v")),
		),
	}
	codes = append(codes, appendStmts...)
	codes = append(codes, []Code{
		Id("query").Op(":=").Qual(p.Package("fmt"), "Sprintf").Call(Lit(p.SQL.Query), Qual(p.Package("strings"), "Join").Call(Id("columns"), Lit(", "))),
		If(
			List(Id("_"), Err()).Op(":=").Add(p.Field("tx").Dot("ExecContext").Call(p.Args.Context(), Id("query"), Id("args").Op("..."))),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err"))),
		),
		Return(Nil()),
	}...)
	return codes
}

func (s *DBDataStore) UpdateBy(p *types.UpdateParam) []Code {
	appendStmts := []Code{}
	for idx := range p.Args.Members {
		appendStmts = append(appendStmts, Id("args").Op("=").Append(Id("args"), Id(fmt.Sprintf("a%d", idx))))
	}
	return s.UpdateWithAppendStmts(p, appendStmts)
}

func (*DBDataStore) UpdateByPlural(p *types.UpdateParam) []Code {
	return []Code{
		Id("columns").Op(":=").Index().String().Values(),
		Id("args").Op(":=").Index().Interface().Values(),
		For(
			List(Id("column"), Id("v")).Op(":=").Range().Id("updateMap"),
		).Block(
			Id("columns").Op("=").Append(Id("columns"), Qual(p.Package("fmt"), "Sprintf").Call(Lit("`%s` = ?"), Id("column"))),
			Id("args").Op("=").Append(Id("args"), Id("v")),
		),
		For(
			List(Id("_"), Id("v")).Op(":=").Range().Id("a0"),
		).Block(
			Id("args").Op("=").Append(Id("args"), Id("v")),
		),
		Id("placeholders").Op(":=").Make(Index().String(), Lit(0), Len(Id("a0"))),
		For(
			Range().Id("a0"),
		).Block(
			Id("placeholders").Op("=").Append(Id("placeholders"), Lit("?")),
		),
		Id("query").Op(":=").Qual(p.Package("fmt"), "Sprintf").Call(
			Lit(p.SQL.Query),
			Qual(p.Package("strings"), "Join").Call(Id("columns"), Lit(", ")),
			Qual(p.Package("strings"), "Join").Call(Id("placeholders"), Lit(", ")),
		),
		If(
			List(Id("_"), Err()).Op(":=").Add(p.Field("tx").Dot("ExecContext").Call(p.Args.Context(), Id("query"), Id("args").Op("..."))),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err"))),
		),
		Return(Nil()),
	}
}

func (*DBDataStore) DeleteWithArgs(p *types.DeleteParam, args []Code) []Code {
	return []Code{
		Id("query").Op(":=").Lit(p.SQL.Query),
		If(
			List(Id("_"), Err()).Op(":=").Add(p.Field("tx").Dot("ExecContext").Call(args...)),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err"))),
		),
		Return(Nil()),
	}
}

func (s *DBDataStore) DeleteBy(p *types.DeleteParam) []Code {
	args := []Code{p.Args.Context(), Id("query")}
	args = append(args, p.SQL.Args...)
	return s.DeleteWithArgs(p, args)
}

func (*DBDataStore) DeleteByPlural(p *types.DeleteParam) []Code {
	return []Code{
		Id("args").Op(":=").Index().Interface().Values(),
		Id("placeholders").Op(":=").Make(Index().String(), Lit(0), Len(Id("a0"))),
		For(
			List(Id("_"), Id("v")).Op(":=").Range().Id("a0"),
		).Block(
			Id("args").Op("=").Append(Id("args"), Id("v")),
			Id("placeholders").Op("=").Append(Id("placeholders"), Lit("?")),
		),
		Id("query").Op(":=").Qual(p.Package("fmt"), "Sprintf").Call(
			Lit(p.SQL.Query),
			Qual(p.Package("strings"), "Join").Call(Id("placeholders"), Lit(", ")),
		),
		If(
			List(Id("_"), Err()).Op(":=").Add(
				p.Field("tx").Dot("ExecContext").Call(p.Args.Context(), Id("query"), Id("args").Op("...")),
			),
			Err().Op("!=").Nil(),
		).Block(
			Return(Qual(p.Package("xerrors"), "Errorf").Call(Lit("failure query %s: %w"), Id("query"), Id("err"))),
		),
		Return(Nil()),
	}
}
