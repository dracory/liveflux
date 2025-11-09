describe('Liveflux Dispatch', function() {
    beforeEach(function() {
        // Mock DOM elements
        this.mockComponent = {
            getAttribute: jasmine.createSpy('getAttribute').and.callFake(function(attr) {
                if (attr === 'data-flux-component') return 'test-component';
                if (attr === 'data-flux-component-id') return 'test-id-123';
                return null;
            })
        };

        // Mock querySelectorAll
        spyOn(document, 'querySelectorAll').and.callFake(function(selector) {
            if (selector.includes('test-kind') || selector.includes('test-component')) {
                return [this.mockComponent];
            }
            return [];
        }.bind(this));

        // Mock findComponent function
        window.liveflux.findComponent = jasmine.createSpy('findComponent').and.returnValue(this.mockComponent);

        // Mock events.dispatch and events.on
        window.liveflux.events = {
            dispatch: jasmine.createSpy('dispatch'),
            on: jasmine.createSpy('on')
        };
    });

    describe('dispatchTo', function() {
        it('should dispatch event to specific component', function() {
            const eventData = { test: 'data' };
            
            window.liveflux.dispatchTo(this.mockComponent, 'test-event', eventData);
            
            expect(window.liveflux.events.dispatch).toHaveBeenCalledWith('test-event', {
                test: 'data',
                __target: 'test-component',
                __target_id: 'test-id-123'
            });
        });

        it('should handle missing component', function() {
            const warnSpy = spyOn(console, 'warn');
            
            window.liveflux.dispatchTo(null, 'test-event', {});
            
            expect(warnSpy).toHaveBeenCalledWith('[Liveflux Events] dispatchTo called without component');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });

        it('should handle missing event name', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchTo(this.mockComponent, null, {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchTo called without event name');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });
    });

    describe('dispatchToKind', function() {
        it('should dispatch event to all components with kind', function() {
            const eventData = { test: 'data' };
            
            window.liveflux.dispatchToKind('test-kind', 'test-event', eventData);
            
            expect(window.liveflux.events.dispatch).toHaveBeenCalledWith('test-event', {
                test: 'data',
                __target: 'test-component'
            });
        });

        it('should handle missing component kind', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToKind(null, 'test-event', {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToKind called without component kind');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });

        it('should handle missing event name', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToKind('test-component', null, {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToKind called without event name');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });
    });

    describe('dispatchToKindAndId', function() {
        it('should dispatch event to specific component by kind and ID', function() {
            const eventData = { test: 'data' };
            
            window.liveflux.dispatchToKindAndId('test-component', 'test-id-123', 'test-event', eventData);
            
            expect(window.liveflux.findComponent).toHaveBeenCalledWith('test-component', 'test-id-123');
            expect(window.liveflux.events.dispatch).toHaveBeenCalledWith('test-event', {
                test: 'data',
                __target: 'test-component',
                __target_id: 'test-id-123'
            });
        });

        it('should handle missing component kind', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToKindAndId(null, 'test-id-123', 'test-event', {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToKindAndId called without component kind');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });

        it('should handle missing component ID', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToKindAndId('test-component', null, 'test-event', {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToKindAndId called without component id');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });
    });

    describe('on', function() {
        it('should delegate to events.on', function() {
            const callback = jasmine.createSpy('callback');
            const mockCleanup = jasmine.createSpy('cleanup');
            window.liveflux.events.on.and.returnValue(mockCleanup);
            
            const result = window.liveflux.on('test-event', callback);
            
            expect(window.liveflux.events.on).toHaveBeenCalledWith('test-event', callback);
            expect(result).toBe(mockCleanup);
        });
    });
});
