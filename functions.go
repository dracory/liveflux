package liveflux

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/dracory/str"
)

// NewID generates a URL-safe random ID string.
func NewID() string {
	id, err := str.RandomFromGamma(12, "123456789bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ")
	if err != nil {
		return str.Random(12)
	}
	return id
}

// DefaultKindFromType derives a sensible default kind from a component's Go type.
// Rules:
// - Use the package name and kebab-cased struct name: "<pkg>.<type-kebab>"
// - If struct name matches package name (case-insensitive), just use the package name.
// Examples:
//
//	package counter, type Counter -> "counter"
//	package users, type UserList -> "users.user-list"
func DefaultKindFromType(c ComponentInterface) string {
	if c == nil {
		return ""
	}

	t := reflect.TypeOf(c)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	pkgPath := t.PkgPath()
	if pkgPath == "" {
		return strings.ToLower(toKebab(t.Name()))
	}

	parts := strings.Split(pkgPath, "/")
	pkg := parts[len(parts)-1]
	typeKebab := toKebab(t.Name())
	if strings.EqualFold(pkg, typeKebab) {
		return strings.ToLower(pkg)
	}

	return strings.ToLower(pkg + "." + typeKebab)
}

// toKebab converts CamelCase or mixed strings to kebab-case.
func toKebab(s string) string {
	var b strings.Builder
	var prevLower bool
	var lastHyphen bool
	for i, r := range s {
		if unicode.IsUpper(r) {
			nextLower := i+1 < len(s) && unicode.IsLower(rune(s[i+1]))
			if i > 0 && (prevLower || nextLower) && !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
			b.WriteRune(unicode.ToLower(r))
			prevLower = false
			lastHyphen = false
		} else if r == '_' || r == ' ' || r == '-' {
			// write a single hyphen for any delimiter, but avoid duplicates
			if !lastHyphen && b.Len() > 0 {
				b.WriteByte('-')
				lastHyphen = true
			}
			prevLower = false
		} else {
			b.WriteRune(r)
			prevLower = true
			lastHyphen = false
		}
	}
	return b.String()
}

// IncludeSelectors joins CSS selectors with commas for use with data-flux-include.
// This is a convenience helper for building the attribute value.
//
// Example:
//
//	hb.Button().
//	  Attr(liveflux.DataFluxAction, "save").
//	  Attr(liveflux.DataFluxInclude, liveflux.IncludeSelectors("#extra-fields", ".shared-inputs")).
//	  Text("Save")
func IncludeSelectors(selectors ...string) string {
	return strings.Join(selectors, ", ")
}

// ExcludeSelectors joins CSS selectors with commas for use with data-flux-exclude.
// This is a convenience helper for building the attribute value.
//
// Example:
//
//	hb.Button().
//	  Attr(liveflux.DataFluxAction, "update").
//	  Attr(liveflux.DataFluxInclude, "#user-form").
//	  Attr(liveflux.DataFluxExclude, liveflux.ExcludeSelectors(".sensitive", ".internal")).
//	  Text("Update Profile")
func ExcludeSelectors(selectors ...string) string {
	return strings.Join(selectors, ", ")
}
