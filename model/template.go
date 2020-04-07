package model

import (
	"fmt"
	"strings"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

func (g *Generator) Constructor(h *types.ModelMethodHelper) Code {
	decl := &types.ConstructorDeclare{
		Class:      h.Class,
		MethodName: fmt.Sprintf("New%s", h.Class.Name.CamelName()),
		Args: types.ValueDeclares{
			{
				Name: "value",
				Type: &types.TypeDeclare{
					Type: &types.Type{
						PackageName: h.Package("entity"),
						Name:        h.Class.Name.CamelName(),
					},
					IsPointer: true,
				},
			},
			{
				Name: fmt.Sprintf("%sDAO", h.Class.Name.CamelLowerName()),
				Type: &types.TypeDeclare{
					Type: &types.Type{
						PackageName: h.Package("dao"),
						Name:        h.Class.Name.CamelName(),
					},
				},
			},
		},
		Return: types.ValueDeclares{
			{
				Type: h.ModelType(),
			},
		},
	}
	properties := Dict{
		Id(h.Class.Name.CamelName()):                            Id("value"),
		Id(fmt.Sprintf("%sDAO", h.Class.Name.CamelLowerName())): Id(fmt.Sprintf("%sDAO", h.Class.Name.CamelLowerName())),
	}
	return decl.MethodInterface(h.ImportList).Block(
		Return(Op("&").Id(h.Class.Name.CamelName()).Values(properties)),
	)
}

func (g *Generator) CollectionConstructor(h *types.ModelMethodHelper) Code {
	decl := &types.ConstructorDeclare{
		Class:      h.Class,
		MethodName: fmt.Sprintf("New%s", h.Class.Name.PluralCamelName()),
		Args: types.ValueDeclares{
			{
				Name: "entities",
				Type: &types.TypeDeclare{
					Type: &types.Type{
						PackageName: h.Package("entity"),
						Name:        h.Class.Name.PluralCamelName(),
					},
				},
			},
		},
		Return: types.ValueDeclares{
			{
				Type: h.ModelCollectionType(),
			},
		},
	}
	properties := Dict{
		Id("values"): Make(Index().Id(fmt.Sprintf("*%s", h.Class.Name.CamelName())), Lit(0), Len(Id("entities"))),
	}
	definedPropertyMap := map[string]struct{}{}
	for _, member := range h.Class.RelationMembers() {
		relation := member.Relation
		if relation.Custom {
			continue
		}
		if relation.All {
			continue
		}
		internalMember := h.Class.MemberByName(relation.Internal.SnakeName())
		propertyName := internalMember.Name.PluralCamelLowerName()
		if _, exists := definedPropertyMap[propertyName]; exists {
			continue
		}
		properties[Id(propertyName)] = Id("entities").Dot(internalMember.Name.PluralCamelName()).Call()
		definedPropertyMap[propertyName] = struct{}{}
	}
	return decl.MethodInterface(h.ImportList).Block(
		Return(Op("&").Id(h.Class.Name.PluralCamelName()).Values(properties)),
	)
}

func (g *Generator) Create(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Create"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.DAO().Op("==").Nil()).Block(
				Comment("for testing"),
				Return(Nil()),
			),
			If(h.Receiver().Dot("isAlreadyCreated")).Block(
				Return(Qual(h.Package("xerrors"), "New").Call(Lit("this instance has already created"))),
			),
			If(
				Err().Op(":=").Add(h.DAO().Dot("Create").Call(Id("ctx"), h.Receiver().Dot(h.Class.Name.CamelName()))),
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Create: %w"), Err())),
			),
			h.Receiver().Dot("savedValue").Op("=").Op("*").Add(h.Receiver().Dot(h.Class.Name.CamelName())),
			h.Receiver().Dot("isAlreadyCreated").Op("=").True(),
			Return(Nil()),
		},
	}
}

