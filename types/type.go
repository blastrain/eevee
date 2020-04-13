package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/iancoleman/strcase"
	"go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/plural"
	"golang.org/x/xerrors"
)

type INDEX struct {
	PrimaryKey string       `yaml:"primary_key"`
	UniqueKeys []*UniqueKey `yaml:"unique_keys,omitempty"`
	Keys       []*Key       `yaml:"keys,omitempty"`
}

type UniqueKey struct {
	Columns []string
}

func (k *UniqueKey) MarshalYAML() (interface{}, error) {
	return k.Columns, nil
}

func (k *UniqueKey) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&k.Columns); err != nil {
		return xerrors.Errorf("cannot unmarshal Columns: %w", err)
	}
	return nil
}

type Key struct {
	Columns []string
}

func (k *Key) MarshalYAML() (interface{}, error) {
	return k.Columns, nil
}

func (k *Key) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&k.Columns); err != nil {
		return xerrors.Errorf("cannot unmarshal Columns: %w", err)
	}
	return nil
}

type Class struct {
	Name      Name               `yaml:"name"`
	DataStore string             `yaml:"datastore"`
	Index     *INDEX             `yaml:"index"`
	ReadOnly  bool               `yaml:"read_only,omitempty"`
	Members   []*Member          `yaml:"members"`
	classMap  *map[string]*Class `yaml:"-"`
}

type Member struct {
	Name        Name         `yaml:"name"`
	Type        *TypeDeclare `yaml:"type,omitempty"`
	Extend      bool         `yaml:"extend,omitempty"`
	Render      *Render      `yaml:"render,omitempty"`
	HasMany     bool         `yaml:"has_many,omitempty"`
	Nullable    bool         `yaml:"nullable,omitempty"`
	Description string       `yaml:"desc,omitempty"`
	Example     interface{}  `yaml:"example,omitempty"`
	Relation    *Relation    `yaml:"relation,omitempty"`
}

type Relation struct {
	Custom   bool `yaml:"custom,omitempty"`
	All      bool `yaml:"all,omitempty"`
	To       Name `yaml:"to"`
	Internal Name `yaml:"internal,omitempty"`
	External Name `yaml:"external,omitempty"`
}

type Render struct {
	IsRender    bool              `yaml:"-"`
	IsInline    bool              `yaml:"-"`
	DefaultName string            `yaml:"-"`
	Names       map[string]string `yaml:"-"`
}

// MarshalYAML output parameter for render property
// (1) if IsRender property is true, output the following
// ```
// render: false
// ```
// (2) if IsInline property is true, output the following
// ```
// render:
//   inline: true
// ```
// (3) if DefaultName property is 'example', output the following
// ```
// render: example
// ```
// (4) if Names property has some keys, output the following
// ```
// render:
//  json: camelName
//  yaml: snake_name
// ```
func (r *Render) MarshalYAML() (interface{}, error) {
	if !r.IsRender {
		return false, nil
	}
	if r.IsInline {
		return map[string]bool{"inline": true}, nil
	}
	if r.DefaultName != "" {
		return r.DefaultName, nil
	}
	if len(r.Names) == 0 {
		return nil, nil
	}
	return r.Names, nil
}

func (r *Render) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r.IsRender = true
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return xerrors.Errorf("cannot unmarshal interface{}: %w", err)
	}
	switch typ := v.(type) {
	case string:
		r.DefaultName = typ
	case bool:
		r.IsRender = typ
	case map[string]interface{}:
		names := map[string]string{}
		for k, v := range typ {
			if k == "inline" {
				r.IsInline = v.(bool)
				continue
			}
			names[k] = v.(string)
		}
		r.Names = names
	}
	return nil
}

type Type struct {
	PackageName  string
	Name         string
	ImportPath   string
	As           string
	IsPrimitive  bool
	DefaultValue interface{}
}

func (d *Type) IsInt() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == IntType.Name || typ == Int8Type.Name ||
		typ == Int16Type.Name || typ == Int32Type.Name || typ == Int64Type.Name
}

func (d *Type) IsUint() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == UintType.Name || typ == Uint8Type.Name ||
		typ == Uint16Type.Name || typ == Uint32Type.Name || typ == Uint64Type.Name
}

func (d *Type) IsFloat() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == Float32Type.Name || typ == Float64Type.Name
}

func (d *Type) IsBool() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == BoolType.Name
}

func (d *Type) IsByte() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == ByteType.Name
}

func (d *Type) IsString() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == StringType.Name
}

func (d *Type) IsComplex() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == Complex64Type.Name || typ == Complex128Type.Name
}

func (d *Type) IsRune() bool {
	typ := d.Name
	if d.As != "" {
		typ = d.As
	}
	return typ == RuneType.Name
}

