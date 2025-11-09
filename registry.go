package liveflux

import (
	"fmt"
	"reflect"
)

// registry maps kinds to prototype instances of components.
// We store a prototype (typically a zero-value pointer like &MyComp{}) and
// create new instances via reflection when needed.
var registry = map[string]ComponentInterface{}

// New creates a new component instance by using the type of the provided example.
// The example is not used beyond its type information. The registry is consulted
// via KindOf(example) and, if empty, DefaultKindFromType(example). This avoids
// passing string kinds at call sites.
func New(component ComponentInterface) (ComponentInterface, error) {
	if component == nil {
		return nil, fmt.Errorf("liveflux: New requires non-nil component")
	}

	kind := KindOf(component)
	if kind == "" {
		kind = DefaultKindFromType(component)
	}

	proto, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("liveflux: component '%s' not registered", kind)
	}

	// Instantiate a new component from the registered prototype's type.
	t := reflect.TypeOf(proto)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("liveflux: registered component '%s' must be a pointer type", kind)
	}
	v := reflect.New(t.Elem()).Interface()
	inst, ok := v.(ComponentInterface)
	if !ok {
		return nil, fmt.Errorf("liveflux: type for component '%s' does not implement ComponentInterface", kind)
	}
	inst.SetKind(kind)
	return inst, nil
}

var typeToKind = map[reflect.Type]string{}

// RegisterByKind makes a component constructor available by kind.
// Typically called from init() in the component's package.
func RegisterByKind(kind string, c ComponentInterface) error {
	if kind == "" || c == nil {
		return fmt.Errorf("liveflux: RegisterByKind requires non-empty kind and non-nil constructor")
	}
	if _, exists := registry[kind]; exists {
		return fmt.Errorf("liveflux: component '%s' already registered", kind)
	}
	// Store the prototype in the registry.
	registry[kind] = c
	// Store reverse lookup for ergonomics using the prototype type.
	typeToKind[reflect.TypeOf(c)] = kind
	return nil
}

// newByKind creates a new component instance by its registered kind.
// Internal to the liveflux package to avoid encouraging string usage at call sites.
func newByKind(kind string) (ComponentInterface, error) {
	proto, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("liveflux: component '%s' not registered", kind)
	}
	t := reflect.TypeOf(proto)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("liveflux: registered component '%s' must be a pointer type", kind)
	}
	v := reflect.New(t.Elem()).Interface()
	inst, ok := v.(ComponentInterface)
	if !ok {
		return nil, fmt.Errorf("liveflux: type for component '%s' does not implement ComponentInterface", kind)
	}
	// Set kind once on construction.
	inst.SetKind(kind)
	return inst, nil
}

// Register registers a component constructor using the component's GetKind()
// as the registry kind. Component must implement GetKind() (enforced by interface).
func Register(c ComponentInterface) error {
	if c == nil {
		return fmt.Errorf("liveflux: Register requires non-nil constructor")
	}

	kind := c.GetKind()
	if kind == "" {
		return fmt.Errorf("liveflux: Register could not determine kind (empty)")
	}

	return RegisterByKind(kind, c)
}

// KindOf returns the registered kind for the given component instance's type.
// Returns empty string if not found.
func KindOf(c ComponentInterface) string {
	if c == nil {
		return ""
	}

	t := reflect.TypeOf(c)
	if kind, ok := typeToKind[t]; ok {
		return kind
	}

	return ""
}
