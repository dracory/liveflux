# Liveflux JavaScript Tests

This directory contains Jasmine test specifications for the Liveflux JavaScript modules.

## Test Structure

- `runner.html` - Main test runner HTML file that loads Jasmine and all test specs
- `events.spec.js` - Tests for the liveflux_events.js module
- `dispatch.spec.js` - Tests for the liveflux_dispatch.js module  
- `bootstrap.spec.js` - Tests for the liveflux_bootstrap.js module

## Running Tests

### Method 1: Direct Browser Access
1. Open `runner.html` in your web browser
2. The tests will run automatically and display results

### Method 2: Local Server (Recommended)
1. Start a local HTTP server in this directory:
   ```bash
   # Using Python
   python -m http.server 8000
   
   # Using Node.js (if you have http-server installed)
   npx http-server -p 8000
   
   # Using Go
   go run .
   ```

2. Navigate to `http://localhost:8000/runner.html` in your browser

### Method 3: Live Server in VS Code
1. Install the "Live Server" extension
2. Right-click on `runner.html` and select "Open with Live Server"

## Test Coverage

### Events Module (`events.spec.js`)
- Event listener registration and cleanup
- Event dispatching to listeners and DOM
- Component-specific event handling
- Subscription functionality

### Dispatch Module (`dispatch.spec.js`)
- `dispatchTo()` - Target specific components
- `dispatchToAlias()` - Target all components by alias
- `dispatchToAliasAndId()` - Target specific component by alias and ID
- Error handling for missing parameters
- Event payload construction with targeting data

### Bootstrap Module (`bootstrap.spec.js`)
- Initialization process and state management
- Event listener setup (click, submit)
- Component placeholder mounting
- Wire initialization (async)
- Compatibility with Livewire events
- DOM ready state handling
- Error handling for missing dependencies

## Writing New Tests

1. Create a new `.spec.js` file following the naming convention
2. Add the script reference to `runner.html` before the closing body tag
3. Use Jasmine syntax (`describe`, `it`, `expect`, `spyOn`, etc.)
4. Mock external dependencies (DOM, window.liveflux namespace)
5. Clean up after each test with `beforeEach` and `afterEach`

## Mocking Strategy

The test suite uses a comprehensive mocking strategy:

- **DOM APIs**: Mock `document.addEventListener`, `querySelectorAll`, etc.
- **Liveflux Namespace**: Mock `window.liveflux` with required properties
- **External Dependencies**: Mock functions that depend on other modules
- **Component Elements**: Mock DOM elements with `getAttribute` methods

## Debugging Tests

- Use browser developer tools to inspect test failures
- Add `console.log` statements in test specs for debugging
- Use `debugger;` statements to set breakpoints
- Check the Jasmine HTML reporter for detailed error messages

## Continuous Integration

These tests can be integrated into CI/CD pipelines using headless browsers:

```bash
# Using Chrome headless
google-chrome --headless --disable-gpu --run-all-tests runner.html

# Using Firefox headless  
firefox --headless runner.html
```