func (d *Type) IsTime() bool {
	if d.PackageName == "time" && d.Name == "Time" {
		return true
	}
	if d.PackageName == "" && d.Name == "time.Time" {
		return true
	}
	return false
}

type TypeDeclare struct {
	Type        *Type
	IsPointer   bool
	IsSlice     bool
	classMap    *map[string]*Class
	subClassMap *map[string]*Class
}

func (d *TypeDeclare) IsCustomPrimitiveType() bool {
	return d.Type.As != ""
}

func (d *TypeDeclare) Name() string {
	return d.Type.Name
}

func (d *TypeDeclare) DefaultValue() interface{} {
	if d.IsSchemaClass() {
		return "$default"
	}
	return d.Type.DefaultValue
}

func (d *TypeDeclare) Class() *Class {
	typeName := Name(d.Name()).SnakeName()
	if class, exists := (*d.classMap)[typeName]; exists {
		return class
	}
	if d.subClassMap == nil {
		return nil
	}
	if class, exists := (*d.subClassMap)[typeName]; exists {
		return class
	}
	return nil
}

func (d *TypeDeclare) IsSchemaClass() bool {
	typeName := Name(d.Name()).SnakeName()
	if _, exists := (*d.classMap)[typeName]; exists {
		return true
	}
	return false
}

func (d *TypeDeclare) FormatName(importList ImportList) string {
	if d.Name() == "interface{}" {
		return d.Name()
	}
	return fmt.Sprintf("%#v", d.Code(importList))
}

func (d *TypeDeclare) CollectionName(importList ImportList) string {
	if d.IsSchemaClass() {
		return plural.Plural(d.Name())
	}
	return fmt.Sprintf("[]%#v", d.Code(importList))
}

func (d *TypeDeclare) CodePackage(pkg string, importList ImportList) code.Code {
	if d.Type.PackageName == pkg {
		return d.CodeWithoutPackage(importList)
	}
	return d.Code(importList)
}

func (d *TypeDeclare) CodeWithoutPackage(importList ImportList) code.Code {
	c := code.Id("")
	if d.IsPointer {
		c = c.Add(code.Op("*"))
	}
	if d.IsSlice {
		c = c.Index()
	}
	return c.Id(d.Type.Name)
}

func (d *TypeDeclare) Code(importList ImportList) code.Code {
	c := code.Id("")
	if d.IsPointer {
		c = c.Add(code.Op("*"))
	}
	if d.IsSlice {
		c = c.Index()
	}
	if d.Type.PackageName != "" {
		c = c.Qual(importList.Package(d.Type.PackageName), d.Type.Name)
	} else {
		c = c.Id(d.Type.Name)
	}
	return c
}

func TypeDeclareWithName(name string) *TypeDeclare {
	return &TypeDeclare{
		Type: &Type{
			Name: name,
		},
	}
}

func TypeDeclareWithType(typ *Type) *TypeDeclare {
	return &TypeDeclare{
		Type: typ,
	}
}

func (d *TypeDeclare) MarshalYAML() (interface{}, error) {
	if d.Type.ImportPath != "" || d.Type.PackageName != "" {
		return struct {
			Import      string `yaml:"import,omitempty"`
			PackageName string `yaml:"package_name,omitempty"`
			Name        string `yaml:"name"`
			IsPointer   bool   `yaml:"is_pointer,omitempty"`
		}{
			Import:      d.Type.ImportPath,
			PackageName: d.Type.PackageName,
			Name:        d.Type.Name,
			IsPointer:   d.IsPointer,
		}, nil
	}
	return d.Type.Name, nil
}

func (d *TypeDeclare) ValueToCode(v interface{}) code.Code {
	switch vv := v.(type) {
	case uint64:
		switch {
		case d.Type.IsBool():
			return code.Lit(vv == 0)
		case d.Type.IsString():
			return code.Lit(fmt.Sprint(vv))
		default:
			return code.Lit(int(vv))
		}
	case int64:
		switch {
		case d.Type.IsBool():
			return code.Lit(vv == 0)
		case d.Type.IsString():
			return code.Lit(fmt.Sprint(vv))
		default:
			return code.Lit(int(vv))
		}
	case string:
		switch {
		case d.Type.IsBool():
			bval, _ := strconv.ParseBool(vv)
			return code.Lit(bval)
		case d.Type.IsInt(), d.Type.IsUint():
			ival, _ := strconv.Atoi(vv)
			return code.Lit(ival)
		case d.Type.IsString():
			return code.Lit(vv)
		}
	}
	return code.Lit(v)
}

