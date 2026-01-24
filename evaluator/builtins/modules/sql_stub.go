//go:build !sql

package modules

import "gitlab.com/bark-lang/bark/object"

// InitSQL returns empty map when SQL support is not compiled in
func InitSQL() map[string]*object.Builtin {
	return map[string]*object.Builtin{}
}
