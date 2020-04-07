package class_test

import (
	"path/filepath"
	"testing"

	"go.knocknote.io/eevee/class"
	"go.knocknote.io/eevee/config"
	"go.knocknote.io/eevee/schema"
	"go.knocknote.io/eevee/types"
)

func TestClassWrite(t *testing.T) {
	schemata, err := schema.NewReader().SchemaFromPath(filepath.Join("testdata", "schema"))
	if err != nil {
		t.Fatalf("%+v", err)
	}
	writer, err := class.NewWriter(filepath.Join("testdata", "class"))
	if err != nil {
		t.Fatalf("%+v", err)
	}
	cfg := &config.Config{}
	for _, schema := range schemata {
		if err := writer.Write(cfg, schema.ToClass()); err != nil {
			t.Fatalf("%+v", err)
		}
	}
}

func TestClassRead(t *testing.T) {
	cfg := &config.Config{
		ClassPath: filepath.Join("testdata", "class"),
	}
	classes, err := class.NewReader().ClassByConfig(cfg)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	var userClass *types.Class
	for _, class := range classes {
		if class.Name.SnakeName() == "user" {
			userClass = class
			break
		}
	}
	if userClass == nil {
		t.Fatal("cannot load user class")
	}
	t.Run("userFields", func(t *testing.T) {
		member := userClass.MemberByName("user_fields")
		if member == nil {
			t.Fatal("cannot setup member")
		}
		if !member.HasMany {
			t.Fatal("cannot setup userFields.HasMany")
		}
		if !member.Extend {
			t.Fatal("cannot setup userFields.Extend")
		}
		if member.Relation == nil {
			t.Fatal("cannot setup userFields.Relation")
		}
		if member.Relation.To.SnakeName() != "user_field" {
			t.Fatal("cannot setup userFields.Relation.To", member.Relation.To.SnakeName())
		}
		if member.Relation.Internal.SnakeName() != "id" {
			t.Fatal("cannot setup userFields.Relation.Internal")
		}
		if member.Relation.External.SnakeName() != "user_id" {
			t.Fatal("cannot setup userFields.Relation.External")
		}
		class := member.Type.Class()
		if class == nil {
			t.Fatal("cannot resolve class reference")
		}
		if class.Name.SnakeName() != "user_field" {
			t.Fatal("cannot resolve class reference")
		}
	})
	skillMember := userClass.MemberByName("skill")
	if skillMember == nil {
		t.Fatal("cannot setup skill member")
	}
	if !skillMember.Render.IsInline {
		t.Fatal("cannot setup skill skill.Render.IsInline")
	}
	groupMember := userClass.MemberByName("group")
	if groupMember == nil {
		t.Fatal("cannot setup group member")
	}
	if groupMember.Render.Names["json"] != "group" {
		t.Fatal("cannot setup skill skill.Render.Names")
	}
	worldMember := userClass.MemberByName("world")
	if worldMember == nil {
		t.Fatal("cannot setup world member")
	}
	if worldMember.Render.IsRender {
		t.Fatal("cannot setup skill skill.Render.IsRender")
	}
}
