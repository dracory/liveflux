(function(){
  if(!window.liveflux){
    console.log('[Liveflux Util] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const { dataFluxParam, dataFluxIndicator, dataFluxSelect } = liveflux;
  const REQUEST_CLASS = 'flux-request';
  const INDICATOR_ORIGINAL_DISPLAY_ATTR = 'data-liveflux-indicator-original-display';
  const dataParamPrefix = `${dataFluxParam}-`;
  const SELECT_LOG_PREFIX = '[Liveflux Select]';

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
    
    // Helper function to serialize a single field
    const serializeField = (field) => {
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
    };
    
    // If element itself is a form field, serialize it
    const tagName = el.tagName;
    if((tagName === 'INPUT' || tagName === 'SELECT' || tagName === 'TEXTAREA') && el.name){
      serializeField(el);
    }
    
    // Serialize child form fields
    const elements = el.querySelectorAll('input[name], select[name], textarea[name]');
    elements.forEach(serializeField);
    
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

  function readSelectAttribute(el){
    if(!el || typeof el.getAttribute !== 'function') return '';
    const candidates = [];
    if(dataFluxSelect){ candidates.push(dataFluxSelect); }
    candidates.push('flux-select');
    for(const name of candidates){
      if(!name) continue;
      const value = el.getAttribute(name);
      if(typeof value === 'string' && value.trim()){ return value.trim(); }
    }
    return '';
  }

  function getComponentRootSelector(){
    const kindAttr = liveflux.dataFluxComponentKind || 'data-flux-component-kind';
    const idAttr = liveflux.dataFluxComponentID || 'data-flux-component-id';
    return `[${kindAttr}][${idAttr}]`;
  }

  function isComponentRootNode(node){
    if(!node || typeof node.matches !== 'function') return false;
    return node.matches(getComponentRootSelector());
  }

  function extractSelectedFragment(html, selectors){
    if(!selectors || !selectors.trim()) return html;
    if(typeof DOMParser === 'undefined'){
      console.warn(`${SELECT_LOG_PREFIX} DOMParser unavailable; returning full response`);
      return html;
    }

    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const selectorList = selectors.split(',').map((s)=>s.trim()).filter(Boolean);

    for(const selector of selectorList){
      try {
        const match = doc.querySelector(selector);
        if(match){
          if(liveflux.debugSelect){
            console.debug(`${SELECT_LOG_PREFIX} Extracted fragment using selector "${selector}"`);
          }
          return match.outerHTML;
        }
      } catch (err){
        console.warn(`${SELECT_LOG_PREFIX} Invalid selector "${selector}"`, err);
      }
    }

    console.warn(`${SELECT_LOG_PREFIX} No matches found for selectors: ${selectorList.join(', ')}`);
    return html;
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

    const selector = rootSelector || getComponentRootSelector();

    // 1. Try nearest root (standard case) - read from data attributes
    let root = btn.closest(selector);
    if(root){
      const comp = root.getAttribute(liveflux.dataFluxComponentKind || 'data-flux-component-kind');
      const id = root.getAttribute(liveflux.dataFluxComponentID || 'data-flux-component-id');
      if(comp && id){
        return { comp: comp, id: id, root: root };
      }
    }

    // 2. Try explicit attributes on button (for buttons outside component root)
    const explicitComp = btn.getAttribute('data-flux-target-kind');
    const explicitId = btn.getAttribute('data-flux-target-id');
    if(explicitComp && explicitId){
      return { comp: explicitComp, id: explicitId, root: null };
    }

    return null;
  }

  function resolveIndicators(trigger, root){
    const targets = new Set();
    if(trigger){
      targets.add(trigger);
    }

    const attrName = dataFluxIndicator || 'data-flux-indicator';
    const attrFallback = 'flux-indicator';
    const indicatorAttr = trigger ? (trigger.getAttribute(attrName) || trigger.getAttribute(attrFallback)) : null;

    const selectors = [];
    if(indicatorAttr){
      indicatorAttr.split(',').forEach(function(sel){
        const trimmed = (sel || '').trim();
        if(trimmed){ selectors.push(trimmed); }
      });
    }

    selectors.forEach(function(selector){
      if(selector === 'this' && trigger){
        targets.add(trigger);
        return;
      }
      try {
        document.querySelectorAll(selector).forEach(function(node){ targets.add(node); });
      } catch(err) {
        console.error('[Liveflux] Invalid indicator selector "' + selector + '"', err);
      }
    });

    if(!indicatorAttr){
      const fallbackSelector = '.flux-indicator';
      if(root){
        root.querySelectorAll(fallbackSelector).forEach(function(node){ targets.add(node); });
      }
      if(trigger && trigger.matches(fallbackSelector)){
        targets.add(trigger);
      }
    }

    return Array.from(targets.values()).filter(Boolean);
  }

  function startRequestIndicators(trigger, root){
    const elements = resolveIndicators(trigger, root);
    elements.forEach(function(el){
      if(!el.classList.contains('flux-indicator') && !el.hasAttribute(INDICATOR_ORIGINAL_DISPLAY_ATTR)){
        let currentDisplay = '';
        if(typeof window !== 'undefined' && window.getComputedStyle){
          currentDisplay = window.getComputedStyle(el).display;
        }
        if(!currentDisplay){
          currentDisplay = el.style.display || '';
        }
        if(currentDisplay === 'none'){
          el.setAttribute(INDICATOR_ORIGINAL_DISPLAY_ATTR, el.style.display || '');
          el.style.display = 'inline-block';
        }
      }
      el.classList.add(REQUEST_CLASS);
    });
    return elements;
  }

  function endRequestIndicators(elements){
    if(!elements) return;
    elements.forEach(function(el){
      el.classList.remove(REQUEST_CLASS);
      if(el.hasAttribute(INDICATOR_ORIGINAL_DISPLAY_ATTR)){
        const originalDisplay = el.getAttribute(INDICATOR_ORIGINAL_DISPLAY_ATTR);
        const shouldRestore = el.style.display === 'inline-block';
        if(shouldRestore){
          if(originalDisplay){
            el.style.display = originalDisplay;
          } else {
            el.style.removeProperty('display');
          }
        }
        el.removeAttribute(INDICATOR_ORIGINAL_DISPLAY_ATTR);
      }
    });
  }

  // Expose on liveflux
  liveflux.executeScripts = executeScripts;
  liveflux.serializeElement = serializeElement;
  liveflux.readParams = readParams;
  liveflux.readSelectAttribute = readSelectAttribute;
  liveflux.getComponentRootSelector = getComponentRootSelector;
  liveflux.isComponentRootNode = isComponentRootNode;
  liveflux.extractSelectedFragment = extractSelectedFragment;
  liveflux.collectAllFields = collectAllFields;
  liveflux.resolveComponentMetadata = resolveComponentMetadata;
  liveflux.resolveIndicators = resolveIndicators;
  liveflux.startRequestIndicators = startRequestIndicators;
  liveflux.endRequestIndicators = endRequestIndicators;

})();
