package repository

import (
	"fmt"
	"strings"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

func (r *Generator) Constructor(h *types.RepositoryMethodHelper, constructor *types.ConstructorDeclare, subClassConstructorMap map[*types.Class]*types.ConstructorDeclare) (*types.ConstructorDeclare, Code) {
	funcName := fmt.Sprintf("New%s", h.Class.Name.CamelName())
	allArgs := types.ValueDeclares{}
	args := []Code{}
	for _, arg := range constructor.Args {
		args = append(args, Id(arg.Name))
		allArgs = append(allArgs, arg)
	}
	properties := Dict{
		Id(h.DAOName()): Id("dao").Dot(funcName).Call(args[0:]...),
	}
	for subClass, constructor := range subClassConstructorMap {
		args := []Code{}
		for _, arg := range constructor.Args {
			args = append(args, Id(arg.Name))
			allArgs = append(allArgs, arg)
		}
		properties[Id(subClass.Name.CamelLowerName())] = Id(constructor.MethodName).Call(args...)
	}
	decl := &types.ConstructorDeclare{
		Class:      h.Class,
		MethodName: funcName,
		Args:       types.ValueDeclares{},
		Return:     types.ValueDeclares{},
	}
	valueNameMap := map[string]struct{}{}
	methodArgs := []Code{}
	for _, arg := range allArgs {
		if _, exists := valueNameMap[arg.Name]; exists {
			continue
		}
		decl.Args = append(decl.Args, arg)
		methodArgs = append(methodArgs, arg.Code(h.ImportList))
		valueNameMap[arg.Name] = struct{}{}
	}
	return decl, Func().Id(funcName).Params(methodArgs...).Op("*").Id(h.ReceiverClassName()).Block(
		Return(Op("&").Add(Id(h.ReceiverClassName())).Values(properties)),
	)
}

func (r *Generator) ConstructorMock(h *types.RepositoryMethodHelper) (*types.ConstructorDeclare, Code) {
	mockName := fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	funcName := fmt.Sprintf("New%s", mockName)
	decl := &types.ConstructorDeclare{
		Class:      h.Class,
		MethodName: funcName,
		Args:       types.ValueDeclares{},
		Return:     types.ValueDeclares{},
	}
	return decl, Func().Id(funcName).Params().Op("*").Id(mockName).Block(
		Return(Op("&").Add(Id(mockName)).Values(Dict{
			Id("expect"): Id(fmt.Sprintf("New%sExpect", h.Class.Name.CamelName())).Call(),
		})),
	)
}

func (r *Generator) EXPECT(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "EXPECT"
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName(fmt.Sprintf("*%sExpect", h.Class.Name.CamelName())),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Return(h.Receiver().Dot("expect")),
		},
	}
}

func (r *Generator) expectMethod(h *types.RepositoryMethodHelper, mtd *types.Method) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = mtd.Decl.MethodName
	decl.ReceiverClassName = fmt.Sprintf("%sExpect", h.Class.Name.CamelName())
	decl.Args = mtd.Decl.Args
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName(fmt.Sprintf("*%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)),
		},
	}
	values := Dict{
		Id("expect"):  h.Receiver(),
		Id("actions"): Index().Func().Params(mtd.Decl.Args.Code(h.ImportList)...).Values(),
	}
	for _, arg := range mtd.Decl.Args {
		values[Id(arg.Name)] = Id(arg.Name)
	}
	lowerName := types.Name(mtd.Decl.MethodName).CamelLowerName()
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Id("exp").Op(":=").Op("&").Id(fmt.Sprintf("%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)).Values(values),
			h.Receiver().Dot(lowerName).Op("=").Append(h.Receiver().Dot(lowerName), Id("exp")),
			Return(Id("exp")),
		},
	}
}

