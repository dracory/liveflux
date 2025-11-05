describe('Liveflux Wire', function() {
    describe('initWire', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <div data-flux-root="1" 
                     data-flux-component="counter" 
                     data-flux-component-id="counter-123">
                    <p>Counter content</p>
                </div>
                <div data-flux-root="1" 
                     data-flux-component="todo" 
                     data-flux-component-id="todo-456">
                    <p>Todo content</p>
                </div>
                <div data-flux-root="1" 
                     data-flux-component-id="incomplete-789">
                    <p>Missing component alias</p>
                </div>
            `;
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should initialize $wire on all valid roots', function() {
            window.liveflux.initWire();
            
            const roots = document.querySelectorAll('[data-flux-root]');
            const counterRoot = roots[0];
            const todoRoot = roots[1];
            
            expect(counterRoot.$wire).toBeDefined();
            expect(counterRoot.$wire.id).toBe('counter-123');
            expect(counterRoot.$wire.alias).toBe('counter');
            
            expect(todoRoot.$wire).toBeDefined();
            expect(todoRoot.$wire.id).toBe('todo-456');
            expect(todoRoot.$wire.alias).toBe('todo');
        });

        it('should skip roots with missing component data', function() {
            window.liveflux.initWire();
            
            const roots = document.querySelectorAll('[data-flux-root]');
            const incompleteRoot = roots[2];
            
            expect(incompleteRoot.$wire).toBeUndefined();
        });

        it('should read component metadata from data attributes', function() {
            window.liveflux.initWire();
            
            const root = document.querySelector('[data-flux-component="counter"]');
            
            expect(root.$wire).toBeDefined();
            expect(root.$wire.id).toBe('counter-123');
            expect(root.$wire.alias).toBe('counter');
        });
    });

    describe('$wire API', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <div data-flux-root="1" 
                     data-flux-component="test-component" 
                     data-flux-component-id="test-123">
                    <p>Test content</p>
                </div>
            `;
            document.body.appendChild(testContainer);
            window.liveflux.initWire();
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should expose id property', function() {
            const root = document.querySelector('[data-flux-root]');
            expect(root.$wire.id).toBe('test-123');
        });

        it('should expose alias property', function() {
            const root = document.querySelector('[data-flux-root]');
            expect(root.$wire.alias).toBe('test-component');
        });

        it('should expose on method', function() {
            const root = document.querySelector('[data-flux-root]');
            expect(typeof root.$wire.on).toBe('function');
        });

        it('should expose dispatch method', function() {
            const root = document.querySelector('[data-flux-root]');
            expect(typeof root.$wire.dispatch).toBe('function');
        });

        it('should expose dispatchSelf method', function() {
            const root = document.querySelector('[data-flux-root]');
            expect(typeof root.$wire.dispatchSelf).toBe('function');
        });

        it('should expose dispatchTo method', function() {
            const root = document.querySelector('[data-flux-root]');
            expect(typeof root.$wire.dispatchTo).toBe('function');
        });

        it('should expose call method', function() {
            const root = document.querySelector('[data-flux-root]');
            expect(typeof root.$wire.call).toBe('function');
        });
    });

    describe('createWire', function() {
        it('should create wire object with correct properties', function() {
            const mockRoot = document.createElement('div');
            const wire = window.liveflux.createWire('test-id', 'test-alias', mockRoot);
            
            expect(wire.id).toBe('test-id');
            expect(wire.alias).toBe('test-alias');
            expect(typeof wire.on).toBe('function');
            expect(typeof wire.dispatch).toBe('function');
            expect(typeof wire.dispatchSelf).toBe('function');
            expect(typeof wire.dispatchTo).toBe('function');
            expect(typeof wire.call).toBe('function');
        });
    });
});
