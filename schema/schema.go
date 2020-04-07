package schema

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"go.knocknote.io/eevee/plural"
	"go.knocknote.io/eevee/types"
	"github.com/knocknote/vitess-sqlparser/sqlparser"
	"golang.org/x/xerrors"
)

type Schema struct {
	Name    string       `yaml:"name"`
	Index   *types.INDEX `yaml:"index"`
	Columns []*Column    `yaml:"members"`
}

type Column struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Nullable bool   `yaml:"-"`
}

func (c *Column) ToMember() *types.Member {
	var decl *types.TypeDeclare
	if c.Type == "time.Time" {
		decl = types.TypeDeclareWithType(&types.Type{
			PackageName:  "time",
			Name:         "Time",
			ImportPath:   "time",
			DefaultValue: time.Time{},
		})
	} else {
		decl = types.TypeDeclareWithName(c.Type)
	}
	decl.IsPointer = c.Nullable
	return &types.Member{
		Name: types.Name(c.Name),
		Type: decl,
	}
}

func (s *Schema) FileName() string {
	return s.Name
}

func (s *Schema) ToClass() *types.Class {
	members := []*types.Member{}
	for _, column := range s.Columns {
		members = append(members, column.ToMember())
	}
	return &types.Class{
		Name:    types.Name(s.Name),
		Index:   s.Index,
		Members: members,
	}
}

type Reader struct {
	unsignedPattern *regexp.Regexp
	floatPattern    *regexp.Regexp
	bigintPattern   *regexp.Regexp
	charPattern     *regexp.Regexp
	datetimePattern *regexp.Regexp
	intPattern      *regexp.Regexp
	enumPattern     *regexp.Regexp
	textPattern     *regexp.Regexp
}

func NewReader() *Reader {
	return &Reader{
		unsignedPattern: regexp.MustCompile(`UNSIGNED`),
		floatPattern:    regexp.MustCompile(`float`),
		bigintPattern:   regexp.MustCompile(`bigint`),
		charPattern:     regexp.MustCompile(`(var)?char`),
		datetimePattern: regexp.MustCompile(`datetime`),
		intPattern:      regexp.MustCompile(`int`),
		enumPattern:     regexp.MustCompile(`enum`),
		textPattern:     regexp.MustCompile(`text`),
	}
}

func (r *Reader) isStringType(mysqlType string) bool {
	if r.charPattern.MatchString(mysqlType) {
		return true
	}
	if r.enumPattern.MatchString(mysqlType) {
		return true
	}
	if r.textPattern.MatchString(mysqlType) {
		return true
	}
	return false
}

func (r *Reader) isUint64Type(mysqlType string) bool {
	if r.unsignedPattern.MatchString(mysqlType) &&
		r.bigintPattern.MatchString(mysqlType) {
		return true
	}
	return false
}

func (r *Reader) isInt64Type(mysqlType string) bool {
	return r.bigintPattern.MatchString(mysqlType)
}

func (r *Reader) isUint32Type(mysqlType string) bool {
	return r.unsignedPattern.MatchString(mysqlType)
}

func (r *Reader) isIntType(mysqlType string) bool {
	return r.intPattern.MatchString(mysqlType)
}

func (r *Reader) isTimeType(mysqlType string) bool {
	return r.datetimePattern.MatchString(mysqlType)
}

func (r *Reader) isFloat32Type(mysqlType string) bool {
	return r.floatPattern.MatchString(mysqlType)
}

func (r *Reader) convertMySQLTypeToGOType(mysqlType string) string {
	switch {
	case r.isStringType(mysqlType):
		return "string"
	case r.isUint64Type(mysqlType):
		return "uint64"
	case r.isInt64Type(mysqlType):
		return "int64"
	case r.isUint32Type(mysqlType):
		return "uint32"
	case r.isIntType(mysqlType):
		return "int"
	case r.isTimeType(mysqlType):
		return "time.Time"
	case r.isFloat32Type(mysqlType):
		return "float32"
	}
	return mysqlType
}

func (r *Reader) isNullableColumn(column *sqlparser.ColumnDef) bool {
	for _, opt := range column.Options {
		switch opt.Type {
		case sqlparser.ColumnOptionNotNull:
			return false
		case sqlparser.ColumnOptionNull:
			return true
		case sqlparser.ColumnOptionDefaultValue:
			return opt.Value == "NULL"
		}
	}
	return false
}

func (r *Reader) parseSQL(stmt sqlparser.Statement) (*Schema, error) {
	createTable, ok := stmt.(*sqlparser.CreateTable)
	if !ok {
		return nil, xerrors.New("only supported create table")
	}
	tableName := createTable.NewName.Name.String()
	columns := []*Column{}
	columnMap := map[string]*Column{}
	for _, column := range createTable.Columns {
		columnName := column.Name
		columnType := r.convertMySQLTypeToGOType(column.Type)
		column := &Column{
			Name:     columnName,
			Type:     columnType,
			Nullable: r.isNullableColumn(column),
		}
		columnMap[columnName] = column
		columns = append(columns, column)
	}
	index := &types.INDEX{}
	for _, constraint := range createTable.Constraints {
		switch constraint.Type {
		case sqlparser.ConstraintPrimaryKey:
			index.PrimaryKey = constraint.Keys[0].String()
		case sqlparser.ConstraintUniq, sqlparser.ConstraintUniqKey, sqlparser.ConstraintUniqIndex:
			uniqueKey := &types.UniqueKey{
				Columns: []string{},
			}
			for _, key := range constraint.Keys {
				uniqueKey.Columns = append(uniqueKey.Columns, key.String())
			}
			index.UniqueKeys = append(index.UniqueKeys, uniqueKey)
		case sqlparser.ConstraintKey, sqlparser.ConstraintIndex:
			key := &types.Key{
				Columns: []string{},
			}
			for _, k := range constraint.Keys {
				key.Columns = append(key.Columns, k.String())
			}
			index.Keys = append(index.Keys, key)
		}
	}
	return &Schema{
		Name:    plural.Singular(tableName),
		Columns: columns,
		Index:   index,
	}, nil
}

func (r *Reader) readSchema(path string) (*Schema, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, xerrors.Errorf("cannot read file %s: %w", path, err)
	}
	stmt, err := sqlparser.Parse(string(bytes))
	if err != nil {
		return nil, xerrors.Errorf("cannot parse SQL [%s]: %w", string(bytes), err)
	}
	return r.parseSQL(stmt)
}

func (r *Reader) SchemaFromPath(path string) ([]*Schema, error) {
	if filepath.Ext(path) == "sql" {
		schema, err := r.readSchema(path)
		if err != nil {
			return nil, xerrors.Errorf("cannot read schema file %s: %w", path, err)
		}
		return []*Schema{schema}, nil
	}

	schemata := []*Schema{}

	sqlFilePattern := regexp.MustCompile(`\.sql$`)
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !sqlFilePattern.MatchString(path) {
			return nil
		}
		schema, err := r.readSchema(path)
		if err != nil {
			return xerrors.Errorf("cannot read schema: %w", err)
		}
		schemata = append(schemata, schema)
		return nil
	}); err != nil {
		return nil, xerrors.Errorf("interrupt walk in %s: %w", path, err)
	}

	return schemata, nil
}
