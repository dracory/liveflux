(function(){
  const g = window; g.__lw = g.__lw || {};

  /**
   * Finds placeholders marked with data-flux-mount="1" and replaces them with
   * server-rendered component HTML, executing any scripts inside.
   * @returns {void}
   */
  g.__lw.mountPlaceholders = function(){
    document.querySelectorAll('[data-flux-mount="1"], [flux-mount="1"]').forEach((el)=>{
      const component = el.getAttribute('data-flux-component') || el.getAttribute('flux-component');

      if(!component) return;

      const params = g.__lw.readParams(el);

      params.liveflux_component_type = component;
      
      g.__lw.post(params).then((result)=>{
        const html = result.html || result;
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        const newNode = tmp.firstElementChild;
        if(newNode){ 
          el.replaceWith(newNode); 
          g.__lw.executeScripts(newNode);
          // Initialize $wire for the newly mounted component
          if(g.__lw.initWire) g.__lw.initWire();
        }
      }).catch((err)=>{ console.error(component+' mount', err); });
    });
  };
})();