func (r *Generator) expectReturn(h *types.RepositoryMethodHelper, mtd *types.Method) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Return"
	decl.ReceiverClassName = fmt.Sprintf("%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)
	decl.Args = mtd.Decl.Return
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName(fmt.Sprintf("*%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)),
		},
	}
	blocks := []Code{}
	for _, arg := range decl.Args {
		blocks = append(blocks, h.Receiver().Dot(arg.Name).Op("=").Id(arg.Name))
	}
	blocks = append(blocks, Return(h.Receiver()))
	return &types.Method{
		Decl: decl,
		Body: blocks,
	}
}

func (r *Generator) expectDo(h *types.RepositoryMethodHelper, mtd *types.Method) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Do"
	if len(mtd.Decl.Args) > 0 {
		snippet := fmt.Sprintf("%#v", Func().Id("a").Params(mtd.Decl.Args.Code(h.ImportList)...).Block())
		newLineCount := strings.Count(snippet, "\n")
		args := snippet[len("func a") : len(snippet)-(len("{}")+newLineCount)]
		decl.Args = types.ValueDeclares{
			{
				Name: "action",
				Type: types.TypeDeclareWithName(fmt.Sprintf("func%s", args)),
			},
		}
	} else {
		decl.Args = types.ValueDeclares{
			{
				Name: "action",
				Type: types.TypeDeclareWithName("func()"),
			},
		}
	}
	decl.ReceiverClassName = fmt.Sprintf("%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName(fmt.Sprintf("*%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			h.Receiver().Dot("actions").Op("=").Append(h.Receiver().Dot("actions"), Id("action")),
			Return(h.Receiver()),
		},
	}
}

func (r *Generator) expectOutOfOrder(h *types.RepositoryMethodHelper, mtd *types.Method) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "OutOfOrder"
	decl.ReceiverClassName = fmt.Sprintf("%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName(fmt.Sprintf("*%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			h.Receiver().Dot("isOutOfOrder").Op("=").True(),
			Return(h.Receiver()),
		},
	}
}

func (r *Generator) expectAnyTimes(h *types.RepositoryMethodHelper, mtd *types.Method) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "AnyTimes"
	decl.ReceiverClassName = fmt.Sprintf("%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName(fmt.Sprintf("*%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			h.Receiver().Dot("isAnyTimes").Op("=").True(),
			Return(h.Receiver()),
		},
	}
}

func (r *Generator) expectTimes(h *types.RepositoryMethodHelper, mtd *types.Method) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Times"
	decl.ReceiverClassName = fmt.Sprintf("%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)
	decl.Args = types.ValueDeclares{
		{
			Name: "n",
			Type: types.TypeDeclareWithType(types.Int),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: types.TypeDeclareWithName(fmt.Sprintf("*%s%sExpect", h.Class.Name.CamelName(), mtd.Decl.MethodName)),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			h.Receiver().Dot("requiredTimes").Op("=").Id("n"),
			Return(h.Receiver()),
		},
	}
}

func (r *Generator) Create(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Create"
	decl.Args = types.ValueDeclares{
		{
			Name: "ctx",
			Type: h.ContextType(),
		},
		{
			Name: "value",
			Type: h.EntityClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: h.ModelClassType(),
		},
		{
			Type: types.TypeDeclareWithType(types.Error),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(Err().Op(":=").Add(h.DAO().Dot("Create").Call(Id("ctx"), Id("value"))), Err().Op("!=").Nil()).Block(
				Return(Nil(), Qual(h.Package("xerrors"), "Errorf").Call(Lit("cannot Create: %w"), Err())),
			),
			Id("v").Op(":=").Add(h.Receiver().Dot("ToModel").Call(Id("value"))),
			Id("v").Dot("SetSavedValue").Call(Id("value")),
			Id("v").Dot("SetAlreadyCreated").Call(True()),
			Return(Id("v"), Nil()),
		},
	}
}

