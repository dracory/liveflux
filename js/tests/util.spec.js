describe('Liveflux Util', function() {
    describe('resolveComponentMetadata', function() {
        let testContainer;
        let rootSelector;

        beforeEach(function() {
            // Create a test container instead of replacing body
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            rootSelector = window.liveflux.getComponentRootSelector();

            testContainer.innerHTML = `
                <div data-flux-component-kind="test-component" data-flux-component-id="test-id-123">
                    <button id="btn-inside" data-flux-action="increment">Click me</button>
                </div>
                <button id="btn-outside" 
                    data-flux-target-kind="external-component" 
                    data-flux-target-id="external-id-456"
                    data-flux-action="submit">
                    External Button
                </button>
            `;
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should resolve metadata from nearest root using data attributes', function() {
            const btn = document.getElementById('btn-inside');
            
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            
            expect(metadata).not.toBeNull();
            expect(metadata.comp).toBe('test-component');
            expect(metadata.id).toBe('test-id-123');
            expect(metadata.root).toBe(document.querySelector(rootSelector));
        });

        it('should resolve metadata from explicit button attributes', function() {
            const btn = document.getElementById('btn-outside');
            
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            
            expect(metadata).not.toBeNull();
            expect(metadata.comp).toBe('external-component');
            expect(metadata.id).toBe('external-id-456');
            expect(metadata.root).toBeNull();
        });

        it('should return null when button is null', function() {
            const metadata = window.liveflux.resolveComponentMetadata(null, rootSelector);
            expect(metadata).toBeNull();
        });

        it('should return null when no metadata is found', function() {
            const btn = document.createElement('button');
            testContainer.appendChild(btn);
            
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            expect(metadata).toBeNull();
        });

        it('should handle missing component attribute', function() {
            testContainer.innerHTML = `
                <div data-flux-component-id="test-id-123">
                    <button id="btn-incomplete" data-flux-action="test">Click</button>
                </div>
            `;
            
            const btn = document.getElementById('btn-incomplete');
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            
            expect(metadata).toBeNull();
        });

        it('should handle missing component-id attribute', function() {
            testContainer.innerHTML = `
                <div data-flux-component-kind="test-component">
                    <button id="btn-incomplete" data-flux-action="test">Click</button>
                </div>
            `;
            
            const btn = document.getElementById('btn-incomplete');
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            
            expect(metadata).toBeNull();
        });
    });

    describe('serializeElement', function() {
        let testContainer;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <form id="test-form">
                    <input type="text" name="username" value="john">
                    <input type="email" name="email" value="john@example.com">
                    <input type="checkbox" name="subscribe" checked>
                    <select name="country">
                        <option value="us" selected>USA</option>
                        <option value="uk">UK</option>
                    </select>
                </form>
            `;
            document.body.appendChild(testContainer);
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should serialize form fields', function() {
            const form = document.getElementById('test-form');
            const fields = window.liveflux.serializeElement(form);
            
            expect(fields.username).toBe('john');
            expect(fields.email).toBe('john@example.com');
            expect(fields.subscribe).toBe('on');
            expect(fields.country).toBe('us');
        });

        it('should handle empty form', function() {
            testContainer.innerHTML = '<form id="empty-form"></form>';
            const form = document.getElementById('empty-form');
            const fields = window.liveflux.serializeElement(form);
            
            expect(Object.keys(fields).length).toBe(0);
        });
    });

    describe('collectAllFields', function() {
        let testContainer;
        let rootSelector;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <div data-flux-component-kind="test" data-flux-component-id="123">
                    <form id="main-form">
                        <input type="text" name="field1" value="value1">
                    </form>
                    <button id="btn-with-include" 
                        data-flux-include="#external-input"
                        data-flux-action="submit">
                        Submit
                    </button>
                </div>
                <input type="text" id="external-input" name="field2" value="value2">
            `;
            document.body.appendChild(testContainer);
            rootSelector = window.liveflux.getComponentRootSelector();
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should collect fields from form', function() {
            const btn = document.getElementById('btn-with-include');
            const root = document.querySelector(rootSelector);
            const form = document.getElementById('main-form');
            
            const fields = window.liveflux.collectAllFields(btn, root, form);
            
            expect(fields.field1).toBe('value1');
        });

        it('should include fields from data-flux-include selector', function() {
            const btn = document.getElementById('btn-with-include');
            const root = document.querySelector(rootSelector);
            const form = document.getElementById('main-form');
            
            const fields = window.liveflux.collectAllFields(btn, root, form);
            
            expect(fields.field1).toBe('value1');
            expect(fields.field2).toBe('value2');
        });

        it('should handle null button', function() {
            const fields = window.liveflux.collectAllFields(null, null, null);
            expect(fields).toEqual({});
        });
    });

    describe('request indicators', function() {
        let trigger, root, indicator, testContainer, rootSelector;

        beforeEach(function() {
            testContainer = document.createElement('div');
            testContainer.id = 'test-container';
            testContainer.innerHTML = `
                <div id="component" data-flux-component-kind="test" data-flux-component-id="abc">
                    <button id="action" data-flux-indicator="this, #spinner">Action</button>
                    <span id="spinner" class="flux-indicator"></span>
                </div>
            `;
            document.body.appendChild(testContainer);
            trigger = document.getElementById('action');
            root = document.getElementById('component');
            indicator = document.getElementById('spinner');
            rootSelector = window.liveflux.getComponentRootSelector();
        });

        afterEach(function() {
            if (testContainer && testContainer.parentNode) {
                testContainer.parentNode.removeChild(testContainer);
            }
        });

        it('should add request classes to trigger and referenced indicators', function() {
            const els = window.liveflux.startRequestIndicators(trigger, root);

            expect(Array.isArray(els)).toBeTrue();
            expect(trigger.classList.contains('flux-request')).toBeTrue();
            expect(indicator.classList.contains('flux-request')).toBeTrue();

            window.liveflux.endRequestIndicators(els);
        });

        it('should fall back to .flux-indicator elements when attribute missing', function() {
            trigger.removeAttribute('data-flux-indicator');
            indicator.className = 'flux-indicator';

            const els = window.liveflux.startRequestIndicators(trigger, root);

            expect(indicator.classList.contains('flux-request')).toBeTrue();

            window.liveflux.endRequestIndicators(els);

            expect(indicator.classList.contains('flux-request')).toBeFalse();
        });

        it('should restore inline style when non-indicator element was hidden inline', function() {
            indicator.className = '';
            indicator.style.display = 'none';

            const els = window.liveflux.startRequestIndicators(trigger, root);

            expect(indicator.style.display).toBe('inline-block');

            window.liveflux.endRequestIndicators(els);

            expect(indicator.style.display).toBe('none');
        });

        it('should remove display override when element had no inline display', function() {
            indicator.className = '';
            indicator.style.display = '';
            const originalComputed = window.getComputedStyle(indicator).display;
            indicator.style.display = 'none';

            const els = window.liveflux.startRequestIndicators(trigger, root);

            expect(indicator.style.display).toBe('inline-block');

            indicator.style.display = '';
            window.liveflux.endRequestIndicators(els);

            expect(indicator.style.display).toBe('');
            expect(window.getComputedStyle(indicator).display).toBe(originalComputed);
        });
    });

    describe('isComponentRootNode', function() {
        let element;

        beforeEach(function() {
            element = document.createElement('div');
        });

        it('returns true when both component kind and id attributes are present', function() {
            element.setAttribute('data-flux-component-kind', 'test.kind');
            element.setAttribute('data-flux-component-id', 'abc123');

            expect(window.liveflux.isComponentRootNode(element)).toBeTrue();
        });

        it('returns false when component kind attribute is missing', function() {
            element.setAttribute('data-flux-component-id', 'abc123');

            expect(window.liveflux.isComponentRootNode(element)).toBeFalse();
        });

        it('returns false when component id attribute is missing', function() {
            element.setAttribute('data-flux-component-kind', 'test.kind');

            expect(window.liveflux.isComponentRootNode(element)).toBeFalse();
        });

        it('returns false for null or non-element input', function() {
            expect(window.liveflux.isComponentRootNode(null)).toBeFalse();
        });
    });

    describe('getComponentRootSelector', function() {
        let originalKindAttr;
        let originalIdAttr;

        beforeEach(function() {
            originalKindAttr = window.liveflux.dataFluxComponentKind;
            originalIdAttr = window.liveflux.dataFluxComponentID;
        });

        afterEach(function() {
            window.liveflux.dataFluxComponentKind = originalKindAttr;
            window.liveflux.dataFluxComponentID = originalIdAttr;
        });

        it('returns default selector when custom attributes are not set', function() {
            window.liveflux.dataFluxComponentKind = undefined;
            window.liveflux.dataFluxComponentID = undefined;

            expect(window.liveflux.getComponentRootSelector()).toBe('[data-flux-component-kind][data-flux-component-id]');
        });

        it('returns selector using custom attribute names', function() {
            window.liveflux.dataFluxComponentKind = 'custom-kind';
            window.liveflux.dataFluxComponentID = 'custom-id';

            expect(window.liveflux.getComponentRootSelector()).toBe('[custom-kind][custom-id]');
        });
    });
});
