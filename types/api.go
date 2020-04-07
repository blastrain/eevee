package types

import (
	"go.knocknote.io/eevee/code"
)

type APIResponseHelper struct {
	Class        *Class
	ReceiverName string
	ImportList   ImportList
}

func (h *APIResponseHelper) Receiver() *code.Statement {
	return code.Id(h.ReceiverName)
}

func (h *APIResponseHelper) Field(name string) *code.Statement {
	return h.Receiver().Dot(name)
}

func (h *APIResponseHelper) MethodCall(name string, args ...code.Code) *code.Statement {
	return h.Field(name).Call(args...)
}

func (h *APIResponseHelper) GetClass() *Class {
	return h.Class
}

func (h *APIResponseHelper) GetImportList() ImportList {
	return h.ImportList
}

func (h *APIResponseHelper) CreateMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: h.Class.Name.CamelName(),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *APIResponseHelper) CreateCollectionMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: h.Class.Name.PluralCamelName(),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *APIResponseHelper) Package(name string) string {
	return h.ImportList.Package(name)
}

func (h *APIResponseHelper) IsModelPackage() bool {
	return false
}
