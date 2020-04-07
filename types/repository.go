package types

import (
	"fmt"

	"go.knocknote.io/eevee/code"
)

type RepositoryMethodHelper struct {
	AppName      string
	Class        *Class
	ReceiverName string
	ImportList   ImportList
}

func (h *RepositoryMethodHelper) Receiver() *code.Statement {
	return code.Id(h.ReceiverName)
}

func (h *RepositoryMethodHelper) ReceiverClassName() string {
	return fmt.Sprintf("%sImpl", h.Class.Name.CamelName())
}

func (h *RepositoryMethodHelper) CreateMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: h.ReceiverClassName(),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *RepositoryMethodHelper) DAO() *code.Statement {
	return h.Receiver().Dot(h.DAOName())
}

func (h *RepositoryMethodHelper) DAOName() string {
	return fmt.Sprintf("%sDAO", h.Class.Name.CamelLowerName())
}

func (h *RepositoryMethodHelper) Package(name string) string {
	return h.ImportList.Package(name)
}

func (h *RepositoryMethodHelper) ContextType() *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			PackageName: "context",
			Name:        "Context",
		},
	}
}

func (h *RepositoryMethodHelper) EntityClassType() *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			PackageName: h.Package("entity"),
			Name:        h.Class.Name.CamelName(),
		},
		IsPointer: true,
	}
}

func (h *RepositoryMethodHelper) EntityCollectionClassType() *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			PackageName: h.Package("entity"),
			Name:        h.Class.Name.PluralCamelName(),
		},
	}
}

func (h *RepositoryMethodHelper) ModelClassType() *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			PackageName: h.Package("model"),
			Name:        h.Class.Name.CamelName(),
		},
		IsPointer: true,
	}
}

func (h *RepositoryMethodHelper) ModelCollectionClassType() *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			PackageName: h.Package("model"),
			Name:        h.Class.Name.PluralCamelName(),
		},
		IsPointer: true,
	}
}

func (h *RepositoryMethodHelper) FieldsToFind() code.Dict {
	propertyMap := code.Dict{}
	for _, member := range h.Class.RelationMembers() {
		relation := member.Relation
		if relation.Custom || relation.All {
			continue
		}
		internalMember := h.Class.MemberByName(relation.Internal.SnakeName())
		propertyMap[code.Id(internalMember.Name.PluralCamelLowerName())] = code.Id("entities").Dot(internalMember.Name.PluralCamelName()).Call()
	}
	return propertyMap
}
