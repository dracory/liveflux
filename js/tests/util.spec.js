describe('Liveflux Util', function() {
    describe('resolveComponentMetadata', function() {
        beforeEach(function() {
            // Create a test DOM structure
            document.body.innerHTML = `
                <div data-flux-root="1" data-flux-component="test-component" data-flux-component-id="test-id-123">
                    <button id="btn-inside" data-flux-action="increment">Click me</button>
                </div>
                <button id="btn-outside" 
                    data-flux-component-type="external-component" 
                    data-flux-component-id="external-id-456"
                    data-flux-action="submit">
                    External Button
                </button>
                <button id="btn-with-root-id" 
                    data-flux-root-id="test-id-123"
                    data-flux-action="update">
                    Button with Root ID
                </button>
            `;
        });

        afterEach(function() {
            document.body.innerHTML = '';
        });

        it('should resolve metadata from nearest root using data attributes', function() {
            const btn = document.getElementById('btn-inside');
            const rootSelector = '[data-flux-root]';
            
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            
            expect(metadata).not.toBeNull();
            expect(metadata.comp).toBe('test-component');
            expect(metadata.id).toBe('test-id-123');
            expect(metadata.root).not.toBeNull();
            expect(metadata.root.getAttribute('data-flux-root')).toBe('1');
        });

        it('should resolve metadata from explicit button attributes', function() {
            const btn = document.getElementById('btn-outside');
            const rootSelector = '[data-flux-root]';
            
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            
            expect(metadata).not.toBeNull();
            expect(metadata.comp).toBe('external-component');
            expect(metadata.id).toBe('external-id-456');
            expect(metadata.root).toBeNull();
        });

        it('should resolve metadata from root ID reference', function() {
            const btn = document.getElementById('btn-with-root-id');
            const rootSelector = '[data-flux-root]';
            
            // Add an ID to the root element
            const root = document.querySelector('[data-flux-root]');
            root.id = 'test-id-123';
            
            const metadata = window.liveflux.resolveComponentMetadata(btn, rootSelector);
            
            expect(metadata).not.toBeNull();
            expect(metadata.comp).toBe('test-component');
            expect(metadata.id).toBe('test-id-123');
            expect(metadata.root).not.toBeNull();
        });

        it('should return null when button is null', function() {
            const metadata = window.liveflux.resolveComponentMetadata(null, '[data-flux-root]');
            expect(metadata).toBeNull();
        });

        it('should return null when no metadata is found', function() {
            const btn = document.createElement('button');
            document.body.appendChild(btn);
            
            const metadata = window.liveflux.resolveComponentMetadata(btn, '[data-flux-root]');
            expect(metadata).toBeNull();
        });

        it('should handle missing component attribute', function() {
            document.body.innerHTML = `
                <div data-flux-root="1" data-flux-component-id="test-id-123">
                    <button id="btn-incomplete" data-flux-action="test">Click</button>
                </div>
            `;
            
            const btn = document.getElementById('btn-incomplete');
            const metadata = window.liveflux.resolveComponentMetadata(btn, '[data-flux-root]');
            
            expect(metadata).toBeNull();
        });

        it('should handle missing component-id attribute', function() {
            document.body.innerHTML = `
                <div data-flux-root="1" data-flux-component="test-component">
                    <button id="btn-incomplete" data-flux-action="test">Click</button>
                </div>
            `;
            
            const btn = document.getElementById('btn-incomplete');
            const metadata = window.liveflux.resolveComponentMetadata(btn, '[data-flux-root]');
            
            expect(metadata).toBeNull();
        });
    });

    describe('serializeElement', function() {
        beforeEach(function() {
            document.body.innerHTML = `
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
        });

        afterEach(function() {
            document.body.innerHTML = '';
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
            document.body.innerHTML = '<form id="empty-form"></form>';
            const form = document.getElementById('empty-form');
            const fields = window.liveflux.serializeElement(form);
            
            expect(Object.keys(fields).length).toBe(0);
        });
    });

    describe('collectAllFields', function() {
        beforeEach(function() {
            document.body.innerHTML = `
                <div data-flux-root="1" data-flux-component="test" data-flux-component-id="123">
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
        });

        afterEach(function() {
            document.body.innerHTML = '';
        });

        it('should collect fields from form', function() {
            const btn = document.getElementById('btn-with-include');
            const root = document.querySelector('[data-flux-root]');
            const form = document.getElementById('main-form');
            
            const fields = window.liveflux.collectAllFields(btn, root, form);
            
            expect(fields.field1).toBe('value1');
        });

        it('should include fields from data-flux-include selector', function() {
            const btn = document.getElementById('btn-with-include');
            const root = document.querySelector('[data-flux-root]');
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
});
