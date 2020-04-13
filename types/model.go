package types

import (
	"fmt"
	"sort"

	"go.knocknote.io/eevee/code"
)

type ModelMethodHelper struct {
	Class        *Class
	ReceiverName string
	ImportList   ImportList
}

func (h *ModelMethodHelper) Receiver() *code.Statement {
	return code.Id(h.ReceiverName)
}

func (h *ModelMethodHelper) Field(name string) *code.Statement {
	return h.Receiver().Dot(name)
}

func (h *ModelMethodHelper) MethodCall(name string, args ...code.Code) *code.Statement {
	return h.Field(name).Call(args...)
}

func (h *ModelMethodHelper) ModelName() string {
	return h.Class.Name.CamelName()
}

func (h *ModelMethodHelper) ModelCollectionName() string {
	return h.Class.Name.PluralCamelName()
}

func (h *ModelMethodHelper) GetImportList() ImportList {
	return h.ImportList
}

func (h *ModelMethodHelper) ModelType() *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			Name: h.ModelName(),
		},
		IsPointer: true,
	}
}

func (h *ModelMethodHelper) GetClass() *Class {
	return h.Class
}

func (h *ModelMethodHelper) ModelCollectionType() *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			Name: h.ModelCollectionName(),
		},
		IsPointer: true,
	}
}

func (h *ModelMethodHelper) CreateMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: h.ModelName(),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *ModelMethodHelper) CreateCollectionMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: h.ModelCollectionName(),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *ModelMethodHelper) CreateMultipleCollectionMethodDeclare() *MethodDeclare {
	return &MethodDeclare{
		Class:             h.Class,
		ReceiverName:      h.ReceiverName,
		ReceiverClassName: fmt.Sprintf("%sCollection", h.ModelCollectionName()),
		ImportList:        h.ImportList,
		Args:              ValueDeclares{},
		Return:            ValueDeclares{},
	}
}

func (h *ModelMethodHelper) CollectionProperties() ValueDeclares {
	propertyMap := map[string]*ValueDeclare{}
	for _, member := range h.Class.RelationMembers() {
		relation := member.Relation
		if relation.Custom {
			continue
		}
		name := member.CollectionName()
		memberClass := member.Type.Class()
		typeName := memberClass.Name.PluralCamelName()
		propertyMap[name.CamelLowerName()] = &ValueDeclare{
			Name: name.CamelLowerName(),
			Type: &TypeDeclare{
				Type: &Type{
					Name: typeName,
				},
				IsPointer: true,
			},
		}
		if relation.All {
			continue
		}
		internalMember := h.Class.MemberByName(relation.Internal.SnakeName())
		propertyMap[internalMember.Name.PluralCamelLowerName()] = &ValueDeclare{
			Name: internalMember.Name.PluralCamelLowerName(),
			Type: &TypeDeclare{
				Type: &Type{
					Name: "[]" + internalMember.Type.FormatName(h.ImportList),
				},
			},
		}
	}
	var sortedKeys []string
	for k := range propertyMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	properties := ValueDeclares{}
	for _, k := range sortedKeys {
		properties = append(properties, propertyMap[k])
	}
	return properties
}

func (h *ModelMethodHelper) IsModelPackage() bool {
	return true
}

func (h *ModelMethodHelper) DAO() *code.Statement {
	return h.Receiver().Dot(h.DAOName())
}

func (h *ModelMethodHelper) DAOName() string {
	return fmt.Sprintf("%sDAO", h.Class.Name.CamelLowerName())
}

func (h *ModelMethodHelper) Package(name string) string {
	return h.ImportList.Package(name)
}