func (d *TypeDeclare) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return xerrors.Errorf("cannot unmarshal interface{}: %w", err)
	}
	d.Type = &Type{}
	if name, ok := v.(string); ok {
		d.Type.Name = name
	} else {
		for k, v := range v.(map[string]interface{}) {
			switch k {
			case "import":
				d.Type.ImportPath = v.(string)
			case "package_name":
				d.Type.PackageName = v.(string)
			case "name":
				d.Type.Name = v.(string)
			case "is_pointer":
				d.IsPointer = v.(bool)
			}
		}
	}
	if typ, exists := PrimitiveTypes[d.Type.Name]; exists {
		d.Type.PackageName = typ.PackageName
		d.Type.As = typ.As
		d.Type.IsPrimitive = true
		d.Type.DefaultValue = typ.DefaultValue
	} else if d.Type.Name == "Time" && d.Type.PackageName == "time" {
		if d.IsPointer {
			d.Type.DefaultValue = nil
		} else {
			d.Type.DefaultValue = time.Time{}
		}
	}
	return nil
}

var (
	IntType        = &Type{Name: "int", IsPrimitive: true, DefaultValue: int(0)}
	Int8Type       = &Type{Name: "int8", IsPrimitive: true, DefaultValue: int8(0)}
	Int16Type      = &Type{Name: "int16", IsPrimitive: true, DefaultValue: int16(0)}
	Int32Type      = &Type{Name: "int32", IsPrimitive: true, DefaultValue: int32(0)}
	Int64Type      = &Type{Name: "int64", IsPrimitive: true, DefaultValue: int64(0)}
	UintType       = &Type{Name: "uint", IsPrimitive: true, DefaultValue: uint(0)}
	Uint8Type      = &Type{Name: "uint8", IsPrimitive: true, DefaultValue: uint8(0)}
	Uint16Type     = &Type{Name: "uint16", IsPrimitive: true, DefaultValue: uint16(0)}
	Uint32Type     = &Type{Name: "uint32", IsPrimitive: true, DefaultValue: uint32(0)}
	Uint64Type     = &Type{Name: "uint64", IsPrimitive: true, DefaultValue: uint64(0)}
	Float32Type    = &Type{Name: "float32", IsPrimitive: true, DefaultValue: float32(0)}
	Float64Type    = &Type{Name: "float64", IsPrimitive: true, DefaultValue: float64(0)}
	BoolType       = &Type{Name: "bool", IsPrimitive: true, DefaultValue: false}
	StringType     = &Type{Name: "string", IsPrimitive: true, DefaultValue: ""}
	Complex64Type  = &Type{Name: "complex64", IsPrimitive: true, DefaultValue: 0}
	Complex128Type = &Type{Name: "complex128", IsPrimitive: true, DefaultValue: 0}
	ByteType       = &Type{Name: "byte", IsPrimitive: true, DefaultValue: 0}
	RuneType       = &Type{Name: "rune", IsPrimitive: true, DefaultValue: ""}
	ErrorType      = &Type{Name: "error"}
	PrimitiveTypes = map[string]*Type{
		IntType.Name:        IntType,
		Int8Type.Name:       Int8Type,
		Int16Type.Name:      Int16Type,
		Int32Type.Name:      Int32Type,
		Int64Type.Name:      Int64Type,
		UintType.Name:       UintType,
		Uint8Type.Name:      Uint8Type,
		Uint16Type.Name:     Uint16Type,
		Uint32Type.Name:     Uint32Type,
		Uint64Type.Name:     Uint64Type,
		Float32Type.Name:    Float32Type,
		Float64Type.Name:    Float64Type,
		BoolType.Name:       BoolType,
		StringType.Name:     StringType,
		Complex64Type.Name:  Complex64Type,
		Complex128Type.Name: Complex128Type,
		ByteType.Name:       ByteType,
		RuneType.Name:       RuneType,
	}
)

type Members []*Member

func (m Members) JoinedName() string {
	names := []string{}
	for _, member := range m {
		names = append(names, member.Name.SnakeName())
	}
	return strings.Join(names, ":")
}

func (m Members) Names() []Name {
	names := []Name{}
	for _, member := range m {
		names = append(names, member.Name)
	}
	return names
}

func (c *Class) FileName() string {
	return c.Name.SnakeName()
}

func (c *Class) ExtendMembers() Members {
	members := Members{}
	for _, member := range c.Members {
		if member.Relation != nil {
			continue
		}
		if member.Extend {
			members = append(members, member)
		}
	}
	return members
}

func (c *Class) RelationMembers() Members {
	members := Members{}
	for _, member := range c.Members {
		if member.Relation != nil {
			members = append(members, member)
		}
	}
	return members
}

func (c *Class) DependencyMembers() []*Member {
	members := []*Member{}
	for _, member := range c.RelationMembers() {
		relation := member.Relation
		if relation.Custom || relation.All {
			continue
		}
		class := member.Type.Class()
		if class == nil {
			continue
		}
		members = append(members, member)
	}
	return members
}

