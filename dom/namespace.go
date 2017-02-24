package dom

import (
	"strings"
)

// Namespace defines a Namespace entity
type Namespace struct {
	Name         string
	Abbreviation string
}

// SplitFQName splits the name into namespace (if any) and local name
func SplitFQName(fqName string) (string, string) {
	index := strings.Index(fqName, ":")
	if index == -1 {
		return "", fqName
	}
	return fqName[:index], fqName[index+1:]
}
