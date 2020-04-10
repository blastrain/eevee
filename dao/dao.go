package dao

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"go.knocknote.io/eevee/code"
	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/config"
	. "go.knocknote.io/eevee/plugin/dao"
	"go.knocknote.io/eevee/types"
	"golang.org/x/tools/imports"
	"golang.org/x/xerrors"
)

type Declare struct {
	Imports     types.ImportList
	Constructor *types.ConstructorDeclare
	Methods     []*types.MethodDeclare
}

type Generator struct {
	appName      string
	packageName  string
	receiverName string
	importList   types.ImportList
	datastores   map[string]*DataStore
	cfg          *config.Config
}

type DataStore struct {
	name      string
	plugins   map[string]struct{}
	pluginMap map[string][]func(types.DAOContext) ([]Code, error)
}

func (s *DataStore) overrideHook(pluginName interface{}, hookName string) error {
	name, ok := pluginName.(string)
	if !ok {
		return xerrors.Errorf("'%s' hook is required string. but passed %v", hookName, pluginName)
	}
	plugin, exists := PluginMap(name)
	if !exists {
		return xerrors.Errorf("cannot find dao plugin by %s", name)
	}
	s.plugins[name] = struct{}{}
	s.pluginMap[hookName] = []func(types.DAOContext) ([]Code, error){plugin[hookName]}
	return nil
}

func (s *DataStore) addHook(pluginName interface{}, hookName string) error {
	switch name := pluginName.(type) {
	case string:
		plugin, exists := PluginMap(name)
		if !exists {
			return xerrors.Errorf("cannot find dao plugin %s", name)
		}
		s.plugins[name] = struct{}{}
		s.pluginMap[hookName] = append(s.pluginMap[hookName], plugin[hookName])
	case []interface{}:
		for _, n := range name {
			if _, ok := n.(string); !ok {
				return xerrors.Errorf("'%s' hook is required string or []string. but passed %v", hookName, name)
			}
			plugin, exists := PluginMap(n.(string))
			if !exists {
				return xerrors.Errorf("cannot find dao plugin %s", n)
			}
			s.plugins[n.(string)] = struct{}{}
			s.pluginMap[hookName] = append(s.pluginMap[hookName], plugin[hookName])
		}
	default:
		return xerrors.Errorf("'%s' hook is required string or []string. but passed %v", hookName, pluginName)
	}
	return nil
}

func (s *DataStore) hookMap(kind string, pluginName interface{}) map[string]func() error {
	return map[string]func() error{
		fmt.Sprintf("%s-declare", kind): func() error {
			return s.overrideHook(pluginName, fmt.Sprintf("%s-declare", kind))
		},
		kind: func() error {
			return s.overrideHook(pluginName, kind)
		},
		fmt.Sprintf("before-%s", kind): func() error {
			return s.addHook(pluginName, fmt.Sprintf("before-%s", kind))
		},
		fmt.Sprintf("after-%s", kind): func() error {
			return s.addHook(pluginName, fmt.Sprintf("after-%s", kind))
		},
	}
}

func (s *DataStore) hookByPoint(point string, pluginName interface{}) error {
	maps := []map[string]func() error{
		s.hookMap("constructor", pluginName),
		s.hookMap("create", pluginName),
		s.hookMap("update", pluginName),
		s.hookMap("delete", pluginName),
		s.hookMap("find-all", pluginName),
		s.hookMap("count", pluginName),
		s.hookMap("findby", pluginName),
		s.hookMap("findby-plural", pluginName),
		s.hookMap("updateby", pluginName),
		s.hookMap("updateby-plural", pluginName),
		s.hookMap("deleteby", pluginName),
		s.hookMap("deleteby-plural", pluginName),
	}
	mergedMap := map[string]func() error{}
	for _, m := range maps {
		for k, v := range m {
			mergedMap[k] = v
		}
	}
	fn := mergedMap[point]
	if fn == nil {
		return xerrors.Errorf("unknown hook name %s", point)
	}
	return fn()
}

func NewPluginMap() map[string][]func(types.DAOContext) ([]Code, error) {
	return map[string][]func(c types.DAOContext) ([]Code, error){
		ConstructorDeclarePlugin:    []func(types.DAOContext) ([]Code, error){},
		ConstructorPlugin:           []func(types.DAOContext) ([]Code, error){},
		CreateDeclarePlugin:         []func(types.DAOContext) ([]Code, error){},
		BeforeCreatePlugin:          []func(types.DAOContext) ([]Code, error){},
		CreatePlugin:                []func(types.DAOContext) ([]Code, error){},
		AfterCreatePlugin:           []func(types.DAOContext) ([]Code, error){},
		UpdateDeclarePlugin:         []func(types.DAOContext) ([]Code, error){},
		BeforeUpdatePlugin:          []func(types.DAOContext) ([]Code, error){},
		UpdatePlugin:                []func(types.DAOContext) ([]Code, error){},
		AfterUpdatePlugin:           []func(types.DAOContext) ([]Code, error){},
		DeleteDeclarePlugin:         []func(types.DAOContext) ([]Code, error){},
		BeforeDeletePlugin:          []func(types.DAOContext) ([]Code, error){},
		DeletePlugin:                []func(types.DAOContext) ([]Code, error){},
		AfterDeletePlugin:           []func(types.DAOContext) ([]Code, error){},
		FindAllDeclarePlugin:        []func(types.DAOContext) ([]Code, error){},
		BeforeFindAllPlugin:         []func(types.DAOContext) ([]Code, error){},
		FindAllPlugin:               []func(types.DAOContext) ([]Code, error){},
		AfterFindAllPlugin:          []func(types.DAOContext) ([]Code, error){},
		CountDeclarePlugin:          []func(types.DAOContext) ([]Code, error){},
		BeforeCountPlugin:           []func(types.DAOContext) ([]Code, error){},
		CountPlugin:                 []func(types.DAOContext) ([]Code, error){},
		AfterCountPlugin:            []func(types.DAOContext) ([]Code, error){},
		FindByDeclarePlugin:         []func(types.DAOContext) ([]Code, error){},
		BeforeFindByPlugin:          []func(types.DAOContext) ([]Code, error){},
		FindByPlugin:                []func(types.DAOContext) ([]Code, error){},
		AfterFindByPlugin:           []func(types.DAOContext) ([]Code, error){},
		FindByPluralDeclarePlugin:   []func(types.DAOContext) ([]Code, error){},
		BeforeFindByPluralPlugin:    []func(types.DAOContext) ([]Code, error){},
		FindByPluralPlugin:          []func(types.DAOContext) ([]Code, error){},
		AfterFindByPluralPlugin:     []func(types.DAOContext) ([]Code, error){},
		UpdateByDeclarePlugin:       []func(types.DAOContext) ([]Code, error){},
		BeforeUpdateByPlugin:        []func(types.DAOContext) ([]Code, error){},
		UpdateByPlugin:              []func(types.DAOContext) ([]Code, error){},
		AfterUpdateByPlugin:         []func(types.DAOContext) ([]Code, error){},
		UpdateByPluralDeclarePlugin: []func(types.DAOContext) ([]Code, error){},
		BeforeUpdateByPluralPlugin:  []func(types.DAOContext) ([]Code, error){},
		UpdateByPluralPlugin:        []func(types.DAOContext) ([]Code, error){},
		AfterUpdateByPluralPlugin:   []func(types.DAOContext) ([]Code, error){},
		DeleteByDeclarePlugin:       []func(types.DAOContext) ([]Code, error){},
		BeforeDeleteByPlugin:        []func(types.DAOContext) ([]Code, error){},
		DeleteByPlugin:              []func(types.DAOContext) ([]Code, error){},
		AfterDeleteByPlugin:         []func(types.DAOContext) ([]Code, error){},
		DeleteByPluralDeclarePlugin: []func(types.DAOContext) ([]Code, error){},
		BeforeDeleteByPluralPlugin:  []func(types.DAOContext) ([]Code, error){},
		DeleteByPluralPlugin:        []func(types.DAOContext) ([]Code, error){},
		AfterDeleteByPluralPlugin:   []func(types.DAOContext) ([]Code, error){},
	}
}

