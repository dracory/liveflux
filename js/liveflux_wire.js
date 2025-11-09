(function(){
  if(!window.liveflux){
    console.log('[Liveflux Wire] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const { dataFluxRoot, dataFluxComponentID } = liveflux;
  const rootSelector = `[${dataFluxRoot}]`;
  const rootSelectorWithFallback = `${rootSelector}, [flux-root]`;

  function createWire(componentId, componentKind, rootEl){
    return {
      on: function(eventName, callback){
        if(liveflux.events && typeof liveflux.events.onComponent === 'function'){
          return liveflux.events.onComponent(componentId, eventName, callback);
        }
        return function(){};
      },
      dispatch: function(eventName, data){
        const dispatchFn = liveflux.dispatch || (liveflux.events && liveflux.events.dispatch);
        if(typeof dispatchFn === 'function') dispatchFn(eventName, data);
      },
      dispatchSelf: function(eventName, data){
        const d = Object.assign({}, data||{}, { __self: true });
        const dispatchFn = liveflux.dispatch || (liveflux.events && liveflux.events.dispatch);
        if(typeof dispatchFn === 'function') dispatchFn(eventName, d);
      },
      dispatchTo: function(targetKind, eventName, data){
        const d = Object.assign({}, data||{}, { __target: targetKind });
        const dispatchFn = liveflux.dispatch || (liveflux.events && liveflux.events.dispatch);
        if(typeof dispatchFn === 'function') dispatchFn(eventName, d);
      },
      call: function(action, data){
        action = action || 'submit';
        const params = Object.assign({}, data || {}, {
          liveflux_component_kind: componentKind,
          liveflux_component_id: componentId,
          liveflux_action: action
        });
        const indicatorEls = liveflux.startRequestIndicators(rootEl, rootEl);

        return liveflux.post(params).then(function(result){
          const html = result.html || result;
          const tmp = document.createElement('div');
          tmp.innerHTML = html;
          const newNode = tmp.firstElementChild;
          if(newNode && rootEl){
            rootEl.replaceWith(newNode);
            rootEl = newNode;
            liveflux.executeScripts(newNode);
            if(liveflux.initWire) liveflux.initWire();
          }
          return result;
        }).finally(function(){
          liveflux.endRequestIndicators(indicatorEls);
        });
      },
      id: componentId,
      kind: componentKind
    };
  }

  function initWire(){
    const roots = document.querySelectorAll(rootSelectorWithFallback);
    roots.forEach(function(root){
      const comp = root.getAttribute(liveflux.dataFluxComponent || 'data-flux-component');
      const id = root.getAttribute(liveflux.dataFluxComponentID || 'data-flux-component-id');
      if(!comp || !id) return;
      root.$wire = createWire(id, comp, root);
    });
  }

  liveflux.createWire = createWire;
  liveflux.initWire = initWire;

})();
