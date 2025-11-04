# Form-less Submission Example

This example demonstrates the **form-less submission** feature in Liveflux, which allows components to collect form data from arbitrary DOM elements using `data-flux-include` and `data-flux-exclude` attributes.

## Features Demonstrated

### 1. Shared Filters (data-flux-include)

Two separate components (`ProductList` and `ArticleList`) share the same filter inputs located outside their component roots. Both use:

```html
<button data-flux-action="refresh" data-flux-include="#global-filters">
  Refresh
</button>
```

This collects values from the `#global-filters` element and submits them along with the component's own data.

### 2. Multi-Step Form (multiple includes)

The `MultiStepForm` component demonstrates including fields from multiple sections:

```html
<button data-flux-action="submit" 
        data-flux-include="#step-1, #step-2">
  Complete Registration
</button>
```

This collects all input fields from both `#step-1` and `#step-2` sections, even though they're outside the component root.

### 3. Excluding Sensitive Fields (data-flux-exclude)

The `ExcludeExample` component shows how to include a form but exclude specific fields:

```html
<button data-flux-action="update-profile"
        data-flux-include="#user-form"
        data-flux-exclude=".sensitive">
  Update Profile (without password)
</button>
```

This includes all fields from `#user-form` except those with the `sensitive` class (like the password field).

## Running the Example

```bash
# From the repository root
go run ./examples/formless

# Or with Task
task examples:formless:run
```

Then open http://localhost:8080 in your browser.

## How It Works

### Client-Side

The `collectAllFields()` function in `liveflux_util.js`:

1. Serializes the default scope (form or component root)
2. Processes `data-flux-include` selectors and merges their fields
3. Processes `data-flux-exclude` selectors and removes those fields
4. Merges button parameters (highest precedence)

### Server-Side

No changes required! The handler already processes all submitted fields via `r.Form`.

## Benefits

- **Initialization safety**: Using `<div>` containers instead of `<form>` elements prevents accidental native form submissions when Liveflux fails to initialize
- **Flexible composition**: Share form fragments across multiple components
- **Reduced wrapper overhead**: No need to wrap everything in `<form>` tags
- **Fine-grained control**: Include or exclude specific fields as needed

## Field Precedence

When the same field name appears in multiple sources, the precedence (lowest to highest) is:

1. Component root fields
2. Associated form fields
3. Included elements (left to right in selector list)
4. Excluded elements (removed)
5. Button `data-flux-param-*` attributes
6. Button `name`/`value` (if applicable)

## See Also

- [Form-less Submission Design Document](../../docs/form-less.md)
- [Liveflux Documentation](../../docs/)
