let originalDataFluxSelect;

beforeAll(function() {
    originalDataFluxSelect = window.liveflux.dataFluxSelect;
    window.liveflux.dataFluxSelect = window.liveflux.dataFluxSelect || 'data-flux-select';
});

afterAll(function() {
    if (typeof originalDataFluxSelect === 'undefined') {
        delete window.liveflux.dataFluxSelect;
    } else {
        window.liveflux.dataFluxSelect = originalDataFluxSelect;
    }
});

describe('Liveflux data-flux-select utilities', function() {
    describe('readSelectAttribute', function() {
        it('returns data attribute when present', function() {
            const btn = document.createElement('button');
            btn.setAttribute('data-flux-select', '#details');

            const value = window.liveflux.readSelectAttribute(btn);

            expect(value).toBe('#details');
        });

        it('falls back to flux-select attribute', function() {
            const btn = document.createElement('button');
            btn.setAttribute('flux-select', '.card');

            const value = window.liveflux.readSelectAttribute(btn);

            expect(value).toBe('.card');
        });

        it('returns empty string when no select attribute exists', function() {
            const btn = document.createElement('button');

            const value = window.liveflux.readSelectAttribute(btn);

            expect(value).toBe('');
        });
    });

    describe('extractSelectedFragment', function() {
        it('returns the first matching selector fragment', function() {
            const html = '<div><section id="summary">Summary</section><div id="details">Details</div></div>';

            const fragment = window.liveflux.extractSelectedFragment(html, '#details');

            expect(fragment).toBe('<div id="details">Details</div>');
        });

        it('picks the first match from comma-separated selectors', function() {
            const html = '<div><article class="primary">Primary</article><article class="secondary">Secondary</article></div>';

            const fragment = window.liveflux.extractSelectedFragment(html, '#missing, .secondary');

            expect(fragment).toBe('<article class="secondary">Secondary</article>');
        });

        it('falls back to original HTML when no selector matches', function() {
            const html = '<div><p>No match</p></div>';

            const fragment = window.liveflux.extractSelectedFragment(html, '#unknown');

            expect(fragment).toBe(html);
        });
    });
});

describe('Liveflux data-flux-select integration', function() {
    let originalPost;
    let extractSpy;
    let testContainer;

    beforeEach(function() {
        originalPost = window.liveflux.post;
        window.liveflux.post = jasmine.createSpy('post').and.returnValue(Promise.resolve({
            html: '<div data-flux-root="1" data-flux-component="test" data-flux-component-id="123">'
                + '<div id="partial">Filtered</div>'
                + '</div>'
        }));

        extractSpy = spyOn(window.liveflux, 'extractSelectedFragment').and.callThrough();

        testContainer = document.createElement('div');
        testContainer.id = 'select-test-container';
        document.body.appendChild(testContainer);
    });

    afterEach(function() {
        window.liveflux.post = originalPost;
        if (extractSpy && extractSpy.and && extractSpy.and.originalFn) {
            window.liveflux.extractSelectedFragment = extractSpy.and.originalFn;
        }
        if (testContainer && testContainer.parentNode) {
            testContainer.parentNode.removeChild(testContainer);
        }
        extractSpy = null;
    });

    it('filters action responses before swap when trigger defines data-flux-select', function(done) {
        testContainer.innerHTML = `
            <div data-flux-root="1" data-flux-component="test" data-flux-component-id="123">
                <div id="partial">Original</div>
                <button id="select-action" data-flux-action="refresh" data-flux-select="#partial">Refresh</button>
                <div class="wrapper">static</div>
            </div>
        `;

        const btn = document.getElementById('select-action');
        const originalRoot = document.querySelector('[data-flux-component-id="123"]');
        const event = {
            target: btn,
            preventDefault: jasmine.createSpy('preventDefault')
        };

        window.liveflux.handleActionClick(event);

        setTimeout(function() {
            expect(extractSpy).toHaveBeenCalledWith(
                '<div data-flux-root="1" data-flux-component="test" data-flux-component-id="123">'
                + '<div id="partial">Filtered</div>'
                + '</div>',
                '#partial'
            );

            const fragment = document.querySelector('#partial');
            expect(fragment && fragment.textContent.trim()).toBe('Filtered');
            const currentRoot = document.querySelector('[data-flux-component-id="123"]');
            expect(currentRoot).toBe(originalRoot);
            expect(document.querySelector('.wrapper').textContent.trim()).toBe('static');
            done();
        }, 50);
    });

    it('filters form submissions when submitter defines data-flux-select', function(done) {
        testContainer.innerHTML = `
            <div data-flux-root="1" data-flux-component="test" data-flux-component-id="123">
                <form id="select-form">
                    <button type="submit" data-flux-action="save" data-flux-select="#partial">Save</button>
                </form>
                <div id="partial">Original</div>
                <div class="wrapper">static</div>
            </div>
        `;

        const form = document.getElementById('select-form');
        const originalRoot = document.querySelector('[data-flux-component-id="123"]');
        const event = {
            target: form,
            preventDefault: jasmine.createSpy('preventDefault')
        };

        window.liveflux.handleFormSubmit(event);

        setTimeout(function() {
            expect(extractSpy).toHaveBeenCalledWith(jasmine.any(String), '#partial');

            const fragment = document.querySelector('#partial');
            expect(fragment && fragment.textContent.trim()).toBe('Filtered');
            const currentRoot = document.querySelector('[data-flux-component-id="123"]');
            expect(currentRoot).toBe(originalRoot);
            expect(document.querySelector('.wrapper').textContent.trim()).toBe('static');
            done();
        }, 50);
    });
});
