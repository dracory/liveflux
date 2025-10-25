/**
 * Liveflux Dispatch
 * Exposes the main methods for dispatching events.
 * 
 * Notes:
 * - This file is loaded in the browser and is not used in the Go code.
 * - It is used to add dispatch methods to the liveflux namespace.
 * - It depends on liveflux_namespace_create.js.
 * - It depends on liveflux_find.js.
 * 
 * Dev notes:
 * - The functions are sorted alphabetically.
 */
(function(){
  if(!window.liveflux){
    console.log('[Liveflux Dispatch] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const {
    dataFluxComponent,
    dataFluxComponentID,
    dataFluxRoot,
  } = liveflux;

  /**
   * Dispatches an event to all listeners and as a browser event.
   * @param {string} eventName - The name of the event to dispatch.
   * @param {Object} data - Optional data to pass with the event.
   * @returns {void}
   */
  function dispatch(eventName, data){
    if(window.liveflux && window.liveflux.events && window.liveflux.events.dispatch){
      window.liveflux.events.dispatch(eventName, data);
    }
  }

  /**
   * Dispatches an event targeted to a specific component.
   * @param {HTMLElement} component - The component element to dispatch the event to.
   * @param {string} eventName - Event name to dispatch.
   * @param {Object} data - Optional data to pass with the event.
   * @returns {void}
   */
  function dispatchTo(component, eventName, data){
    if(!component){
      console.warn('[Liveflux Events] dispatchTo called without component');
      return;
    }

    if(!eventName){
      console.warn('[Liveflux Events] dispatchTo called without event name');
      return;
    }
    
    const componentAlias = component.getAttribute(dataFluxComponent);
    const componentId = component.getAttribute(dataFluxComponentID);

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

    if(window.liveflux && window.liveflux.events && window.liveflux.events.dispatch){
      window.liveflux.events.dispatch(eventName, payload);
    }
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

    const components = document.querySelectorAll(`[${dataFluxRoot}][${dataFluxComponent}="${componentAlias}"]`);
    if(components.length === 0){
      console.warn('[Liveflux Events] dispatchToAlias called without component');
      return;
    }

    components.forEach(function(component){
      const payload = Object.assign({}, data || {});
      if(componentAlias){
        payload.__target = componentAlias;
      }
      if(window.liveflux && window.liveflux.events && window.liveflux.events.dispatch){
        window.liveflux.events.dispatch(eventName, payload);
      }
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

    const component = liveflux.findComponent ? liveflux.findComponent(componentAlias, componentId) : null;
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

    if(liveflux.events && liveflux.events.dispatch){
      liveflux.events.dispatch(eventName, payload);
    }
  }

  /**
   * Registers a global event listener for a specific event name.
   * @param {string} eventName - The name of the event to listen for.
   * @param {Function} callback - The callback function to execute when the event is dispatched.
   * @returns {Function} - A cleanup function to remove the listener.
   */
  function on(eventName, callback){
    if(window.liveflux && window.liveflux.events && window.liveflux.events.on){
      return window.liveflux.events.on(eventName, callback);
    }
    return function(){};
  }
  
  // Check if liveflux namespace exists
  // Add functions to liveflux namespace
  liveflux.dispatch = dispatch;
  liveflux.dispatchTo = dispatchTo;
  liveflux.dispatchToAlias = dispatchToAlias;
  liveflux.dispatchToAliasAndId = dispatchToAliasAndId;
  liveflux.on = on;
})();