func (c *Class) DependencyClasses() []*Class {
	classes := []*Class{}
	for _, member := range c.DependencyMembers() {
		classes = append(classes, member.Type.Class())
	}
	return classes
}

func (c *Class) dependencyClassesRecursive() []*Class {
	classes := []*Class{}
	for _, class := range c.DependencyClasses() {
		classes = append(classes, class)
		classes = append(classes, class.dependencyClassesRecursive()...)
	}
	return classes
}

func (c *Class) AllDependencyClasses() []*Class {
	return c.dependencyClassesRecursive()
}

func (c *Class) MarshalRelation() ([]byte, error) {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return nil, xerrors.Errorf("cannot marshal %s class: %w", c.Name.SnakeName(), err)
	}
	return bytes, nil
}

func (c *Class) PrimaryKey() *Member {
	return c.MemberByName(c.Index.PrimaryKey)
}

func (c *Class) UniqueKeys() []Members {
	uniqueKeys := []Members{}
	for _, uniqueKey := range c.Index.UniqueKeys {
		members := Members{}
		for _, column := range uniqueKey.Columns {
			members = append(members, c.MemberByName(column))
		}
		uniqueKeys = append(uniqueKeys, members)
	}
	return uniqueKeys
}

func (c *Class) Keys() []Members {
	keys := []Members{}
	for _, key := range c.Index.Keys {
		members := Members{}
		for _, column := range key.Columns {
			members = append(members, c.MemberByName(column))
		}
		keys = append(keys, members)
	}
	return keys
}

func (c *Class) MemberByName(name string) *Member {
	for _, member := range c.Members {
		if member.Name.SnakeName() == name {
			return member
		}
	}
	return nil
}

func (c *Class) SetClassMap(classMap *map[string]*Class) {
	c.classMap = classMap
	for _, member := range c.Members {
		if member.Type == nil {
			member.Type = &TypeDeclare{Type: &Type{}}
		}
		member.Type.classMap = classMap
	}
}

func (c *Class) SetSubClassMap(subClassMap *map[string]*Class) {
	for _, member := range c.Members {
		if member.Type == nil {
			member.Type = &TypeDeclare{Type: &Type{}}
		}
		member.Type.subClassMap = subClassMap
	}
}

func (c *Class) ResolveTypeReference() error {
	for _, member := range c.Members {
		if member.Type.Name() != "" {
			continue
		}
		if member.Relation == nil {
			continue
		}
		class, exists := (*member.Type.classMap)[member.Relation.To.SnakeName()]
		if !exists {
			return xerrors.Errorf("not found class by %s", Name(member.Type.Name()).SnakeName())
		}
		member.Type.Type.Name = class.Name.CamelName()
	}
	return nil
}

func (c *Class) Merge(schema *Class) {
	c.Index = schema.Index
	schemaMemberMap := map[string]*Member{}
	for _, member := range schema.Members {
		schemaMemberMap[member.Name.SnakeName()] = member
	}

	memberNameMap := map[string]struct{}{}
	extendMembers := []*Member{}
	mergedMembers := []*Member{}
	for _, member := range c.Members {
		if member.Extend {
			extendMembers = append(extendMembers, member)
			continue
		}
		if _, exists := schemaMemberMap[member.Name.SnakeName()]; !exists {
			// remove member
			continue
		}
		mergedMembers = append(mergedMembers, member)
		memberNameMap[member.Name.SnakeName()] = struct{}{}
	}
	for _, member := range schema.Members {
		if _, exists := memberNameMap[member.Name.SnakeName()]; exists {
			continue
		}

		mergedMembers = append(mergedMembers, member)
	}
	mergedMembers = append(mergedMembers, extendMembers...)
	c.Members = mergedMembers
}

func (c *Class) TestData() *TestData {
	defaultObject := c.DefaultTestObject()
	return &TestData{
		Single: TestObjectMap{
			"default": defaultObject,
		},
		Collection: map[string][]*TestObject{
			"defaults": {defaultObject},
		},
	}
}

func (c *Class) DefaultTestObject() *TestObject {
	mapValue := map[string]interface{}{}
	for _, member := range c.Members {
		name := member.Name.SnakeName()
		example := member.Example
		if example == nil {
			mapValue[name] = member.Type.DefaultValue()
		} else {
			mapValue[name] = example
		}
	}
	return &TestObject{
		MapValue: mapValue,
	}
}

func (m *Member) RenderProtocols() []string {
	protocols := []string{"json"}
	if m.Render == nil {
		return protocols
	}
	for proto := range m.Render.Names {
		if proto == "json" {
			continue
		}
		protocols = append(protocols, proto)
	}
	return protocols
}

