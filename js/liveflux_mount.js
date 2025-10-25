(function(){
  if(!window.liveflux){
    console.log('[Liveflux Mount] liveflux namespace not found');
    return;
  }

  function mountPlaceholders(){
    document.querySelectorAll('[data-flux-mount="1"], [flux-mount="1"]').forEach((el)=>{
      const component = el.getAttribute('data-flux-component') || el.getAttribute('flux-component');
      if(!component) return;
      const params = window.liveflux.readParams(el);
      params.liveflux_component_type = component;
      window.liveflux.post(params).then((result)=>{
        const html = result.html || result;
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        const newNode = tmp.firstElementChild;
        if(newNode){
          el.replaceWith(newNode);
          window.liveflux.executeScripts(newNode);
          if(window.liveflux.initWire) window.liveflux.initWire();
        }
      }).catch((err)=>{ console.error(component+' mount', err); });
    });
  }

  window.liveflux.mountPlaceholders = mountPlaceholders;
})();