func NewGenerator(cfg *config.Config) *Generator {
	g := &Generator{
		appName:      cfg.ModulePath,
		packageName:  cfg.DAOPackageName(),
		receiverName: "d",
		datastores:   map[string]*DataStore{},
		importList:   types.ImportList{},
		cfg:          cfg,
	}
	g.buildDataStores()
	return g
}

func (g *Generator) pluginsByDataStoreName(name string) map[string]interface{} {
	if g.cfg.DAO == nil {
		return nil
	}
	if g.cfg.DAO.DataStore == nil {
		return nil
	}
	ds, exists := g.cfg.DAO.DataStore[name]
	if !exists {
		return nil
	}
	return ds.Hooks
}

func (g *Generator) buildDataStores() error {
	// get currently registered all datastores
	for _, name := range DataStoreNames() {
		daoPluginMap := NewPluginMap()
		pluginMap, _ := DataStoreMap(name)
		for k, v := range pluginMap {
			daoPluginMap[k] = append(daoPluginMap[k], v)
		}
		datastore := &DataStore{
			name:      name,
			plugins:   map[string]struct{}{},
			pluginMap: daoPluginMap,
		}
		for hookPoint, pluginName := range g.pluginsByDataStoreName(name) {
			if err := datastore.hookByPoint(hookPoint, pluginName); err != nil {
				return xerrors.Errorf("failed to hook: %w", err)
			}
		}
		g.datastores[name] = datastore
	}
	return nil
}

func (g *Generator) newDataAccessParam(class *types.Class) types.DataAccessParam {
	return types.DataAccessParam{
		Class:      class,
		ClassName:  func() *Statement { return Id(class.Name.CamelName()) },
		Receiver:   func() *Statement { return Id(g.receiverName) },
		ImportList: g.importList,
	}
}

func (g *Generator) newConstructorParam(class *types.Class) *types.ConstructorParam {
	return &types.ConstructorParam{
		DataAccessParam: g.newDataAccessParam(class),
		ImplName:        fmt.Sprintf("%sImpl", class.Name.CamelName()),
		Args: &types.ConstructorParamArgs{
			Context: func() *Statement { return Id("ctx") },
		},
	}
}

func (g *Generator) newCreateParam(class *types.Class) *types.CreateParam {
	return &types.CreateParam{
		DataAccessParam: g.newDataAccessParam(class),
		Args: &types.CreateParamArgs{
			Context: func() *Statement { return Id("ctx") },
			Value:   func() *Statement { return Id("value") },
		},
	}
}

func (g *Generator) newUpdateParam(class *types.Class) *types.UpdateParam {
	return &types.UpdateParam{
		DataAccessParam: g.newDataAccessParam(class),
		Args: &types.UpdateParamArgs{
			Context:   func() *Statement { return Id("ctx") },
			Value:     func() *Statement { return Id("value") },
			UpdateMap: func() *Statement { return Id("updateMap") },
		},
	}
}

func (g *Generator) newDeleteParam(class *types.Class) *types.DeleteParam {
	return &types.DeleteParam{
		DataAccessParam: g.newDataAccessParam(class),
		Args: &types.DeleteParamArgs{
			Context: func() *Statement { return Id("ctx") },
			Value:   func() *Statement { return Id("value") },
		},
	}
}

func (g *Generator) newFindParam(class *types.Class) *types.FindParam {
	return &types.FindParam{
		DataAccessParam: g.newDataAccessParam(class),
		Args: &types.FindParamArgs{
			Context: func() *Statement { return Id("ctx") },
			Members: []*types.Member{},
		},
	}
}

func (g *Generator) newCountParam(class *types.Class) *types.CountParam {
	return &types.CountParam{
		DataAccessParam: g.newDataAccessParam(class),
		Args: &types.CountParamArgs{
			Context: func() *Statement { return Id("ctx") },
			Members: []*types.Member{},
		},
	}
}

