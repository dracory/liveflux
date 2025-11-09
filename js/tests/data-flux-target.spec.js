describe('Liveflux Target Updates', function() {
  let originalExecuteScripts;

  beforeEach(function() {
    // Mock executeScripts
    originalExecuteScripts = window.liveflux.executeScripts;
    if (!window.liveflux.executeScripts || !window.liveflux.executeScripts.and) {
      window.liveflux.executeScripts = jasmine.createSpy('executeScripts');
    } else {
      window.liveflux.executeScripts.calls.reset();
    }
  });

  afterEach(function() {
    if (typeof originalExecuteScripts === 'undefined') {
      delete window.liveflux.executeScripts;
    } else {
      window.liveflux.executeScripts = originalExecuteScripts;
    }
  });

  describe('hasTargetTemplates', function() {
    it('should detect template with data-flux-target', function() {
      const html = '<template data-flux-target="#foo">content</template>';
      expect(window.liveflux.hasTargetTemplates(html)).toBe(true);
    });

    it('should detect template with data-flux-component', function() {
      const html = '<template data-flux-component-kind="cart">content</template>';
      expect(window.liveflux.hasTargetTemplates(html)).toBe(true);
    });

    it('should return false for regular HTML', function() {
      const html = '<div>regular content</div>';
      expect(window.liveflux.hasTargetTemplates(html)).toBe(false);
    });

    it('should return false for empty string', function() {
      expect(window.liveflux.hasTargetTemplates('')).toBe(false);
    });
  });

  describe('applyTargets', function() {
    let testContainer;

    beforeEach(function() {
      testContainer = document.createElement('div');
      testContainer.id = 'test-container';
      document.body.appendChild(testContainer);
    });

    afterEach(function() {
      if (testContainer && testContainer.parentNode) {
        testContainer.parentNode.removeChild(testContainer);
      }
    });

    it('should apply single target fragment with replace mode', function() {
      // Setup component root
      const root = document.createElement('div');
      root.setAttribute('data-flux-root', '1');
      root.setAttribute('data-flux-component-kind', 'cart');
      root.setAttribute('data-flux-component-id', 'abc123');
      root.innerHTML = '<div id="total">$99</div>';
      testContainer.appendChild(root);

      // Template response
      const html = `
        <template data-flux-target="#total" data-flux-swap="replace">
          <div id="total">$125</div>
        </template>
      `;

      const result = window.liveflux.applyTargets(html, root);
      
      expect(result).toBeNull(); // Success
      expect(root.querySelector('#total').textContent).toBe('$125');
      expect(window.liveflux.executeScripts).toHaveBeenCalled();
    });

    it('should apply multiple fragments in order', function() {
      const root = document.createElement('div');
      root.setAttribute('data-flux-root', '1');
      root.innerHTML = `
        <div id="total">$99</div>
        <ul class="items"></ul>
      `;
      testContainer.appendChild(root);

      const html = `
        <template data-flux-target="#total" data-flux-swap="replace">
          <div id="total">$125</div>
        </template>
        <template data-flux-target=".items" data-flux-swap="inner">
          <li>Item 1</li>
        </template>
      `;

      const result = window.liveflux.applyTargets(html, root);
      
      expect(result).toBeNull();
      expect(root.querySelector('#total').textContent).toBe('$125');
      expect(root.querySelector('.items').innerHTML).toContain('Item 1');
    });

    it('should apply full component replacement first', function() {
      const root = document.createElement('div');
      root.setAttribute('data-flux-root', '1');
      root.setAttribute('data-flux-component-kind', 'cart');
      root.setAttribute('data-flux-component-id', 'abc123');
      root.innerHTML = '<div>old content</div>';
      testContainer.appendChild(root);

      const html = `
        <template data-flux-component-kind="cart" data-flux-component-id="abc123">
          <div data-flux-root="1" data-flux-component-kind="cart" data-flux-component-id="abc123">
            <div id="new">new content</div>
          </div>
        </template>
      `;

      const result = window.liveflux.applyTargets(html, root);
      
      expect(result).toBeNull();
      const newRoot = testContainer.querySelector('[data-flux-root]');
      expect(newRoot.querySelector('#new').textContent).toBe('new content');
    });

    it('should validate component metadata', function() {
      const root = document.createElement('div');
      root.setAttribute('data-flux-root', '1');
      root.setAttribute('data-flux-component-kind', 'cart');
      root.setAttribute('data-flux-component-id', 'abc123');
      root.innerHTML = '<div id="total">$99</div>';
      testContainer.appendChild(root);

      // Mismatched component ID
      const html = `
        <template data-flux-target="#total" data-flux-swap="replace" data-flux-component-kind="cart" data-flux-component-id="wrong">
          <div id="total">$125</div>
        </template>
      `;

      spyOn(console, 'warn');
      const result = window.liveflux.applyTargets(html, root);
      
      expect(result).toBe(html); // Fallback
      expect(console.warn).toHaveBeenCalledWith(jasmine.stringContaining('Component ID mismatch'));
    });

    it('should support inner swap mode', function() {
      const root = document.createElement('div');
      root.innerHTML = '<div id="container"><p>old</p></div>';
      testContainer.appendChild(root);

      const html = `
        <template data-flux-target="#container" data-flux-swap="inner">
          <div><p>new</p></div>
        </template>
      `;

      window.liveflux.applyTargets(html, root);
      
      // Inner mode replaces innerHTML with fragment.innerHTML
      expect(root.querySelector('#container').innerHTML).toContain('<p>new</p>');
      expect(root.querySelector('#container').id).toBe('container'); // Container still exists
    });

    it('should support beforeend swap mode (append)', function() {
      const root = document.createElement('div');
      root.innerHTML = '<ul id="list"><li>Item 1</li></ul>';
      testContainer.appendChild(root);

      const html = `
        <template data-flux-target="#list" data-flux-swap="beforeend">
          <li>Item 2</li>
        </template>
      `;

      window.liveflux.applyTargets(html, root);
      
      const items = root.querySelectorAll('#list li');
      expect(items.length).toBe(2);
      expect(items[1].textContent).toBe('Item 2');
    });

    it('should return fallback HTML if selector not found', function() {
      const root = document.createElement('div');
      root.innerHTML = '<div id="exists">content</div>';
      testContainer.appendChild(root);

      const html = `
        <template data-flux-target="#missing">
          <div>new</div>
        </template>
      `;

      spyOn(console, 'warn');
      const result = window.liveflux.applyTargets(html, root);
      
      expect(result).toBe(html); // Fallback
      expect(console.warn).toHaveBeenCalledWith(jasmine.stringContaining('Selector not found'));
    });

    it('should return original HTML if no templates found', function() {
      const root = document.createElement('div');
      testContainer.appendChild(root);

      const html = '<div>regular content</div>';
      const result = window.liveflux.applyTargets(html, root);
      
      expect(result).toBe(html);
    });
  });

  describe('enableTargetSupport', function() {
    it('should add X-Liveflux-Target header', function() {
      window.liveflux.enableTargetSupport();
      
      expect(window.liveflux.headers['X-Liveflux-Target']).toBe('enabled');
    });

    it('should create headers object if missing', function() {
      delete window.liveflux.headers;
      window.liveflux.enableTargetSupport();
      
      expect(window.liveflux.headers).toBeDefined();
      expect(window.liveflux.headers['X-Liveflux-Target']).toBe('enabled');
    });
  });

  describe('disableTargetSupport', function() {
    it('should remove X-Liveflux-Target header', function() {
      window.liveflux.headers = { 'X-Liveflux-Target': 'enabled' };
      window.liveflux.disableTargetSupport();
      
      expect(window.liveflux.headers['X-Liveflux-Target']).toBeUndefined();
    });

    it('should handle missing headers object', function() {
      delete window.liveflux.headers;
      expect(function() { window.liveflux.disableTargetSupport(); }).not.toThrow();
    });
  });
});
