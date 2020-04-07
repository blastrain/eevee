package types

import (
	"fmt"
	"regexp"

	"github.com/iancoleman/strcase"
	"go.knocknote.io/eevee/plural"
)

type Name string

type specialCase struct {
	lower        string
	upper        string
	startPattern *regexp.Regexp
	midPattern   *regexp.Regexp
	endPattern   *regexp.Regexp
}

var (
	specialCases = []*specialCase{
		{lower: "id", upper: "ID"},
		{lower: "url", upper: "URL"},
		{lower: "os", upper: "OS"},
	}
)

func init() {
	for _, sc := range specialCases {
		sc.startPattern = regexp.MustCompile(fmt.Sprintf(`^%s_`, sc.lower))
		sc.midPattern = regexp.MustCompile(fmt.Sprintf(`_%s_`, sc.lower))
		sc.endPattern = regexp.MustCompile(fmt.Sprintf(`_%s([0-9]?)s?$`, sc.lower))
	}
}

func matchedSpecialCase(s string) *specialCase {
	for _, sc := range specialCases {
		if sc.lower == s {
			return sc
		}
	}
	return nil
}

func matchedPluralSpecialCase(s string) *specialCase {
	for _, sc := range specialCases {
		if sc.lower+"s" == s {
			return sc
		}
	}
	return nil
}

func (n Name) normalize(name string) string {
	s := name
	for _, sc := range specialCases {
		s = sc.startPattern.ReplaceAllString(s, fmt.Sprintf(`^%s_`, sc.upper))
		s = sc.midPattern.ReplaceAllString(s, fmt.Sprintf(`_%s_`, sc.upper))
		s = sc.endPattern.ReplaceAllString(s, fmt.Sprintf(`%s$1`, sc.upper))
	}
	return s
}

func (n Name) normalizePlural(name string) string {
	s := name
	for _, sc := range specialCases {
		s = sc.startPattern.ReplaceAllString(s, fmt.Sprintf(`^%s_`, sc.upper))
		s = sc.midPattern.ReplaceAllString(s, fmt.Sprintf(`_%s_`, sc.upper))
		s = sc.endPattern.ReplaceAllString(s, fmt.Sprintf(`%s${1}s`, sc.upper))
	}
	return s
}

func (n Name) CamelName() string {
	sc := matchedSpecialCase(string(n))
	if sc != nil {
		return sc.upper
	}
	return strcase.ToCamel(n.normalize(string(n)))
}

func (n Name) PluralCamelName() string {
	name := plural.Plural(string(n))
	sc := matchedPluralSpecialCase(string(name))
	if sc != nil {
		return sc.upper + "s"
	}
	return strcase.ToCamel(n.normalizePlural(string(name)))
}

func (n Name) CamelLowerName() string {
	return strcase.ToLowerCamel(n.normalize(string(n)))
}

func (n Name) PluralCamelLowerName() string {
	return strcase.ToLowerCamel(n.normalizePlural(plural.Plural(string(n))))
}

func (n Name) SnakeName() string {
	return strcase.ToSnake(string(n))
}

func (n Name) PluralSnakeName() string {
	return strcase.ToSnake(plural.Plural(string(n)))
}
