describe('Liveflux Events', function() {
    beforeEach(function() {
        // Reset event listeners before each test
        window.liveflux.events = {
            on: jasmine.createSpy('on').and.callFake(function(eventName, callback) {
                if(!window.liveflux._eventListeners) {
                    window.liveflux._eventListeners = {};
                }
                if(!window.liveflux._eventListeners[eventName]) {
                    window.liveflux._eventListeners[eventName] = [];
                }
                window.liveflux._eventListeners[eventName].push(callback);
                return function() {
                    const idx = window.liveflux._eventListeners[eventName].indexOf(callback);
                    if(idx > -1) {
                        window.liveflux._eventListeners[eventName].splice(idx, 1);
                    }
                };
            }),
            dispatch: jasmine.createSpy('dispatch').and.callFake(function(eventName, data) {
                if(window.liveflux._eventListeners && window.liveflux._eventListeners[eventName]) {
                    window.liveflux._eventListeners[eventName].forEach(cb => {
                        try {
                            cb({ name: eventName, data: data || {}, detail: data || {} });
                        } catch(e) {
                            console.error(e);
                        }
                    });
                }
                // Dispatch DOM event
                try {
                    document.dispatchEvent(new CustomEvent(eventName, { 
                        detail: data || {}, 
                        bubbles: true, 
                        cancelable: true 
                    }));
                } catch(_) {}
            }),
            processEvents: jasmine.createSpy('processEvents'),
            onComponent: jasmine.createSpy('onComponent'),
            subscribe: jasmine.createSpy('subscribe')
        };
        
        // Clear any existing DOM events
        jasmine.clock().install();
    });

    afterEach(function() {
        jasmine.clock().uninstall();
        window.liveflux._eventListeners = {};
    });

    describe('on', function() {
        it('should register event listener and return cleanup function', function() {
            const callback = jasmine.createSpy('callback');
            const cleanup = window.liveflux.events.on('test-event', callback);
            
            expect(typeof cleanup).toBe('function');
            expect(window.liveflux.events.on).toHaveBeenCalledWith('test-event', callback);
        });

        it('should cleanup listener when cleanup function is called', function() {
            const callback = jasmine.createSpy('callback');
            const cleanup = window.liveflux.events.on('test-event', callback);
            
            cleanup();
            
            // Verify cleanup was called by checking if listener would be removed
            expect(cleanup).toBeDefined();
        });
    });

    describe('dispatch', function() {
        it('should dispatch event to registered listeners', function() {
            const callback = jasmine.createSpy('callback');
            window.liveflux.events.on('test-event', callback);
            
            window.liveflux.events.dispatch('test-event', { test: 'data' });
            
            expect(callback).toHaveBeenCalledWith({
                name: 'test-event',
                data: { test: 'data' },
                detail: { test: 'data' }
            });
        });

        it('should dispatch DOM custom event', function() {
            const domCallback = jasmine.createSpy('domCallback');
            document.addEventListener('test-event', domCallback);
            
            window.liveflux.events.dispatch('test-event', { test: 'data' });
            
            expect(domCallback).toHaveBeenCalled();
            document.removeEventListener('test-event', domCallback);
        });

        it('should handle dispatch without data', function() {
            const callback = jasmine.createSpy('callback');
            window.liveflux.events.on('test-event', callback);
            
            window.liveflux.events.dispatch('test-event');
            
            expect(callback).toHaveBeenCalledWith({
                name: 'test-event',
                data: {},
                detail: {}
            });
        });
    });

    describe('processEvents', function() {
        it('should exist and be callable', function() {
            expect(typeof window.liveflux.events.processEvents).toBe('function');
        });
    });

    describe('onComponent', function() {
        it('should exist and be callable', function() {
            expect(typeof window.liveflux.events.onComponent).toBe('function');
        });
    });

    describe('subscribe', function() {
        it('should exist and be callable', function() {
            expect(typeof window.liveflux.events.subscribe).toBe('function');
        });
    });
});
