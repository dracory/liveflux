(function(){
  if(!window.liveflux){
    console.log('[Liveflux Events] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;

  // Internal registries
  const eventListeners = {};
  const componentEventListeners = {};

  function on(eventName, callback){
    if(!eventListeners[eventName]) eventListeners[eventName] = [];
    eventListeners[eventName].push(callback);
    return function(){
      const idx = eventListeners[eventName].indexOf(callback);
      if(idx > -1) eventListeners[eventName].splice(idx, 1);
    };
  }

  /**
   * Dispatches an event to all listeners and as a browser event.
   * @param {string} eventName - The name of the event to dispatch.
   * @param {Object} data - Optional data to pass with the event.
   * @returns {void}
   */
  function dispatch(eventName, data){
    const payload = data || {};

    // global listeners
    if(eventListeners[eventName]){
      eventListeners[eventName].forEach(cb=>{ try{ cb({ name:eventName, data:payload, detail:payload }); }catch(e){ console.error(e); } });
    }
    
    console.log('[Liveflux Events] dispatch called with event name:', eventName, 'data:', payload);
    
    // component listeners
    for(const cid in componentEventListeners){
      const map = componentEventListeners[cid] || {};
      const listeners = map[eventName] || [];
      listeners.forEach(cb=>{ try{ cb({ name:eventName, data:payload, detail:payload }); }catch(e){ console.error(e); } });
    }
    
    // DOM CustomEvent
    try {
      document.dispatchEvent(new CustomEvent(eventName, { detail: payload, bubbles:true, cancelable:true }));
    } catch(_) {}
  }

  function processEvents(response, componentId, componentAlias){
    const hdr = response.headers.get('X-Liveflux-Events');
    if(!hdr) return;
    try{
      const events = JSON.parse(hdr);
      if(!Array.isArray(events)) return;
      events.forEach((ev)=>{
        if(!ev || !ev.name) return;
        const data = ev.data || {};
        // handle targeting
        let payload = data;
        const targetAlias = payload.__target;
        const targetId = payload.__target_id;

        if(targetAlias || targetId){
          payload = Object.assign({}, payload);
          delete payload.__target;
          delete payload.__target_id;

          let handled = false;
          if(targetAlias && targetId && typeof liveflux.dispatchToAliasAndId === 'function'){
            try { liveflux.dispatchToAliasAndId(targetAlias, targetId, ev.name, payload); handled = true; }
            catch(e){ console.error('[Liveflux Events] dispatchToAliasAndId error', e); }
          }
          if(!handled && targetAlias && typeof liveflux.dispatchToAlias === 'function'){
            try { liveflux.dispatchToAlias(targetAlias, ev.name, payload); handled = true; }
            catch(e){ console.error('[Liveflux Events] dispatchToAlias error', e); }
          }
          if(!handled && targetId && typeof liveflux.findComponent === 'function' && typeof liveflux.dispatchTo === 'function'){
            const lookupAlias = targetAlias || componentAlias;
            try {
              const targetRoot = lookupAlias ? liveflux.findComponent(lookupAlias, targetId) : null;
              if(targetRoot){
                liveflux.dispatchTo(targetRoot, ev.name, payload);
                handled = true;
              }
            } catch(e){ console.error('[Liveflux Events] dispatchTo target error', e); }
          }

          if(handled){
            return;
          }
        }

        if(payload.__self){
          const listeners = componentEventListeners[componentId] && componentEventListeners[componentId][ev.name];
          if(listeners && listeners.length){
            const selfPayload = Object.assign({}, payload);
            delete selfPayload.__self;
            listeners.forEach(cb=>{ try{ cb({ name:ev.name, data:selfPayload, detail:selfPayload }); }catch(e){ console.error(e); } });
          }
          return;
        }
        dispatch(ev.name, payload);
      });
    } catch(e){ console.error('[Liveflux Events] parse error', e); }
  }

  function onComponent(componentId, eventName, callback){
    if(!componentEventListeners[componentId]) componentEventListeners[componentId] = {};
    if(!componentEventListeners[componentId][eventName]) componentEventListeners[componentId][eventName] = [];
    componentEventListeners[componentId][eventName].push(callback);
    return function(){
      const list = componentEventListeners[componentId][eventName] || [];
      const idx = list.indexOf(callback);
      if(idx > -1) list.splice(idx,1);
    };
  }

  function subscribe(componentAlias, componentId, eventName, targetMethod, timeoutMs){
    var delay = typeof timeoutMs === 'number' ? timeoutMs : 0;
    setTimeout(function(){
      var root = (window.liveflux && window.liveflux.findComponent) ? window.liveflux.findComponent(componentAlias, componentId) : null;
      if(!root) return;
      function ready(){
        if(!root.$wire){ setTimeout(ready, 50); return; }
        root.$wire.on(eventName, function(){ root.$wire.call(targetMethod); });
      }
      ready();
    }, delay);
  }

  // Expose as module
  window.liveflux.events = {
    on, dispatch, processEvents, onComponent, subscribe
  };
  // Convenience top-level
  window.liveflux.on = on;
  window.liveflux.dispatch = dispatch;
  window.liveflux.subscribe = subscribe;

})();
