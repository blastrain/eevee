package dao

import (
	"sync"

	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

const (
	// ImportPlugin name for hook of Imports
	ImportsPlugin = "imports"
	// StructFieldsPlugin name for hook of StructFields
	StructFieldsPlugin = "struct-fields"
	// ConstructorDeclarePlugin name for hook of ConstructorDeclare
	ConstructorDeclarePlugin = "constructor-declare"
	// ConstructorPlugin name for hook of Constructor
	ConstructorPlugin = "constructor"
	// CreateDeclarePlugin name for hook of CreateDeclare
	CreateDeclarePlugin = "create-declare"
	// BeforeCreatePlugin name for hook of BeforeCreate
	BeforeCreatePlugin = "before-create"
	// CreatePlugin name for hook of Create
	CreatePlugin = "create"
	// AfterCreatePlugin name for hook of AfterCreate
	AfterCreatePlugin = "after-create"
	// UpdateDeclarePlugin name for hook of UpdateDeclare
	UpdateDeclarePlugin = "update-declare"
	// BeforeUpdatePlugin name for hook of BeforeUpdate
	BeforeUpdatePlugin = "before-update"
	// UpdatePlugin name for hook of Update
	UpdatePlugin = "update"
	// AfterUpdatePlugin name for hook of AfterUpdate
	AfterUpdatePlugin = "after-update"
	// DeleteDeclarePlugin name for hook of DeleteDeclare
	DeleteDeclarePlugin = "delete-declare"
	// BeforeDeletePlugin name for hook of BeforeDelete
	BeforeDeletePlugin = "before-delete"
	// DeletePlugin name for hook of Delete
	DeletePlugin = "delete"
	// AfterDeletePlugin name for hook of AfterDelete
	AfterDeletePlugin = "after-delete"
	// FindAllDeclarePlugin name for hook of FindAllDeclare
	FindAllDeclarePlugin = "find-all-declare"
	// BeforeFindAllPlugin name for hook of BeforeFindAll
	BeforeFindAllPlugin = "before-find-all"
	// FindAllPlugin name for hook of FindAll
	FindAllPlugin = "find-all"
	// AfterFindAllPlugin name for hook of AfterFindAll
	AfterFindAllPlugin = "after-find-all"
	// CountDeclarePlugin name for hook of CountDeclare
	CountDeclarePlugin = "count-declare"
	// BeforeCountPlugin name for hook of BeforeCount
	BeforeCountPlugin = "before-count"
	// CountPlugin name for hook of Count
	CountPlugin = "count"
	// AfterCountPlugin name for hook of AfterCount
	AfterCountPlugin = "after-count"
	// FindByDeclarePlugin name for hook of FindByDeclare
	FindByDeclarePlugin = "findby-declare"
	// BeforeFindByPlugin name for hook of BeforeFindBy
	BeforeFindByPlugin = "before-findby"
	// FindByPlugin name for hook of FindBy
	FindByPlugin = "findby"
	// AfterFindByPlugin name for hook of AfterFindBy
	AfterFindByPlugin = "after-findby"
	// FindByPluralDeclarePlugin name for hook of FindByPluralDeclare
	FindByPluralDeclarePlugin = "findby-plural-declare"
	// BeforeFindByPluralPlugin name for hook of BeforeFindByPlural
	BeforeFindByPluralPlugin = "before-findby-plural"
	// FindByPluralPlugin name for hook of FindByPlural
	FindByPluralPlugin = "findby-plural"
	// AfterFindByPluralPlugin name for hook of AfterFindByPlural
	AfterFindByPluralPlugin = "after-findby-plural"
	// UpdateByDeclarePlugin name for hook of UpdateByDeclare
	UpdateByDeclarePlugin = "updateby-declare"
	// BeforeUpdateByPlugin name for hook of BeforeUpdateBy
	BeforeUpdateByPlugin = "before-updateby"
	// UpdateByPlugin name for hook of UpdateBy
	UpdateByPlugin = "updateby"
	// AfterUpdateByPlugin name for hook of AfterUpdateBy
	AfterUpdateByPlugin = "after-updateby"
	// UpdateByPluralDeclarePlugin name for hook of UpdateByPluralDeclare
	UpdateByPluralDeclarePlugin = "updateby-plural-declare"
	// BeforeUpdateByPluralPlugin name for hook of BeforeUpdateByPlural
	BeforeUpdateByPluralPlugin = "before-updateby-plural"
	// UpdateByPluralPlugin name for hook of UpdateByPlural
	UpdateByPluralPlugin = "updateby-plural"
	// AfterUpdateByPluralPlugin name for hook of AfterUpdateByPlural
	AfterUpdateByPluralPlugin = "after-updateby-plural"
	// DeleteByDeclarePlugin name for hook of DeleteByDeclare
	DeleteByDeclarePlugin = "deleteby-declare"
	// BeforeDeleteByPlugin name for hook of BeforeDeleteBy
	BeforeDeleteByPlugin = "before-deleteby"
	// DeleteByPlugin name for hook of DeleteBy
	DeleteByPlugin = "deleteby"
	// AfterDeleteByPlugin name for hook of AfterDeleteBy
	AfterDeleteByPlugin = "after-deleteby"
	// DeleteByPluralDeclarePlugin name for hook of DeleteByPluralDeclare
	DeleteByPluralDeclarePlugin = "deleteby-plural-declare"
	// BeforeDeleteByPluralPlugin name for hook of BeforeDeleteByPlural
	BeforeDeleteByPluralPlugin = "before-deleteby-plural"
	// DeleteByPluralPlugin name for hook of DeleteByPlural
	DeleteByPluralPlugin = "deleteby-plural"
	// AfterDeleteByPluralPlugin name for hook of AfterDeleteByPlural
	AfterDeleteByPluralPlugin = "after-deleteby-plural"
)

type DataStorePlugin interface {
	// StructFields hook definition of data accessor structure
	StructFields(*types.Class, types.StructFieldList) types.StructFieldList
	// Imports hook import statement
	Imports(types.ImportList) types.ImportList
	// ConstructorDeclare hook declaration for NewXX interface
	ConstructorDeclare(*types.ConstructorDeclare) error
	// Constructor initialize data accessor
	Constructor(*types.ConstructorParam) []Code
	// Create exec insert query to database in default
	Create(*types.CreateParam) []Code
	// Update exec update query with 'where id = ?' to database in default
	Update(*types.UpdateParam) []Code
	// Delete exec delete query with 'where id = ?' to database in default
	Delete(*types.DeleteParam) []Code
	// FindAll exec select query without where statement to database in default
	FindAll(*types.FindParam) []Code
	// Count exec count query to database in default
	Count(*types.CountParam) []Code
	// FindBy exec select query to database in default
	FindBy(*types.FindParam) []Code
	// FindByPlural exec select query to database in default
	FindByPlural(*types.FindParam) []Code
	// UpdateBy exec update query to database in default
	UpdateBy(*types.UpdateParam) []Code
	// UpdateByPlural exec update query to database in default
	UpdateByPlural(*types.UpdateParam) []Code
	// DeleteBy exec delete query to database in default
	DeleteBy(*types.DeleteParam) []Code
	// DeleteByPlural exec delete query to database in default
	DeleteByPlural(*types.DeleteParam) []Code
}

type DAOPlugin interface {
	DataStorePlugin
	// CreateDeclare hook declaration for Create interface
	CreateDeclare(*types.MethodDeclare) error
	// BeforeCreate insert some codes as first statement for Create
	BeforeCreate(*types.CreateParam) []Code
	// AfterCreate insert some codes in 'defer' function for Create
	AfterCreate(*types.CreateParam) []Code
	// UpdateDeclare hook declaration for Update interface
	UpdateDeclare(*types.MethodDeclare) error
	// BeforeUpdate insert some codes as first statement for Update
	BeforeUpdate(*types.UpdateParam) []Code
	// AfterUpdate insert some codes in 'defer' function for Update
	AfterUpdate(*types.UpdateParam) []Code
	// DeleteDeclare hook declaration for Delete interface
	DeleteDeclare(*types.MethodDeclare) error
	// BeforeDelete insert some codes as first statement for Delete
	BeforeDelete(*types.DeleteParam) []Code
	// AfterDelete insert some codes in 'defer' function for Delete
	AfterDelete(*types.DeleteParam) []Code
	// FindAllDeclare hook declaration for FindAll interface
	FindAllDeclare(*types.MethodDeclare) error
	// BeforeFindAll insert some codes as first statement for FindAll
	BeforeFindAll(*types.FindParam) []Code
	// AfterFindAll insert some codes in 'defer' function for FindAll
	AfterFindAll(*types.FindParam) []Code
	// CountDeclare hook declaration for Count interface
	CountDeclare(*types.MethodDeclare) error
	// BeforeCount insert some codes as first statement for Count
	BeforeCount(*types.CountParam) []Code
	// AfterCount insert some codes in 'defer' function for Count
	AfterCount(*types.CountParam) []Code
	// FindByDeclare hook declaration for FindBy interface
	FindByDeclare(*types.MethodDeclare) error
	// BeforeFindBy insert some codes as first statement for FindBy
	BeforeFindBy(*types.FindParam) []Code
	// AfterFindBy insert some codes in 'defer' function for FindBy
	AfterFindBy(*types.FindParam) []Code
	// FindByPluralDeclare hook declaration for FindByPlural interface
	FindByPluralDeclare(*types.MethodDeclare) error
	// BeforeFindByPlural insert some codes as first statement for FindByPlural
	BeforeFindByPlural(*types.FindParam) []Code
	// AfterFindByPlural insert some codes in 'defer' function for FindByPlural
	AfterFindByPlural(*types.FindParam) []Code
	// UpdateByDeclare hook declaration for UpdateBy interface
	UpdateByDeclare(*types.MethodDeclare) error
	// BeforeUpdateBy insert some codes as first statement for UpdateBy
	BeforeUpdateBy(*types.UpdateParam) []Code
	// AfterUpdateBy insert some codes in 'defer' function for UpdateBy
	AfterUpdateBy(*types.UpdateParam) []Code
	// UpdateByPluralDeclare hook declaration for UpdateByPlural interface
	UpdateByPluralDeclare(*types.MethodDeclare) error
	// BeforeUpdateByPlural insert some codes as first statement for UpdateByPlural
	BeforeUpdateByPlural(*types.UpdateParam) []Code
	// AfterUpdateByPlural insert some codes in 'defer' function for UpdateByPlural
	AfterUpdateByPlural(*types.UpdateParam) []Code
	// DeleteByDeclare hook declaration for DeleteBy interface
	DeleteByDeclare(*types.MethodDeclare) error
	// BeforeDeleteBy insert some codes as first statement for DeleteBy
	BeforeDeleteBy(*types.DeleteParam) []Code
	// AfterDeleteBy insert some codes in 'defer' function for DeleteBy
	AfterDeleteBy(*types.DeleteParam) []Code
	// DeleteByPluralDeclare hook declaration for DeleteByPlural interface
	DeleteByPluralDeclare(*types.MethodDeclare) error
	// BeforeDeleteByPlural insert some codes as first statement for DeleteByPlural
	BeforeDeleteByPlural(*types.DeleteParam) []Code
	// AfterDeleteByPlural insert some codes in 'defer' function for DeleteByPlural
	AfterDeleteByPlural(*types.DeleteParam) []Code
}

var (
	datastoresMu sync.RWMutex
	pluginsMu    sync.RWMutex
	datastores   = map[string]DataStorePlugin{}
	plugins      = map[string]DAOPlugin{}
)

func RegisterDataStore(name string, datastore DataStorePlugin) {
	datastoresMu.Lock()
	defer datastoresMu.Unlock()
	datastores[name] = datastore
}

func Register(name string, plugin DAOPlugin) {
	pluginsMu.Lock()
	defer pluginsMu.Unlock()
	plugins[name] = plugin
}

type DAOPluginMap map[string]func(types.DAOContext) ([]Code, error)

func DataStoreMap(name string) (DAOPluginMap, bool) {
	datastore, ok := datastores[name]
	if !ok {
		return nil, ok
	}
	return map[string]func(types.DAOContext) ([]Code, error){
		ConstructorDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, datastore.ConstructorDeclare(c.(*types.ConstructorDeclare))
		},
		ConstructorPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.Constructor(c.(*types.ConstructorParam)), nil
		},
		CreatePlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.Create(c.(*types.CreateParam)), nil
		},
		UpdatePlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.Update(c.(*types.UpdateParam)), nil
		},
		DeletePlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.Delete(c.(*types.DeleteParam)), nil
		},
		FindAllPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.FindAll(c.(*types.FindParam)), nil
		},
		CountPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.Count(c.(*types.CountParam)), nil
		},
		FindByPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.FindBy(c.(*types.FindParam)), nil
		},
		FindByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.FindByPlural(c.(*types.FindParam)), nil
		},
		UpdateByPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.UpdateBy(c.(*types.UpdateParam)), nil
		},
		UpdateByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.UpdateByPlural(c.(*types.UpdateParam)), nil
		},
		DeleteByPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.DeleteBy(c.(*types.DeleteParam)), nil
		},
		DeleteByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return datastore.DeleteByPlural(c.(*types.DeleteParam)), nil
		},
	}, ok
}