func (g *Generator) Update(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Update"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	conditions := []Code{}
	for _, member := range h.Class.Members {
		if member.Relation != nil || member.Extend {
			continue
		}
		conditions = append(conditions, If(h.Receiver().Dot("savedValue").Dot(member.Name.CamelName()).
			Op("!=").Add(h.Receiver().Dot(member.Name.CamelName()))).Block(
			Id("isRequiredUpdate").Op("=").True(),
		))
	}
	body := []Code{
		If(h.DAO().Op("==").Nil()).Block(
			Comment("for testing"),
			Return(Nil()),
		),
		Id("isRequiredUpdate").Op(":=").False(),
	}
	body = append(body, conditions...)
	body = append(body, []Code{
		If(Op("!").Id("isRequiredUpdate")).Block(Return(Nil())),
		If(Err().Op(":=").Add(
			h.DAO().Dot("Update").Call(Id("ctx"), h.Receiver().Dot(h.Class.Name.CamelName())),
		), Err().Op("!=").Nil()).Block(
			Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Update: %w"), Err())),
		),
		h.Receiver().Dot("savedValue").Op("=").Op("*").Add(h.Receiver().Dot(h.Class.Name.CamelName())),
		Return(Nil()),
	}...)
	return &types.Method{
		Decl: decl,
		Body: body,
	}
}

func (g *Generator) Delete(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Delete"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.DAO().Op("==").Nil()).Block(
				Comment("for testing"),
				Return(Nil()),
			),
			If(Err().Op(":=").Add(h.DAO().Dot("DeleteByID").Call(Id("ctx"), h.Receiver().Dot("ID"))), Err().Op("!=").Nil()).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Delete: %w"), Err())),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) SetAlreadyCreated(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "SetAlreadyCreated"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "isAlreadyCreated",
		Type: types.TypeDeclareWithType(types.Bool),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			h.Receiver().Dot("isAlreadyCreated").Op("=").Id("isAlreadyCreated"),
		},
	}
}

func (g *Generator) SetSavedValue(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "SetSavedValue"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "savedValue",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: h.Package("entity"),
				Name:        h.Class.Name.CamelName(),
			},
			IsPointer: true,
		},
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			h.Receiver().Dot("savedValue").Op("=").Op("*").Id("savedValue"),
		},
	}
}

func (g *Generator) SetConverter(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "SetConverter"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "conv",
		Type: types.TypeDeclareWithName("ModelConverter"),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			h.Receiver().Dot("conv").Op("=").Id("conv"),
		},
	}
}

func (g *Generator) Save(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Save"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Dot("isAlreadyCreated")).Block(
				If(Err().Op(":=").Add(h.Receiver().Dot("Update").Call(Id("ctx"))), Err().Op("!=").Nil()).Block(
					Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Update: %w"), Err())),
				),
				Return(Nil()),
			),
			If(Err().Op(":=").Add(h.Receiver().Dot("Create").Call(Id("ctx"))), Err().Op("!=").Nil()).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Create: %w"), Err())),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) CreateForCollection(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Create"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(
				Err().Op(":=").Add(h.Receiver().Dot("EachWithError").Call(
					Func().Params(Id("v").Op("*").Id(h.Class.Name.CamelName())).Id("error").Block(
						If(Err().Op(":=").Id("v").Dot("Create").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
							Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Create: %w"), Err())),
						),
						Return(Nil()),
					),
				)),
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(
					Lit(fmt.Sprintf("interrupt iteration for %s: %%w", h.Class.Name.PluralCamelName())),
					Err(),
				)),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) UpdateForCollection(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Update"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(
				Err().Op(":=").Add(h.Receiver().Dot("EachWithError").Call(
					Func().Params(Id("v").Op("*").Id(h.Class.Name.CamelName())).Id("error").Block(
						If(Err().Op(":=").Id("v").Dot("Update").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
							Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Update: %w"), Err())),
						),
						Return(Nil()),
					),
				)),
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(
					Lit(fmt.Sprintf("interrupt iteration for %s: %%w", h.Class.Name.PluralCamelName())),
					Err(),
				)),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) SaveForCollection(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Save"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(
				Err().Op(":=").Add(h.Receiver().Dot("EachWithError").Call(
					Func().Params(Id("v").Op("*").Id(h.Class.Name.CamelName())).Id("error").Block(
						If(Err().Op(":=").Id("v").Dot("Save").Call(Id("ctx")), Err().Op("!=").Nil()).Block(
							Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to Save: %w"), Err())),
						),
						Return(Nil()),
					),
				)),
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(
					Lit(fmt.Sprintf("interrupt iteration for %s: %%w", h.Class.Name.PluralCamelName())),
					Err(),
				)),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) NewCollection(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = fmt.Sprintf("new%s", h.Class.Name.PluralCamelName())
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "values",
		Type: types.TypeDeclareWithName(fmt.Sprintf("[]*%s", h.Class.Name.CamelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	dict := Dict{
		Id("values"): Id("values"),
	}
	for _, property := range h.CollectionProperties() {
		dict[Id(property.Name)] = h.Receiver().Dot(property.Name)
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Return(Op("&").Id(h.ModelCollectionName()).Values(dict)),
		},
	}
}

func (g *Generator) Each(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Each"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "iter",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s)", h.Class.Name.CamelName())),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return()),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				Id("iter").Call(Id("value")),
			),
		},
	}
}