func (r *Generator) mockCode(
	h *types.RepositoryMethodHelper,
	methodName string,
	args types.ValueDeclares,
	returns types.ValueDeclares) []Code {
	errIndex := len(returns) - 1
	blocks := []Code{
		If(Len(h.Receiver().Dot("expect").Dot(methodName)).Op("==").Lit(0)).Block(
			Id(fmt.Sprintf("r%d", errIndex)).Op("=").
				Qual(h.Package("xerrors"), "New").Call(
				Lit(fmt.Sprintf("cannot find mock method for %s.%s", h.Class.Name.CamelName(), types.Name(methodName).CamelName())),
			),
			Return(),
		),
	}
	actionArgs := []Code{}
	forBlocks := []Code{}
	for _, arg := range args {
		isSlice := strings.HasPrefix(arg.Type.Name(), "[]")
		if isSlice {
			forBlocks = append(forBlocks, []Code{
				If(Len(Id("exp").Dot(arg.Name)).Op("!=").Len(Id(arg.Name))).Block(
					Continue(),
				),
				If(Id("exp").Dot("isOutOfOrder")).Block(
					Id("isMatched").Op(":=").Func().Params().Bool().Block(
						For(
							List(Id("_"), Id("exp")).Op(":=").Range().Id("exp").Dot(arg.Name),
						).Block(
							Id("found").Op(":=").False(),
							For(
								List(Id("idx"), Id("act")).Op(":=").Range().Id(arg.Name),
							).Block(
								If(Id("exp").Op("!=").Id("act")).Block(
									Continue(),
								),
								Id(arg.Name).Op("=").Append(
									Id(arg.Name).Index(Op(":").Id("idx")),
									Id(arg.Name).Index(Id("idx").Op("+").Lit(1).Op(":")).Op("..."),
								),
								Id("found").Op("=").True(),
								Break(),
							),
							If(Op("!").Id("found")).Block(
								Return(False()),
							),
						),
						Return(True()),
					).Call(),
					If(Op("!").Id("isMatched")).Block(Continue()),
				).Else().Block(
					If(Op("!").Qual("reflect", "DeepEqual").Call(Id("exp").Dot(arg.Name), Id(arg.Name))).Block(
						Continue(),
					),
				),
			}...)
		} else {
			forBlocks = append(forBlocks,
				If(Op("!").Qual("reflect", "DeepEqual").Call(Id("exp").Dot(arg.Name), Id(arg.Name))).Block(
					Continue(),
				),
			)
		}
		actionArgs = append(actionArgs, Id(arg.Name))
	}
	returnBlocks := []Code{}
	for _, r := range returns {
		returnBlocks = append(returnBlocks, Id(r.Name).Op("=").Id("exp").Dot(r.Name))
	}
	returnBlocks = append(returnBlocks, Return())
	forBlocks = append(forBlocks, []Code{
		For(
			List(Id("_"), Id("action")).Op(":=").Range().Add(Id("exp").Dot("actions")),
		).Block(
			Id("action").Call(actionArgs...),
		),
		If(Id("exp").Dot("isAnyTimes")).Block(returnBlocks...),
		If(Id("exp").Dot("requiredTimes").Op(">").Lit(1).Op("&&").Id("exp").Dot("calledTimes").Op(">").Id("exp").Dot("requiredTimes")).Block(
			Id(fmt.Sprintf("r%d", errIndex)).Op("=").
				Qual(h.Package("xerrors"), "Errorf").Call(
				Lit("invalid call times. requiredTimes: [%d] calledTimes: [%d]"),
				Id("exp").Dot("requiredTimes"),
				Id("exp").Dot("calledTimes"),
			),
			Return(),
		),
		Id("exp").Dot("calledTimes").Op("++"),
	}...)
	forBlocks = append(forBlocks, returnBlocks...)
	blocks = append(blocks,
		For(
			List(Id("_"), Id("exp")).Op(":=").Range().Add(h.Receiver().Dot("expect").Dot(methodName)),
		).Block(forBlocks...),
	)
	if len(args) > 0 {
		errFormat := fmt.Sprintf("invalid argument %s", h.Class.Name.CamelName())
		for _, arg := range args {
			errFormat += fmt.Sprintf(" %s:[%%+v]", arg.Name)
		}
		errArgs := []Code{Lit(errFormat)}
		for _, arg := range args {
			errArgs = append(errArgs, Id(arg.Name))
		}
		blocks = append(blocks,
			Id(fmt.Sprintf("r%d", errIndex)).Op("=").
				Qual(h.Package("xerrors"), "Errorf").Call(errArgs...),
		)
	}
	blocks = append(blocks, Return())
	return blocks
}

