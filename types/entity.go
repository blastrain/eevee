package types

import (
	. "go.knocknote.io/eevee/code"
)

type EntityMethodHelper struct {
	Class        *Class
	ReceiverName string
	ImportList   ImportList
}

func (h *EntityMethodHelper) CreateMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: h.Class.Name.CamelName(),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *EntityMethodHelper) CreatePluralMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: h.Class.Name.PluralCamelName(),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *EntityMethodHelper) Field(name string) *Statement {
	return Id(h.ReceiverName).Dot(name)
}

func (h *EntityMethodHelper) Package(name string) string {
	return h.ImportList.Package(name)
}
