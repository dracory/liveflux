package liveflux

import (
	"fmt"
	"reflect"
)

// registry maps aliases to prototype instances of components.
// We store a prototype (typically a zero-value pointer like &MyComp{}) and
// create new instances via reflection when needed.
var registry = map[string]ComponentInterface{}

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

	proto, ok := registry[alias]
	if !ok {
		return nil, fmt.Errorf("liveflux: component '%s' not registered", alias)
	}

	// Instantiate a new component from the registered prototype's type.
	t := reflect.TypeOf(proto)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("liveflux: registered component '%s' must be a pointer type", alias)
	}
	v := reflect.New(t.Elem()).Interface()
	inst, ok := v.(ComponentInterface)
	if !ok {
		return nil, fmt.Errorf("liveflux: type for component '%s' does not implement ComponentInterface", alias)
	}
	inst.SetAlias(alias)
	return inst, nil
}

var typeToAlias = map[reflect.Type]string{}

// RegisterByAlias makes a component constructor available by alias.
// Typically called from init() in the component's package.
func RegisterByAlias(alias string, c ComponentInterface) error {
	if alias == "" || c == nil {
		return fmt.Errorf("liveflux: RegisterByAlias requires non-empty alias and non-nil constructor")
	}
	if _, exists := registry[alias]; exists {
		return fmt.Errorf("liveflux: component '%s' already registered", alias)
	}
	// Store the prototype in the registry.
	registry[alias] = c
	// Store reverse lookup for ergonomics using the prototype type.
	typeToAlias[reflect.TypeOf(c)] = alias
	return nil
}

// newByAlias creates a new component instance by its registered alias.
// Internal to the liveflux package to avoid encouraging string usage at call sites.
func newByAlias(alias string) (ComponentInterface, error) {
	proto, ok := registry[alias]
	if !ok {
		return nil, fmt.Errorf("liveflux: component '%s' not registered", alias)
	}
	t := reflect.TypeOf(proto)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("liveflux: registered component '%s' must be a pointer type", alias)
	}
	v := reflect.New(t.Elem()).Interface()
	inst, ok := v.(ComponentInterface)
	if !ok {
		return nil, fmt.Errorf("liveflux: type for component '%s' does not implement ComponentInterface", alias)
	}
	// Set alias once on construction.
	inst.SetAlias(alias)
	return inst, nil
}

// Register registers a component constructor using the component's GetAlias()
// as the registry alias. Component must implement GetAlias() (enforced by interface).
func Register(c ComponentInterface) error {
	if c == nil {
		return fmt.Errorf("liveflux: Register requires non-nil constructor")
	}

	alias := c.GetAlias()
	if alias == "" {
		return fmt.Errorf("liveflux: Register could not determine alias (empty)")
	}

	return RegisterByAlias(alias, c)
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
