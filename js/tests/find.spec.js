describe('Liveflux Find', function() {
    beforeEach(function() {
        // Re-initialize find module
        delete window.liveflux.findComponent;
        delete window.liveflux.findAllComponents;
        
        // Load find module
        (function(){
            if(!window.liveflux){
                console.log('[Liveflux Find] liveflux namespace not found');
                return;
            }

            const liveflux = window.liveflux;
            const { dataFluxComponentKind, dataFluxComponentID } = liveflux;

            function queryComponentRoots(){
                const selector = liveflux.getComponentRootSelector ? liveflux.getComponentRootSelector() : `[${dataFluxComponentKind}][${dataFluxComponentID}]`;
                return Array.from(document.querySelectorAll(selector));
            }

            function findAllComponents(predicate){
                const elements = queryComponentRoots();
                if(typeof predicate === 'function'){
                    return elements.filter(predicate);
                }
                return elements;
            }

            function findComponent(componentKind, componentId){
                if(!componentKind || !componentId){
                    return null;
                }
                return findAllComponents(function(el){
                    return el.getAttribute(dataFluxComponentKind) === componentKind && el.getAttribute(dataFluxComponentID) === componentId;
                })[0] || null;
            }
            
            liveflux.findAllComponents = findAllComponents;
            liveflux.findComponent = findComponent;
        })();

        const liveflux = window.liveflux;
        this.component = {
            id: 'test-element',
            getAttribute: jasmine.createSpy('getAttribute').and.callFake(function(attr){
                if(attr === (liveflux.dataFluxComponentKind || 'data-flux-component-kind')) return 'test-component';
                if(attr === (liveflux.dataFluxComponentID || 'data-flux-component-id')) return 'test-id-123';
                return null;
            })
        };

        this.rootSelector = liveflux.getComponentRootSelector ? liveflux.getComponentRootSelector() : '[data-flux-component-kind][data-flux-component-id]';

        spyOn(document, 'querySelectorAll').and.callFake(function(selector){
            if(selector === this.rootSelector){
                return [this.component];
            }
            return [];
        }.bind(this));
    });

    describe('findAllComponents', function() {
        it('should return all components when no predicate provided', function() {
            const otherComponent = { id: 'other', getAttribute: jasmine.createSpy('getAttribute').and.returnValue(null) };
            document.querySelectorAll.and.returnValue([this.component, otherComponent]);

            const result = window.liveflux.findAllComponents();

            expect(document.querySelectorAll).toHaveBeenCalledWith(this.rootSelector);
            expect(result).toEqual([this.component, otherComponent]);
        });

        it('should filter components when predicate provided', function() {
            const otherComponent = {
                id: 'other',
                getAttribute: jasmine.createSpy('getAttribute').and.callFake(function(attr){
                    if(attr === (window.liveflux.dataFluxComponentKind || 'data-flux-component-kind')) return 'other-kind';
                    if(attr === (window.liveflux.dataFluxComponentID || 'data-flux-component-id')) return 'other-id';
                    return null;
                })
            };
            document.querySelectorAll.and.returnValue([this.component, otherComponent]);

            const result = window.liveflux.findAllComponents(function(el){
                return el.getAttribute(window.liveflux.dataFluxComponentKind || 'data-flux-component-kind') === 'test-component';
            });

            expect(result).toEqual([this.component]);
        });
    });

    describe('findComponent', function() {
        it('should find component by kind and ID', function() {
            const result = window.liveflux.findComponent('test-component', 'test-id-123');
            
            expect(document.querySelectorAll).toHaveBeenCalledWith(this.rootSelector);
            expect(result).toBe(this.component);
        });

        it('should return null when component is not found', function() {
            document.querySelectorAll.and.returnValue([]);

            const result = window.liveflux.findComponent('nonexistent', 'nonexistent-id');
            
            expect(document.querySelectorAll).toHaveBeenCalledWith(this.rootSelector);
            expect(result).toBeNull();
        });

        it('should handle empty ID', function() {
            const result = window.liveflux.findComponent('test-component', '');
            
            expect(document.querySelectorAll).not.toHaveBeenCalled();
            expect(result).toBeNull();
        });

        it('should handle empty kind', function() {
            const result = window.liveflux.findComponent('', 'test-id-123');
            
            expect(document.querySelectorAll).not.toHaveBeenCalled();
            expect(result).toBeNull();
        });

        it('should exist as a function', function() {
            expect(typeof window.liveflux.findComponent).toBe('function');
        });

        it('should return null when component kind or id missing', function() {
            expect(window.liveflux.findComponent(null, 'id')).toBeNull();
            expect(window.liveflux.findComponent('kind', null)).toBeNull();
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
