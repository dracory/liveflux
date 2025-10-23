(function(){
  // Initialize global namespace
  const g = window; g.__lw = g.__lw || {};

  // Event listeners registry: maps event names to arrays of listener functions
  g.__lw.eventListeners = g.__lw.eventListeners || {};

  // Component-specific event listeners: maps component IDs to event listeners
  g.__lw.componentEventListeners = g.__lw.componentEventListeners || {};

  /**
   * Registers a global event listener for a specific event name.
   * @param {string} eventName - The name of the event to listen for.
   * @param {Function} callback - The callback function to execute when the event is dispatched.
   * @returns {Function} - A cleanup function to remove the listener.
   */
  g.__lw.on = function(eventName, callback){
    if(!g.__lw.eventListeners[eventName]){
      g.__lw.eventListeners[eventName] = [];
    }
    g.__lw.eventListeners[eventName].push(callback);

    // Return cleanup function
    return function(){
      const idx = g.__lw.eventListeners[eventName].indexOf(callback);
      if(idx > -1){
        g.__lw.eventListeners[eventName].splice(idx, 1);
      }
    };
  };

  /**
   * Dispatches an event to all registered listeners and as a browser event.
   * @param {string} eventName - The name of the event to dispatch.
   * @param {Object} data - Optional data to pass with the event.
   * @returns {void}
   */
  g.__lw.dispatch = function(eventName, data){
    console.log('[Liveflux Events] Dispatching event:', eventName, 'with data:', data);
    
    const event = {
      name: eventName,
      data: data || {},
      detail: data || {}
    };

    // Trigger global listeners
    if(g.__lw.eventListeners[eventName]){
      console.log('[Liveflux Events] Found', g.__lw.eventListeners[eventName].length, 'global listeners for', eventName);
      g.__lw.eventListeners[eventName].forEach(function(callback){
        try {
          callback(event);
        } catch(err){
          console.error('[Liveflux Events] Event listener error:', err);
        }
      });
    } else {
      console.log('[Liveflux Events] No global listeners registered for', eventName);
    }

    // Trigger component-specific listeners for all components
    var componentListenerCount = 0;
    for(var componentId in g.__lw.componentEventListeners){
      if(g.__lw.componentEventListeners[componentId][eventName]){
        var listeners = g.__lw.componentEventListeners[componentId][eventName];
        componentListenerCount += listeners.length;
        console.log('[Liveflux Events] Triggering', listeners.length, 'component listeners for', componentId);
        listeners.forEach(function(callback){
          try {
            callback(event);
          } catch(err){
            console.error('[Liveflux Events] Component listener error:', err);
          }
        });
      }
    }
    if(componentListenerCount > 0){
      console.log('[Liveflux Events] Total component listeners triggered:', componentListenerCount);
    }

    // Dispatch as browser custom event for Alpine/vanilla JS integration
    const customEvent = new CustomEvent(eventName, {
      detail: data || {},
      bubbles: true,
      cancelable: true
    });
    console.log('[Liveflux Events] Dispatching browser custom event:', eventName);
    document.dispatchEvent(customEvent);
  };

  /**
   * Processes events from the server response header.
   * @param {Response} response - The fetch response object.
   * @param {string} componentId - The ID of the component that triggered the request.
   * @param {string} componentAlias - The alias of the component that triggered the request.
   * @returns {void}
   */
  g.__lw.processEvents = function(response, componentId, componentAlias){
    const eventsHeader = response.headers.get('X-Liveflux-Events');
    console.log('[Liveflux Events] Processing events from response', {
      componentId: componentId,
      componentAlias: componentAlias,
      eventsHeader: eventsHeader
    });
    
    if(!eventsHeader) {
      console.log('[Liveflux Events] No events header found');
      return;
    }

    try {
      const events = JSON.parse(eventsHeader);
      console.log('[Liveflux Events] Parsed events:', events);
      
      if(!Array.isArray(events)) {
        console.warn('[Liveflux Events] Events is not an array:', events);
        return;
      }

      events.forEach(function(event){
        if(!event.name) {
          console.warn('[Liveflux Events] Event missing name:', event);
          return;
        }

        const data = event.data || {};
        console.log('[Liveflux Events] Processing event:', event.name, 'with data:', data);

        // Check if event is targeted to a specific component
        if(data.__target){
          console.log('[Liveflux Events] Event targeted to:', data.__target);
          // Only dispatch if this component matches the target
          if(data.__target !== componentAlias){
            console.log('[Liveflux Events] Skipping - target mismatch');
            return;
          }
          // Remove metadata before dispatching
          delete data.__target;
        }

        // Check if event is self-only
        if(data.__self){
          console.log('[Liveflux Events] Event is self-only');
          // Only dispatch to the component that sent it
          if(g.__lw.componentEventListeners[componentId] && 
             g.__lw.componentEventListeners[componentId][event.name]){
            const listeners = g.__lw.componentEventListeners[componentId][event.name];
            console.log('[Liveflux Events] Triggering', listeners.length, 'component listeners');
            listeners.forEach(function(callback){
              try {
                callback({ name: event.name, data: data, detail: data });
              } catch(err){
                console.error('[Liveflux Events] Component event listener error:', err);
              }
            });
          } else {
            console.log('[Liveflux Events] No component listeners found for self-only event');
          }
          return;
        }

        // Dispatch globally
        console.log('[Liveflux Events] Dispatching globally:', event.name);
        g.__lw.dispatch(event.name, data);
      });
    } catch(err){
      console.error('[Liveflux Events] Failed to process events:', err);
    }
  };

  /**
   * Registers a component-specific event listener.
   * Used within component scripts via $wire.on().
   * @param {string} componentId - The ID of the component.
   * @param {string} eventName - The name of the event to listen for.
   * @param {Function} callback - The callback function to execute.
   * @returns {Function} - A cleanup function to remove the listener.
   */
  g.__lw.onComponent = function(componentId, eventName, callback){
    console.log('[Liveflux Events] Registering component listener:', {
      componentId: componentId,
      eventName: eventName
    });
    
    if(!g.__lw.componentEventListeners[componentId]){
      g.__lw.componentEventListeners[componentId] = {};
    }
    if(!g.__lw.componentEventListeners[componentId][eventName]){
      g.__lw.componentEventListeners[componentId][eventName] = [];
    }
    g.__lw.componentEventListeners[componentId][eventName].push(callback);
    
    console.log('[Liveflux Events] Component listeners for', eventName, ':', 
      g.__lw.componentEventListeners[componentId][eventName].length);

    // Return cleanup function
    return function(){
      const listeners = g.__lw.componentEventListeners[componentId][eventName];
      const idx = listeners.indexOf(callback);
      if(idx > -1){
        listeners.splice(idx, 1);
      }
    };
  };

  // Expose Liveflux global object for external scripts
  if(!window.Liveflux){
    window.Liveflux = {
      on: g.__lw.on,
      dispatch: g.__lw.dispatch
    };
  }
})();
