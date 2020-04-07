package dao

import (
	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

type DefaultPlugin struct {
}

func (*DefaultPlugin) Imports(pkgs types.ImportList) types.ImportList { return pkgs }
func (*DefaultPlugin) StructFields(class *types.Class, fields types.StructFieldList) types.StructFieldList {
	return fields
}
func (*DefaultPlugin) ConstructorDeclare(d *types.ConstructorDeclare) error { return nil }
func (*DefaultPlugin) Constructor(p *types.ConstructorParam) []Code         { return []Code{} }
func (*DefaultPlugin) CreateDeclare(d *types.MethodDeclare) error           { return nil }
func (*DefaultPlugin) Create(*types.CreateParam) []Code                     { return []Code{} }
func (*DefaultPlugin) BeforeCreate(*types.CreateParam) []Code               { return []Code{} }
func (*DefaultPlugin) AfterCreate(*types.CreateParam) []Code                { return []Code{} }
func (*DefaultPlugin) UpdateDeclare(d *types.MethodDeclare) error           { return nil }
func (*DefaultPlugin) Update(*types.UpdateParam) []Code                     { return []Code{} }
func (*DefaultPlugin) BeforeUpdate(*types.UpdateParam) []Code               { return []Code{} }
func (*DefaultPlugin) AfterUpdate(*types.UpdateParam) []Code                { return []Code{} }
func (*DefaultPlugin) DeleteDeclare(d *types.MethodDeclare) error           { return nil }
func (*DefaultPlugin) Delete(*types.DeleteParam) []Code                     { return []Code{} }
func (*DefaultPlugin) BeforeDelete(p *types.DeleteParam) []Code             { return []Code{} }
func (*DefaultPlugin) AfterDelete(p *types.DeleteParam) []Code              { return []Code{} }
func (*DefaultPlugin) FindAllDeclare(d *types.MethodDeclare) error          { return nil }
func (*DefaultPlugin) FindAll(*types.FindParam) []Code                      { return []Code{} }
func (*DefaultPlugin) BeforeFindAll(p *types.FindParam) []Code              { return []Code{} }
func (*DefaultPlugin) AfterFindAll(p *types.FindParam) []Code               { return []Code{} }
func (*DefaultPlugin) CountDeclare(d *types.MethodDeclare) error            { return nil }
func (*DefaultPlugin) Count(*types.CountParam) []Code                       { return []Code{} }
func (*DefaultPlugin) BeforeCount(p *types.CountParam) []Code               { return []Code{} }
func (*DefaultPlugin) AfterCount(p *types.CountParam) []Code                { return []Code{} }
func (*DefaultPlugin) FindByDeclare(d *types.MethodDeclare) error           { return nil }
func (*DefaultPlugin) FindBy(*types.FindParam) []Code                       { return []Code{} }
func (*DefaultPlugin) BeforeFindBy(p *types.FindParam) []Code               { return []Code{} }
func (*DefaultPlugin) AfterFindBy(p *types.FindParam) []Code                { return []Code{} }
func (*DefaultPlugin) FindByPluralDeclare(d *types.MethodDeclare) error     { return nil }
func (*DefaultPlugin) FindByPlural(*types.FindParam) []Code                 { return []Code{} }
func (*DefaultPlugin) BeforeFindByPlural(p *types.FindParam) []Code         { return []Code{} }
func (*DefaultPlugin) AfterFindByPlural(p *types.FindParam) []Code          { return []Code{} }
func (*DefaultPlugin) UpdateByDeclare(d *types.MethodDeclare) error         { return nil }
func (*DefaultPlugin) UpdateBy(*types.UpdateParam) []Code                   { return []Code{} }
func (*DefaultPlugin) BeforeUpdateBy(p *types.UpdateParam) []Code           { return []Code{} }
func (*DefaultPlugin) AfterUpdateBy(p *types.UpdateParam) []Code            { return []Code{} }
func (*DefaultPlugin) UpdateByPluralDeclare(d *types.MethodDeclare) error   { return nil }
func (*DefaultPlugin) UpdateByPlural(*types.UpdateParam) []Code             { return []Code{} }
func (*DefaultPlugin) BeforeUpdateByPlural(p *types.UpdateParam) []Code     { return []Code{} }
func (*DefaultPlugin) AfterUpdateByPlural(p *types.UpdateParam) []Code      { return []Code{} }
func (*DefaultPlugin) DeleteByDeclare(d *types.MethodDeclare) error         { return nil }
func (*DefaultPlugin) DeleteBy(*types.DeleteParam) []Code                   { return []Code{} }
func (*DefaultPlugin) BeforeDeleteBy(p *types.DeleteParam) []Code           { return []Code{} }
func (*DefaultPlugin) AfterDeleteBy(p *types.DeleteParam) []Code            { return []Code{} }
func (*DefaultPlugin) DeleteByPluralDeclare(d *types.MethodDeclare) error   { return nil }
func (*DefaultPlugin) DeleteByPlural(*types.DeleteParam) []Code             { return []Code{} }
func (*DefaultPlugin) BeforeDeleteByPlural(p *types.DeleteParam) []Code     { return []Code{} }
func (*DefaultPlugin) AfterDeleteByPlural(p *types.DeleteParam) []Code      { return []Code{} }