func (g *Generator) EachIndex(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "EachIndex"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "iter",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(int, *%s)", h.Class.Name.CamelName())),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return()),
			For(List(Id("idx"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				Id("iter").Call(Id("idx"), Id("value")),
			),
		},
	}
}

func (g *Generator) EachWithError(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "EachWithError"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "iter",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) error", h.Class.Name.CamelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Err().Op(":=").Id("iter").Call(Id("value")), Err().Op("!=").Nil()).Block(
					Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to iteration: %w"), Err())),
				),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) EachIndexWithError(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "EachIndexWithError"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "iter",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(int, *%s) error", h.Class.Name.CamelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			For(List(Id("idx"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Err().Op(":=").Id("iter").Call(Id("idx"), Id("value")), Err().Op("!=").Nil()).Block(
					Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to iteration: %w"), Err())),
				),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) Map(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Map"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "mapFunc",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) *%s", h.ModelName(), h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			Id("mappedValues").Op(":=").Index().Op("*").Id(h.ModelName()).Values(),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				Id("mappedValue").Op(":=").Id("mapFunc").Call(Id("value")),
				If(Id("mappedValue").Op("!=").Nil()).Block(
					Id("mappedValues").Op("=").Append(Id("mappedValues"), Id("mappedValue")),
				),
			),
			Return(h.Receiver().Dot(fmt.Sprintf("new%s", h.Class.Name.PluralCamelName())).Call(Id("mappedValues"))),
		},
	}
}

func (g *Generator) Any(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Any"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "cond",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) bool", h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Bool),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(False())),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Id("cond").Call(Id("value"))).Block(
					Return(True()),
				),
			),
			Return(False()),
		},
	}
}

func (g *Generator) Some(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Some"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "cond",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) bool", h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Bool),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Return(h.Receiver().Dot("Any").Call(Id("cond"))),
		},
	}
}

func (g *Generator) IsIncluded(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "IsIncluded"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "cond",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) bool", h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Bool),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Return(h.Receiver().Dot("Any").Call(Id("cond"))),
		},
	}
}

func (g *Generator) All(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "All"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "cond",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) bool", h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Bool),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(False())),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Op("!").Id("cond").Call(Id("value"))).Block(
					Return(False()),
				),
			),
			Return(True()),
		},
	}
}

func (g *Generator) Sort(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Sort"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "compare",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s, *%s) bool", h.ModelName(), h.ModelName())),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return()),
			Qual("sort", "Slice").Call(
				h.Field("values"), Func().Params(List(Id("i"), Id("j").Id("int"))).Id("bool").Block(
					Return(
						Id("compare").Call(h.Field("values").Index(Id("i")), h.Field("values").Index(Id("j"))),
					),
				),
			),
		},
	}
}

