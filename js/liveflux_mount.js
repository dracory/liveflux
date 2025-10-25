(function(){
  if(!window.liveflux){
    console.log('[Liveflux Mount] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const { dataFluxMount, dataFluxComponent } = liveflux;
  const mountSelector = `[${dataFluxMount}="1"]`;
  const mountSelectorWithFallback = `${mountSelector}, [flux-mount="1"]`;

  function mountPlaceholders(){
    document.querySelectorAll(mountSelectorWithFallback).forEach((el)=>{
      const component = el.getAttribute(dataFluxComponent) || el.getAttribute('flux-component');
      if(!component) return;
      const params = liveflux.readParams(el);
      params.liveflux_component_type = component;
      liveflux.post(params).then((result)=>{
        const html = result.html || result;
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        const newNode = tmp.firstElementChild;
        if(newNode){
          el.replaceWith(newNode);
          liveflux.executeScripts(newNode);
          if(liveflux.initWire) liveflux.initWire();
        }
      }).catch((err)=>{ console.error(component+' mount', err); });
    });
  }

  liveflux.mountPlaceholders = mountPlaceholders;
})();
