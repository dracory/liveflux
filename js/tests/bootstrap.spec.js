describe('Liveflux Bootstrap', function() {
    describe('initialization', function() {
        it('should set bootstrap init flag', function() {
            expect(window.liveflux.__bootstrapInitDone).toBe(true);
        });

        it('should expose bootstrapInit function', function() {
            expect(typeof window.liveflux.bootstrapInit).toBe('function');
        });

        it('should not re-run module initialization', function() {
            // The module should only run once due to __bootstrapInitDone flag
            const flag = window.liveflux.__bootstrapInitDone;
            expect(flag).toBe(true);
        });
    });

    describe('bootstrapInit function', function() {
        let originalHandleActionClick;
        let originalHandleFormSubmit;
        let originalMountPlaceholders;
        let originalInitWire;

        beforeAll(function() {
            originalHandleActionClick = window.liveflux.handleActionClick;
            originalHandleFormSubmit = window.liveflux.handleFormSubmit;
            originalMountPlaceholders = window.liveflux.mountPlaceholders;
            originalInitWire = window.liveflux.initWire;
        });

        beforeEach(function() {
            // Mock the liveflux methods
            window.liveflux.handleActionClick = jasmine.createSpy('handleActionClick');
            window.liveflux.handleFormSubmit = jasmine.createSpy('handleFormSubmit');
            window.liveflux.mountPlaceholders = jasmine.createSpy('mountPlaceholders');
            window.liveflux.initWire = jasmine.createSpy('initWire');
        });

        afterEach(function() {
            window.liveflux.handleActionClick = originalHandleActionClick;
            window.liveflux.handleFormSubmit = originalHandleFormSubmit;
            window.liveflux.mountPlaceholders = originalMountPlaceholders;
            if (typeof originalInitWire === 'undefined') {
                delete window.liveflux.initWire;
            } else {
                window.liveflux.initWire = originalInitWire;
            }
        });

        afterAll(function() {
            window.liveflux.handleActionClick = originalHandleActionClick;
            window.liveflux.handleFormSubmit = originalHandleFormSubmit;
            window.liveflux.mountPlaceholders = originalMountPlaceholders;
            if (typeof originalInitWire === 'undefined') {
                delete window.liveflux.initWire;
            } else {
                window.liveflux.initWire = originalInitWire;
            }
        });

        it('should add event listeners when called', function() {
            spyOn(document, 'addEventListener');
            
            window.liveflux.bootstrapInit();
            
            expect(document.addEventListener).toHaveBeenCalledWith('click', window.liveflux.handleActionClick);
            expect(document.addEventListener).toHaveBeenCalledWith('submit', window.liveflux.handleFormSubmit);
        });

        it('should call mountPlaceholders', function() {
            window.liveflux.bootstrapInit();
            
            expect(window.liveflux.mountPlaceholders).toHaveBeenCalled();
        });

        it('should call initWire asynchronously when available', function(done) {
            window.liveflux.bootstrapInit();
            
            // initWire is called via setTimeout, so we need to wait
            setTimeout(function() {
                expect(window.liveflux.initWire).toHaveBeenCalled();
                done();
            }, 10);
        });

        it('should handle missing initWire gracefully', function() {
            delete window.liveflux.initWire;
            
            // Should not throw error
            expect(function() {
                window.liveflux.bootstrapInit();
            }).not.toThrow();
        });

        it('should dispatch livewire:init compatibility event', function() {
            spyOn(document, 'dispatchEvent');
            
            window.liveflux.bootstrapInit();
            
            expect(document.dispatchEvent).toHaveBeenCalled();
            const eventCall = document.dispatchEvent.calls.argsFor(0)[0];
            expect(eventCall.type).toBe('livewire:init');
        });
    });

    describe('module loading', function() {
        it('should have loaded successfully', function() {
            expect(window.liveflux.__bootstrapInitDone).toBe(true);
            expect(window.liveflux.bootstrapInit).toBeDefined();
        });
    });
});
