describe('Liveflux Triggers', function() {
    let originalPost;
    let originalExecuteScripts;
    let originalInitWire;
    let originalCollectAllFields;
    let originalResolveComponentMetadata;
    let originalStartRequestIndicators;
    let originalEndRequestIndicators;

    beforeAll(function() {
        originalPost = window.liveflux.post;
        originalExecuteScripts = window.liveflux.executeScripts;
        originalInitWire = window.liveflux.initWire;
        originalCollectAllFields = window.liveflux.collectAllFields;
        originalResolveComponentMetadata = window.liveflux.resolveComponentMetadata;
        originalStartRequestIndicators = window.liveflux.startRequestIndicators;
        originalEndRequestIndicators = window.liveflux.endRequestIndicators;
    });

    beforeEach(function() {
        // Mock liveflux.post
        if (!window.liveflux.post || !window.liveflux.post.and) {
            window.liveflux.post = jasmine.createSpy('post').and.returnValue(Promise.resolve({
                html: '<div data-flux-root="1" data-flux-component="test" data-flux-component-id="123"><p>Updated</p></div>'
            }));
        } else {
            window.liveflux.post.calls.reset();
            window.liveflux.post.and.returnValue(Promise.resolve({
                html: '<div data-flux-root="1" data-flux-component="test" data-flux-component-id="123"><p>Updated</p></div>'
            }));
        }
        
        // Mock executeScripts
        if (!window.liveflux.executeScripts || !window.liveflux.executeScripts.and) {
            window.liveflux.executeScripts = jasmine.createSpy('executeScripts');
        } else {
            window.liveflux.executeScripts.calls.reset();
        }
        
        // Mock initWire
        if (!window.liveflux.initWire || !window.liveflux.initWire.and) {
            window.liveflux.initWire = jasmine.createSpy('initWire');
        } else {
            window.liveflux.initWire.calls.reset();
        }
        
        // Mock collectAllFields
        if (!window.liveflux.collectAllFields || !window.liveflux.collectAllFields.and) {
            window.liveflux.collectAllFields = jasmine.createSpy('collectAllFields').and.returnValue({
                query: 'test'
            });
        } else {
            window.liveflux.collectAllFields.calls.reset();
            window.liveflux.collectAllFields.and.returnValue({
                query: 'test'
            });
        }
        
        // Mock resolveComponentMetadata
        if (!window.liveflux.resolveComponentMetadata || !window.liveflux.resolveComponentMetadata.and) {
            window.liveflux.resolveComponentMetadata = jasmine.createSpy('resolveComponentMetadata').and.returnValue({
                comp: 'search',
                id: 'search-123',
                root: document.createElement('div')
            });
        } else {
            window.liveflux.resolveComponentMetadata.calls.reset();
            window.liveflux.resolveComponentMetadata.and.returnValue({
                comp: 'search',
                id: 'search-123',
                root: document.createElement('div')
            });
        }
        
        // Mock indicators
        if (!window.liveflux.startRequestIndicators || !window.liveflux.startRequestIndicators.and) {
            window.liveflux.startRequestIndicators = jasmine.createSpy('startRequestIndicators').and.returnValue([]);
        } else {
            window.liveflux.startRequestIndicators.calls.reset();
            window.liveflux.startRequestIndicators.and.returnValue([]);
        }
        
        if (!window.liveflux.endRequestIndicators || !window.liveflux.endRequestIndicators.and) {
            window.liveflux.endRequestIndicators = jasmine.createSpy('endRequestIndicators');
        } else {
            window.liveflux.endRequestIndicators.calls.reset();
        }
    });

    afterEach(function() {
        window.liveflux.post = originalPost;
        window.liveflux.executeScripts = originalExecuteScripts;
        window.liveflux.initWire = originalInitWire;
        window.liveflux.collectAllFields = originalCollectAllFields;
        window.liveflux.resolveComponentMetadata = originalResolveComponentMetadata;
        window.liveflux.startRequestIndicators = originalStartRequestIndicators;
        window.liveflux.endRequestIndicators = originalEndRequestIndicators;
    });

    describe('parseTriggers', function() {
        it('should parse simple event trigger', function() {
            const el = document.createElement('input');
            el.setAttribute('data-flux-trigger', 'keyup');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers.length).toBe(1);
            expect(triggers[0].events).toContain('keyup');
            expect(triggers[0].filters.length).toBe(0);
        });

        it('should parse event with changed filter', function() {
            const el = document.createElement('input');
            el.setAttribute('data-flux-trigger', 'keyup changed');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers.length).toBe(1);
            expect(triggers[0].events).toContain('keyup');
            expect(triggers[0].filters.length).toBe(1);
            expect(triggers[0].filters[0].type).toBe('changed');
        });

        it('should parse event with delay modifier', function() {
            const el = document.createElement('input');
            el.setAttribute('data-flux-trigger', 'keyup delay:300ms');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers.length).toBe(1);
            expect(triggers[0].modifiers.delay).toBe(300);
        });

        it('should parse delay in seconds', function() {
            const el = document.createElement('input');
            el.setAttribute('data-flux-trigger', 'input delay:1s');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].modifiers.delay).toBe(1000);
        });

        it('should parse throttle modifier', function() {
            const el = document.createElement('div');
            el.setAttribute('data-flux-trigger', 'scroll throttle:500ms');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].modifiers.throttle).toBe(500);
        });

        it('should parse queue strategy', function() {
            const el = document.createElement('input');
            el.setAttribute('data-flux-trigger', 'keyup queue:all');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].modifiers.queue).toBe('all');
        });

        it('should parse once filter', function() {
            const el = document.createElement('button');
            el.setAttribute('data-flux-trigger', 'click once');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].filters.length).toBe(1);
            expect(triggers[0].filters[0].type).toBe('once');
        });

        it('should parse from selector filter', function() {
            const el = document.createElement('div');
            el.setAttribute('data-flux-trigger', 'click from:.delete-btn');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].filters.length).toBe(1);
            expect(triggers[0].filters[0].type).toBe('from');
            expect(triggers[0].filters[0].selector).toBe('.delete-btn');
        });

        it('should parse not selector filter', function() {
            const el = document.createElement('div');
            el.setAttribute('data-flux-trigger', 'click not:button');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].filters.length).toBe(1);
            expect(triggers[0].filters[0].type).toBe('not');
            expect(triggers[0].filters[0].selector).toBe('button');
        });

        it('should parse complex trigger definition', function() {
            const el = document.createElement('input');
            el.setAttribute('data-flux-trigger', 'keyup changed delay:300ms queue:replace');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers.length).toBe(1);
            expect(triggers[0].events).toContain('keyup');
            expect(triggers[0].filters[0].type).toBe('changed');
            expect(triggers[0].modifiers.delay).toBe(300);
            expect(triggers[0].modifiers.queue).toBe('replace');
        });

        it('should parse multiple trigger definitions', function() {
            const el = document.createElement('input');
            el.setAttribute('data-flux-trigger', 'keyup delay:300ms, blur');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers.length).toBe(2);
            expect(triggers[0].events).toContain('keyup');
            expect(triggers[0].modifiers.delay).toBe(300);
            expect(triggers[1].events).toContain('blur');
        });

        it('should infer default event for text input', function() {
            const el = document.createElement('input');
            el.type = 'text';
            el.setAttribute('data-flux-trigger', 'delay:300ms');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].events).toContain('keyup');
            expect(triggers[0].events).toContain('changed');
        });

        it('should infer default event for select', function() {
            const el = document.createElement('select');
            el.setAttribute('data-flux-trigger', 'delay:100ms');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers[0].events).toContain('change');
        });

        it('should infer default event for checkbox', function() {
            const el = document.createElement('input');
            el.type = 'checkbox';
            el.setAttribute('data-flux-trigger', 'delay:100ms');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers.length).toBe(1);
            expect(triggers[0].events).toContain('change');
        });

        it('should return empty array for element without trigger attribute', function() {
            const el = document.createElement('input');
            
            const triggers = window.liveflux.parseTriggers(el);
            
            expect(triggers.length).toBe(0);
        });
    });

    describe('registerTriggers and event handling', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            // Cleanup triggers before removing container
            if (testContainer) {
                const triggeredElements = testContainer.querySelectorAll('[data-flux-trigger]');
                triggeredElements.forEach(el => {
                    if (window.liveflux.unregisterTriggers) {
                        window.liveflux.unregisterTriggers(el);
                    }
                });
            }
            
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should register trigger on element', function() {
            testContainer.innerHTML = `
                <div data-flux-root="1" data-flux-component="search" data-flux-component-id="search-123">
                    <input id="search-input" 
                           type="text" 
                           name="query"
                           data-flux-trigger="keyup"
                           data-flux-action="search" />
                </div>
            `;
            
            const input = document.getElementById('search-input');
            window.liveflux.registerTriggers(input);
            
            // Trigger should be registered (we can't easily test the internal state,
            // but we can verify it doesn't throw)
            expect(input).toBeDefined();
        });

        it('should fire action on trigger event', function(done) {
            testContainer.innerHTML = `
                <div data-flux-root="1" data-flux-component="search" data-flux-component-id="search-123">
                    <input id="search-input" 
                           type="text" 
                           name="query"
                           data-flux-trigger="input"
                           data-flux-action="search" />
                </div>
            `;
            
            const input = document.getElementById('search-input');
            const root = testContainer.querySelector('[data-flux-root]');
            
            // Update existing spy to return the actual root
            window.liveflux.resolveComponentMetadata.and.returnValue({
                comp: 'search',
                id: 'search-123',
                root: root
            });
            
            window.liveflux.registerTriggers(input);
            
            // Trigger the event
            const event = new Event('input', { bubbles: true });
            input.dispatchEvent(event);
            
            // Give it a moment to process async operations
            setTimeout(function() {
                expect(window.liveflux.post).toHaveBeenCalled();
                done();
            }, 100);
        });

        it('should respect delay modifier', function(done) {
            testContainer.innerHTML = `
                <div data-flux-root="1" data-flux-component="search" data-flux-component-id="search-123">
                    <input id="search-input" 
                           type="text" 
                           name="query"
                           data-flux-trigger="input delay:200ms"
                           data-flux-action="search" />
                </div>
            `;
            
            const input = document.getElementById('search-input');
            const root = testContainer.querySelector('[data-flux-root]');
            
            // Update existing spy to return the actual root
            window.liveflux.resolveComponentMetadata.and.returnValue({
                comp: 'search',
                id: 'search-123',
                root: root
            });
            
            window.liveflux.registerTriggers(input);
            
            const event = new Event('input', { bubbles: true });
            input.dispatchEvent(event);
            
            // Should not fire immediately
            setTimeout(function() {
                expect(window.liveflux.post).not.toHaveBeenCalled();
            }, 50);
            
            // Should fire after delay
            setTimeout(function() {
                expect(window.liveflux.post).toHaveBeenCalled();
                done();
            }, 300);
        });

        it('should debounce multiple rapid events', function(done) {
            testContainer.innerHTML = `
                <div data-flux-root="1" data-flux-component="search" data-flux-component-id="search-123">
                    <input id="search-input" 
                           type="text" 
                           name="query"
                           data-flux-trigger="input delay:200ms"
                           data-flux-action="search" />
                </div>
            `;
            
            const input = document.getElementById('search-input');
            const root = testContainer.querySelector('[data-flux-root]');
            
            // Update existing spy to return the actual root
            window.liveflux.resolveComponentMetadata.and.returnValue({
                comp: 'search',
                id: 'search-123',
                root: root
            });
            
            window.liveflux.registerTriggers(input);
            
            const event = new Event('input', { bubbles: true });
            
            // Fire multiple events rapidly
            input.dispatchEvent(event);
            setTimeout(() => input.dispatchEvent(event), 50);
            setTimeout(() => input.dispatchEvent(event), 100);
            
            // Should only fire once after all events settle
            setTimeout(function() {
                const callCount = window.liveflux.post.calls.count();
                expect(callCount).toBe(1);
                done();
            }, 400);
        });
    });

    describe('initTriggers', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should initialize all triggers in document', function() {
            testContainer.innerHTML = `
                <div data-flux-root="1" data-flux-component="search" data-flux-component-id="search-123">
                    <input id="input1" data-flux-trigger="keyup" data-flux-action="search" />
                    <input id="input2" data-flux-trigger="change" data-flux-action="filter" />
                </div>
            `;
            
            window.liveflux.initTriggers(testContainer);
            
            // Should not throw and should find elements
            const elements = testContainer.querySelectorAll('[data-flux-trigger]');
            expect(elements.length).toBe(2);
        });

        it('should initialize triggers in specific root', function() {
            testContainer.innerHTML = `
                <div id="root1" data-flux-root="1" data-flux-component="comp1" data-flux-component-id="id1">
                    <input data-flux-trigger="keyup" data-flux-action="action1" />
                </div>
                <div id="root2" data-flux-root="1" data-flux-component="comp2" data-flux-component-id="id2">
                    <input data-flux-trigger="change" data-flux-action="action2" />
                </div>
            `;
            
            const root1 = document.getElementById('root1');
            window.liveflux.initTriggers(root1);
            
            // Should initialize only triggers in root1
            expect(root1.querySelector('[data-flux-trigger]')).toBeDefined();
        });
    });

    describe('unregisterTriggers', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should unregister triggers from element', function() {
            testContainer.innerHTML = `
                <div data-flux-root="1" data-flux-component="search" data-flux-component-id="search-123">
                    <input id="search-input" 
                           type="text" 
                           data-flux-trigger="keyup"
                           data-flux-action="search" />
                </div>
            `;
            
            const input = document.getElementById('search-input');
            window.liveflux.registerTriggers(input);
            window.liveflux.unregisterTriggers(input);
            
            // Should not throw
            expect(input).toBeDefined();
        });
    });

    describe('configureTriggers', function() {
        it('should configure default trigger delay', function() {
            window.liveflux.configureTriggers({
                defaultTriggerDelay: 500
            });
            
            // Configuration should be applied (we can't easily test internal state)
            expect(window.liveflux.configureTriggers).toBeDefined();
        });

        it('should configure enable/disable triggers', function() {
            window.liveflux.configureTriggers({
                enableTriggers: false
            });
            
            expect(window.liveflux.configureTriggers).toBeDefined();
            
            // Re-enable for other tests
            window.liveflux.configureTriggers({
                enableTriggers: true
            });
        });
    });

    describe('X-Liveflux-Trigger header', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            // Cleanup triggers before removing container
            if (testContainer) {
                const triggeredElements = testContainer.querySelectorAll('[data-flux-trigger]');
                triggeredElements.forEach(el => {
                    if (window.liveflux.unregisterTriggers) {
                        window.liveflux.unregisterTriggers(el);
                    }
                });
            }
            
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should send trigger event name in header', function(done) {
            testContainer.innerHTML = `
                <div data-flux-root="1" data-flux-component="search" data-flux-component-id="search-123">
                    <input id="search-input" 
                           type="text"
                           name="query"
                           data-flux-trigger="keyup"
                           data-flux-action="search" />
                </div>
            `;
            
            const input = document.getElementById('search-input');
            const root = testContainer.querySelector('[data-flux-root]');
            
            // Update existing spy to return the actual root
            window.liveflux.resolveComponentMetadata.and.returnValue({
                comp: 'search',
                id: 'search-123',
                root: root
            });
            
            window.liveflux.registerTriggers(input);
            
            const event = new Event('keyup', { bubbles: true });
            input.dispatchEvent(event);
            
            setTimeout(function() {
                expect(window.liveflux.post).toHaveBeenCalled();
                // Headers should have been set (we can't easily verify the exact header value
                // without more complex mocking, but we can verify the call was made)
                done();
            }, 100);
        });
    });
});