func (g *Generator) newConstructorDeclare(class *types.Class) (*types.ConstructorDeclare, error) {
	declare := &types.ConstructorDeclare{
		Class:      class,
		MethodName: fmt.Sprintf("New%s", class.Name.CamelName()),
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
		},
		Return: []*types.ValueDeclare{
			{
				Type: types.TypeDeclareWithName(class.Name.CamelName()),
			},
		},
		ImportList: g.importList,
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[ConstructorDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for constructor: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newCreateDeclare(class *types.Class) (*types.MethodDeclare, error) {
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        "Create",
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
			{
				Name: "value",
				Type: &types.TypeDeclare{
					Type: &types.Type{
						PackageName: g.importList.Package("entity"),
						Name:        class.Name.CamelName(),
					},
					IsPointer: true,
				},
			},
		},
		Return: []*types.ValueDeclare{
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[CreateDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for create: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newUpdateDeclare(class *types.Class) (*types.MethodDeclare, error) {
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        "Update",
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
			{
				Name: "value",
				Type: &types.TypeDeclare{
					Type: &types.Type{
						PackageName: g.importList.Package("entity"),
						Name:        class.Name.CamelName(),
					},
					IsPointer: true,
				},
			},
		},
		Return: []*types.ValueDeclare{
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[UpdateDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for update: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newDeleteDeclare(class *types.Class) (*types.MethodDeclare, error) {
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        "Delete",
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
			{
				Name: "value",
				Type: &types.TypeDeclare{
					Type: &types.Type{
						PackageName: g.importList.Package("entity"),
						Name:        class.Name.CamelName(),
					},
					IsPointer: true,
				},
			},
		},
		Return: []*types.ValueDeclare{
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[DeleteDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for delete: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newFindAllDeclare(class *types.Class) (*types.MethodDeclare, error) {
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        "FindAll",
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
		},
		Return: []*types.ValueDeclare{
			{
				Name: "r",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("entity"),
					Name:        class.Name.PluralCamelName(),
				}),
			},
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[FindAllDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for findAll: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newCountDeclare(class *types.Class) (*types.MethodDeclare, error) {
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        "Count",
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
		},
		Return: []*types.ValueDeclare{
			{
				Name: "r",
				Type: types.TypeDeclareWithType(types.Int64Type),
			},
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[CountDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for count: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newFindByDeclare(class *types.Class, p *types.FindParam) (*types.MethodDeclare, error) {
	args := types.ValueDeclares{
		{
			Name: "ctx",
			Type: types.TypeDeclareWithType(&types.Type{
				PackageName: g.importList.Package("context"),
				Name:        "Context",
			}),
		},
	}
	argsCamelNames := []string{}
	for idx, member := range p.Args.Members {
		argsCamelNames = append(argsCamelNames, member.Name.CamelName())
		args = append(args, &types.ValueDeclare{
			Name: fmt.Sprintf("a%d", idx),
			Type: member.Type,
		})
	}
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        fmt.Sprintf("FindBy%s", strings.Join(argsCamelNames, "And")),
		ArgMembers:        p.Args.Members,
		Args:              args,
		Return: []*types.ValueDeclare{
			{
				Name: "r",
				Type: &types.TypeDeclare{
					Type: &types.Type{
						PackageName: g.importList.Package("entity"),
						Name:        p.Class.Name.CamelName(),
					},
					IsPointer: true,
				},
			},
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[FindByDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for findBy: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newFindBySliceDeclare(class *types.Class, p *types.FindParam) (*types.MethodDeclare, error) {
	args := types.ValueDeclares{
		{
			Name: "ctx",
			Type: types.TypeDeclareWithType(&types.Type{
				PackageName: g.importList.Package("context"),
				Name:        "Context",
			}),
		},
	}
	argsCamelNames := []string{}
	for idx, member := range p.Args.Members {
		argsCamelNames = append(argsCamelNames, member.Name.CamelName())
		args = append(args, &types.ValueDeclare{
			Name: fmt.Sprintf("a%d", idx),
			Type: member.Type,
		})
	}
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        fmt.Sprintf("FindBy%s", strings.Join(argsCamelNames, "And")),
		Args:              args,
		ArgMembers:        p.Args.Members,
		Return: []*types.ValueDeclare{
			{
				Name: "r",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("entity"),
					Name:        p.Class.Name.PluralCamelName(),
				}),
			},
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[FindByDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for findBy: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newFindByPluralDeclare(class *types.Class, p *types.FindParam) (*types.MethodDeclare, error) {
	member := p.Args.Members[0]
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        fmt.Sprintf("FindBy%s", member.Name.PluralCamelName()),
		ArgMembers:        p.Args.Members,
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
			{
				Name: "a0",
				Type: &types.TypeDeclare{
					Type:    member.Type.Type,
					IsSlice: true,
				},
			},
		},
		Return: []*types.ValueDeclare{
			{
				Name: "r",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("entity"),
					Name:        p.Class.Name.PluralCamelName(),
				}),
			},
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[FindByPluralDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for findByPlural: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newUpdateByDeclare(class *types.Class, p *types.UpdateParam) (*types.MethodDeclare, error) {
	args := types.ValueDeclares{
		{
			Name: "ctx",
			Type: types.TypeDeclareWithType(&types.Type{
				PackageName: g.importList.Package("context"),
				Name:        "Context",
			}),
		},
	}
	argsCamelNames := []string{}
	for idx, member := range p.Args.Members {
		argsCamelNames = append(argsCamelNames, member.Name.CamelName())
		args = append(args, &types.ValueDeclare{
			Name: fmt.Sprintf("a%d", idx),
			Type: member.Type,
		})
	}
	args = append(args, &types.ValueDeclare{
		Name: "updateMap",
		Type: types.TypeDeclareWithName("map[string]interface{}"),
	})
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        fmt.Sprintf("UpdateBy%s", strings.Join(argsCamelNames, "And")),
		ArgMembers:        p.Args.Members,
		Args:              args,
		Return: types.ValueDeclares{
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[UpdateByDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for updateBy: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newUpdateByPluralDeclare(class *types.Class, p *types.UpdateParam) (*types.MethodDeclare, error) {
	member := p.Args.Members[0]
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        fmt.Sprintf("UpdateBy%s", member.Name.PluralCamelName()),
		ArgMembers:        p.Args.Members,
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
			{
				Name: "a0",
				Type: &types.TypeDeclare{
					Type:    member.Type.Type,
					IsSlice: true,
				},
			},
			{
				Name: "updateMap",
				Type: types.TypeDeclareWithName("map[string]interface{}"),
			},
		},
		Return: types.ValueDeclares{
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[UpdateByPluralDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for updateByPlural: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newDeleteByDeclare(class *types.Class, p *types.DeleteParam) (*types.MethodDeclare, error) {
	args := types.ValueDeclares{
		{
			Name: "ctx",
			Type: types.TypeDeclareWithType(&types.Type{
				PackageName: g.importList.Package("context"),
				Name:        "Context",
			}),
		},
	}
	argsCamelNames := []string{}
	for idx, member := range p.Args.Members {
		args = append(args, &types.ValueDeclare{
			Name: fmt.Sprintf("a%d", idx),
			Type: member.Type,
		})
		argsCamelNames = append(argsCamelNames, member.Name.CamelName())
	}
	methodName := fmt.Sprintf("DeleteBy%s", strings.Join(argsCamelNames, "And"))
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        methodName,
		ArgMembers:        p.Args.Members,
		Args:              args,
		Return: types.ValueDeclares{
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[DeleteByDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for deleteBy: %w", err)
		}
	}
	return declare, nil
}

func (g *Generator) newDeleteByPluralDeclare(class *types.Class, p *types.DeleteParam) (*types.MethodDeclare, error) {
	member := p.Args.Members[0]
	declare := &types.MethodDeclare{
		Class:             class,
		ReceiverName:      g.receiverName,
		ReceiverClassName: fmt.Sprintf("%sImpl", class.Name.CamelName()),
		ImportList:        g.importList,
		MethodName:        fmt.Sprintf("DeleteBy%s", member.Name.PluralCamelName()),
		ArgMembers:        p.Args.Members,
		Args: types.ValueDeclares{
			{
				Name: "ctx",
				Type: types.TypeDeclareWithType(&types.Type{
					PackageName: g.importList.Package("context"),
					Name:        "Context",
				}),
			},
			{
				Name: "a0",
				Type: &types.TypeDeclare{
					Type:    member.Type.Type,
					IsSlice: true,
				},
			},
		},
		Return: types.ValueDeclares{
			{
				Name: "e",
				Type: types.TypeDeclareWithType(types.ErrorType),
			},
		},
	}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[DeleteByPluralDeclarePlugin] {
		if _, err := fn(declare); err != nil {
			return nil, xerrors.Errorf("failed to declaration for deleteByPlural: %w", err)
		}
	}
	return declare, nil
}

type MethodGenerator struct {
	decl  *types.MethodDeclare
	hooks []Code
}

func (g *MethodGenerator) Generate(importList types.ImportList) *Statement {
	return g.decl.MethodInterface(importList).Block(g.hooks...)
}

type MethodGenerators []*MethodGenerator

func (g *Generator) getHookCodes(class *types.Class, kind string, param types.DAOContext) []Code {
	blocks := []Code{}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[fmt.Sprintf("before-%s", kind)] {
		codes, _ := fn(param)
		blocks = append(blocks, Block(codes...))
	}
	deferBlocks := []Code{}
	for _, fn := range datastore.pluginMap[fmt.Sprintf("after-%s", kind)] {
		codes, _ := fn(param)
		deferBlocks = append(deferBlocks, Block(codes...))
	}
	if len(deferBlocks) > 0 {
		blocks = append(blocks, Defer().Func().Params().Block(deferBlocks...).Call())
	}
	for _, fn := range datastore.pluginMap[kind] {
		codes, _ := fn(param)
		blocks = append(blocks, codes...)
	}
	return blocks
}

func (g *Generator) newCreateMethodGenerator(class *types.Class) (*MethodGenerator, error) {
	decl, err := g.newCreateDeclare(class)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for create: %w", err)
	}
	param := g.newCreateParam(class)
	placeholders := []string{}
	columns := []string{}
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	for _, member := range class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		placeholders = append(placeholders, "?")
		columns = append(columns, fmt.Sprintf("`%s`", member.Name))
	}
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
		escapedTableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	param.SQL = &types.SQL{Query: query}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "create", param),
	}, nil
}

func (g *Generator) newUpdateMethodGenerator(class *types.Class) (*MethodGenerator, error) {
	decl, err := g.newUpdateDeclare(class)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for update: %w", err)
	}
	param := g.newUpdateParam(class)
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	columns := []string{}
	for _, member := range class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		if member.Name.SnakeName() == "id" {
			continue
		}
		columns = append(columns, fmt.Sprintf("`%s` = ?", member.Name.SnakeName()))
	}
	param.SQL = &types.SQL{
		Query: fmt.Sprintf(`UPDATE %s SET %s WHERE id = ?`, escapedTableName, strings.Join(columns, ", ")),
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "update", param),
	}, nil
}

func (g *Generator) newDeleteMethodGenerator(class *types.Class) (*MethodGenerator, error) {
	decl, err := g.newDeleteDeclare(class)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for delete: %w", err)
	}
	param := g.newDeleteParam(class)
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	param.SQL = &types.SQL{
		Query: fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, escapedTableName),
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "delete", param),
	}, nil
}

func (g *Generator) newFindAllMethodGenerator(class *types.Class) (*MethodGenerator, error) {
	decl, err := g.newFindAllDeclare(class)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for findAll: %w", err)
	}
	param := g.newFindParam(class)
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	columns := []string{}
	scanValues := []Code{}
	for _, member := range class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		columns = append(columns, fmt.Sprintf("`%s`", member.Name))
		scanValues = append(scanValues, Op("&").Id("value").Dot(member.Name.CamelName()))
	}
	param.SQL = &types.SQL{
		Query:      fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(columns, ", "), escapedTableName),
		ScanValues: scanValues,
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "find-all", param),
	}, nil
}

func (g *Generator) newCountMethodGenerator(class *types.Class) (*MethodGenerator, error) {
	decl, err := g.newCountDeclare(class)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for count: %w", err)
	}
	param := g.newCountParam(class)
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	param.SQL = &types.SQL{
		Query: fmt.Sprintf(`COUNT(*) FROM %s`, escapedTableName),
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "count", param),
	}, nil
}

func (g *Generator) createSQLForFindBy(class *types.Class, param *types.FindParam) *types.SQL {
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	columns := []string{}
	scanValues := []Code{}
	for _, member := range class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		columns = append(columns, fmt.Sprintf("`%s`", member.Name))
		scanValues = append(scanValues, Line().Op("&").Id("value").Dot(member.Name.CamelName()))
	}
	scanValues = append(scanValues, Line())
	conditions := []string{}
	argNames := []Code{}
	for idx, member := range param.Args.Members {
		conditions = append(conditions, fmt.Sprintf("`%s` = ?", member.Name.SnakeName()))
		argNames = append(argNames, Id(fmt.Sprintf("a%d", idx)))
	}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`,
		strings.Join(columns, ", "),
		escapedTableName,
		strings.Join(conditions, " AND "),
	)
	return &types.SQL{
		Query:      query,
		Args:       argNames,
		ScanValues: scanValues,
	}
}

func (g *Generator) newFindByMethodGenerator(class *types.Class, param *types.FindParam) (*MethodGenerator, error) {
	decl, err := g.newFindByDeclare(class, param)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for findBy: %w", err)
	}
	param.SQL = g.createSQLForFindBy(class, param)
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "findby", param),
	}, nil
}

func (g *Generator) newFindBySliceMethodGenerator(class *types.Class, param *types.FindParam) (*MethodGenerator, error) {
	decl, err := g.newFindBySliceDeclare(class, param)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for findBySlice: %w", err)
	}
	param.SQL = g.createSQLForFindBy(class, param)
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "findby", param),
	}, nil
}

func (g *Generator) newFindByPluralMethodGenerator(class *types.Class, param *types.FindParam) (*MethodGenerator, error) {
	decl, err := g.newFindByPluralDeclare(class, param)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for findByPlural: %w", err)
	}
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	columns := []string{}
	scanValues := []Code{}
	for _, member := range class.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		columns = append(columns, fmt.Sprintf("`%s`", member.Name))
		scanValues = append(scanValues, Line().Op("&").Id("value").Dot(member.Name.CamelName()))
	}
	scanValues = append(scanValues, Line())
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s IN (%%s)`,
		strings.Join(columns, ", "),
		escapedTableName,
		fmt.Sprintf("`%s`", param.Args.Members[0].Name.SnakeName()),
	)
	param.SQL = &types.SQL{
		Query:      query,
		ScanValues: scanValues,
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "findby-plural", param),
	}, nil
}

