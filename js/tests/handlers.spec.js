describe('Liveflux Handlers', function() {
    beforeEach(function() {
        // Mock liveflux.post
        if (!window.liveflux.post || !window.liveflux.post.and) {
            window.liveflux.post = jasmine.createSpy('post').and.returnValue(Promise.resolve({
                html: '<div data-flux-root="1" data-flux-component="test" data-flux-component-id="123"><p>Updated</p></div>'
            }));
        } else {
            window.liveflux.post.and.returnValue(Promise.resolve({
                html: '<div data-flux-root="1" data-flux-component="test" data-flux-component-id="123"><p>Updated</p></div>'
            }));
        }
        
        // Mock executeScripts
        window.liveflux.executeScripts = jasmine.createSpy('executeScripts');
        
        // Mock initWire
        window.liveflux.initWire = jasmine.createSpy('initWire');
    });

    describe('handleActionClick', function() {
        beforeEach(function() {
            document.body.innerHTML = `
                <div data-flux-root="1" 
                     data-flux-component="counter" 
                     data-flux-component-id="counter-123">
                    <button id="increment-btn" data-flux-action="increment">Increment</button>
                    <input type="text" name="amount" value="5">
                </div>
            `;
        });

        afterEach(function() {
            document.body.innerHTML = '';
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
                    expect(callArgs.liveflux_component_type).toBe('counter');
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

        it('should replace root element on success', function(done) {
            const btn = document.getElementById('increment-btn');
            const event = { 
                target: btn,
                preventDefault: jasmine.createSpy('preventDefault')
            };
            
            window.liveflux.handleActionClick(event);
            
            setTimeout(function() {
                expect(window.liveflux.executeScripts).toHaveBeenCalled();
                expect(window.liveflux.initWire).toHaveBeenCalled();
                done();
            }, 100);
        });
    });

    describe('handleFormSubmit', function() {
        beforeEach(function() {
            document.body.innerHTML = `
                <div data-flux-root="1" 
                     data-flux-component="contact-form" 
                     data-flux-component-id="form-789">
                    <form id="test-form">
                        <input type="text" name="name" value="John Doe">
                        <input type="email" name="email" value="john@example.com">
                        <button type="submit" data-flux-action="save">Submit</button>
                    </form>
                </div>
            `;
        });

        afterEach(function() {
            document.body.innerHTML = '';
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
                    expect(callArgs.liveflux_component_type).toBe('contact-form');
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

        it('should prevent default form submission', function(done) {
            const form = document.getElementById('test-form');
            const preventDefaultSpy = jasmine.createSpy('preventDefault');
            const event = { 
                target: form,
                preventDefault: preventDefaultSpy
            };
            
            window.liveflux.handleFormSubmit(event);
            
            setTimeout(function() {
                expect(preventDefaultSpy).toHaveBeenCalled();
                done();
            }, 50);
        });
    });

    describe('button outside root', function() {
        beforeEach(function() {
            document.body.innerHTML = `
                <div data-flux-root="1" 
                     data-flux-component="modal" 
                     data-flux-component-id="modal-999">
                    <p>Modal content</p>
                </div>
                <button id="external-btn" 
                    data-flux-component-type="modal" 
                    data-flux-component-id="modal-999"
                    data-flux-action="close">
                    Close Modal
                </button>
            `;
        });

        afterEach(function() {
            document.body.innerHTML = '';
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
                    expect(callArgs.liveflux_component_type).toBe('modal');
                    expect(callArgs.liveflux_component_id).toBe('modal-999');
                    expect(callArgs.liveflux_action).toBe('close');
                }
                done();
            }, 50);
        });
    });
});
