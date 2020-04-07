package eevee

import (
	"go.knocknote.io/eevee/plugin/dao"
	"go.knocknote.io/eevee/types"
)

type OctilleryDataAccessHandler struct {
	dao.DefaultPlugin
}

func init() {
	dao.Register("octillery", &OctilleryDataAccessHandler{})
}

func (*OctilleryDataAccessHandler) Imports(pkgs types.ImportList) types.ImportList {
	for _, decl := range []*types.ImportDeclare{
		{
			Path: "go.knocknote.io/octillery/database/sql",
			Name: "sql",
		},
	} {
		pkgs[decl.Name] = decl
	}
	return pkgs
}
