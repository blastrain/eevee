package graph

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"go.knocknote.io/eevee/config"
	_ "go.knocknote.io/eevee/static"
	"go.knocknote.io/eevee/types"
	"github.com/rakyll/statik/fs"
	"golang.org/x/xerrors"
)

type Graph struct {
	GraphClasses []*GraphClass
	Classes      []*types.Class
}

type GraphClass struct {
	*types.Class
	Dependencies      []*DependencyClass
	ReferencedClasses []*types.Class
}

type DependencyClass struct {
	*types.Class
	Members []*DependencyMember
}

type DependencyMember struct {
	*types.Member
	IsReferencedAsExternalMember bool
}

func dependencyClasses(class *types.Class) []*DependencyClass {
	classes := []*DependencyClass{}
	for _, member := range class.DependencyMembers() {
		externalName := member.Relation.External.SnakeName()
		members := []*DependencyMember{}
		depClass := member.Type.Class()
		for _, depMember := range depClass.Members {
			isExternalMember := false
			if externalName == depMember.Name.SnakeName() {
				isExternalMember = true
			}
			members = append(members, &DependencyMember{
				Member:                       depMember,
				IsReferencedAsExternalMember: isExternalMember,
			})
		}
		classes = append(classes, &DependencyClass{
			Class:   depClass,
			Members: members,
		})
		classes = append(classes, dependencyClasses(depClass)...)
	}
	return classes
}

func Generate(cfg *config.Config, classes []*types.Class) error {
	graphClasses := []*GraphClass{}
	referencedClassMap := map[string][]*types.Class{}
	for _, class := range classes {
		className := class.Name.SnakeName()
		depClasses := dependencyClasses(class)
		if _, exists := referencedClassMap[className]; !exists {
			referencedClassMap[className] = []*types.Class{}
		}
		for _, depClass := range class.DependencyClasses() {
			depClassName := depClass.Name.SnakeName()
			if _, exists := referencedClassMap[depClassName]; !exists {
				referencedClassMap[depClassName] = []*types.Class{}
			}
			referencedClassMap[depClassName] = append(referencedClassMap[depClassName], class)
		}
		graphClasses = append(graphClasses, &GraphClass{
			Class:        class,
			Dependencies: depClasses,
		})
	}
	for _, graphClass := range graphClasses {
		graphClass.ReferencedClasses = referencedClassMap[graphClass.Name.SnakeName()]
	}
	statikFS, err := fs.New()
	if err != nil {
		return xerrors.Errorf("failed to create statik fs: %w", err)
	}
	graphFile, err := statikFS.Open("/graph.tmpl")
	if err != nil {
		return xerrors.Errorf("failed to open graph.tmpl: %w", err)
	}
	graphBytes, err := ioutil.ReadAll(graphFile)
	if err != nil {
		return xerrors.Errorf("failed to read from graph.tmpl: %w", err)
	}
	tmpl, err := template.New("").Parse(string(graphBytes))
	if err != nil {
		return xerrors.Errorf("failed to parse html template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, &Graph{
		GraphClasses: graphClasses,
		Classes:      classes,
	}); err != nil {
		return xerrors.Errorf("failed to execute template: %w", err)
	}
	if err := os.MkdirAll(cfg.GraphPath, 0755); err != nil {
		return xerrors.Errorf("failed to create directory %s: %w", cfg.GraphPath, err)
	}
	graphPath := filepath.Join(cfg.GraphPath, "index.html")
	if err := ioutil.WriteFile(graphPath, buf.Bytes(), 0644); err != nil {
		return xerrors.Errorf("failed to write index.html to %s: %w", graphPath, err)
	}
	vizFile, err := statikFS.Open("/viz.js")
	if err != nil {
		return xerrors.Errorf("failed to open viz.js: %w", err)
	}
	vizBytes, err := ioutil.ReadAll(vizFile)
	if err != nil {
		return xerrors.Errorf("failed to read from viz.js: %w", err)
	}
	vizPath := filepath.Join(cfg.GraphPath, "viz.js")
	if err := ioutil.WriteFile(vizPath, vizBytes, 0644); err != nil {
		return xerrors.Errorf("failed to write viz.js to %s: %w", vizPath, err)
	}
	return nil
}