func (g *Generator) SortStable(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "SortStable"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "compare",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s, *%s) bool", h.ModelName(), h.ModelName())),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return()),
			Qual("sort", "SliceStable").Call(
				h.Field("values"), Func().Params(List(Id("i"), Id("j").Id("int"))).Id("bool").Block(
					Return(
						Id("compare").Call(h.Field("values").Index(Id("i")), h.Field("values").Index(Id("j"))),
					),
				),
			),
		},
	}
}

func (g *Generator) Find(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Find"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "cond",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) bool", h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Id("cond").Call(Id("value"))).Block(
					Return(Id("value")),
				),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) Filter(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Filter"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "filter",
		Type: types.TypeDeclareWithName(fmt.Sprintf("func(*%s) bool", h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			Id("filteredValues").Op(":=").Index().Op("*").Id(h.ModelName()).Values(),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Id("filter").Call(Id("value"))).Block(
					Id("filteredValues").Op("=").Append(Id("filteredValues"), Id("value")),
				),
			),
			Return(h.Receiver().Dot(fmt.Sprintf("new%s", h.Class.Name.PluralCamelName())).Call(Id("filteredValues"))),
		},
	}
}

func (g *Generator) IsEmpty(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "IsEmpty"
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Bool),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(True())),
			If(Len(h.Field("values")).Op("==").Lit(0)).Block(Return(True())),
			Return(False()),
		},
	}
}

func (g *Generator) At(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "At"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "idx",
		Type: types.TypeDeclareWithType(types.Int),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			If(Id("idx").Op("<").Lit(0)).Block(Return(Nil())),
			If(Len(h.Field("values")).Op(">").Id("idx")).Block(
				Return(h.Field("values").Index(Id("idx"))),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) First(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "First"
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			If(Len(h.Field("values")).Op(">").Lit(0)).Block(
				Return(h.Field("values").Index(Lit(0))),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) Last(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Last"
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			If(Len(h.Field("values")).Op(">").Lit(0)).Block(
				Return(h.Field("values").Index(Len(h.Field("values")).Op("-").Lit(1))),
			),
			Return(Nil()),
		},
	}
}

func (g *Generator) Compact(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Compact"
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			Id("compactedValues").Op(":=").Index().Op("*").Id(h.ModelName()).Values(),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(Id("value").Op("==").Nil()).Block(Continue()),
				Id("compactedValues").Op("=").Append(Id("compactedValues"), Id("value")),
			),
			Return(h.Receiver().Dot(fmt.Sprintf("new%s", h.Class.Name.PluralCamelName())).Call(Id("compactedValues"))),
		},
	}
}

func (g *Generator) Add(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Add"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "args",
		Type: types.TypeDeclareWithName(fmt.Sprintf("...*%s", h.ModelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			For(List(Id("_"), Id("value")).Op(":=").Range().Id("args")).Block(
				h.Field("values").Op("=").Append(h.Field("values"), Id("value")),
			),
			Return(h.Receiver()),
		},
	}
}

func (g *Generator) Merge(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Merge"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "args",
		Type: types.TypeDeclareWithName(fmt.Sprintf("...*%s", h.ModelCollectionName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			For(List(Id("_"), Id("arg")).Op(":=").Range().Id("args")).Block(
				For(List(Id("_"), Id("value")).Op(":=").Range().Id("arg").Dot("values")).Block(
					h.Field("values").Op("=").Append(h.Field("values"), Id("value")),
				),
			),
			Return(h.Receiver()),
		},
	}
}

func (g *Generator) Len(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = "Len"
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Int),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Lit(0))),
			Return(Len(h.Field("values"))),
		},
	}
}

func (g *Generator) MergeCollection(h *types.ModelMethodHelper) *types.Method {
	decl := h.CreateMultipleCollectionMethodDeclare()
	decl.MethodName = "Merge"
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Id("m").Op("==").Nil()).Block(Return(Nil())),
			If(Len(Op("*").Add(h.Receiver())).Op("==").Lit(0)).Block(Return(Nil())),
			If(Len(Op("*").Add(h.Receiver())).Op("==").Lit(1)).Block(Return(Parens(Op("*").Add(h.Receiver())).Index(Lit(0)))),
			Id("values").Op(":=").Index().Op("*").Id(h.ModelName()).Values(),
			For(List(Id("_"), Id("collection")).Op(":=").Range().Add(Op("*").Add(h.Receiver()))).Block(
				For(List(Id("_"), Id("value")).Op(":=").Range().Id("collection").Dot("values")).Block(
					Id("values").Op("=").Append(Id("values"), Id("value")),
				),
			),
			Return(Parens(Op("*").Add(h.Receiver())).Index(Lit(0)).Dot(fmt.Sprintf("new%s", h.Class.Name.PluralCamelName())).Call(Id("values"))),
		},
	}
}

