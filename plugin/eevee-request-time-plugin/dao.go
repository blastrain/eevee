package eevee

import (
	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/plugin/dao"
	"go.knocknote.io/eevee/types"
)

type DAORequestTimePlugin struct {
	dao.DefaultPlugin
}

func init() {
	dao.Register("request-time", &DAORequestTimePlugin{})
}

func (*DAORequestTimePlugin) BeforeCreate(p *types.CreateParam) []Code {
	createdAt := p.Class.MemberByName("created_at")
	updatedAt := p.Class.MemberByName("updated_at")
	if createdAt == nil || updatedAt == nil {
		return nil
	}
	code := []Code{
		Id("v").Op(":=").Id("ctx").Dot("Value").Call(Lit("REQUEST_TIME")),
		List(Id("requestTime"), Id("ok")).Op(":=").Id("v").Assert(Qual(p.Package("time"), "Time")),
		If(
			Add(Op("!")).Id("ok"),
		).Block(
			Return(Qual(p.Package("xerrors"), "New").Call(Lit("cannot convert time.Time value from ctx.Value(`REQUEST_TIME`)"))),
		),
	}
	if createdAt.Type.IsPointer {
		code = append(code, Id("value").Dot("CreatedAt").Op("=").Op("&").Id("requestTime"))
	} else {
		code = append(code, Id("value").Dot("CreatedAt").Op("=").Id("requestTime"))
	}
	if updatedAt.Type.IsPointer {
		code = append(code, Id("value").Dot("UpdatedAt").Op("=").Op("&").Id("requestTime"))
	} else {
		code = append(code, Id("value").Dot("UpdatedAt").Op("=").Id("requestTime"))
	}
	return code
}

func (*DAORequestTimePlugin) BeforeUpdate(p *types.UpdateParam) []Code {
	updatedAt := p.Class.MemberByName("updated_at")
	if updatedAt == nil {
		return nil
	}
	code := []Code{
		Id("v").Op(":=").Id("ctx").Dot("Value").Call(Lit("REQUEST_TIME")),
		List(Id("requestTime"), Id("ok")).Op(":=").Id("v").Assert(Qual(p.Package("time"), "Time")),
		If(
			Add(Op("!")).Id("ok"),
		).Block(
			Return(Qual(p.Package("xerrors"), "New").Call(Lit("cannot convert time.Time value from ctx.Value(`REQUEST_TIME`)"))),
		),
	}
	if updatedAt.Type.IsPointer {
		code = append(code, Id("value").Dot("UpdatedAt").Op("=").Op("&").Id("requestTime"))
	} else {
		code = append(code, Id("value").Dot("UpdatedAt").Op("=").Id("requestTime"))
	}
	return code
}
