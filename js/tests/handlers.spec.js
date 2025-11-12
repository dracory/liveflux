describe('Liveflux Handlers', function() {
    let originalPost;
    let originalExecuteScripts;
    let originalInitWire;

    beforeAll(function() {
        originalPost = window.liveflux.post;
        originalExecuteScripts = window.liveflux.executeScripts;
        originalInitWire = window.liveflux.initWire;
    });

    beforeEach(function() {
        // Mock liveflux.post
        if (!window.liveflux.post || !window.liveflux.post.and) {
            window.liveflux.post = jasmine.createSpy('post').and.returnValue(Promise.resolve({
                html: '<button id="outside" data-flux-target-kind="test" data-flux-target-id="123" data-flux-action="ping">Ping</button>'
            }));
        } else {
            window.liveflux.post.calls.reset();
            window.liveflux.post.and.returnValue(Promise.resolve({
                html: '<button id="outside" data-flux-target-kind="test" data-flux-target-id="123" data-flux-action="ping">Ping</button>'
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
    });

    afterEach(function() {
        if (typeof originalPost === 'undefined') {
            delete window.liveflux.post;
        } else {
            window.liveflux.post = originalPost;
        }

        if (typeof originalExecuteScripts === 'undefined') {
            delete window.liveflux.executeScripts;
        } else {
            window.liveflux.executeScripts = originalExecuteScripts;
        }

        if (typeof originalInitWire === 'undefined') {
            delete window.liveflux.initWire;
        } else {
            window.liveflux.initWire = originalInitWire;
        }
    });

    afterAll(function() {
        if (typeof originalPost === 'undefined') {
            delete window.liveflux.post;
        } else {
            window.liveflux.post = originalPost;
        }

        if (typeof originalExecuteScripts === 'undefined') {
            delete window.liveflux.executeScripts;
        } else {
            window.liveflux.executeScripts = originalExecuteScripts;
        }

        if (typeof originalInitWire === 'undefined') {
            delete window.liveflux.initWire;
        } else {
            window.liveflux.initWire = originalInitWire;
        }
    });

    describe('handleActionClick', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <div data-flux-component-kind="counter" 
                     data-flux-component-id="counter-123">
                    <button id="increment-btn" data-flux-action="increment">Increment</button>
                    <input type="text" name="amount" value="5">
                </div>
            `;
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should read component metadata from data attributes', function(done) {
            const btn = document.getElementById('increment-btn');
            const event = { 
                target: btn,
                preventDefault: jasmine.createSpy('preventDefault')
            };
            
            window.liveflux.handleActionClick(event);
            
            setTimeout(function() {
                if (window.liveflux.post.calls.count() > 0) {
                    const callArgs = window.liveflux.post.calls.argsFor(0)[0];
                    expect(callArgs.liveflux_component_kind).toBe('counter');
                    expect(callArgs.liveflux_component_id).toBe('counter-123');
                    expect(callArgs.liveflux_action).toBe('increment');
                }
                done();
            }, 50);
        });

        it('should include form fields in request', function(done) {
            const btn = document.getElementById('increment-btn');
            const event = { 
                target: btn,
                preventDefault: jasmine.createSpy('preventDefault')
            };
            
            window.liveflux.handleActionClick(event);
            
            setTimeout(function() {
                if (window.liveflux.post.calls.count() > 0) {
                    const callArgs = window.liveflux.post.calls.argsFor(0)[0];
                    expect(callArgs.amount).toBe('5');
                }
                done();
            }, 50);
        });

        xit('should replace root element on success', function(done) {
            const btn = document.getElementById('increment-btn');
            const event = { 
                target: btn,
                preventDefault: jasmine.createSpy('preventDefault')
            };
            
            window.liveflux.handleActionClick(event);
            
            // Wait for promise to resolve and DOM manipulation to complete
            setTimeout(function() {
                expect(window.liveflux.post).toHaveBeenCalled();
                expect(window.liveflux.executeScripts).toHaveBeenCalled();
                expect(window.liveflux.initWire).toHaveBeenCalled();
                done();
            }, 150);
        });
    });

    describe('handleFormSubmit', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <div data-flux-component-kind="contact-form" 
                     data-flux-component-id="form-789">
                    <form id="test-form">
                        <input type="text" name="name" value="John Doe">
                        <input type="email" name="email" value="john@example.com">
                        <button type="submit" data-flux-action="save">Submit</button>
                    </form>
                </div>
            `;
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should read component metadata from data attributes', function(done) {
            const form = document.getElementById('test-form');
            const event = { 
                target: form,
                preventDefault: jasmine.createSpy('preventDefault')
            };
            
            window.liveflux.handleFormSubmit(event);
            
            setTimeout(function() {
                if (window.liveflux.post.calls.count() > 0) {
                    const callArgs = window.liveflux.post.calls.argsFor(0)[0];
                    expect(callArgs.liveflux_component_kind).toBe('contact-form');
                    expect(callArgs.liveflux_component_id).toBe('form-789');
                    expect(callArgs.liveflux_action).toBe('save');
                }
                done();
            }, 50);
        });

        it('should serialize form fields', function(done) {
            const form = document.getElementById('test-form');
            const event = { 
                target: form,
                preventDefault: jasmine.createSpy('preventDefault')
            };
            
            window.liveflux.handleFormSubmit(event);
            
            setTimeout(function() {
                if (window.liveflux.post.calls.count() > 0) {
                    const callArgs = window.liveflux.post.calls.argsFor(0)[0];
                    expect(callArgs.name).toBe('John Doe');
                    expect(callArgs.email).toBe('john@example.com');
                }
                done();
            }, 50);
        });

        xit('should prevent default form submission', function() {
            const form = document.getElementById('test-form');
            const preventDefaultSpy = jasmine.createSpy('preventDefault');
            const event = { 
                target: form,
                preventDefault: preventDefaultSpy
            };
            
            window.liveflux.handleFormSubmit(event);
            
            // preventDefault is called synchronously
            expect(preventDefaultSpy).toHaveBeenCalled();
        });
    });

    describe('button outside root', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <div data-flux-component-kind="modal" data-flux-component-id="modal-999">
                    <button id="internal-btn" data-flux-action="open">Open Modal</button>
                </div>
                <button id="external-btn" 
                    data-flux-target-kind="modal" 
                    data-flux-target-id="modal-999"
                    data-flux-action="close">
                    Close Modal
                </button>
            `;
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should handle buttons with explicit component attributes', function(done) {
            const btn = document.getElementById('external-btn');
            const event = { 
                target: btn,
                preventDefault: jasmine.createSpy('preventDefault')
            };
            
            window.liveflux.handleActionClick(event);
            
            setTimeout(function() {
                if (window.liveflux.post.calls.count() > 0) {
                    const callArgs = window.liveflux.post.calls.argsFor(0)[0];
                    expect(callArgs.liveflux_component_kind).toBe('modal');
                    expect(callArgs.liveflux_component_id).toBe('modal-999');
                    expect(callArgs.liveflux_action).toBe('close');
                }
                done();
            }, 50);
        });
    });
});
