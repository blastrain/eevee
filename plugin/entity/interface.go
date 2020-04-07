package entity

import "go.knocknote.io/eevee/types"

const (
	// AddMethodsPlugin name for hook of AddMethods
	AddMethodsPlugin = "add-methods"
)

type EntityPlugin interface {
	Imports(types.ImportList) types.ImportList
	AddMethods(*types.EntityMethodHelper) types.Methods
}

var (
	plugins = map[string]EntityPlugin{}
)

func Register(name string, plugin EntityPlugin) {
	plugins[name] = plugin
}

func Plugin(name string) (EntityPlugin, bool) {
	plugin, ok := plugins[name]
	return plugin, ok
}

func Plugins() []EntityPlugin {
	plgs := []EntityPlugin{}
	for _, plg := range plugins {
		plgs = append(plgs, plg)
	}
	return plgs
}
