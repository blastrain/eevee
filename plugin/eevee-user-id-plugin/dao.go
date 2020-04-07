package eevee

import (
	"fmt"
	"strings"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/plugin/dao"
	"go.knocknote.io/eevee/types"
)

type DAOUserIDPlugin struct {
	dao.DefaultPlugin
	DataStore *dao.DBDataStore
}

func init() {
	dao.Register("user-id", &DAOUserIDPlugin{DataStore: &dao.DBDataStore{}})
}

func (h *DAOUserIDPlugin) isUserClass(class *types.Class) bool {
	name := class.Name.SnakeName()
	return strings.HasPrefix(name, "user_")
}

func (h *DAOUserIDPlugin) newMembers(members types.Members) types.Members {
	newMembers := types.Members{}
	for _, member := range members {
		if member.Name.SnakeName() == "user_id" {
			continue
		}
		newMembers = append(newMembers, member)
	}
	return newMembers
}

func (h *DAOUserIDPlugin) newArgs(d *types.MethodDeclare) ([]string, types.ValueDeclares) {
	newArgMembers := h.newMembers(d.ArgMembers)
	argsCamelNames := []string{}
	args := types.ValueDeclares{
		{
			Name: "ctx",
			Type: &types.TypeDeclare{
				Type: &types.Type{
					PackageName: d.ImportList.Package("context"),
					Name:        "Context",
				},
			},
		},
	}
	for idx, member := range newArgMembers {
		argsCamelNames = append(argsCamelNames, member.Name.CamelName())
		args = append(args, &types.ValueDeclare{
			Name: fmt.Sprintf("a%d", idx),
			Type: member.Type,
		})
	}
	return argsCamelNames, args
}

func (h *DAOUserIDPlugin) StructFields(class *types.Class, fields types.StructFieldList) types.StructFieldList {
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
	if h.isUserClass(class) {
		values = append(values, &types.ValueDeclare{
			Name: "userID",
			Type: &types.TypeDeclare{
				Type: &types.Type{
					Name: "uint64",
				},
			},
		})
	}
	for _, value := range values {
		fields[value.Name] = value
	}
	return fields
}

func (h *DAOUserIDPlugin) ConstructorDeclare(d *types.ConstructorDeclare) error {
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
	if !h.isUserClass(d.Class) {
		return nil
	}
	d.Args = append(d.Args, &types.ValueDeclare{
		Name: "userID",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				Name: "uint64",
			},
		},
	})
	return nil
}

func (h *DAOUserIDPlugin) Constructor(p *types.ConstructorParam) []Code {
	if !h.isUserClass(p.Class) {
		return h.DataStore.Constructor(p)
	}
	return []Code{
		Return(Op("&").Id(p.ImplName).Values(Dict{
			Id("tx"):     Id("tx"),
			Id("userID"): Id("userID"),
		})),
	}
}

func (h *DAOUserIDPlugin) FindByDeclare(d *types.MethodDeclare) error {
	if !h.isUserClass(d.Class) {
		return nil
	}
	argsCamelNames, args := h.newArgs(d)
	if len(argsCamelNames) == 0 {
		d.MethodName = "FindAll"
	} else {
		d.MethodName = fmt.Sprintf("FindBy%s", strings.Join(argsCamelNames, "And"))
	}
	d.Args = args
	return nil
}

func (h *DAOUserIDPlugin) FindBy(p *types.FindParam) []Code {
	if !h.isUserClass(p.Class) {
		return h.DataStore.FindBy(p)
	}
	if p.Args.Members[0].Name.SnakeName() != "user_id" {
		return h.DataStore.FindBy(p)
	}
	queryArgs := []Code{Code(p.Args.Context()), Id("query"), p.Field("userID")}
	p.SQL.Args = p.SQL.Args[:len(p.SQL.Args)-1] // remove argument for userID
	queryArgs = append(queryArgs, p.SQL.Args...)
	return h.DataStore.FindWithQueryArgs(p, queryArgs)
}

func (h *DAOUserIDPlugin) UpdateByDeclare(d *types.MethodDeclare) error {
	if !h.isUserClass(d.Class) {
		return nil
	}
	argsCamelNames, args := h.newArgs(d)
	args = append(args, &types.ValueDeclare{
		Name: "updateMap",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				Name: "map[string]interface{}",
			},
		},
	})
	d.MethodName = fmt.Sprintf("UpdateBy%s", strings.Join(argsCamelNames, "And"))
	d.Args = args
	return nil
}

func (h *DAOUserIDPlugin) UpdateBy(p *types.UpdateParam) []Code {
	if !h.isUserClass(p.Class) {
		return h.DataStore.UpdateBy(p)
	}
	p.Args.Members = h.newMembers(p.Args.Members)
	appendStmts := []Code{Id("args").Op("=").Append(Id("args"), p.Field("userID"))}
	for idx := range p.Args.Members {
		appendStmts = append(appendStmts, Id("args").Op("=").Append(Id("args"), Id(fmt.Sprintf("a%d", idx))))
	}
	return h.DataStore.UpdateWithAppendStmts(p, appendStmts)
}

func (h *DAOUserIDPlugin) DeleteByDeclare(d *types.MethodDeclare) error {
	if !h.isUserClass(d.Class) {
		return nil
	}
	argsCamelNames, args := h.newArgs(d)
	d.MethodName = fmt.Sprintf("DeleteBy%s", strings.Join(argsCamelNames, "And"))
	d.Args = args
	return nil
}

func (h *DAOUserIDPlugin) DeleteBy(p *types.DeleteParam) []Code {
	if !h.isUserClass(p.Class) {
		return h.DataStore.DeleteBy(p)
	}
	args := []Code{p.Args.Context(), Id("query"), p.Field("userID")}
	p.SQL.Args = p.SQL.Args[:len(p.SQL.Args)-1] // remove argument for userID
	args = append(args, p.SQL.Args...)
	return h.DataStore.DeleteWithArgs(p, args)
}