func (g *Generator) newUpdateByMethodGenerator(class *types.Class, param *types.UpdateParam) (*MethodGenerator, error) {
	decl, err := g.newUpdateByDeclare(class, param)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for updateBy: %w", err)
	}
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	columns := []string{}
	for _, member := range param.Args.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			continue
		}
		columns = append(columns, fmt.Sprintf("`%s` = ?", member.Name.SnakeName()))
	}
	query := fmt.Sprintf(`UPDATE %s SET %%s WHERE %s`,
		escapedTableName,
		strings.Join(columns, " AND "),
	)
	param.SQL = &types.SQL{
		Query: query,
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "updateby", param),
	}, nil
}

func (g *Generator) newUpdateByPluralMethodGenerator(class *types.Class, param *types.UpdateParam) (*MethodGenerator, error) {
	decl, err := g.newUpdateByPluralDeclare(class, param)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for updateByPlural: %w", err)
	}
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	member := param.Args.Members[0]
	query := fmt.Sprintf(`UPDATE %s SET %%s WHERE %s`,
		escapedTableName,
		fmt.Sprintf("`%s` IN (%%s)", member.Name.SnakeName()),
	)
	param.SQL = &types.SQL{
		Query: query,
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "updateby-plural", param),
	}, nil
}

func (g *Generator) newDeleteByMethodGenerator(class *types.Class, param *types.DeleteParam) (*MethodGenerator, error) {
	decl, err := g.newDeleteByDeclare(class, param)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for deleteBy: %w", err)
	}
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	columns := []string{}
	args := []Code{}
	for idx, member := range param.Args.Members {
		columns = append(columns, fmt.Sprintf("`%s` = ?", member.Name.SnakeName()))
		args = append(args, Id(fmt.Sprintf("a%d", idx)))
	}
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s`,
		escapedTableName,
		strings.Join(columns, " AND "),
	)
	param.SQL = &types.SQL{
		Query: query,
		Args:  args,
	}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "deleteby", param),
	}, nil
}

func (g *Generator) newDeleteByPluralMethodGenerator(class *types.Class, param *types.DeleteParam) (*MethodGenerator, error) {
	decl, err := g.newDeleteByPluralDeclare(class, param)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for deleteByPlural: %w", err)
	}
	escapedTableName := fmt.Sprintf("`%s`", class.Name.PluralSnakeName())
	columns := []string{}
	member := param.Args.Members[0]
	columns = append(columns)
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s`,
		escapedTableName,
		fmt.Sprintf("`%s` IN (%%s)", member.Name.SnakeName()),
	)
	param.SQL = &types.SQL{Query: query}
	return &MethodGenerator{
		decl:  decl,
		hooks: g.getHookCodes(class, "deleteby-plural", param),
	}, nil
}