func (m *Member) RenderNameByProtocol(requiredProtocol string) string {
	if m.Render != nil {
		if !m.Render.IsRender {
			return "-"
		}
		for proto, name := range m.Render.Names {
			if proto == requiredProtocol {
				return name
			}
		}
		if m.Render.DefaultName != "" {
			return m.Render.DefaultName
		}
	}
	return m.Name.CamelLowerName()
}

func (m *Member) CamelType() string {
	return strcase.ToCamel(m.Type.Type.Name)
}

func (m *Member) IsCollectionType() bool {
	if m.HasMany {
		return true
	}
	if m.Relation != nil {
		if m.Relation.All {
			return true
		}
	}
	return false
}

func (m *Member) CollectionName() Name {
	if m.IsCollectionType() {
		return m.Name
	}
	return Name(m.Name.PluralSnakeName())
}

func (m *Member) ModelCollectionTypeName(importList ImportList) string {
	if m.IsCollectionType() {
		if m.Type.IsSchemaClass() {
			return fmt.Sprintf("%sCollection", m.Type.CollectionName(importList))
		}
	} else {
		if m.Type.IsSchemaClass() {
			return fmt.Sprintf("*%s", m.Type.CollectionName(importList))
		}
	}
	return m.Type.CollectionName(importList)
}

func (c *Class) Validate() error {
	if c.Name.SnakeName() == "" {
		return xerrors.New("undefined class name. required name property")
	}
	for idx, member := range c.Members {
		if err := member.Validate(); err != nil {
			return xerrors.Errorf("invalid the %d member: %w", idx, err)
		}
	}
	return nil
}

func (m *Member) Validate() error {
	if m.Name.SnakeName() == "" {
		return xerrors.New("undefined member name. required name property")
	}
	if m.Relation != nil {
		if err := m.Relation.Validate(); err != nil {
			return xerrors.Errorf("invalid %s member's relation: %w", m.Name.SnakeName(), err)
		}
	}
	if m.Type == nil && m.Relation == nil {
		return xerrors.Errorf("undefined %s member type. required type property", m.Name.SnakeName())
	}
	return nil
}

func (r *Relation) Validate() error {
	if r.To.SnakeName() == "" {
		return xerrors.New("must be declared 'to' property in relation")
	}
	if r.Custom || r.All {
		if r.Internal != "" {
			return xerrors.New("if defined custom property, ignored internal property")
		}
		if r.External != "" {
			return xerrors.New("if defined custom property, ignored external property")
		}
	} else {
		if r.Internal == "" {
			return xerrors.New("undefined internal member in relation")
		}
		if r.External == "" {
			return xerrors.New("undefined external member in relation")
		}
	}
	return nil
}

type API struct {
	Name        Name      `yaml:"name"`
	Description string    `yaml:"desc"`
	URI         string    `yaml:"uri"`
	Method      string    `yaml:"method"`
	Request     *Request  `yaml:"request"`
	Response    *Response `yaml:"response"`
}

type Request struct {
	Params RequestParams `yaml:"parameters"`
}

type RequestParams []*RequestParam

func (r RequestParams) HasInBodyParam() bool {
	for _, p := range r {
		if p.In == InBody {
			return true
		}
	}
	return false
}

func (r RequestParams) HasInQueryParam() bool {
	for _, p := range r {
		if p.In == InQuery {
			return true
		}
	}
	return false
}

type InType string

const (
	InHeader InType = "header"
	InPath          = "path"
	InQuery         = "query"
	InBody          = "body"
)

type RequestParam struct {
	Name     Name        `yaml:"name"`
	Type     string      `yaml:"type"`
	In       InType      `yaml:"in"`
	Desc     string      `yaml:"desc"`
	Example  interface{} `yaml:"example"`
	Render   *Render     `yaml:"render,omitempty"`
	Required bool        `yaml:"required"`
}

func (p *RequestParam) RenderName() string {
	if p.Render == nil {
		return p.Name.CamelLowerName()
	}
	return p.Render.DefaultName
}

type APIClass struct {
	Class      `yaml:",inline"`
	Include    []*Include `yaml:"include"`
	IncludeAll bool       `yaml:"include_all"`
}

type Include struct {
	Name    Name       `yaml:"name"`
	Only    []Name     `yaml:"only"`
	Except  []Name     `yaml:"except"`
	Include []*Include `yaml:"include"`
}