func (r *Generator) toModelMockCode(
	h *types.RepositoryMethodHelper,
	methodName string,
	args types.ValueDeclares,
	returns types.ValueDeclares) []Code {
	blocks := []Code{
		If(Len(h.Receiver().Dot("expect").Dot(methodName)).Op("==").Lit(0)).Block(
			Qual("log", "Printf").Call(Lit(fmt.Sprintf("cannot find mock method for %s.%s", h.Class.Name.CamelName(), types.Name(methodName).CamelName()))),
			Return(),
		),
	}
	actionArgs := []Code{}
	forBlocks := []Code{}
	for _, arg := range args {
		forBlocks = append(forBlocks,
			If(Op("!").Qual("reflect", "DeepEqual").Call(Id("exp").Dot(arg.Name), Id(arg.Name))).Block(
				Continue(),
			),
		)
		actionArgs = append(actionArgs, Id(arg.Name))
	}
	returnBlocks := []Code{}
	for _, r := range returns {
		returnBlocks = append(returnBlocks, Id(r.Name).Op("=").Id("exp").Dot(r.Name))
	}
	returnBlocks = append(returnBlocks, Return())
	forBlocks = append(forBlocks, []Code{
		For(
			List(Id("_"), Id("action")).Op(":=").Range().Add(Id("exp").Dot("actions")),
		).Block(
			Id("action").Call(actionArgs...),
		),
		If(Id("exp").Dot("isAnyTimes")).Block(returnBlocks...),
		If(Id("exp").Dot("requiredTimes").Op(">").Lit(1).Op("&&").Id("exp").Dot("calledTimes").Op(">").Id("exp").Dot("requiredTimes")).Block(
			Qual("log", "Printf").Call(
				Lit("invalid call times. requiredTimes: [%d] calledTimes: [%d]"),
				Id("exp").Dot("requiredTimes"),
				Id("exp").Dot("calledTimes"),
			),
			Return(),
		),
		Id("exp").Dot("calledTimes").Op("++"),
	}...)
	forBlocks = append(forBlocks, returnBlocks...)
	blocks = append(blocks,
		For(
			List(Id("_"), Id("exp")).Op(":=").Range().Add(h.Receiver().Dot("expect").Dot(methodName)),
		).Block(forBlocks...),
	)
	if len(args) > 0 {
		errFormat := fmt.Sprintf("invalid argument %s", h.Class.Name.CamelName())
		for _, arg := range args {
			errFormat += fmt.Sprintf(" %s:[%%+v]", arg.Name)
		}
		errArgs := []Code{Lit(errFormat)}
		for _, arg := range args {
			errArgs = append(errArgs, Id(arg.Name))
		}
		blocks = append(blocks,
			Qual("log", "Printf").Call(errArgs...),
		)
	}
	blocks = append(blocks, Return())
	return blocks
}