func (g *Generator) newFindByMethodGeneratorsFromPrimaryKey(class *types.Class, primaryKey *types.Member) ([]*MethodGenerator, error) {
	p := g.newFindParam(class)
	p.Args.Members = append(p.Args.Members, primaryKey)
	p.IsSingleReturnValue = true
	findByGen, err := g.newFindByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create FindByMethodGenerator: %w", err)
	}
	if len(p.Args.Members) != 1 {
		return []*MethodGenerator{findByGen}, nil
	}
	p = g.newFindParam(class)
	p.Args.Members = append(p.Args.Members, primaryKey)
	findByPluralGen, err := g.newFindByPluralMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create FindByPluralMethodGenerator: %w", err)
	}
	return []*MethodGenerator{findByGen, findByPluralGen}, nil
}

func (g *Generator) newFindByMethodGeneratorsFromUniqueKey(class *types.Class, uniqueKey types.Members) ([]*MethodGenerator, error) {
	p := g.newFindParam(class)
	p.Args.Members = append(p.Args.Members, uniqueKey...)
	p.IsSingleReturnValue = true
	findByGen, err := g.newFindByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create FindByMethodGenerator: %w", err)
	}
	generators := []*MethodGenerator{findByGen}
	if len(p.Args.Members) == 1 {
		p = g.newFindParam(class)
		p.Args.Members = append(p.Args.Members, uniqueKey...)
		findByPluralGen, err := g.newFindByPluralMethodGenerator(class, p)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindByPluralMethodGenerator: %w", err)
		}
		generators = append(generators, findByPluralGen)
	}
	if len(uniqueKey) < 2 {
		return generators, nil
	}
	uniqueKey = uniqueKey[:len(uniqueKey)-1]
	for i := len(uniqueKey); i > 0; i-- {
		p := g.newFindParam(class)
		p.Args.Members = append(p.Args.Members, uniqueKey...)
		generator, err := g.newFindBySliceMethodGenerator(class, p)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindBySliceMethodGenerator: %w", err)
		}
		generators = append(generators, generator)
		if len(uniqueKey) == 1 {
			p = g.newFindParam(class)
			p.Args.Members = append(p.Args.Members, uniqueKey...)
			findByPluralGen, err := g.newFindByPluralMethodGenerator(class, p)
			if err != nil {
				return nil, xerrors.Errorf("cannot create FindByPluralMethodGenerator: %w", err)
			}
			generators = append(generators, findByPluralGen)
		}
		uniqueKey = uniqueKey[:len(uniqueKey)-1]
	}
	return generators, nil
}

func (g *Generator) newFindByMethodGeneratorsFromKey(class *types.Class, key types.Members) ([]*MethodGenerator, error) {
	p := g.newFindParam(class)
	p.Args.Members = append(p.Args.Members, key...)
	findByGen, err := g.newFindBySliceMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create FindBySliceMethodGenerator: %w", err)
	}
	generators := []*MethodGenerator{findByGen}
	if len(p.Args.Members) == 1 {
		p = g.newFindParam(class)
		p.Args.Members = append(p.Args.Members, key...)
		findByPluralGen, err := g.newFindByPluralMethodGenerator(class, p)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindByPluralMethodGenerator: %w", err)
		}
		generators = append(generators, findByPluralGen)
	}
	if len(key) < 2 {
		return generators, nil
	}
	key = key[:len(key)-1]
	for i := len(key); i > 0; i-- {
		p := g.newFindParam(class)
		p.Args.Members = append(p.Args.Members, key...)
		generator, err := g.newFindBySliceMethodGenerator(class, p)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindBySliceMethodGenerator: %w", err)
		}
		generators = append(generators, generator)
		if len(key) == 1 {
			p = g.newFindParam(class)
			p.Args.Members = append(p.Args.Members, key...)
			findByPluralGen, err := g.newFindByPluralMethodGenerator(class, p)
			if err != nil {
				return nil, xerrors.Errorf("cannot create FindByPluralMethodGenerator: %w", err)
			}
			generators = append(generators, findByPluralGen)
		}
		key = key[:len(key)-1]
	}
	return generators, nil
}

func (g *Generator) newUpdateByMethodGeneratorsFromPrimaryKey(class *types.Class, primaryKey *types.Member) ([]*MethodGenerator, error) {
	p := g.newUpdateParam(class)
	p.Args.Members = append(p.Args.Members, primaryKey)
	updateByGen, err := g.newUpdateByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create UpdateByMethodGenerator: %w", err)
	}
	if len(p.Args.Members) != 1 {
		return []*MethodGenerator{updateByGen}, nil
	}
	p = g.newUpdateParam(class)
	p.Args.Members = append(p.Args.Members, primaryKey)
	updateByPluralGen, err := g.newUpdateByPluralMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create UpdateByPluralMethodGenerator: %w", err)
	}
	return []*MethodGenerator{updateByGen, updateByPluralGen}, nil
}

func (g *Generator) newUpdateByMethodGeneratorsFromUniqueKey(class *types.Class, uniqueKey types.Members) ([]*MethodGenerator, error) {
	p := g.newUpdateParam(class)
	p.Args.Members = append(p.Args.Members, uniqueKey...)
	updateByGen, err := g.newUpdateByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create UpdateByMethodGenerator: %w", err)
	}
	if len(p.Args.Members) != 1 {
		return []*MethodGenerator{updateByGen}, nil
	}
	p = g.newUpdateParam(class)
	p.Args.Members = append(p.Args.Members, uniqueKey...)
	updateByPluralGen, err := g.newUpdateByPluralMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create UpdateByPluralMethodGenerator: %w", err)
	}
	return []*MethodGenerator{updateByGen, updateByPluralGen}, nil
}

func (g *Generator) newUpdateByMethodGeneratorsFromKey(class *types.Class, key types.Members) ([]*MethodGenerator, error) {
	p := g.newUpdateParam(class)
	p.Args.Members = append(p.Args.Members, key...)
	updateByGen, err := g.newUpdateByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create UpdateByMethodGenerator: %w", err)
	}
	if len(p.Args.Members) != 1 {
		return []*MethodGenerator{updateByGen}, nil
	}
	p = g.newUpdateParam(class)
	p.Args.Members = append(p.Args.Members, key...)
	updateByPluralGen, err := g.newUpdateByPluralMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create UpdateByPluralMethodGenerator: %w", err)
	}
	return []*MethodGenerator{updateByGen, updateByPluralGen}, nil
}

