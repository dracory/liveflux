(function(){
  if(!window.liveflux){
    console.log('[Liveflux Util] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const { dataFluxParam } = liveflux;
  const dataParamPrefix = `${dataFluxParam}-`;

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
      if(attr.name.startsWith(dataParamPrefix)){
        out[attr.name.substring(dataParamPrefix.length)] = attr.value;
      } else if (attr.name.startsWith('flux-param-')){
        out[attr.name.substring('flux-param-'.length)] = attr.value;
      }
    }
    return out;
  }

  // collectAllFields implements the form-less submission feature.
  // It collects fields from the default scope (form or root), then merges
  // fields from elements specified in data-flux-include, removes fields
  // from data-flux-exclude, and finally merges button params.
  function collectAllFields(btn, root, assocForm){
    if(!btn) return {};

    // 1. Serialize default scope (form or root)
    let fields = assocForm 
      ? serializeElement(assocForm) 
      : (root ? serializeElement(root) : {});

    // 2. Process data-flux-include
    const includeAttr = btn.getAttribute('data-flux-include') || 
                        btn.getAttribute('flux-include');
    if(includeAttr){
      const selectors = includeAttr.split(',').map(function(s){ return s.trim(); });
      selectors.forEach(function(selector){
        if(!selector) return;
        try {
          const elements = document.querySelectorAll(selector);
          if(elements.length === 0){
            console.warn('[Liveflux] Include selector "' + selector + '" matched no elements');
          }
          elements.forEach(function(el){
            const included = serializeElement(el);
            // Later sources override (last-write-wins)
            Object.assign(fields, included);
          });
        } catch(e) {
          console.error('[Liveflux] Invalid include selector "' + selector + '":', e);
        }
      });
    }

    // 3. Process data-flux-exclude
    const excludeAttr = btn.getAttribute('data-flux-exclude') || 
                        btn.getAttribute('flux-exclude');
    if(excludeAttr){
      const excludeSelectors = excludeAttr.split(',').map(function(s){ return s.trim(); });
      excludeSelectors.forEach(function(selector){
        if(!selector) return;
        try {
          const elements = document.querySelectorAll(selector);
          elements.forEach(function(el){
            const excluded = serializeElement(el);
            Object.keys(excluded).forEach(function(key){
              delete fields[key];
            });
          });
        } catch(e) {
          console.error('[Liveflux] Invalid exclude selector "' + selector + '":', e);
        }
      });
    }

    // 4. Merge button params (highest precedence)
    const btnParams = readParams(btn);
    if(btn.name){
      btnParams[btn.name] = btn.value;
    }
    Object.assign(fields, btnParams);

    return fields;
  }

  // resolveComponentMetadata attempts to find component type and ID from various sources.
  // Returns { comp: string, id: string, root: Element|null } or null if not found.
  function resolveComponentMetadata(btn, rootSelector){
    if(!btn) return null;

    // 1. Try nearest root (standard case) - read from data attributes
    let root = btn.closest(rootSelector);
    if(root){
      const comp = root.getAttribute(liveflux.dataFluxComponent || 'data-flux-component');
      const id = root.getAttribute(liveflux.dataFluxComponentID || 'data-flux-component-id');
      if(comp && id){
        return { comp: comp, id: id, root: root };
      }
    }

    // 2. Try explicit attributes on button (for buttons outside component root)
    const explicitComp = btn.getAttribute('data-flux-component-type');
    const explicitId = btn.getAttribute('data-flux-component-id');
    if(explicitComp && explicitId){
      return { comp: explicitComp, id: explicitId, root: null };
    }

    // 3. Try data attribute pointing to root by ID
    const rootId = btn.getAttribute('data-flux-root-id');
    if(rootId){
      root = document.getElementById(rootId);
      if(root){
        const comp = root.getAttribute(liveflux.dataFluxComponent || 'data-flux-component');
        const id = root.getAttribute(liveflux.dataFluxComponentID || 'data-flux-component-id');
        if(comp && id){
          return { comp: comp, id: id, root: root };
        }
      }
    }

    return null;
  }

  // Expose on liveflux
  liveflux.executeScripts = executeScripts;
  liveflux.serializeElement = serializeElement;
  liveflux.readParams = readParams;
  liveflux.collectAllFields = collectAllFields;
  liveflux.resolveComponentMetadata = resolveComponentMetadata;

})();
