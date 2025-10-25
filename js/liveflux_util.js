(function(){
  if(!window.liveflux){
    console.log('[Liveflux Util] liveflux namespace not found');
    return;
  }

  // Default endpoint if not set
  if (!window.liveflux.endpoint) { window.liveflux.endpoint = '/liveflux'; }

  function executeScripts(root){
    if(!root) return;
    const scripts = root.querySelectorAll('script');
    scripts.forEach((old)=>{
      const s = document.createElement('script');
      for (const attr of old.attributes) s.setAttribute(attr.name, attr.value);
      s.text = old.textContent || '';
      old.replaceWith(s);
    });
  }

  function serializeElement(el){
    const params = {};
    if(!el) return params;
    const elements = el.querySelectorAll('input[name], select[name], textarea[name]');
    elements.forEach((field)=>{
      const name = field.name; if(!name) return;
      const type = (field.type||'').toLowerCase();
      if((type === 'checkbox' || type === 'radio') && !field.checked) return;
      if(field.tagName === 'SELECT' && field.multiple){
        const selected = Array.from(field.options).filter(o=>o.selected).map(o=>o.value);
        if(selected.length === 0) return;
        params[name] = selected[selected.length-1];
        return;
      }
      params[name] = field.value ?? '';
    });
    return params;
  }

  function readParams(el){
    const out = {}; if(!el || !el.attributes) return out;
    for(const attr of el.attributes){
      if(!attr.name) continue;
      if(attr.name.startsWith('data-flux-param-')){
        out[attr.name.substring('data-flux-param-'.length)] = attr.value;
      } else if (attr.name.startsWith('flux-param-')){
        out[attr.name.substring('flux-param-'.length)] = attr.value;
      }
    }
    return out;
  }

  // Expose on liveflux
  window.liveflux.executeScripts = executeScripts;
  window.liveflux.serializeElement = serializeElement;
  window.liveflux.readParams = readParams;

})();