func (g *Generator) newDeleteByMethodGeneratorsFromPrimaryKey(class *types.Class, primaryKey *types.Member) ([]*MethodGenerator, error) {
	p := g.newDeleteParam(class)
	p.Args.Members = append(p.Args.Members, primaryKey)
	deleteByGen, err := g.newDeleteByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create DeleteByMethodGenerator: %w", err)
	}
	if len(p.Args.Members) != 1 {
		return []*MethodGenerator{deleteByGen}, nil
	}
	p = g.newDeleteParam(class)
	p.Args.Members = append(p.Args.Members, primaryKey)
	deleteByPluralGen, err := g.newDeleteByPluralMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create DeleteByPluralMethodGenerator: %w", err)
	}
	return []*MethodGenerator{deleteByGen, deleteByPluralGen}, nil
}

func (g *Generator) newDeleteByMethodGeneratorsFromUniqueKey(class *types.Class, uniqueKey types.Members) ([]*MethodGenerator, error) {
	p := g.newDeleteParam(class)
	p.Args.Members = append(p.Args.Members, uniqueKey...)
	deleteByGen, err := g.newDeleteByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create DeleteByMethodGenerator: %w", err)
	}
	if len(p.Args.Members) != 1 {
		return []*MethodGenerator{deleteByGen}, nil
	}
	p = g.newDeleteParam(class)
	p.Args.Members = append(p.Args.Members, uniqueKey...)
	deleteByPluralGen, err := g.newDeleteByPluralMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create DeleteByPluralMethodGenerator: %w", err)
	}
	return []*MethodGenerator{deleteByGen, deleteByPluralGen}, nil
}

func (g *Generator) newDeleteByMethodGeneratorsFromKey(class *types.Class, key types.Members) ([]*MethodGenerator, error) {
	p := g.newDeleteParam(class)
	p.Args.Members = append(p.Args.Members, key...)
	deleteByGen, err := g.newDeleteByMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create DeleteByMethodGenerator: %w", err)
	}
	if len(p.Args.Members) != 1 {
		return []*MethodGenerator{deleteByGen}, nil
	}
	p = g.newDeleteParam(class)
	p.Args.Members = append(p.Args.Members, key...)
	deleteByPluralGen, err := g.newDeleteByPluralMethodGenerator(class, p)
	if err != nil {
		return nil, xerrors.Errorf("cannot create DeleteByPluralMethodGenerator: %w", err)
	}
	return []*MethodGenerator{deleteByGen, deleteByPluralGen}, nil
}

func (g *Generator) newMethodGenerators(class *types.Class) (MethodGenerators, error) {
	gens := MethodGenerators{}
	if !class.ReadOnly {
		gen, err := g.newCreateMethodGenerator(class)
		if err != nil {
			return nil, xerrors.Errorf("cannot create CreateMethodGenerator: %w", err)
		}
		gens = append(gens, gen)
	}
	if !class.ReadOnly {
		gen, err := g.newUpdateMethodGenerator(class)
		if err != nil {
			return nil, xerrors.Errorf("cannot create UpdateMethodGenerator: %w", err)
		}
		gens = append(gens, gen)
	}
	if !class.ReadOnly {
		gen, err := g.newDeleteMethodGenerator(class)
		if err != nil {
			return nil, xerrors.Errorf("cannot create DeleteMethodGenerator: %w", err)
		}
		gens = append(gens, gen)
	}
	{
		gen, err := g.newFindAllMethodGenerator(class)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindAllMethodGenerator: %w", err)
		}
		gens = append(gens, gen)
	}
	{
		gen, err := g.newCountMethodGenerator(class)
		if err != nil {
			return nil, xerrors.Errorf("cannot create CountMethodGenerator: %w", err)
		}
		gens = append(gens, gen)
	}
	primaryKey := class.PrimaryKey()
	if primaryKey != nil {
		findByGens, err := g.newFindByMethodGeneratorsFromPrimaryKey(class, primaryKey)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindByMethodGenerators from primary key: %w", err)
		}
		gens = append(gens, findByGens...)
		if !class.ReadOnly {
			updateByGens, err := g.newUpdateByMethodGeneratorsFromPrimaryKey(class, primaryKey)
			if err != nil {
				return nil, xerrors.Errorf("cannot create UpdateByMethodGenerators from primary key: %w", err)
			}
			gens = append(gens, updateByGens...)
			deleteByGens, err := g.newDeleteByMethodGeneratorsFromPrimaryKey(class, primaryKey)
			if err != nil {
				return nil, xerrors.Errorf("cannot create DeleteByMethodGenerators from primary key: %w", err)
			}
			gens = append(gens, deleteByGens...)
		}
	}
	for _, uniqueKey := range class.UniqueKeys() {
		findByGens, err := g.newFindByMethodGeneratorsFromUniqueKey(class, uniqueKey)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindByMethodGenerators from unique key: %w", err)
		}
		gens = append(gens, findByGens...)
		if !class.ReadOnly {
			updateByGens, err := g.newUpdateByMethodGeneratorsFromUniqueKey(class, uniqueKey)
			if err != nil {
				return nil, xerrors.Errorf("cannot create UpdateByMethodGenerators from unique key: %w", err)
			}
			gens = append(gens, updateByGens...)
			deleteByGens, err := g.newDeleteByMethodGeneratorsFromUniqueKey(class, uniqueKey)
			if err != nil {
				return nil, xerrors.Errorf("cannot create DeleteByMethodGenerators from unique key: %w", err)
			}
			gens = append(gens, deleteByGens...)
		}
	}
	for _, key := range class.Keys() {
		findByGens, err := g.newFindByMethodGeneratorsFromKey(class, key)
		if err != nil {
			return nil, xerrors.Errorf("cannot create FindByMethodGenerators from key: %w", err)
		}
		gens = append(gens, findByGens...)
		if !class.ReadOnly {
			updateByGens, err := g.newUpdateByMethodGeneratorsFromKey(class, key)
			if err != nil {
				return nil, xerrors.Errorf("cannot create UpdateByMethodGenerators from key: %w", err)
			}
			gens = append(gens, updateByGens...)
			deleteByGens, err := g.newDeleteByMethodGeneratorsFromKey(class, key)
			if err != nil {
				return nil, xerrors.Errorf("cannot create DeleteByMethodGenerators from key: %w", err)
			}
			gens = append(gens, deleteByGens...)
		}
	}
	return gens, nil
}

func (g *Generator) generateDeclareImpl(class *types.Class) *Statement {
	fields := map[string]*types.ValueDeclare{}
	datastore := DataStoreByName(class.DataStore)
	datastore.StructFields(class, fields)
	g.importList = datastore.Imports(types.DefaultImportList(g.cfg.ModulePath, g.cfg.ContextImportPath()))
	for pluginName := range g.datastores[class.DataStore].plugins {
		plg, ok := Plugin(pluginName)
		if ok {
			plg.StructFields(class, fields)
		}
	}
	fieldNames := []string{}
	for _, field := range fields {
		fieldNames = append(fieldNames, field.Name)
	}
	sort.Strings(fieldNames)
	codes := []Code{}
	for _, fieldName := range fieldNames {
		field := fields[fieldName]
		if field.Type.Type.PackageName != "" {
			field.Type.Type.PackageName = g.importList.Package(field.Type.Type.PackageName)
		}
		codes = append(codes, field.Code(g.importList))
	}
	implName := fmt.Sprintf("%sImpl", class.Name.CamelName())
	return GoType().Id(implName).Struct(codes...)
}