func (g *Generator) Unique(h *types.ModelMethodHelper, member *types.Member) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = fmt.Sprintf("Unique%s", member.Name.CamelName())
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			Id("filterMap").Op(":=").Map(member.Type.Code(h.ImportList)).Struct().Values(),
			Return(
				h.Receiver().Dot("Filter").Call(
					Func().Params(Id("value").Op("*").Id(h.ModelName())).Bool().Block(
						If(
							List(Id("_"), Id("exists")).Op(":=").Id("filterMap").Index(Id("value").Dot(member.Name.CamelName())),
							Id("exists"),
						).Block(Return(False())),
						Id("filterMap").Index(Id("value").Dot(member.Name.CamelName())).Op("=").Struct().Values(),
						Return(True()),
					),
				),
			),
		},
	}
}

func (g *Generator) GroupBy(h *types.ModelMethodHelper, member *types.Member) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = fmt.Sprintf("GroupBy%s", member.Name.CamelName())
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithName(fmt.Sprintf("map[%#v]*%s", member.Type.Code(h.ImportList), h.ModelCollectionName())),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			Id("values").Op(":=").Map(member.Type.Code(h.ImportList)).Op("*").Id(h.ModelCollectionName()).Values(),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				If(
					List(Id("_"), Id("exists")).Op(":=").Id("values").Index(Id("value").Dot(member.Name.CamelName())),
					Op("!").Id("exists"),
				).Block(
					Id("values").Index(Id("value").Dot(member.Name.CamelName())).Op("=").Op("&").Id(h.ModelCollectionName()).Values(),
				),
				Id("values").Index(Id("value").Dot(member.Name.CamelName())).Dot("Add").Call(Id("value")),
			),
			Return(Id("values")),
		},
	}
}

func (g *Generator) FirstBy(h *types.ModelMethodHelper, members types.Members) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	names := []string{}
	for _, name := range members.Names() {
		names = append(names, name.CamelName())
	}
	decl.MethodName = fmt.Sprintf("FirstBy%s", strings.Join(names, "And"))
	blocks := []Code{}
	for idx, member := range members {
		decl.Args = append(decl.Args, &types.ValueDeclare{
			Name: fmt.Sprintf("a%d", idx),
			Type: member.Type,
		})
		blocks = append(blocks, If(Id("value").Dot(member.Name.CamelName()).Op("!=").Id(fmt.Sprintf("a%d", idx))).Block(Continue()))
	}
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelType(),
	})
	blocks = append(blocks, Return(Id("value")))
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(blocks...),
			Return(Nil()),
		},
	}
}

func (g *Generator) FilterBy(h *types.ModelMethodHelper, members types.Members) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	names := []string{}
	for _, name := range members.Names() {
		names = append(names, name.CamelName())
	}
	decl.MethodName = fmt.Sprintf("FilterBy%s", strings.Join(names, "And"))
	blocks := []Code{}
	for idx, member := range members {
		decl.Args = append(decl.Args, &types.ValueDeclare{
			Name: fmt.Sprintf("a%d", idx),
			Type: member.Type,
		})
		blocks = append(blocks, If(Id("value").Dot(member.Name.CamelName()).Op("!=").Id(fmt.Sprintf("a%d", idx))).Block(Continue()))
	}
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionType(),
	})
	blocks = append(blocks, Id("values").Op("=").Append(Id("values"), Id("value")))
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			Id("values").Op(":=").Index().Op("*").Id(h.ModelName()).Values(),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(blocks...),
			Return(h.Receiver().Dot(fmt.Sprintf("new%s", h.Class.Name.PluralCamelName())).Call(Id("values"))),
		},
	}
}