func (r *Generator) CreateMock(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Create"
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.Args = types.ValueDeclares{
		{
			Name: "ctx",
			Type: h.ContextType(),
		},
		{
			Name: "value",
			Type: h.EntityClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Name: "r0",
			Type: h.ModelClassType(),
		},
		{
			Name: "r1",
			Type: types.TypeDeclareWithType(types.Error),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: r.mockCode(h, "create", decl.Args, decl.Return),
	}
}

func (r *Generator) Creates(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "Creates"
	decl.Args = types.ValueDeclares{
		{
			Name: "ctx",
			Type: h.ContextType(),
		},
		{
			Name: "entities",
			Type: h.EntityCollectionClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: h.ModelCollectionClassType(),
		},
		{
			Type: types.TypeDeclareWithType(types.Error),
		},
	}
	className := h.Class.Name.CamelName()
	return &types.Method{
		Decl: decl,
		Body: []Code{
			For(List(Id("_"), Id("v")).Op(":=").Range().Id("entities")).Block(
				If(List(Id("_"), Err()).Op(":=").Add(h.Receiver().Dot("Create").Call(Id("ctx"), Id("v"))), Err().Op("!=").Nil()).Block(
					Return(Nil(), Qual(h.Package("xerrors"), "Errorf").Call(Lit("cannot Create: %w"), Err())),
				),
			),
			Id("values").Op(":=").Add(h.Receiver().Dot("ToModels").Call(Id("entities"))),
			Id("values").Dot("Each").Call(
				Func().Params(Id("v").Op("*").Qual(h.Package("model"), className)).Block(
					Id("v").Dot("SetSavedValue").Call(Id("v").Dot(h.Class.Name.CamelName())),
					Id("v").Dot("SetAlreadyCreated").Call(True()),
				),
			),
			Return(Id("values"), Nil()),
		},
	}
}

func (r *Generator) CreatesMock(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.MethodName = "Creates"
	decl.Args = types.ValueDeclares{
		{
			Name: "ctx",
			Type: h.ContextType(),
		},
		{
			Name: "entities",
			Type: h.EntityCollectionClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Name: "r0",
			Type: h.ModelCollectionClassType(),
		},
		{
			Name: "r1",
			Type: types.TypeDeclareWithType(types.Error),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: r.mockCode(h, "creates", decl.Args, decl.Return),
	}
}

func (r *Generator) ToModel(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "ToModel"
	decl.Args = types.ValueDeclares{
		{
			Name: "value",
			Type: h.EntityClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: h.ModelClassType(),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Return(
				h.Receiver().Dot("createCollection").Call(
					Qual(h.Package("entity"), h.Class.Name.PluralCamelName()).Values(Id("value")),
				).Dot("First").Call(),
			),
		},
	}
}

func (r *Generator) ToModelMock(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.MethodName = "ToModel"
	decl.Args = types.ValueDeclares{
		{
			Name: "value",
			Type: h.EntityClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Name: "r0",
			Type: h.ModelClassType(),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: r.toModelMockCode(h, "toModel", decl.Args, decl.Return),
	}
}

func (r *Generator) ToModels(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "ToModels"
	decl.Args = types.ValueDeclares{
		{
			Name: "values",
			Type: h.EntityCollectionClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Type: h.ModelCollectionClassType(),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Return(h.Receiver().Dot("createCollection").Call(Id("values"))),
		},
	}
}

func (r *Generator) ToModelsMock(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.MethodName = "ToModels"
	decl.Args = types.ValueDeclares{
		{
			Name: "values",
			Type: h.EntityCollectionClassType(),
		},
	}
	decl.Return = types.ValueDeclares{
		{
			Name: "r0",
			Type: h.ModelCollectionClassType(),
		},
	}
	return &types.Method{
		Decl: decl,
		Body: r.toModelMockCode(h, "toModels", decl.Args, decl.Return),
	}
}

func (r *Generator) FindBy(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = method.MethodName
	decl.Args = method.Args
	collectionClassName := h.Class.Name.PluralCamelName()
	className := h.Class.Name.CamelName()
	for _, retDecl := range method.Return {
		name := retDecl.Type.Type.Name
		if name == className {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Type: h.ModelClassType(),
			})
		} else if name == collectionClassName {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Type: h.ModelCollectionClassType(),
			})
		} else {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Type: types.TypeDeclareWithType(&types.Type{
					Name: name,
				}),
			})
		}
	}
	args := []Code{}
	for _, arg := range method.Args {
		args = append(args, Id(arg.Name))
	}
	blocks := []Code{}
	call := h.DAO().Dot(method.MethodName).Call(args...)
	name := method.Return[0].Type.Type.Name
	if types.Name(name).PluralCamelName() == name {
		var extendCode Code
		if !h.Class.ReadOnly {
			className := h.Class.Name.CamelName()
			extendCode = Id("collection").Dot("Each").Call(
				Func().Params(Id("v").Op("*").Qual(h.Package("model"), className)).Block(
					Id("v").Dot("SetSavedValue").Call(Id("v").Dot(h.Class.Name.CamelName())),
					Id("v").Dot("SetAlreadyCreated").Call(True()),
				),
			)
		}
		blocks = append(blocks, []Code{
			List(Id("values"), Err()).Op(":=").Add(call),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), Qual(h.Package("xerrors"), "Errorf").Call(Lit(fmt.Sprintf("failed to %s: %%w", method.MethodName)), Err()))),
			Id("collection").Op(":=").Add(h.Receiver().Dot("createCollection").Call(Id("values"))),
			extendCode,
			Return(Id("collection"), Nil()),
		}...)
	} else {
		extendCodes := []Code{
			Return(Id("v"), Nil()),
		}
		if !h.Class.ReadOnly {
			extendCodes = []Code{
				Id("v").Dot("SetSavedValue").Call(Id("v").Dot(h.Class.Name.CamelName())),
				Id("v").Dot("SetAlreadyCreated").Call(True()),
				Return(Id("v"), Nil()),
			}
		}
		blocks = append(blocks, []Code{
			List(Id("value"), Err()).Op(":=").Add(call),
			If(Err().Op("!=").Nil()).Block(Return(Nil(), Qual(h.Package("xerrors"), "Errorf").Call(Lit(fmt.Sprintf("failed to %s: %%w", method.MethodName)), Err()))),
			If(Id("value").Op("==").Nil()).Block(Return(Nil(), Nil())),
			Id("v").Op(":=").Add(h.Receiver().Dot("createCollection").Call(
				Id("entity").Dot(h.Class.Name.PluralCamelName()).Values(Id("value")),
			).Dot("First").Call()),
		}...)
		blocks = append(blocks, extendCodes...)
	}
	return &types.Method{
		Decl: decl,
		Body: blocks,
	}
}