func (i *Include) BuilderCode(h *APIResponseHelper) code.Code {
	onlyNames := []code.Code{}
	for _, only := range i.Only {
		onlyNames = append(onlyNames, code.Lit(only.SnakeName()))
	}
	exceptNames := []code.Code{}
	for _, except := range i.Except {
		exceptNames = append(exceptNames, code.Lit(except.SnakeName()))
	}
	block := []code.Code{}
	if len(onlyNames) > 0 {
		block = append(block, code.Id("optBuilder").Dot("Only").Call(onlyNames...))
	}
	if len(exceptNames) > 0 {
		block = append(block, code.Id("optBuilder").Dot("Except").Call(exceptNames...))
	}
	for _, include := range i.Include {
		block = append(block, include.BuilderCode(h))
	}
	return code.Id("optBuilder").Dot("IncludeWithCallback").Call(
		code.Lit(i.Name.SnakeName()),
		code.Func().Params(code.Id("optBuilder").Op("*").Qual(h.Package("model"), "RenderOptionBuilder")).Block(
			block...,
		),
	)
}

type Response struct {
	SubTypes []*APIClass
	Type     *APIClass
}

func (r *Response) BuilderCode(h *APIResponseHelper) []code.Code {
	block := []code.Code{
		code.Id("optBuilder").Op(":=").Qual(h.Package("model"), "NewRenderOptionBuilder").Call(),
	}
	if r.Type.IncludeAll {
		block = append(block, code.Id("optBuilder").Dot("IncludeAll").Call())
		return block
	}
	subTypeBuilderCodeMap := map[string][]code.Code{}
	for _, subType := range r.SubTypes {
		codes := []code.Code{}
		for _, include := range subType.Include {
			codes = append(codes, include.BuilderCode(h))
		}
		subTypeBuilderCodeMap[subType.Class.Name.SnakeName()] = codes
	}
	for _, include := range r.Type.Include {
		member := r.Type.MemberByName(include.Name.SnakeName())
		if member == nil {
			continue
		}
		typeName := Name(member.Type.Name()).SnakeName()
		if codes, exists := subTypeBuilderCodeMap[typeName]; exists {
			block = append(block, code.Id("optBuilder").Dot("IncludeWithCallback").Call(
				code.Lit(include.Name.SnakeName()),
				code.Func().Params(code.Id("optBuilder").Op("*").Qual(h.Package("model"), "RenderOptionBuilder")).Block(
					codes...,
				),
			))
		} else {
			block = append(block, include.BuilderCode(h))
		}
	}
	return block
}

func (r *Response) classToAPIClass(class *Class) *APIClass {
	for _, subType := range r.SubTypes {
		if subType.Class.Name.SnakeName() == class.Name.SnakeName() {
			return subType
		}
	}
	return nil
}

func (r *Response) renderMember(parentMember *Member, includes []*Include) map[string]interface{} {
	class := parentMember.Type.Class()
	includes = r.includes(class, includes)
	rendered := map[string]interface{}{}
	for _, include := range includes {
		if len(include.Only) > 0 {
			for _, only := range include.Only {
				member := class.MemberByName(only.SnakeName())
				if member == nil {
					continue
				}
				key := member.RenderNameByProtocol("json")
				if member.Example != nil {
					rendered[key] = member.Example
				} else {
					rendered[key] = member.Type.DefaultValue()
				}
			}
		} else if len(include.Except) > 0 {
			exceptMap := map[string]struct{}{}
			for _, except := range include.Except {
				exceptMap[except.SnakeName()] = struct{}{}
			}
			for _, member := range class.Members {
				if member.Relation != nil {
					continue
				}
				if member.Render != nil && !member.Render.IsRender {
					continue
				}
				if _, exists := exceptMap[member.Name.SnakeName()]; exists {
					continue
				}
				key := member.RenderNameByProtocol("json")
				if member.Example != nil {
					rendered[key] = member.Example
				} else {
					rendered[key] = member.Type.DefaultValue()
				}
			}
		}
		if len(include.Include) > 0 {
			value := r.renderIncludes(class, r.includes(class, include.Include))
			member := class.MemberByName(include.Name.SnakeName())
			if member != nil && member.Render != nil && member.Render.IsInline {
				for _, v := range value {
					for k, vv := range v.(map[string]interface{}) {
						rendered[k] = vv
					}
				}
			} else {
				for k, v := range value {
					rendered[k] = v
				}
			}
		}
	}
	return rendered
}

func (r *Response) includes(class *Class, includes []*Include) []*Include {
	apiClass := r.classToAPIClass(class)
	if apiClass != nil {
		return apiClass.Include
	}
	return includes
}

func (r *Response) renderIncludes(class *Class, includes []*Include) map[string]interface{} {
	rendered := map[string]interface{}{}
	for _, include := range includes {
		member := class.MemberByName(include.Name.SnakeName())
		if member == nil {
			continue
		}
		value := r.renderMember(member, r.includes(class, []*Include{include}))
		key := member.RenderNameByProtocol("json")
		if member.HasMany {
			rendered[key] = []interface{}{value}
		} else {
			rendered[key] = value
		}
	}
	return rendered
}

