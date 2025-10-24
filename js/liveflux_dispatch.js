/**
 * Liveflux Dispatch
 * Exposes the main methods for dispatching events.
 * 
 * Notes:
 * - This file is loaded in the browser and is not used in the Go code.
 * - It is used to add dispatch methods to the liveflux namespace.
 * - It depends om liveflux_namespace_create.js.
 * - It depends om liveflux_find.js.
 * 
 * Dev notes:
 * - The functions are sorted alphabetically.
 */
(function(){
  /**
   * Dispatches an event to all listeners and as a browser event.
   * @param {string} eventName - The name of the event to dispatch.
   * @param {Object} data - Optional data to pass with the event.
   * @returns {void}
   */
  function dispatch(eventName, data){
    window.__lw.dispatch(eventName, data);
  }

  /**
   * Dispatches an event targeted to a specific component.
   * @param {HTMLElement} component - The component element to dispatch the event to.
   * @param {Object} data - Optional data to pass with the event.
   * @returns {void}
   */
  function dispatchTo(component, data){
    if(!component){
      console.warn('[Liveflux Events] dispatchTo called without component');
      return;
    }

    if(!data){
      console.warn('[Liveflux Events] dispatchTo called without data');
      return;
    }
    
    const componentAlias = component.getAttribute('data-flux-component');
    const componentId = component.getAttribute('data-flux-component-id');

    if(!componentAlias){
      console.warn('[Liveflux Events] dispatchTo called without component alias');
      return;
    }
    
    if(!componentId){
      console.warn('[Liveflux Events] dispatchTo called without component id');
      return;
    }
    
    const payload = Object.assign({}, data || {});
    if(componentAlias){
      payload.__target = componentAlias;
    }
    if(componentId){
      payload.__target_id = componentId;
    }

    window.__lw.dispatch(eventName, payload);
  }

  /**
   * Dispatches an event targeted to all components of specific component alias.
   * @param {string} componentAlias - Alias of the target component.
   * @param {string} eventName - Event name to dispatch.
   * @param {Object} [data] - Optional payload for the event.
   * @returns {void}
   */
  function dispatchToAlias(componentAlias, eventName, data){
    if(!componentAlias){
      console.warn('[Liveflux Events] dispatchToAlias called without component alias');
      return;
    }

    if(!eventName){
      console.warn('[Liveflux Events] dispatchToAlias called without event name');
      return;
    }

    const components = document.querySelectorAll('[data-flux-root][data-flux-component="' + componentAlias + '"]');
    if(components.length === 0){
      console.warn('[Liveflux Events] dispatchToAlias called without component');
      return;
    }

    components.forEach(function(component){
      const payload = Object.assign({}, data || {});
      if(componentAlias){
        payload.__target = componentAlias;
      }
      g.__lw.dispatch(eventName, payload);
    });
  }

  /**
   * Dispatches an event targeted to a specific component alias and ID.
   * @param {string} componentAlias - Alias of the target component.
   * @param {string} componentId - ID of the target component instance.
   * @param {string} eventName - Event name to dispatch.
   * @param {Object} [data] - Optional payload for the event.
   * @returns {void}
   */
  function dispatchToAliasAndId(componentAlias, componentId, eventName, data){
    if(!componentAlias){
      console.warn('[Liveflux Events] dispatchToAliasAndId called without component alias');
      return;
    }

    if(!componentId){
      console.warn('[Liveflux Events] dispatchToAliasAndId called without component id');
      return;
    }

    if(!eventName){
      console.warn('[Liveflux Events] dispatchToAliasAndId called without event name');
      return;
    }

    const component = findComponent(componentAlias, componentId);
    if(!component){
      console.warn('[Liveflux Events] dispatchToAliasAndId called without component');
      return;
    }

    const payload = Object.assign({}, data || {});
    if(componentAlias){
      payload.__target = componentAlias;
    }
    if(componentId){
      payload.__target_id = componentId;
    }

    window.__lw.dispatch(eventName, payload);
  }

  /**
   * Finds a component by alias and ID.
   * @param {string} componentAlias - Alias of the target component.
   * @param {string} componentId - ID of the target component instance.
   * @returns {HTMLElement|null} - The component element if found, otherwise null.
   */
  function findComponent(componentAlias, componentId){
    return document.querySelector('[data-flux-root][data-flux-component="' + componentAlias + '"][data-flux-component-id="' + componentId + '"]');
  }

  /**
   * Registers a global event listener for a specific event name.
   * @param {string} eventName - The name of the event to listen for.
   * @param {Function} callback - The callback function to execute when the event is dispatched.
   * @returns {Function} - A cleanup function to remove the listener.
   */
  function on(eventName, callback){
    return window.__lw.on(eventName, callback);
  }
  
  // Check if liveflux namespace exists
  if(!window.liveflux){
    console.log('[Liveflux Dispatch] liveflux namespace not found');
  }
  
  // Add functions to liveflux namespace
  window.liveflux.dispatch = dispatch;
  window.liveflux.dispatchTo = dispatchTo;
  window.liveflux.dispatchToAlias = dispatchToAlias;
  window.liveflux.dispatchToAliasAndId = dispatchToAliasAndId;
  window.liveflux.on = on;
})();
