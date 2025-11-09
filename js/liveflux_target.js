(function(){
  if(!window.liveflux){
    console.log('[Liveflux Target] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const TARGET_LOG_PREFIX = '[Liveflux Target]';

  /**
   * Applies targeted fragment updates from a template-based response.
   * @param {string} html - The HTML response containing <template> elements
   * @param {HTMLElement} componentRoot - The component root element to update
   * @returns {string|null} - Returns null if targets applied successfully, or the original HTML for fallback
   */
  function applyTargets(html, componentRoot) {
    if (!html || !componentRoot) {
      console.warn(`${TARGET_LOG_PREFIX} Missing html or componentRoot`);
      return html;
    }

    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const templates = doc.querySelectorAll('template[data-flux-target], template[data-flux-component-kind]');
    
    if (templates.length === 0) {
      // No templates found, treat as full component replacement
      return html;
    }
    
    let appliedCount = 0;
    let newComponentRoot = componentRoot;
    
    // Process component template first (full replacement)
    const componentTemplate = doc.querySelector('template[data-flux-component-kind]:not([data-flux-target])');
    if (componentTemplate) {
      const newRoot = componentTemplate.content.firstElementChild;
      if (newRoot) {
        componentRoot.replaceWith(newRoot);
        liveflux.executeScripts(newRoot);
        if (liveflux.initTriggers) liveflux.initTriggers(newRoot);
        newComponentRoot = newRoot; // Update reference for subsequent selectors
        appliedCount++;
        console.log(`${TARGET_LOG_PREFIX} Applied full component replacement`);
      }
    }
    
    // Process targeted fragments in document order
    templates.forEach(template => {
      // Skip the component template we already processed
      if (template.hasAttribute('data-flux-component-kind') && !template.hasAttribute('data-flux-target')) {
        return;
      }
      
      const selector = template.dataset.fluxTarget;
      if (!selector) {
        console.warn(`${TARGET_LOG_PREFIX} Template missing data-flux-target attribute`);
        return;
      }

      const swapMode = template.dataset.fluxSwap || 'replace';
      const targetComponent = template.dataset.fluxComponent;
      const targetComponentId = template.dataset.fluxComponentId;
      try {
        // Validate component metadata if present
        if (targetComponent || targetComponentId) {
          const rootComponent = newComponentRoot.getAttribute('data-flux-component-kind');
          const rootComponentId = newComponentRoot.getAttribute('data-flux-component-id');
          
          if (targetComponent && targetComponent !== rootComponent) {
            console.warn(`${TARGET_LOG_PREFIX} Component mismatch: expected ${rootComponent}, got ${targetComponent}`);
            return;
          }
          
          if (targetComponentId && targetComponentId !== rootComponentId) {
            console.warn(`${TARGET_LOG_PREFIX} Component ID mismatch: expected ${rootComponentId}, got ${targetComponentId}`);
            return;
          }
        }

        const searchRoot = (targetComponent || targetComponentId) ? newComponentRoot : document;

        const target = searchRoot.querySelector(selector);
        if (!target) {
          console.warn(`${TARGET_LOG_PREFIX} Selector not found: ${selector}`);
          return;
        }
        
        const fragment = template.content.firstElementChild;
        if (!fragment) {
          console.warn(`${TARGET_LOG_PREFIX} Template content is empty for selector: ${selector}`);
          return;
        }
        
        // Apply swap mode
        switch (swapMode) {
          case 'replace':
            target.replaceWith(fragment);
            liveflux.executeScripts(fragment);
            if (liveflux.initTriggers) liveflux.initTriggers(fragment);
            break;
          case 'inner':
            target.innerHTML = fragment.innerHTML;
            liveflux.executeScripts(target);
            if (liveflux.initTriggers) liveflux.initTriggers(target);
            break;
          case 'beforebegin':
          case 'afterbegin':
          case 'beforeend':
          case 'afterend':
            target.insertAdjacentElement(swapMode, fragment);
            liveflux.executeScripts(fragment);
            if (liveflux.initTriggers) liveflux.initTriggers(fragment);
            break;
          default:
            console.warn(`${TARGET_LOG_PREFIX} Unknown swap mode: ${swapMode}`);
            return;
        }
        
        appliedCount++;
        console.log(`${TARGET_LOG_PREFIX} Applied: ${selector} (mode: ${swapMode})`);
      } catch (e) {
        console.error(`${TARGET_LOG_PREFIX} Error applying selector: ${selector}`, e);
      }
    });
    
    if (appliedCount === 0) {
      console.warn(`${TARGET_LOG_PREFIX} No targets applied, falling back to full render`);
      return html; // Signal to caller to do full replacement
    }
    
    return null; // Targets applied successfully
  }

  /**
   * Checks if the response contains template-based fragments
   * @param {string} html - The HTML response
   * @returns {boolean}
   */
  function hasTargetTemplates(html) {
    if (!html) return false;
    return html.includes('<template data-flux-target') || html.includes('<template data-flux-component-kind');
  }

  /**
   * Enables target support by adding the handshake header
   */
  function enableTargetSupport() {
    if (!liveflux.headers) {
      liveflux.headers = {};
    }
    liveflux.headers['X-Liveflux-Target'] = 'enabled';
    console.log(`${TARGET_LOG_PREFIX} Target support enabled`);
  }

  /**
   * Disables target support by removing the handshake header
   */
  function disableTargetSupport() {
    if (liveflux.headers && liveflux.headers['X-Liveflux-Target']) {
      delete liveflux.headers['X-Liveflux-Target'];
      console.log(`${TARGET_LOG_PREFIX} Target support disabled`);
    }
  }

  // Expose API
  liveflux.applyTargets = applyTargets;
  liveflux.hasTargetTemplates = hasTargetTemplates;
  liveflux.enableTargetSupport = enableTargetSupport;
  liveflux.disableTargetSupport = disableTargetSupport;

  // Auto-enable target support by default
  // Users can disable it by calling liveflux.disableTargetSupport()
  if (liveflux.autoEnableTargets !== false) {
    enableTargetSupport();
  }
})();