func (r *Response) renderAll(class *Class) map[string]interface{} {
	rendered := map[string]interface{}{}
	for _, member := range class.Members {
		key := member.RenderNameByProtocol("json")
		subClass := member.Type.Class()
		if subClass != nil {
			rendered[key] = r.renderAll(subClass)
		} else if member.Example != nil {
			rendered[key] = member.Example
		} else {
			rendered[key] = member.Type.DefaultValue()
		}
	}
	return rendered
}

func (r *Response) Render() interface{} {
	class := &r.Type.Class
	if r.Type.IncludeAll {
		return r.renderAll(class)
	}
	return r.renderIncludes(class, r.includes(class, r.Type.Include))
}

type Attribute struct {
	Name    string
	Type    string
	Desc    string
	Example interface{}
}

func (r *Response) attributes(class *Class, prefix string) []*Attribute {
	attrs := []*Attribute{}
	for _, member := range class.Members {
		if member.Render != nil && !member.Render.IsRender {
			continue
		}
		var name string
		if prefix != "" {
			name = fmt.Sprintf("%s.", prefix)
		}
		name += member.RenderNameByProtocol("json")
		if member.HasMany {
			name += "[0]"
		}
		subClass := member.Type.Class()
		if subClass != nil {
			attrs = append(attrs, r.attributes(subClass, name)...)
		} else if member.Example != nil {
			attrs = append(attrs, &Attribute{
				Name:    name,
				Type:    member.Type.Name(),
				Desc:    member.Description,
				Example: member.Example,
			})
		} else {
			attrs = append(attrs, &Attribute{
				Name:    name,
				Type:    member.Type.Name(),
				Desc:    member.Description,
				Example: member.Type.DefaultValue(),
			})
		}
	}
	return attrs
}

func (r *Response) Attributes() []*Attribute {
	return r.attributes(&r.Type.Class, "")
}

func (r *Response) RenderJSON() string {
	renderedMap := r.Render()
	bytes, err := json.MarshalIndent(renderedMap, "", strings.Repeat(" ", 4))
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (r *Response) ResolveClassReference(classMap *map[string]*Class, subClassMap *map[string]*Class) {
	defer func() {
		r.Type.SetClassMap(classMap)
		r.Type.SetSubClassMap(subClassMap)
	}()
	if len(r.Type.Members) > 0 {
		return
	}
	r.Type.Class = *(*classMap)[r.Type.Name.SnakeName()]
}

func (r *Response) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var res struct {
		SubTypes []*APIClass `yaml:"subtypes"`
		Type     *APIClass   `yaml:"type"`
	}
	if err := unmarshal(&res); err != nil {
		return xerrors.Errorf("cannot unmarshal Response: %w", err)
	}
	r.SubTypes = res.SubTypes
	r.Type = res.Type
	subtypeMap := map[string]struct{}{}
	for _, subtype := range r.SubTypes {
		subtypeMap[subtype.Name.SnakeName()] = struct{}{}
	}
	for _, member := range r.Type.Members {
		name := Name(member.Type.Name())
		if _, exists := subtypeMap[name.SnakeName()]; exists {
			member.Type.Type.Name = name.CamelName()
		}
	}
	return nil
}

type TestObjectMap map[string]*TestObject
type TestObjectDecl struct {
	Name        string `yaml:"-"`
	*TestObject `yaml:",inline,anchor"`
}

func (m TestObjectMap) MarshalYAML() (interface{}, error) {
	newMap := map[string]*TestObjectDecl{}
	for k, v := range m {
		newMap[k] = &TestObjectDecl{Name: k, TestObject: v}
	}
	return newMap, nil
}

type TestData struct {
	Single     TestObjectMap            `yaml:"single"`
	Collection map[string][]*TestObject `yaml:"collection"`
}

type TestObject struct {
	*TestObject `yaml:",omitempty,inline,alias"`
	MapValue    map[string]interface{} `yaml:",omitempty,inline"`
}

func (o *TestObject) MergedMapValue() map[string]interface{} {
	mapValue := map[string]interface{}{}
	if o.TestObject != nil {
		for k, v := range o.TestObject.MergedMapValue() {
			mapValue[k] = v
		}
	}
	for k, v := range o.MapValue {
		mapValue[k] = v
	}
	return mapValue
}

func (d *TestData) MergeDefault(class *Class) {
	data := class.TestData()
	defaultMapValue := data.Single["default"].MapValue
	for k := range d.Single["default"].MapValue {
		if _, exists := defaultMapValue[k]; exists {
			continue
		}
		if member := class.MemberByName(k); member == nil {
			// remove value
			delete(d.Single["default"].MapValue, k)
		}
	}
	for k, v := range defaultMapValue {
		if _, exists := d.Single["default"].MapValue[k]; exists {
			continue
		}
		// add value
		d.Single["default"].MapValue[k] = v
	}
}