func (g *Generator) generateConstructor(class *types.Class, decl *types.ConstructorDeclare) *Statement {
	codes := []Code{}
	datastore := g.datastores[class.DataStore]
	for _, fn := range datastore.pluginMap[ConstructorPlugin] {
		c, _ := fn(g.newConstructorParam(class))
		codes = append(codes, c...)
	}
	return decl.MethodInterface(g.importList).Block(codes...)
}

func (g *Generator) existsFile(class *types.Class, path string) bool {
	_, err := os.Stat(filepath.Join(path, fmt.Sprintf("%s.go", class.Name.SnakeName())))
	return err == nil
}

type Decl struct {
	Name string
	Code Code
}

type FuncDecl struct {
	*Decl
	Comment string
}

func (d *FuncDecl) IsGenerated() bool {
	return d.Comment == code.FuncGeneratedMarker
}

type FuncDecls []*FuncDecl

func (d FuncDecls) GeneratedFuncs() FuncDecls {
	decls := FuncDecls{}
	for _, f := range d {
		if !f.IsGenerated() {
			continue
		}
		decls = append(decls, f)
	}
	return decls
}

func (d FuncDecls) FuncNameMap() map[string]*FuncDecl {
	nameMap := map[string]*FuncDecl{}
	for _, fn := range d {
		nameMap[fn.Name] = fn
	}
	return nameMap
}

type PackageDecl struct {
	Funcs      FuncDecls
	Interfaces []*types.MethodDeclare
	Structs    []*Decl
	Imports    types.ImportList
}

func (d *PackageDecl) InterfaceMap() map[string]*types.MethodDeclare {
	m := map[string]*types.MethodDeclare{}
	for _, decl := range d.Interfaces {
		m[decl.MethodName] = decl
	}
	return m
}

func (g *Generator) formatNode(node ast.Node, fset *token.FileSet) (string, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "", xerrors.Errorf("cannot format node: %w", err)
	}
	return buf.String(), nil
}

func (g *Generator) nodeToCode(node ast.Node, fset *token.FileSet) (Code, error) {
	code, err := g.formatNode(node, fset)
	if err != nil {
		return nil, xerrors.Errorf("cannot format node: %w", err)
	}
	return Id(code), nil
}

func (g *Generator) parseImport(decl *ast.GenDecl) (types.ImportList, error) {
	importList := types.ImportList{}
	for _, spec := range decl.Specs {
		importSpec := spec.(*ast.ImportSpec)
		var name string
		if importSpec.Name == nil {
			path, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return nil, xerrors.Errorf("cannot unquote from %s: %w", importSpec.Path.Value, err)
			}
			splitted := strings.Split(path, "/")
			name = splitted[len(splitted)-1]
		} else {
			name = importSpec.Name.String()
		}
		path, _ := strconv.Unquote(importSpec.Path.Value)
		importList[name] = &types.ImportDeclare{
			Name: name,
			Path: path,
		}
	}
	return importList, nil
}

func (g *Generator) exprToTypeDeclare(expr ast.Expr) *types.TypeDeclare {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return types.TypeDeclareWithType(&types.Type{
			PackageName: g.importList.Package(e.X.(*ast.Ident).String()),
			Name:        e.Sel.String(),
		})

	case *ast.Ident:
		return types.TypeDeclareWithType(&types.Type{
			Name: e.String(),
		})
	case *ast.StarExpr:
		typ := g.exprToTypeDeclare(e.X)
		typ.IsPointer = true
		return typ
	case *ast.ArrayType:
		index := "[]"
		if e.Len != nil {
			index = fmt.Sprintf("[%s]", e.Len.(*ast.Ident).String())
		}
		typeName := g.exprToTypeDeclare(e.Elt).FormatName(g.importList)
		return types.TypeDeclareWithName(fmt.Sprintf("%s%s", index, typeName))
	case *ast.ChanType:
		typeName := g.exprToTypeDeclare(e.Value).FormatName(g.importList)
		chanName := ""
		if e.Dir == ast.SEND {
			chanName = "chan<-"
		} else {
			chanName = "<-chan"
		}
		return types.TypeDeclareWithName(fmt.Sprintf("%s %s", chanName, typeName))
	case *ast.FuncType:
	case *ast.InterfaceType:
		return types.TypeDeclareWithName("interface{}")
	case *ast.MapType:
		keyName := g.exprToTypeDeclare(e.Key).FormatName(g.importList)
		valueName := g.exprToTypeDeclare(e.Value).FormatName(g.importList)
		return types.TypeDeclareWithName(fmt.Sprintf("map[%s]%s", keyName, valueName))
	case *ast.StructType:
		return types.TypeDeclareWithName("struct{}")
	}
	return nil
}

func (g *Generator) parseType(pkg *PackageDecl, decl *ast.GenDecl, fset *token.FileSet) error {
	for _, spec := range decl.Specs {
		typeSpec := spec.(*ast.TypeSpec)
		switch t := typeSpec.Type.(type) {
		case *ast.StructType:
			code, err := g.nodeToCode(typeSpec, fset)
			if err != nil {
				return xerrors.Errorf("cannot convert struct to Code: %w", err)
			}
			pkg.Structs = append(pkg.Structs, &Decl{
				Name: typeSpec.Name.String(),
				Code: code,
			})
		case *ast.InterfaceType:
			for _, field := range t.Methods.List {
				method := &types.MethodDeclare{
					Args:   types.ValueDeclares{},
					Return: types.ValueDeclares{},
				}
				for idx, param := range field.Type.(*ast.FuncType).Params.List {
					method.Args = append(method.Args, &types.ValueDeclare{
						Name: fmt.Sprintf("a%d", idx),
						Type: g.exprToTypeDeclare(param.Type),
					})
				}
				for _, param := range field.Type.(*ast.FuncType).Results.List {
					method.Return = append(method.Return, &types.ValueDeclare{Type: g.exprToTypeDeclare(param.Type)})
				}
				name := field.Names[0].String()
				method.MethodName = name
				pkg.Interfaces = append(pkg.Interfaces, method)
			}
		}
	}
	return nil
}

func (g *Generator) parseFile(class *types.Class, path string) (*PackageDecl, error) {
	pkg := &PackageDecl{
		Funcs:      FuncDecls{},
		Interfaces: []*types.MethodDeclare{},
		Structs:    []*Decl{},
	}
	src := filepath.Join(path, fmt.Sprintf("%s.go", class.Name.SnakeName()))
	if _, err := os.Stat(src); err != nil {
		return pkg, nil
	}
	bytes, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, xerrors.Errorf("cannot read file %s: %w", src, err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", bytes, parser.Mode(parser.ParseComments))
	if err != nil {
		return nil, xerrors.Errorf("cannot parse to %s: %w", string(bytes), err)
	}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.IMPORT:
				importList, err := g.parseImport(d)
				if err != nil {
					return nil, xerrors.Errorf("cannot parse import statement: %w", err)
				}
				pkg.Imports = importList
			case token.CONST:
				log.Println("unsupported global const block")
			case token.TYPE:
				if err := g.parseType(pkg, d, fset); err != nil {
					return nil, xerrors.Errorf("cannot parse type statement: %w", err)
				}
			case token.VAR:
				log.Println("unsupported global var block")
			}
		case *ast.FuncDecl:
			code, err := g.nodeToCode(d, fset)
			if err != nil {
				return nil, xerrors.Errorf("cannot convert ast.Node to Code for function: %w", err)
			}
			pkg.Funcs = append(pkg.Funcs, &FuncDecl{
				Decl: &Decl{
					Name: d.Name.String(),
					Code: code,
				},
				Comment: strings.TrimRight(d.Doc.Text(), "\n"),
			})
		default:
		}
	}
	return pkg, nil
}

