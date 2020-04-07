package types

import (
	"go.knocknote.io/eevee/code"
)

type SQL struct {
	Query      string
	Args       []code.Code
	ScanValues []code.Code
}

type DAOContext interface{}

type DataAccessParam struct {
	Class      *Class
	ClassName  func() *code.Statement
	Receiver   func() *code.Statement
	ImportList ImportList
	SQL        *SQL
}

func (p *DataAccessParam) Package(name string) string {
	return p.ImportList.Package(name)
}

func (p *DataAccessParam) Field(name string) *code.Statement {
	return p.Receiver().Dot(name)
}

type ConstructorParam struct {
	DataAccessParam
	ImplName string
	Args     *ConstructorParamArgs
}

type ConstructorParamArgs struct {
	Context func() *code.Statement
}

type CreateParam struct {
	DataAccessParam
	Args *CreateParamArgs
}

type CreateParamArgs struct {
	Context func() *code.Statement
	Value   func() *code.Statement
}

type UpdateParam struct {
	DataAccessParam
	Args *UpdateParamArgs
}

type UpdateParamArgs struct {
	Context   func() *code.Statement
	Value     func() *code.Statement
	UpdateMap func() *code.Statement
	Members   []*Member
}

type DeleteParam struct {
	DataAccessParam
	Args *DeleteParamArgs
}

type DeleteParamArgs struct {
	Context func() *code.Statement
	Value   func() *code.Statement
	Members []*Member
}

type FindParam struct {
	DataAccessParam
	Args                *FindParamArgs
	IsSingleReturnValue bool
}

type FindParamArgs struct {
	Context func() *code.Statement
	Members []*Member
}

type CountParam struct {
	DataAccessParam
	Args *CountParamArgs
}

type CountParamArgs struct {
	Context func() *code.Statement
	Members []*Member
}
