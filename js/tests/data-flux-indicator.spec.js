'use strict';

const REQUEST_CLASS = 'flux-request';

describe('Liveflux data-flux-indicator helpers', function() {
    let testContainer;
    let root;
    let trigger;

    beforeEach(function() {
        testContainer = document.createElement('div');
        testContainer.id = 'indicator-test-container';
        document.body.appendChild(testContainer);

        root = document.createElement('div');
        root.setAttribute('data-flux-component-kind', 'demo.component');
        root.setAttribute('data-flux-component-id', 'demo-1');

        trigger = document.createElement('button');
        trigger.type = 'button';
        trigger.setAttribute('data-flux-action', 'fetch');
        trigger.textContent = 'Trigger';

        root.appendChild(trigger);
        testContainer.appendChild(root);
    });

    afterEach(function() {
        if (testContainer && testContainer.parentNode) {
            testContainer.parentNode.removeChild(testContainer);
        }
        testContainer = null;
        root = null;
        trigger = null;
    });

    it('activates trigger and referenced indicator targets', function() {
        const indicator = document.createElement('span');
        indicator.id = 'secondary-indicator';
        root.appendChild(indicator);

        trigger.setAttribute('data-flux-indicator', 'this, #secondary-indicator');

        const elements = window.liveflux.startRequestIndicators(trigger, root);

        expect(elements).toEqual(jasmine.arrayContaining([trigger, indicator]));
        expect(trigger.classList.contains(REQUEST_CLASS)).toBeTrue();
        expect(indicator.classList.contains(REQUEST_CLASS)).toBeTrue();

        window.liveflux.endRequestIndicators(elements);

        expect(trigger.classList.contains(REQUEST_CLASS)).toBeFalse();
        expect(indicator.classList.contains(REQUEST_CLASS)).toBeFalse();
    });

    it('falls back to .flux-indicator elements when attribute is missing', function() {
        trigger.removeAttribute('data-flux-indicator');

        const fallback = document.createElement('div');
        fallback.className = 'flux-indicator';
        root.appendChild(fallback);

        const elements = window.liveflux.startRequestIndicators(trigger, root);

        expect(elements).toEqual(jasmine.arrayContaining([trigger, fallback]));
        expect(fallback.classList.contains(REQUEST_CLASS)).toBeTrue();

        window.liveflux.endRequestIndicators(elements);
        expect(fallback.classList.contains(REQUEST_CLASS)).toBeFalse();
    });

    it('temporarily shows targeted elements hidden via inline display when they lack .flux-indicator', function() {
        const hidden = document.createElement('div');
        hidden.id = 'hidden-indicator';
        hidden.style.display = 'none';
        root.appendChild(hidden);

        trigger.setAttribute('data-flux-indicator', '#hidden-indicator');

        const elements = window.liveflux.startRequestIndicators(trigger, root);

        expect(hidden.style.display).toBe('inline-block');
        expect(hidden.getAttribute('data-liveflux-indicator-original-display')).toBe('none');

        window.liveflux.endRequestIndicators(elements);

        expect(hidden.style.display).toBe('none');
        expect(hidden.hasAttribute('data-liveflux-indicator-original-display')).toBeFalse();
    });

    it('does not override display for elements already marked with .flux-indicator', function() {
        const indicator = document.createElement('div');
        indicator.id = 'local-indicator';
        indicator.className = 'flux-indicator';
        indicator.style.display = 'none';
        root.appendChild(indicator);

        trigger.setAttribute('data-flux-indicator', '#local-indicator');

        const elements = window.liveflux.startRequestIndicators(trigger, root);

        expect(indicator.style.display).toBe('none');
        expect(indicator.hasAttribute('data-liveflux-indicator-original-display')).toBeFalse();
        expect(indicator.classList.contains(REQUEST_CLASS)).toBeTrue();

        window.liveflux.endRequestIndicators(elements);
        expect(indicator.classList.contains(REQUEST_CLASS)).toBeFalse();
    });
});
