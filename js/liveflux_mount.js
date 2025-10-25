(function(){
  const g = window; g.liveflux = g.liveflux || {};

  function mountPlaceholders(){
    document.querySelectorAll('[data-flux-mount="1"], [flux-mount="1"]').forEach((el)=>{
      const component = el.getAttribute('data-flux-component') || el.getAttribute('flux-component');
      if(!component) return;
      const params = g.liveflux.readParams(el);
      params.liveflux_component_type = component;
      g.liveflux.post(params).then((result)=>{
        const html = result.html || result;
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        const newNode = tmp.firstElementChild;
        if(newNode){
          el.replaceWith(newNode);
          g.liveflux.executeScripts(newNode);
          if(g.liveflux.initWire) g.liveflux.initWire();
        }
      }).catch((err)=>{ console.error(component+' mount', err); });
    });
  }

  g.liveflux.mountPlaceholders = mountPlaceholders;
})();