func (r *Generator) FindByMock(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	methodName := method.MethodName
	decl := h.CreateMethodDeclare()
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.MethodName = methodName
	decl.Args = method.Args
	collectionClassName := h.Class.Name.PluralCamelName()
	className := h.Class.Name.CamelName()
	for idx, retDecl := range method.Return {
		name := retDecl.Type.Type.Name
		if name == className {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Name: fmt.Sprintf("r%d", idx),
				Type: h.ModelClassType(),
			})
		} else if name == collectionClassName {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Name: fmt.Sprintf("r%d", idx),
				Type: h.ModelCollectionClassType(),
			})
		} else {
			decl.Return = append(decl.Return, &types.ValueDeclare{
				Name: fmt.Sprintf("r%d", idx),
				Type: types.TypeDeclareWithType(&types.Type{
					Name: name,
				}),
			})
		}
	}
	return &types.Method{
		Decl: decl,
		Body: r.mockCode(h, types.Name(methodName).CamelLowerName(), decl.Args, decl.Return),
	}
}

func (r *Generator) UpdateBy(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = method.MethodName
	decl.Args = method.Args
	decl.Return = method.Return
	args := []Code{}
	for _, arg := range method.Args {
		args = append(args, Id(arg.Name))
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(
				Err().Op(":=").Add(h.DAO().Dot(method.MethodName).Call(args...)),
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to update: %w"), Err())),
			),
			Return(Nil()),
		},
	}
}

