(function(){
  if(!window.liveflux){
    console.log('[Liveflux Triggers] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const TRIGGER_LOG_PREFIX = '[Liveflux Triggers]';
  
  // Store trigger state per element using WeakMap for automatic cleanup
  const triggerRegistry = new WeakMap();
  const valueCache = new WeakMap();
  const pendingTimers = new WeakMap();
  
  // Configuration
  const config = {
    defaultTriggerDelay: 0,
    enableTriggers: true
  };

  /**
   * Default event mapping based on element type
   */
  const DEFAULT_EVENTS = {
    'INPUT:text': 'keyup changed',
    'INPUT:search': 'keyup changed',
    'INPUT:email': 'keyup changed',
    'INPUT:url': 'keyup changed',
    'INPUT:tel': 'keyup changed',
    'INPUT:password': 'keyup changed',
    'INPUT:number': 'keyup changed',
    'INPUT:checkbox': 'change',
    'INPUT:radio': 'change',
    'TEXTAREA': 'keyup changed',
    'SELECT': 'change',
    'BUTTON': 'click',
    'A': 'click',
    'FORM': 'submit'
  };

  /**
   * Parse trigger attribute into structured data
   * @param {HTMLElement} el - Element with data-flux-trigger
   * @returns {Array} Array of trigger definitions
   */
  function parseTriggers(el) {
    if (!el) return [];
    
    const triggerAttr = el.getAttribute('data-flux-trigger') || el.getAttribute('flux-trigger');
    if (!triggerAttr) return [];

    const definitions = [];
    const parts = triggerAttr.split(',').map(s => s.trim()).filter(Boolean);

    parts.forEach(part => {
      const tokens = part.split(/\s+/);
      const definition = {
        events: [],
        filters: [],
        modifiers: {}
      };

      tokens.forEach(token => {
        if (!token) return;

        // Check for modifiers (contains colon)
        if (token.includes(':')) {
          const [key, value] = token.split(':');
          if (key === 'delay' || key === 'throttle') {
            definition.modifiers[key] = parseDuration(value);
          } else if (key === 'queue') {
            definition.modifiers.queue = value;
          } else if (key === 'from' || key === 'not') {
            definition.filters.push({ type: key, selector: value });
          }
        }
        // Check for filters
        else if (token === 'changed' || token === 'once') {
          definition.filters.push({ type: token });
        }
        // Otherwise it's an event name
        else {
          definition.events.push(token);
        }
      });

      // If no events specified, infer default based on element type
      if (definition.events.length === 0) {
        const defaultEvents = inferDefaultEvents(el);
        definition.events = defaultEvents.split(/\s+/).filter(Boolean);
      }

      definitions.push(definition);
    });

    return definitions;
  }

  /**
   * Infer default event(s) based on element type
   */
  function inferDefaultEvents(el) {
    const tagName = el.tagName;
    const type = (el.type || '').toLowerCase();
    
    const key = `${tagName}:${type}`;
    if (DEFAULT_EVENTS[key]) {
      return DEFAULT_EVENTS[key];
    }
    
    if (DEFAULT_EVENTS[tagName]) {
      return DEFAULT_EVENTS[tagName];
    }
    
    return 'click'; // Ultimate fallback
  }

  /**
   * Parse duration string (e.g., "300ms", "1s") to milliseconds
   */
  function parseDuration(str) {
    if (!str) return 0;
    
    const match = str.match(/^(\d+(?:\.\d+)?)(ms|s)?$/);
    if (!match) {
      console.warn(`${TRIGGER_LOG_PREFIX} Invalid duration: ${str}`);
      return 0;
    }
    
    const value = parseFloat(match[1]);
    const unit = match[2] || 'ms';
    
    return unit === 's' ? value * 1000 : value;
  }

  /**
   * Serialize element values for change detection
   */
  function serializeForComparison(el, metadata) {
    const form = el.closest('form');
    const fields = liveflux.collectAllFields 
      ? liveflux.collectAllFields(el, metadata.root, form)
      : liveflux.serializeElement(el);
    
    return JSON.stringify(fields);
  }

  /**
   * Check if value has changed since last trigger
   */
  function hasChanged(el, metadata) {
    const currentValue = serializeForComparison(el, metadata);
    const lastValue = valueCache.get(el);
    
    // Cache should be initialized during registration
    // If it's not cached, something went wrong - cache it now and don't trigger
    if (lastValue === undefined) {
      valueCache.set(el, currentValue);
      return false;
    }
    
    // Compare with cached value
    if (currentValue !== lastValue) {
      valueCache.set(el, currentValue);
      return true;
    }
    
    return false;
  }

  /**
   * Evaluate filters for a trigger
   */
  function evaluateFilters(filters, event, el, metadata) {
    for (const filter of filters) {
      if (filter.type === 'changed') {
        if (!hasChanged(el, metadata)) {
          return false;
        }
      }
      else if (filter.type === 'once') {
        const state = triggerRegistry.get(el);
        if (state && state.fired) {
          return false;
        }
      }
      else if (filter.type === 'from') {
        if (!event.target.matches(filter.selector)) {
          return false;
        }
      }
      else if (filter.type === 'not') {
        if (event.target.matches(filter.selector)) {
          return false;
        }
      }
    }
    
    return true;
  }

  /**
   * Create debounced function
   */
  function debounce(fn, delay) {
    let timeoutId;
    return function(...args) {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => fn(...args), delay);
      return timeoutId;
    };
  }

  /**
   * Create throttled function
   */
  function throttle(fn, delay) {
    let lastCall = 0;
    let timeoutId;
    
    return function(...args) {
      const now = Date.now();
      const timeSinceLastCall = now - lastCall;
      
      if (timeSinceLastCall >= delay) {
        lastCall = now;
        fn(...args);
      } else {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(() => {
          lastCall = Date.now();
          fn(...args);
        }, delay - timeSinceLastCall);
      }
      
      return timeoutId;
    };
  }

  /**
   * Fire the action for a trigger
   */
  function fireTriggerAction(el, eventName, metadata) {
    if (!metadata) {
      console.warn(`${TRIGGER_LOG_PREFIX} No component metadata found`);
      return;
    }

    const action = el.getAttribute('data-flux-action') || 
                   el.getAttribute('flux-action') ||
                   el.closest('[data-flux-action], [flux-action]')?.getAttribute('data-flux-action') ||
                   el.closest('[data-flux-action], [flux-action]')?.getAttribute('flux-action');
    
    if (!action) {
      console.warn(`${TRIGGER_LOG_PREFIX} No action found for trigger on`, el);
      return;
    }

    // Collect fields
    const form = el.closest('form');
    const fields = liveflux.collectAllFields 
      ? liveflux.collectAllFields(el, metadata.root, form)
      : liveflux.serializeElement(el);

    const params = Object.assign({}, fields, {
      liveflux_component_type: metadata.comp,
      liveflux_component_id: metadata.id,
      liveflux_action: action
    });

    // Store trigger event name for header
    const triggerEventName = eventName;

    // Start indicators
    const indicatorEls = liveflux.startRequestIndicators(el, metadata.root);

    // Make request with trigger header
    const originalHeaders = liveflux.headers || {};
    liveflux.headers = Object.assign({}, originalHeaders, {
      'X-Liveflux-Trigger': triggerEventName
    });

    liveflux.post(params).then((result) => {
      // Restore original headers
      liveflux.headers = originalHeaders;
      const rawHtml = result.html || result;
      
      // Check if response contains target templates
      if (liveflux.hasTargetTemplates && liveflux.hasTargetTemplates(rawHtml)) {
        const fallback = liveflux.applyTargets(rawHtml, metadata.root);
        if (fallback) {
          // Targets failed, do full replacement
          const tmp = document.createElement('div');
          tmp.innerHTML = fallback;
          const newNode = tmp.firstElementChild;
          if (newNode && metadata.root) {
            metadata.root.replaceWith(newNode);
            liveflux.executeScripts(newNode);
          }
        }
        if (liveflux.initWire) liveflux.initWire();
        // Re-init triggers after DOM update
        if (liveflux.initTriggers) liveflux.initTriggers();
        return;
      }
      
      // Traditional full replacement
      const tmp = document.createElement('div');
      tmp.innerHTML = rawHtml;
      const newNode = tmp.firstElementChild;
      if (newNode && metadata.root) {
        metadata.root.replaceWith(newNode);
        liveflux.executeScripts(newNode);
        if (liveflux.initWire) liveflux.initWire();
        // Re-init triggers after DOM update
        if (liveflux.initTriggers) liveflux.initTriggers();
      }
    }).catch((err) => {
      // Restore original headers on error
      liveflux.headers = originalHeaders;
      console.error(`${TRIGGER_LOG_PREFIX} Action failed:`, err);
    }).finally(() => {
      liveflux.endRequestIndicators(indicatorEls);
    });

    // Mark as fired for 'once' filter
    const state = triggerRegistry.get(el);
    if (state) {
      state.fired = true;
    }
  }

  /**
   * Create event handler for a trigger definition
   */
  function createTriggerHandler(el, definition, metadata) {
    const handler = (event) => {
      // Evaluate filters
      if (!evaluateFilters(definition.filters, event, el, metadata)) {
        return;
      }

      // For form submit, prevent default
      if (definition.events.includes('submit')) {
        event.preventDefault();
      }

      // Apply timing modifiers
      const delay = definition.modifiers.delay || config.defaultTriggerDelay;
      const throttleDelay = definition.modifiers.throttle;
      const queueStrategy = definition.modifiers.queue || 'replace';

      const fireAction = () => {
        // Handle queue strategy
        if (queueStrategy === 'replace') {
          // Cancel any pending request for this element
          const timers = pendingTimers.get(el);
          if (timers) {
            timers.forEach(clearTimeout);
            timers.clear();
          }
        }
        
        fireTriggerAction(el, event.type, metadata);
      };

      if (throttleDelay) {
        const throttled = throttle(fireAction, throttleDelay);
        const timerId = throttled();
        
        // Track timer for cleanup
        if (!pendingTimers.has(el)) {
          pendingTimers.set(el, new Set());
        }
        if (timerId) {
          pendingTimers.get(el).add(timerId);
        }
      } else if (delay) {
        const debounced = debounce(fireAction, delay);
        const timerId = debounced();
        
        // Track timer for cleanup
        if (!pendingTimers.has(el)) {
          pendingTimers.set(el, new Set());
        }
        if (timerId) {
          pendingTimers.get(el).add(timerId);
        }
      } else {
        fireAction();
      }
    };

    return handler;
  }

  /**
   * Register triggers for an element
   */
  function registerTriggers(el) {
    if (!el || !config.enableTriggers) return;

    // Check if already registered
    if (triggerRegistry.has(el)) {
      return;
    }

    const definitions = parseTriggers(el);
    if (definitions.length === 0) return;

    // Resolve component metadata
    const rootSelector = `[${liveflux.dataFluxRoot || 'data-flux-root'}], [flux-root]`;
    const metadata = liveflux.resolveComponentMetadata(el, rootSelector);
    
    if (!metadata) {
      console.warn(`${TRIGGER_LOG_PREFIX} No component metadata found for element`, el);
      return;
    }

    // Initialize value cache for 'changed' filter
    definitions.forEach(definition => {
      const hasChangedFilter = definition.filters.some(f => f.type === 'changed');
      if (hasChangedFilter) {
        const initialValue = serializeForComparison(el, metadata);
        valueCache.set(el, initialValue);
      }
    });

    const listeners = [];

    definitions.forEach(definition => {
      const handler = createTriggerHandler(el, definition, metadata);
      
      definition.events.forEach(eventName => {
        el.addEventListener(eventName, handler);
        listeners.push({ eventName, handler });
      });
    });

    // Store state
    triggerRegistry.set(el, {
      listeners,
      fired: false
    });

    console.log(`${TRIGGER_LOG_PREFIX} Registered ${listeners.length} trigger(s) for`, el);
  }

  /**
   * Unregister triggers for an element
   */
  function unregisterTriggers(el) {
    if (!el) return;

    const state = triggerRegistry.get(el);
    if (!state) return;

    // Remove event listeners
    state.listeners.forEach(({ eventName, handler }) => {
      el.removeEventListener(eventName, handler);
    });

    // Clear timers
    const timers = pendingTimers.get(el);
    if (timers) {
      timers.forEach(clearTimeout);
      timers.clear();
    }

    // Clean up state
    triggerRegistry.delete(el);
    valueCache.delete(el);
    pendingTimers.delete(el);
  }

  /**
   * Initialize triggers for all elements with data-flux-trigger
   */
  function initTriggers(root) {
    if (!config.enableTriggers) return;

    const searchRoot = root || document;
    const elements = searchRoot.querySelectorAll('[data-flux-trigger], [flux-trigger]');
    
    elements.forEach(registerTriggers);
    
    console.log(`${TRIGGER_LOG_PREFIX} Initialized ${elements.length} trigger element(s)`);
  }

  /**
   * Clean up triggers in a subtree (before DOM removal)
   */
  function cleanupTriggers(root) {
    if (!root) return;

    const elements = root.querySelectorAll('[data-flux-trigger], [flux-trigger]');
    elements.forEach(unregisterTriggers);
    
    // Also check if root itself has triggers
    if (root.hasAttribute && (root.hasAttribute('data-flux-trigger') || root.hasAttribute('flux-trigger'))) {
      unregisterTriggers(root);
    }
  }

  /**
   * Configure trigger system
   */
  function configureTriggers(options) {
    if (options.defaultTriggerDelay !== undefined) {
      config.defaultTriggerDelay = options.defaultTriggerDelay;
    }
    if (options.enableTriggers !== undefined) {
      config.enableTriggers = options.enableTriggers;
    }
  }

  // Expose API
  liveflux.parseTriggers = parseTriggers;
  liveflux.registerTriggers = registerTriggers;
  liveflux.unregisterTriggers = unregisterTriggers;
  liveflux.initTriggers = initTriggers;
  liveflux.cleanupTriggers = cleanupTriggers;
  liveflux.configureTriggers = configureTriggers;

})();
