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
            if (selector.includes('test-component')) {
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

    describe('dispatchToAlias', function() {
        it('should dispatch event to all components with alias', function() {
            const eventData = { test: 'data' };
            
            window.liveflux.dispatchToAlias('test-component', 'test-event', eventData);
            
            expect(window.liveflux.events.dispatch).toHaveBeenCalledWith('test-event', {
                test: 'data',
                __target: 'test-component'
            });
        });

        it('should handle missing component alias', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToAlias(null, 'test-event', {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToAlias called without component alias');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });

        it('should handle missing event name', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToAlias('test-component', null, {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToAlias called without event name');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });
    });

    describe('dispatchToAliasAndId', function() {
        it('should dispatch event to specific component by alias and ID', function() {
            const eventData = { test: 'data' };
            
            window.liveflux.dispatchToAliasAndId('test-component', 'test-id-123', 'test-event', eventData);
            
            expect(window.liveflux.findComponent).toHaveBeenCalledWith('test-component', 'test-id-123');
            expect(window.liveflux.events.dispatch).toHaveBeenCalledWith('test-event', {
                test: 'data',
                __target: 'test-component',
                __target_id: 'test-id-123'
            });
        });

        it('should handle missing component alias', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToAliasAndId(null, 'test-id-123', 'test-event', {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToAliasAndId called without component alias');
            expect(window.liveflux.events.dispatch).not.toHaveBeenCalled();
        });

        it('should handle missing component ID', function() {
            spyOn(console, 'warn');
            
            window.liveflux.dispatchToAliasAndId('test-component', null, 'test-event', {});
            
            expect(console.warn).toHaveBeenCalledWith('[Liveflux Events] dispatchToAliasAndId called without component id');
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
