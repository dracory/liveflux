(function(){
  const g = window; g.liveflux = g.liveflux || {}; g.__lw = g.__lw || {};

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

  function dispatch(eventName, data){
    const payload = data || {};
    // global listeners
    if(eventListeners[eventName]){
      eventListeners[eventName].forEach(cb=>{ try{ cb({ name:eventName, data:payload, detail:payload }); }catch(e){ console.error(e); } });
    }
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
        if(data.__target || data.__target_id){
          if(data.__target && data.__target !== componentAlias) return;
          if(data.__target_id && data.__target_id !== componentId) return;
          delete data.__target; delete data.__target_id;
        }
        if(data.__self){
          if(componentEventListeners[componentId] && componentEventListeners[componentId][ev.name]){
            componentEventListeners[componentId][ev.name].forEach(cb=>{ try{ cb({ name:ev.name, data:data, detail:data }); }catch(e){ console.error(e); } });
          }
          return;
        }
        dispatch(ev.name, data);
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
      var root = (g.liveflux && g.liveflux.findComponent) ? g.liveflux.findComponent(componentAlias, componentId) : null;
      if(!root) return;
      function ready(){
        if(!root.$wire){ setTimeout(ready, 50); return; }
        root.$wire.on(eventName, function(){ root.$wire.call(targetMethod); });
      }
      ready();
    }, delay);
  }

  // Expose as module
  g.liveflux.events = {
    on, dispatch, processEvents, onComponent, subscribe
  };
  // Convenience top-level
  if(!g.liveflux.on) g.liveflux.on = on;
  if(!g.liveflux.dispatch) g.liveflux.dispatch = dispatch;
  if(!g.liveflux.subscribe) g.liveflux.subscribe = subscribe;

  // Back-compat bridges
  g.__lw.on = on;
  g.__lw.dispatch = dispatch;
  g.__lw.processEvents = processEvents;
  g.__lw.onComponent = onComponent;
})();