func (g *Generator) collectionBySchemaType(h *types.ModelMethodHelper, member *types.Member) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	var methodName string
	if member.IsCollectionType() {
		methodName = fmt.Sprintf("%sCollection", member.Name.CamelName())
	} else {
		methodName = member.CollectionName().PluralCamelName()
	}
	decl.MethodName = methodName
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	if member.Relation.Custom {
		if member.IsCollectionType() {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Type: types.TypeDeclareWithName(
					fmt.Sprintf("%sCollection", types.Name(member.Type.Name()).PluralCamelName()),
				),
			})
		} else {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Type: types.TypeDeclareWithName(
					fmt.Sprintf("*%s", types.Name(member.Type.Name()).PluralCamelName()),
				),
			})
		}
	} else {
		decl.Return = append(decl.Return, &types.ValueDeclare{
			Type: types.TypeDeclareWithName(member.ModelCollectionTypeName(h.ImportList)),
		})
	}
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	var (
		valuesType Code
		appendCode Code
	)
	if member.Relation.Custom {
		if member.IsCollectionType() {
			valuesType = Id(fmt.Sprintf("%sCollection", types.Name(member.Type.Name()).PluralCamelName()))
			appendCode = Id("values").Op("=").Append(Id("values"), Id(member.Name.CamelLowerName()))
		} else {
			valuesType = Op("&").Id(types.Name(member.Type.Name()).PluralCamelName())
			appendCode = Id("values").Dot("Add").Call(Id(member.Name.CamelLowerName()))
		}
	} else if member.IsCollectionType() {
		valuesType = Id(fmt.Sprintf("%sCollection", member.Type.CollectionName(h.ImportList)))
		appendCode = Id("values").Op("=").Append(Id("values"), Id(member.Name.CamelLowerName()))
	} else {
		valuesType = Op("&").Id(member.Type.CollectionName(h.ImportList))
		appendCode = Id("values").Dot("Add").Call(Id(member.Name.CamelLowerName()))
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil(), Nil())),
			Id("values").Op(":=").Add(valuesType).Values(),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				List(Id(member.Name.CamelLowerName()), Err()).Op(":=").Id("value").Dot(member.Name.CamelName()).Call(Id("ctx")),
				If(Err().Op("!=").Nil()).Block(Return(Nil(), Qual(h.Package("xerrors"), "Errorf").Call(
					Lit(fmt.Sprintf("failed to get %s: %%w", member.Name.CamelName())), Err(),
				))),
				If(Id(member.Name.CamelLowerName()).Op("==").Nil()).Block(Continue()),
				appendCode,
			),
			Return(Id("values"), Nil()),
		},
	}
}

func (g *Generator) collectionByPrimitiveType(h *types.ModelMethodHelper, member *types.Member) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = member.CollectionName().PluralCamelName()
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithName(member.ModelCollectionTypeName(h.ImportList)),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Op("==").Nil()).Block(Return(Nil())),
			Id("values").Op(":=").Id(member.ModelCollectionTypeName(h.ImportList)).Values(),
			For(List(Id("_"), Id("value")).Op(":=").Range().Add(h.Field("values"))).Block(
				Id("values").Op("=").Append(Id("values"), Id("value").Dot(member.Name.CamelName())),
			),
			Return(Id("values")),
		},
	}
}

func (g *Generator) Collection(h *types.ModelMethodHelper, member *types.Member) *types.Method {
	if member.Type.IsSchemaClass() || (member.Relation != nil && member.Relation.Custom) {
		return g.collectionBySchemaType(h, member)
	}
	return g.collectionByPrimitiveType(h, member)
}

