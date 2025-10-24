(function(){
  const g = window; g.liveflux = g.liveflux || {}; g.__lw = g.__lw || {};

  function createWire(componentId, componentAlias, rootEl){
    return {
      on: function(eventName, callback){
        return g.liveflux.events && g.liveflux.events.onComponent
          ? g.liveflux.events.onComponent(componentId, eventName, callback)
          : (g.__lw.onComponent ? g.__lw.onComponent(componentId, eventName, callback) : function(){});
      },
      dispatch: function(eventName, data){
        (g.liveflux.dispatch || (g.liveflux.events && g.liveflux.events.dispatch) || g.__lw.dispatch || function(){ })(eventName, data);
      },
      dispatchSelf: function(eventName, data){
        const d = Object.assign({}, data||{}, { __self: true });
        (g.liveflux.dispatch || (g.liveflux.events && g.liveflux.events.dispatch) || g.__lw.dispatch || function(){ })(eventName, d);
      },
      dispatchTo: function(targetAlias, eventName, data){
        const d = Object.assign({}, data||{}, { __target: targetAlias });
        (g.liveflux.dispatch || (g.liveflux.events && g.liveflux.events.dispatch) || g.__lw.dispatch || function(){ })(eventName, d);
      },
      call: function(action, data){
        action = action || 'submit';
        const params = Object.assign({}, data || {}, {
          liveflux_component_type: componentAlias,
          liveflux_component_id: componentId,
          liveflux_action: action
        });
        return g.liveflux.post(params).then(function(result){
          const html = result.html || result;
          const tmp = document.createElement('div');
          tmp.innerHTML = html;
          const newNode = tmp.firstElementChild;
          if(newNode && rootEl){
            rootEl.replaceWith(newNode);
            rootEl = newNode;
            g.liveflux.executeScripts(newNode);
            if(g.liveflux.initWire) g.liveflux.initWire();
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

  g.liveflux.createWire = createWire;
  g.liveflux.initWire = initWire;

  // Back-compat
  g.__lw.createWire = createWire;
  g.__lw.initWire = initWire;
})();