func (r *Generator) UpdateByMock(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	methodName := method.MethodName
	decl := h.CreateMethodDeclare()
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.MethodName = methodName
	decl.Args = method.Args
	decl.Return = method.Return
	for idx, r := range decl.Return {
		r.Name = fmt.Sprintf("r%d", idx)
	}
	return &types.Method{
		Decl: decl,
		Body: r.mockCode(h, types.Name(methodName).CamelLowerName(), decl.Args, decl.Return),
	}
}

func (r *Generator) DeleteBy(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = method.MethodName
	decl.Args = method.Args
	decl.Return = method.Return
	args := []Code{}
	for _, arg := range method.Args {
		args = append(args, Id(arg.Name))
	}
	return &types.Method{
		Decl: decl,
		Body: []Code{
			If(
				Err().Op(":=").Add(h.DAO().Dot(method.MethodName).Call(args...)),
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual(h.Package("xerrors"), "Errorf").Call(Lit("failed to delete: %w"), Err())),
			),
			Return(Nil()),
		},
	}
}

func (r *Generator) DeleteByMock(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	methodName := method.MethodName
	decl := h.CreateMethodDeclare()
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.MethodName = methodName
	decl.Args = method.Args
	decl.Return = method.Return
	for idx, r := range decl.Return {
		r.Name = fmt.Sprintf("r%d", idx)
	}
	return &types.Method{
		Decl: decl,
		Body: r.mockCode(h, types.Name(methodName).CamelLowerName(), decl.Args, decl.Return),
	}
}

func (r *Generator) Other(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = method.MethodName
	decl.Args = method.Args
	decl.Return = method.Return
	args := []Code{}
	for _, arg := range method.Args {
		args = append(args, Id(arg.Name))
	}
	returns := []Code{}
	for idx, r := range decl.Return {
		r.Name = fmt.Sprintf("r%d", idx)
		returns = append(returns, Id(fmt.Sprintf("r%d", idx)))
	}
	err := fmt.Sprintf("r%d", len(decl.Return)-1)
	return &types.Method{
		Decl: decl,
		Body: []Code{
			List(returns...).Op("=").Add(h.DAO().Dot(method.MethodName).Call(args...)),
			If(
				Id(err).Op("!=").Nil(),
			).Block(
				Id(err).Op("=").Qual(h.Package("xerrors"), "Errorf").Call(
					Lit(fmt.Sprintf("failed to %s: %%w", method.MethodName)), Id(err),
				),
			),
			Return(),
		},
	}
}

func (r *Generator) OtherMock(h *types.RepositoryMethodHelper, method *types.MethodDeclare) *types.Method {
	methodName := method.MethodName
	decl := h.CreateMethodDeclare()
	decl.ReceiverClassName = fmt.Sprintf("%sMock", h.Class.Name.CamelName())
	decl.MethodName = methodName
	decl.Args = method.Args
	decl.Return = method.Return
	for idx, r := range decl.Return {
		r.Name = fmt.Sprintf("r%d", idx)
	}
	return &types.Method{
		Decl: decl,
		Body: r.mockCode(h, types.Name(methodName).CamelLowerName(), decl.Args, decl.Return),
	}
}

func (g *Generator) createCollection(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "createCollection"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "entities",
		Type: h.EntityCollectionClassType(),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelCollectionClassType(),
	})
	return &types.Method{
		Decl: decl,
		Body: []Code{
			Id("values").Op(":=").Qual(h.Package("model"), fmt.Sprintf("New%s", h.Class.Name.PluralCamelName())).Call(Id("entities")),
			For(Id("i").Op(":=").Lit(0), Id("i").Op("<").Len(Id("entities")), Id("i").Op("+=").Lit(1)).Block(
				Id("values").Dot("Add").Call(h.Receiver().Dot("create").Call(
					Id("entities").Index(Id("i")),
					Id("values"),
				)),
			),
			Return(Id("values")),
		},
	}
}