func (g *Generator) PackageDeclare(class *types.Class, path string) (*Declare, error) {
	importList := types.DefaultImportList(g.cfg.ModulePath, g.cfg.ContextImportPath())
	datastore := DataStoreByName(class.DataStore)
	g.importList = datastore.Imports(g.importList)
	for pluginName := range g.datastores[class.DataStore].plugins {
		plg, ok := Plugin(pluginName)
		if ok {
			g.importList = plg.Imports(g.importList)
		}
	}
	pkg, err := g.parseFile(class, path)
	if err != nil {
		return nil, xerrors.Errorf("cannot parse file %s: %w", path, err)
	}
	for _, importDecl := range pkg.Imports {
		g.importList[importDecl.Name] = importDecl
	}
	for _, importDecl := range importList {
		g.importList[importDecl.Name] = importDecl
	}
	constructorDecl, err := g.newConstructorDeclare(class)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for constructor of %s: %w", class.Name.SnakeName(), err)
	}
	return &Declare{
		Imports:     g.importList,
		Constructor: constructorDecl,
		Methods:     pkg.Interfaces,
	}, nil
}

func (g *Generator) generate(class *types.Class, path string) ([]byte, error) {
	importList := types.DefaultImportList(g.cfg.ModulePath, g.cfg.ContextImportPath())
	datastore := DataStoreByName(class.DataStore)
	g.importList = datastore.Imports(g.importList)
	for pluginName := range g.datastores[class.DataStore].plugins {
		plg, ok := Plugin(pluginName)
		if ok {
			g.importList = plg.Imports(g.importList)
		}
	}
	for _, importDecl := range importList {
		g.importList[importDecl.Name] = importDecl
	}
	pkg, err := g.parseFile(class, path)
	if err != nil {
		return nil, xerrors.Errorf("cannot parse file %s: %w", path, err)
	}
	for _, importDecl := range pkg.Imports {
		g.importList[importDecl.Name] = importDecl
	}
	f := NewFile(g.packageName)
	for _, importDeclare := range g.importList {
		f.ImportName(importDeclare.Path, importDeclare.Name)
	}
	gens, err := g.newMethodGenerators(class)
	if err != nil {
		return nil, xerrors.Errorf("cannot create MethodGenerators for %s: %w", class.Name.SnakeName(), err)
	}
	newGeneratedNameMap := map[string]*MethodGenerator{}
	for _, gen := range gens {
		newGeneratedNameMap[gen.decl.MethodName] = gen
	}
	constructorDecl, err := g.newConstructorDeclare(class)
	if err != nil {
		return nil, xerrors.Errorf("failed to declaration for constructor of %s: %w", class.Name.SnakeName(), err)
	}
	constructorName := constructorDecl.MethodName
	methodNames := []string{}
	generatedFuncCodeMap := map[string]Code{}
	customizedFuncCodes := []Code{}
	interfaceMap := pkg.InterfaceMap()
	for _, fn := range pkg.Funcs {
		if !fn.IsGenerated() {
			continue
		}
		if constructorName == fn.Name {
			continue
		}
		gen, exists := newGeneratedNameMap[fn.Name]
		if exists {
			// overwrite by new generated method interface
			interfaceMap[fn.Name] = gen.decl

			generatedFuncCodeMap[fn.Name] = gen.Generate(g.importList)
			methodNames = append(methodNames, fn.Name)
		} else {
			delete(interfaceMap, fn.Name)
		}
	}
	funcNameMap := pkg.Funcs.FuncNameMap()
	for name, gen := range newGeneratedNameMap {
		if _, exists := interfaceMap[name]; !exists {
			interfaceMap[name] = gen.decl
		}
		if _, exists := funcNameMap[name]; !exists {
			methodNames = append(methodNames, name)
			generatedFuncCodeMap[name] = gen.Generate(g.importList)
		}
	}
	sort.Strings(methodNames)
	for _, fn := range pkg.Funcs {
		if fn.IsGenerated() {
			continue
		}
		if constructorName == fn.Name {
			continue
		}
		methodNames = append(methodNames, fn.Name)
		customizedFuncCodes = append(customizedFuncCodes, fn.Code)
	}
	interfaceCodes := []Code{}
	for _, methodName := range methodNames {
		decl, exists := interfaceMap[methodName]
		if exists {
			interfaceCodes = append(interfaceCodes, decl.Interface(g.importList))
		}
	}
	f.Line()
	f.Add(GoType().Id(class.Name.CamelName()).Interface(interfaceCodes...))
	f.Line()
	f.Add(g.generateDeclareImpl(class))
	f.Line()
	constructorFn, exists := funcNameMap[constructorName]
	if exists {
		if constructorFn.IsGenerated() {
			f.Add(g.generateConstructor(class, constructorDecl))
		} else {
			f.Add(constructorFn.Code)
		}
	} else {
		f.Add(g.generateConstructor(class, constructorDecl))
	}
	for _, methodName := range methodNames {
		funcCode, exists := generatedFuncCodeMap[methodName]
		if exists {
			f.Comment(code.FuncGeneratedMarker)
			f.Add(funcCode)
		}
	}
	for _, funcCode := range customizedFuncCodes {
		f.Line()
		f.Add(funcCode)
	}
	bytes := []byte(fmt.Sprintf("%#v", f))
	source, err := imports.Process("", bytes, nil)
	if err != nil {
		return nil, xerrors.Errorf("failed to format by goimport %s: %w", string(source), err)
	}
	return source, nil
}

func (g *Generator) writeFile(class *types.Class, path string, source []byte) error {
	if err := ioutil.WriteFile(
		filepath.Join(path, fmt.Sprintf("%s.go", class.Name.SnakeName())),
		source, 0644,
	); err != nil {
		return xerrors.Errorf("cannot write file %s: %w", path, err)
	}
	return nil
}

func (g *Generator) Generate(classes []*types.Class) error {
	path := g.cfg.OutputPathWithPackage(g.packageName)
	if err := os.MkdirAll(path, 0755); err != nil {
		return xerrors.Errorf("cannot create directory to %s: %w", path, err)
	}
	for _, class := range classes {
		source, err := g.generate(class, path)
		if err != nil {
			return xerrors.Errorf("cannot generate source for %s: %w", class.Name.SnakeName(), err)
		}
		if err := g.writeFile(class, path, source); err != nil {
			return xerrors.Errorf("cannot write file for %s: %w", class.Name.SnakeName(), err)
		}
	}
	return nil
}
