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
    
    const componentKind = component.getAttribute(dataFluxComponent);
    const componentId   = component.getAttribute(dataFluxComponentID);

    if(!componentKind){
      console.warn('[Liveflux Events] dispatchTo called without component kind');
      return;
    }
    
    if(!componentId){
      console.warn('[Liveflux Events] dispatchTo called without component id');
      return;
    }

    console.log('[Liveflux Events] dispatchTo called with component kind:', componentKind, 'component id:', componentId, 'event name:',eventName, 'data:', data);
    
    const payload = Object.assign({}, data || {});
    if(componentKind){
      payload.__target = componentKind;
    }
    if(componentId){
      payload.__target_id = componentId;
    }

    if(window.liveflux && window.liveflux.events && window.liveflux.events.dispatch){
      window.liveflux.events.dispatch(eventName, payload);
    }
  }

  /**
   * Dispatches an event targeted to all components of specific component kind.
   * @param {string} componentKind - Kind of the target component.
   * @param {string} eventName - Event name to dispatch.
   * @param {Object} [data] - Optional payload for the event.
   * @returns {void}
   */
  function dispatchToKind(componentKind, eventName, data){
    if(!componentKind){
      console.warn('[Liveflux Events] dispatchToKind called without component kind');
      return;
    }

    if(!eventName){
      console.warn('[Liveflux Events] dispatchToKind called without event name');
      return;
    }

    const components = document.querySelectorAll(`[${dataFluxRoot}][${dataFluxComponent}="${componentKind}"]`);
    if(components.length === 0){
      console.warn('[Liveflux Events] dispatchToKind called without component');
      return;
    }

    components.forEach(function(component){
      const actualComponentKind = component.getAttribute(dataFluxComponent);
      const payload = Object.assign({}, data || {});
      if(actualComponentKind){
        payload.__target = actualComponentKind;
      }
      if(window.liveflux && window.liveflux.events && window.liveflux.events.dispatch){
        window.liveflux.events.dispatch(eventName, payload);
      }
    });
  }

  /**
   * Dispatches an event targeted to a specific component kind and ID.
   * @param {string} componentKind - Kind of the target component.
   * @param {string} componentId - ID of the target component instance.
   * @param {string} eventName - Event name to dispatch.
   * @param {Object} [data] - Optional payload for the event.
   * @returns {void}
   */
  function dispatchToKindAndId(componentKind, componentId, eventName, data){
    if(!componentKind){
      console.warn('[Liveflux Events] dispatchToKindAndId called without component kind');
      return;
    }

    if(!componentId){
      console.warn('[Liveflux Events] dispatchToKindAndId called without component id');
      return;
    }

    if(!eventName){
      console.warn('[Liveflux Events] dispatchToKindAndId called without event name');
      return;
    }

    const component = liveflux.findComponent ? liveflux.findComponent(componentKind, componentId) : null;
    if(!component){
      console.warn('[Liveflux Events] dispatchToKindAndId called without component');
      return;
    }

    const payload = Object.assign({}, data || {});
    if(componentKind){
      payload.__target = componentKind;
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
  if(!liveflux.dispatch && liveflux.events && typeof liveflux.events.dispatch === 'function'){
    liveflux.dispatch = liveflux.events.dispatch;
  }
  liveflux.dispatchTo = dispatchTo;
  liveflux.dispatchToKind = dispatchToKind;
  liveflux.dispatchToKindAndId = dispatchToKindAndId;
  liveflux.on = on;
})();
