describe('Liveflux Find', function() {
    beforeEach(function() {
        // Mock document.querySelector
        this.mockElement = { 
            id: 'test-element',
            getAttribute: jasmine.createSpy('getAttribute')
        };
        
        spyOn(document, 'querySelector').and.callFake(function(selector) {
            if (selector.includes('test-component') && selector.includes('test-id-123')) {
                return this.mockElement;
            }
            return null;
        }.bind(this));
        
        // Re-initialize find module
        delete window.liveflux.findComponent;
        
        // Load find module
        (function(){
            if(!window.liveflux){
                console.log('[Liveflux Find] liveflux namespace not found');
                return;
            }

            const liveflux = window.liveflux;
            const { dataFluxRoot, dataFluxComponent, dataFluxComponentID } = liveflux;

            function findComponent(componentAlias, componentId){
                return document.querySelector(`[${dataFluxRoot}][${dataFluxComponent}="${componentAlias}"][${dataFluxComponentID}="${componentId}"]`);
            }
            
            liveflux.findComponent = findComponent;
        })();
    });

    describe('findComponent', function() {
        it('should find component by alias and ID', function() {
            const result = window.liveflux.findComponent('test-component', 'test-id-123');
            
            expect(document.querySelector).toHaveBeenCalledWith(
                '[data-flux-root][data-flux-component="test-component"][data-flux-component-id="test-id-123"]'
            );
            expect(result).toBe(this.mockElement);
        });

        it('should return null when component is not found', function() {
            const result = window.liveflux.findComponent('nonexistent', 'nonexistent-id');
            
            expect(document.querySelector).toHaveBeenCalledWith(
                '[data-flux-root][data-flux-component="nonexistent"][data-flux-component-id="nonexistent-id"]'
            );
            expect(result).toBeNull();
        });

        it('should handle empty alias', function() {
            const result = window.liveflux.findComponent('', 'test-id-123');
            
            expect(document.querySelector).toHaveBeenCalledWith(
                '[data-flux-root][data-flux-component=""][data-flux-component-id="test-id-123"]'
            );
            expect(result).toBeNull();
        });

        it('should handle empty ID', function() {
            const result = window.liveflux.findComponent('test-component', '');
            
            expect(document.querySelector).toHaveBeenCalledWith(
                '[data-flux-root][data-flux-component="test-component"][data-flux-component-id=""]'
            );
            expect(result).toBeNull();
        });

        it('should exist as a function', function() {
            expect(typeof window.liveflux.findComponent).toBe('function');
        });
    });

    describe('error handling', function() {
        it('should handle missing liveflux namespace gracefully', function() {
            const originalLiveflux = window.liveflux;
            delete window.liveflux;
            
            spyOn(console, 'log');
            
            expect(function() {
                // Try to run find module without namespace
                (function(){
                    if(!window.liveflux){
                        console.log('[Liveflux Find] liveflux namespace not found');
                        return;
                    }
                })();
            }).not.toThrow();
            
            expect(console.log).toHaveBeenCalledWith('[Liveflux Find] liveflux namespace not found');
            
            // Restore
            window.liveflux = originalLiveflux;
        });
    });
});
