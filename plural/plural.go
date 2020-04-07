package plural

import (
	"regexp"

	"github.com/jinzhu/inflection"
)

func Singular(name string) string {
	return inflection.Singular(name)
}

var (
	numPattern = regexp.MustCompile(`([0-9]+)$`)
)

func Plural(name string) string {
	if numPattern.MatchString(name) {
		return name + "s"
	}
	return inflection.Plural(name)
}

func Register(singular, plural string) {
	inflection.AddIrregular(singular, plural)
}