func (g *Generator) findBy(h *types.ModelMethodHelper, member *types.Member) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = fmt.Sprintf("Find%s", member.Name.CamelName())
	internalMember := h.Class.MemberByName(member.Relation.Internal.SnakeName())
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: internalMember.Name.CamelLowerName(),
		Type: internalMember.Type,
	})
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "finder",
		Type: types.TypeDeclareWithName(fmt.Sprintf("%sFinder", member.Type.Class().Name.CamelName())),
	})
	var (
		typeName                      string
		filterOrFirstByMethodTemplate string
		fieldName                     string
	)
	if member.IsCollectionType() {
		typeName = member.Type.Class().Name.PluralCamelName()
		filterOrFirstByMethodTemplate = "FilterBy%s"
		fieldName = member.Name.CamelLowerName()
	} else {
		typeName = member.Type.Class().Name.CamelName()
		filterOrFirstByMethodTemplate = "FirstBy%s"
		fieldName = member.Name.PluralCamelLowerName()
	}
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: &types.TypeDeclare{
			Type: &types.Type{
				Name: typeName,
			},
			IsPointer: true,
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	externalName := member.Relation.External.CamelName()
	externalPluralName := member.Relation.External.PluralCamelName()
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Dot(fieldName).Op("!=").Nil()).Block(
				Return(h.Receiver().Dot(fieldName).
					Dot(fmt.Sprintf(filterOrFirstByMethodTemplate, externalName)).Call(Id(internalMember.Name.CamelLowerName())), Nil()),
			),
			List(Id(fieldName), Err()).Op(":=").Id("finder").
				Dot(fmt.Sprintf("FindBy%s", externalPluralName)).Call(Id("ctx"), h.Receiver().Dot(internalMember.Name.PluralCamelLowerName())),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), Qual(h.Package("xerrors"), "Errorf").Call(Lit(fmt.Sprintf("failed to FindBy%s: %%w", externalPluralName)), Err()))),
			If(Id(fieldName).Op("==").Nil()).Block(Return(Nil(), Qual(h.Package("xerrors"), "New").Call(Lit("cannot find record")))),
			h.Receiver().Dot(fieldName).Op("=").Id(fieldName),
			Return(h.Receiver().Dot(fieldName).
				Dot(fmt.Sprintf(filterOrFirstByMethodTemplate, externalName)).Call(Id(internalMember.Name.CamelLowerName())), Nil()),
		},
	}
}

func (g *Generator) findAll(h *types.ModelMethodHelper, member *types.Member) *types.Method {
	decl := h.CreateCollectionMethodDeclare()
	decl.MethodName = fmt.Sprintf("Find%s", member.Name.CamelName())
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "ctx",
		Type: &types.TypeDeclare{
			Type: &types.Type{
				PackageName: "context",
				Name:        "Context",
			},
		},
	})
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "finder",
		Type: types.TypeDeclareWithName(fmt.Sprintf("%sFinder", member.Type.Class().Name.CamelName())),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: &types.TypeDeclare{
			Type: &types.Type{
				Name: member.Type.Class().Name.PluralCamelName(),
			},
			IsPointer: true,
		},
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: types.TypeDeclareWithType(types.Error),
	})
	fieldName := member.Name.PluralCamelLowerName()
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(h.Receiver().Dot(fieldName).Op("!=").Nil()).Block(
				Return(h.Receiver().Dot(fieldName), Nil()),
			),
			List(Id(fieldName), Err()).Op(":=").Id("finder").Dot("FindAll").Call(Id("ctx")),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to FindAll: %%w"), Err()))),
			h.Receiver().Dot(fieldName).Op("=").Id(fieldName),
			Return(Id(fieldName), Nil()),
		},
	}
}

func (g *Generator) Methods() []func(*types.ModelMethodHelper) *types.Method {
	return []func(*types.ModelMethodHelper) *types.Method{
		func(h *types.ModelMethodHelper) *types.Method {
			return g.NewCollection(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Each(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.EachIndex(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.EachWithError(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.EachIndexWithError(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Map(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Any(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Some(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.IsIncluded(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.All(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Sort(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.SortStable(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Find(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Filter(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.IsEmpty(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.At(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.First(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Last(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Compact(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Add(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Merge(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.Len(h)
		},
		func(h *types.ModelMethodHelper) *types.Method {
			return g.MergeCollection(h)
		},
	}
}