func PluginMap(name string) (DAOPluginMap, bool) {
	plugin, ok := plugins[name]
	if !ok {
		return nil, ok
	}
	return map[string]func(types.DAOContext) ([]Code, error){
		ConstructorDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.ConstructorDeclare(c.(*types.ConstructorDeclare))
		},
		ConstructorPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.Constructor(c.(*types.ConstructorParam)), nil
		},
		CreateDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.CreateDeclare(c.(*types.MethodDeclare))
		},
		BeforeCreatePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeCreate(c.(*types.CreateParam)), nil
		},
		CreatePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.Create(c.(*types.CreateParam)), nil
		},
		AfterCreatePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterCreate(c.(*types.CreateParam)), nil
		},
		UpdateDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.UpdateDeclare(c.(*types.MethodDeclare))
		},
		BeforeUpdatePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeUpdate(c.(*types.UpdateParam)), nil
		},
		UpdatePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.Update(c.(*types.UpdateParam)), nil
		},
		AfterUpdatePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterUpdate(c.(*types.UpdateParam)), nil
		},
		DeleteDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.DeleteDeclare(c.(*types.MethodDeclare))
		},
		BeforeDeletePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeDelete(c.(*types.DeleteParam)), nil
		},
		DeletePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.Delete(c.(*types.DeleteParam)), nil
		},
		AfterDeletePlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterDelete(c.(*types.DeleteParam)), nil
		},
		FindAllDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.FindAllDeclare(c.(*types.MethodDeclare))
		},
		BeforeFindAllPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeFindAll(c.(*types.FindParam)), nil
		},
		FindAllPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.FindAll(c.(*types.FindParam)), nil
		},
		AfterFindAllPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterFindAll(c.(*types.FindParam)), nil
		},
		CountDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.CountDeclare(c.(*types.MethodDeclare))
		},
		BeforeCountPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeCount(c.(*types.CountParam)), nil
		},
		CountPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.Count(c.(*types.CountParam)), nil
		},
		AfterCountPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterCount(c.(*types.CountParam)), nil
		},
		FindByDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.FindByDeclare(c.(*types.MethodDeclare))
		},
		BeforeFindByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeFindBy(c.(*types.FindParam)), nil
		},
		FindByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.FindBy(c.(*types.FindParam)), nil
		},
		AfterFindByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterFindBy(c.(*types.FindParam)), nil
		},
		FindByPluralDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.FindByPluralDeclare(c.(*types.MethodDeclare))
		},
		BeforeFindByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeFindByPlural(c.(*types.FindParam)), nil
		},
		FindByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.FindByPlural(c.(*types.FindParam)), nil
		},
		AfterFindByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterFindByPlural(c.(*types.FindParam)), nil
		},
		UpdateByDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.UpdateByDeclare(c.(*types.MethodDeclare))
		},
		BeforeUpdateByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeUpdateBy(c.(*types.UpdateParam)), nil
		},
		UpdateByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.UpdateBy(c.(*types.UpdateParam)), nil
		},
		AfterUpdateByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterUpdateBy(c.(*types.UpdateParam)), nil
		},
		UpdateByPluralDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.UpdateByPluralDeclare(c.(*types.MethodDeclare))
		},
		BeforeUpdateByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeUpdateByPlural(c.(*types.UpdateParam)), nil
		},
		UpdateByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.UpdateByPlural(c.(*types.UpdateParam)), nil
		},
		AfterUpdateByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterUpdateByPlural(c.(*types.UpdateParam)), nil
		},
		DeleteByDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.DeleteByDeclare(c.(*types.MethodDeclare))
		},
		BeforeDeleteByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeDeleteBy(c.(*types.DeleteParam)), nil
		},
		DeleteByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.DeleteBy(c.(*types.DeleteParam)), nil
		},
		AfterDeleteByPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterDeleteBy(c.(*types.DeleteParam)), nil
		},
		DeleteByPluralDeclarePlugin: func(c types.DAOContext) ([]Code, error) {
			return nil, plugin.DeleteByPluralDeclare(c.(*types.MethodDeclare))
		},
		BeforeDeleteByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.BeforeDeleteByPlural(c.(*types.DeleteParam)), nil
		},
		DeleteByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.DeleteByPlural(c.(*types.DeleteParam)), nil
		},
		AfterDeleteByPluralPlugin: func(c types.DAOContext) ([]Code, error) {
			return plugin.AfterDeleteByPlural(c.(*types.DeleteParam)), nil
		},
	}, ok
}

func DataStoreByName(name string) DataStorePlugin {
	return datastores[name]
}

func DataStoreNames() []string {
	names := []string{}
	for name := range datastores {
		names = append(names, name)
	}
	return names
}

func Plugin(name string) (DAOPlugin, bool) {
	plugin, ok := plugins[name]
	return plugin, ok
}

func Plugins() []DAOPlugin {
	plgs := []DAOPlugin{}
	for _, plg := range plugins {
		plgs = append(plgs, plg)
	}
	return plgs
}
