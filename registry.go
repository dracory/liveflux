package liveflux

import (
	"fmt"
	"reflect"
)

// ctor is a factory that returns a new zero-value ComponentInterface instance.
type ctor func() ComponentInterface

var registry = map[string]ctor{}

// New creates a new component instance by using the type of the provided example.
// The example is not used beyond its type information. The registry is consulted
// via AliasOf(example) and, if empty, DefaultAliasFromType(example). This avoids
// passing string aliases at call sites.
func New(component ComponentInterface) (ComponentInterface, error) {
	if component == nil {
		return nil, fmt.Errorf("liveflux: New requires non-nil component")
	}

	alias := AliasOf(component)
	if alias == "" {
		alias = DefaultAliasFromType(component)
	}

	ctor, ok := registry[alias]
	if !ok {
		return nil, fmt.Errorf("liveflux: component '%s' not registered", alias)
	}

	inst := ctor()
	if inst != nil {
		inst.SetAlias(alias)
	}

	return inst, nil
}

var typeToAlias = map[reflect.Type]string{}

// RegisterByAlias makes a component constructor available by alias.
// Typically called from init() in the component's package.
func RegisterByAlias(alias string, c ctor) {
	if alias == "" || c == nil {
		panic("liveflux: RegisterByAlias requires non-empty alias and non-nil constructor")
	}
	if _, exists := registry[alias]; exists {
		panic(fmt.Sprintf("liveflux: component '%s' already registered", alias))
	}
	registry[alias] = c
	// Store reverse lookup for ergonomics
	if inst := c(); inst != nil {
		typeToAlias[reflect.TypeOf(inst)] = alias
	}
}

// newByAlias creates a new component instance by its registered alias.
// Internal to the liveflux package to avoid encouraging string usage at call sites.
func newByAlias(alias string) (ComponentInterface, error) {
	c, ok := registry[alias]
	if !ok {
		return nil, fmt.Errorf("liveflux: component '%s' not registered", alias)
	}
	inst := c()
	if inst != nil {
		// Set alias once on construction.
		inst.SetAlias(alias)
	}
	return inst, nil
}

// Register registers a component constructor using the component's GetAlias()
// as the registry alias. Component must implement GetAlias() (enforced by interface).
func Register(c ctor) {
	if c == nil {
		panic("liveflux: Register requires non-nil constructor")
	}
	inst := c()
	if inst == nil {
		panic("liveflux: Register constructor returned nil instance")
	}
	alias := inst.GetAlias()
	if alias == "" {
		alias = DefaultAliasFromType(inst)
	}
	if alias == "" {
		panic("liveflux: Register could not determine alias (empty)")
	}
	RegisterByAlias(alias, c)
}

// AliasOf returns the registered alias for the given component instance's type.
// Returns empty string if not found.
func AliasOf(c ComponentInterface) string {
	if c == nil {
		return ""
	}

	t := reflect.TypeOf(c)
	if alias, ok := typeToAlias[t]; ok {
		return alias
	}

	return ""
}