type StructFieldList map[string]*ValueDeclare

type ImportDeclare struct {
	Name string
	Path string
}

func (d *ImportDeclare) GetName() string {
	return d.Name
}

func (d *ImportDeclare) GetPath() string {
	return d.Path
}

type ImportList map[string]*ImportDeclare

func (l ImportList) Each(f func(imp code.Import)) {
	for _, imp := range l {
		f(imp)
	}
}

func (l ImportList) Package(name string) string {
	if decl, exists := l[name]; exists {
		return decl.Path
	}
	return name
}

func DefaultImportList(modulePath string, ctxImportPath string) ImportList {
	importList := ImportList{}
	for _, decl := range []*ImportDeclare{
		{
			Path: fmt.Sprintf("%s/entity", modulePath),
			Name: "entity",
		},
		{
			Path: fmt.Sprintf("%s/dao", modulePath),
			Name: "dao",
		},
		{
			Path: fmt.Sprintf("%s/model", modulePath),
			Name: "model",
		},
		{
			Path: fmt.Sprintf("%s/repository", modulePath),
			Name: "repository",
		},
		{
			Path: "golang.org/x/xerrors",
			Name: "xerrors",
		},
		{
			Path: "database/sql",
			Name: "sql",
		},
		{
			Path: "encoding/json",
			Name: "json",
		},
	} {
		importList[decl.Name] = decl
	}
	if ctxImportPath != "" {
		importList["context"] = &ImportDeclare{
			Path: ctxImportPath,
			Name: "context",
		}
	}
	return importList
}

type ValueDeclare struct {
	Name string
	Type *TypeDeclare
}

func (d *ValueDeclare) Code(importList ImportList) code.Code {
	c := code.Id(d.Name)
	c.Add(d.Type.Code(importList))
	return c
}

func (d *ValueDeclare) Interface(importList ImportList) code.Code {
	return d.Type.Code(importList)
}

type ValueDeclares []*ValueDeclare

func (d ValueDeclares) Code(importList ImportList) []code.Code {
	c := []code.Code{}
	for _, v := range d {
		c = append(c, v.Code(importList))
	}
	return c
}

func (d ValueDeclares) Interface(importList ImportList) []code.Code {
	c := []code.Code{}
	for _, v := range d {
		c = append(c, v.Interface(importList))
	}
	return c
}

type ConstructorDeclare struct {
	Class      *Class
	MethodName string
	Args       ValueDeclares
	Return     ValueDeclares
	ImportList ImportList
}

func (d *ConstructorDeclare) MethodInterface(importList ImportList) *code.Statement {
	return code.Func().Id(d.MethodName).Params(d.Args.Code(importList)...).
		Parens(code.List(d.Return.Code(importList)...))
}

func (d *ConstructorDeclare) Package(name string) string {
	return d.ImportList.Package(name)
}

type MethodDeclare struct {
	Class                *Class
	ReceiverName         string
	ReceiverClassName    string
	ImportList           ImportList
	MethodName           string
	ArgMembers           Members
	Args                 ValueDeclares
	Return               ValueDeclares
	IsNotPointerReceiver bool
}

func (d *MethodDeclare) Interface(importList ImportList) *code.Statement {
	returnCodes := d.Return.Interface(importList)
	def := code.Id(d.MethodName).Params(d.Args.Interface(importList)...)
	if len(returnCodes) > 1 {
		return def.Parens(code.List(returnCodes...))
	} else if len(returnCodes) > 0 {
		c := returnCodes[0]
		return def.Id(fmt.Sprintf("%#v", c))
	}
	return def
}

func (d *MethodDeclare) MethodInterface(importList ImportList) *code.Statement {
	receiver := code.Id(d.ReceiverName)
	if !d.IsNotPointerReceiver {
		receiver = receiver.Op("*")
	}
	receiver = receiver.Id(d.ReceiverClassName)
	return code.Func().Params(receiver).Id(d.MethodName).
		Params(d.Args.Code(importList)...).
		Parens(code.List(d.Return.Code(importList)...))
}

type Method struct {
	Decl *MethodDeclare
	Body []code.Code
}

func (m *Method) Generate(importList ImportList) code.Code {
	return m.Decl.MethodInterface(importList).Block(m.Body...)
}

type Methods []*Method

func (m Methods) Generate(importList ImportList) []code.Code {
	codes := []code.Code{}
	for _, mtd := range m {
		codes = append(codes, mtd.Generate(importList))
	}
	return codes
}
