(function(){
  if(!window.liveflux){
    console.log('[Liveflux Wire] liveflux namespace not found');
    return;
  }

  function createWire(componentId, componentAlias, rootEl){
    return {
      on: function(eventName, callback){
        if(window.liveflux.events && typeof window.liveflux.events.onComponent === 'function'){
          return window.liveflux.events.onComponent(componentId, eventName, callback);
        }
        return function(){};
      },
      dispatch: function(eventName, data){
        const dispatchFn = window.liveflux.dispatch || (window.liveflux.events && window.liveflux.events.dispatch);
        if(typeof dispatchFn === 'function') dispatchFn(eventName, data);
      },
      dispatchSelf: function(eventName, data){
        const d = Object.assign({}, data||{}, { __self: true });
        const dispatchFn = window.liveflux.dispatch || (window.liveflux.events && window.liveflux.events.dispatch);
        if(typeof dispatchFn === 'function') dispatchFn(eventName, d);
      },
      dispatchTo: function(targetAlias, eventName, data){
        const d = Object.assign({}, data||{}, { __target: targetAlias });
        const dispatchFn = window.liveflux.dispatch || (window.liveflux.events && window.liveflux.events.dispatch);
        if(typeof dispatchFn === 'function') dispatchFn(eventName, d);
      },
      call: function(action, data){
        action = action || 'submit';
        const params = Object.assign({}, data || {}, {
          liveflux_component_type: componentAlias,
          liveflux_component_id: componentId,
          liveflux_action: action
        });
        return window.liveflux.post(params).then(function(result){
          const html = result.html || result;
          const tmp = document.createElement('div');
          tmp.innerHTML = html;
          const newNode = tmp.firstElementChild;
          if(newNode && rootEl){
            rootEl.replaceWith(newNode);
            rootEl = newNode;
            window.liveflux.executeScripts(newNode);
            if(window.liveflux.initWire) window.liveflux.initWire();
          }
          return result;
        });
      },
      id: componentId,
      alias: componentAlias
    };
  }

  function initWire(){
    const roots = document.querySelectorAll('[data-flux-root], [flux-root]');
    roots.forEach(function(root){
      const comp = root.querySelector('input[name="liveflux_component_type"]');
      const id = root.querySelector('input[name="liveflux_component_id"]');
      if(!comp || !id) return;
      root.$wire = createWire(id.value, comp.value, root);
    });
  }

  window.liveflux.createWire = createWire;
  window.liveflux.initWire = initWire;

})();