func (g *Generator) create(h *types.RepositoryMethodHelper) *types.Method {
	decl := h.CreateMethodDeclare()
	decl.MethodName = "create"
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "entity",
		Type: h.EntityClassType(),
	})
	decl.Args = append(decl.Args, &types.ValueDeclare{
		Name: "values",
		Type: h.ModelCollectionClassType(),
	})
	decl.Return = append(decl.Return, &types.ValueDeclare{
		Type: h.ModelClassType(),
	})
	block := []Code{
		Id("value").Op(":=").Qual(h.Package("model"), fmt.Sprintf("New%s", h.Class.Name.CamelName())).Call(
			Id("entity"), h.Receiver().Dot(fmt.Sprintf("%sDAO", h.Class.Name.CamelLowerName())),
		),
	}
	for _, member := range h.Class.RelationMembers() {
		relation := member.Relation
		if relation.All {
			block = append(block, []Code{
				h.Receiver().Dot(member.Type.Class().Name.CamelLowerName()).Assert(
					Id(fmt.Sprintf("*%sImpl", member.Type.Class().Name.CamelName())),
				).Dot("repo").Op("=").Add(h.Receiver().Dot("repo")),
				Id("value").Dot(member.Name.CamelName()).Op("=").
					Func().Params(Id("ctx").Qual(h.Package("context"), "Context")).
					Params(Op("*").Qual(h.Package("model"), member.Type.Class().Name.PluralCamelName()), Id("error")).Block(
					Return(Id("values").Dot(fmt.Sprintf("Find%s", member.Name.CamelName())).Call(
						Id("ctx"),
						h.Receiver().Dot(member.Type.Class().Name.CamelLowerName()),
					)),
				),
			}...)
			continue
		}
		if relation.Custom {
			continue
		}
		internalMember := h.Class.MemberByName(member.Relation.Internal.SnakeName())
		if member.IsCollectionType() {
			block = append(block, []Code{
				h.Receiver().Dot(member.Type.Class().Name.CamelLowerName()).Assert(
					Id(fmt.Sprintf("*%sImpl", member.Type.Class().Name.CamelName())),
				).Dot("repo").Op("=").Add(h.Receiver().Dot("repo")),
				Id("value").Dot(member.Name.CamelName()).Op("=").
					Func().Params(Id("ctx").Qual(h.Package("context"), "Context")).
					Params(Op("*").Qual(h.Package("model"), member.Type.Class().Name.PluralCamelName()), Id("error")).Block(
					Return(Id("values").Dot(fmt.Sprintf("Find%s", member.Name.CamelName())).Call(
						Id("ctx"),
						Id("value").Dot(internalMember.Name.CamelName()),
						h.Receiver().Dot(member.Type.Class().Name.CamelLowerName()),
					)),
				),
			}...)
		} else {
			block = append(block, []Code{
				h.Receiver().Dot(member.Type.Class().Name.CamelLowerName()).Assert(
					Id(fmt.Sprintf("*%sImpl", member.Type.Class().Name.CamelName())),
				).Dot("repo").Op("=").Add(h.Receiver().Dot("repo")),
				Id("value").Dot(member.Name.CamelName()).Op("=").
					Func().Params(Id("ctx").Qual(h.Package("context"), "Context")).
					Params(Op("*").Qual(h.Package("model"), member.Type.Class().Name.CamelName()), Id("error")).Block(
					Return(Id("values").Dot(fmt.Sprintf("Find%s", member.Name.CamelName())).Call(
						Id("ctx"),
						Id("value").Dot(internalMember.Name.CamelName()),
						h.Receiver().Dot(member.Type.Class().Name.CamelLowerName()),
					)),
				),
			}...)
		}
	}
	block = append(block, []Code{
		Id("value").Dot("SetConverter").Call(h.Receiver().Dot("repo").Assert(Qual(h.Package("model"), "ModelConverter"))),
		Return(Id("value")),
	}...)
	return &types.Method{
		Decl: decl,
		Body: block,
	}
}
