(function(){
  if(!window.liveflux){
    console.log('[Liveflux Handlers] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const { dataFluxAction, dataFluxRoot } = liveflux;

  const actionSelector = `[${dataFluxAction}]`;
  const actionSelectorWithFallback = `${actionSelector}, [flux-action]`;
  const rootSelector = `[${dataFluxRoot}]`;
  const rootSelectorWithFallback = `${rootSelector}, [flux-root]`;
  const SELECT_LOG_PREFIX = '[Liveflux Select]';

  // Track in-flight requests per component ID to prevent concurrent requests
  const pendingRequests = new Map();

  function parseSelectors(selectAttr){
    if(!selectAttr) return [];
    return selectAttr.split(',').map(function(s){ return s.trim(); }).filter(Boolean);
  }

  function isComponentRootNode(node){
    if(!node || typeof node.hasAttribute !== 'function') return false;
    const rootAttr = dataFluxRoot || 'data-flux-root';
    return node.hasAttribute(rootAttr) || node.hasAttribute('flux-root');
  }

  function applySelectedFragment(root, selectors, newNode){
    if(!root || !newNode || !selectors || selectors.length === 0) return false;
    for(const selector of selectors){
      try {
        const target = root.querySelector(selector);
        if(target){
          target.replaceWith(newNode);
          return true;
        }
      } catch(err){
        console.warn(`${SELECT_LOG_PREFIX} Failed to apply selector "${selector}"`, err);
      }
    }
    return false;
  }

  function handleActionClick(e){
    const btn = e.target.closest(actionSelectorWithFallback);
    if(!btn) return;

    // Resolve component metadata with fallback chain
    const metadata = liveflux.resolveComponentMetadata(btn, rootSelectorWithFallback);
    if(!metadata) return;

    const action = btn.getAttribute(dataFluxAction) || btn.getAttribute('flux-action');
    const selectAttr = liveflux.readSelectAttribute ? liveflux.readSelectAttribute(btn) : '';
    const formId = btn.getAttribute('form');
    const assocForm = btn.closest('form') || (formId ? document.getElementById(formId) : null);

    const isSubmitBtn = (btn.tagName === 'BUTTON' || btn.tagName === 'INPUT') && (btn.getAttribute('type') || '').toLowerCase() === 'submit';
    if (assocForm && isSubmitBtn) { return; }

    e.preventDefault();

    // Check if there's already a pending request for this component
    if(pendingRequests.has(metadata.id)){
      console.log('[Liveflux] Skipping action - request already in progress for component:', metadata.id);
      return;
    }

    // Use collectAllFields to support data-flux-include and data-flux-exclude
    const fields = liveflux.collectAllFields(btn, metadata.root, assocForm);

    const params = Object.assign({}, fields, {
      liveflux_component_kind: metadata.comp,
      liveflux_component_id: metadata.id,
      liveflux_action: action
    });

    // Mark this component as having a pending request
    pendingRequests.set(metadata.id, true);

    const indicatorEls = liveflux.startRequestIndicators(btn, metadata.root);

    liveflux.post(params).then((result)=>{
      const rawHtml = result.html || result;
      
      // Check if response contains target templates
      if(liveflux.hasTargetTemplates && liveflux.hasTargetTemplates(rawHtml)){
        const fallback = liveflux.applyTargets(rawHtml, metadata.root);
        if(fallback){
          // Targets failed, do full replacement with fallback HTML
          const tmp = document.createElement('div');
          tmp.innerHTML = fallback;
          const newNode = tmp.firstElementChild;
          if(newNode && metadata.root){
            metadata.root.replaceWith(newNode);
            liveflux.executeScripts(newNode);
          }
        }
        if(liveflux.initWire) liveflux.initWire();
        return;
      }
      
      // Traditional flow with data-flux-select support
      const html = liveflux.extractSelectedFragment ? liveflux.extractSelectedFragment(rawHtml, selectAttr) : rawHtml;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      let newNode = tmp.firstElementChild;
      const selectors = parseSelectors(selectAttr);
      if(newNode && selectors.length && metadata.root && !isComponentRootNode(newNode)){
        const applied = applySelectedFragment(metadata.root, selectors, newNode);
        if(applied){
          liveflux.executeScripts(newNode);
          if(liveflux.initWire) liveflux.initWire();
          return;
        }
        newNode = tmp.firstElementChild;
      }
      if(newNode && metadata.root){
        metadata.root.replaceWith(newNode);
        liveflux.executeScripts(newNode);
        if(liveflux.initWire) liveflux.initWire();
      } else if(newNode && !metadata.root){
        // Button was outside root - try to find root by ID to replace
        const targetRoot = document.querySelector('[data-flux-component-id="' + metadata.id + '"]');
        if(targetRoot){
          if(selectors.length && !isComponentRootNode(newNode)){
            const applied = applySelectedFragment(targetRoot, selectors, newNode);
            if(applied){
              liveflux.executeScripts(newNode);
              if(liveflux.initWire) liveflux.initWire();
              return;
            }
            newNode = tmp.firstElementChild;
          }
          targetRoot.replaceWith(newNode);
          liveflux.executeScripts(newNode);
          if(liveflux.initWire) liveflux.initWire();
        }
      }
    }).catch((err)=>{ console.error('action', err); })
      .finally(()=>{
        liveflux.endRequestIndicators(indicatorEls);
        // Clear the pending request flag
        pendingRequests.delete(metadata.id);
      });
  }

  function handleFormSubmit(e){
    const form = e.target.closest(`${rootSelector} form, [flux-root] form, form`);
    if(!form) return;
    const root = form.closest(rootSelectorWithFallback);
    if(!root) return;
    const comp = root.getAttribute(liveflux.dataFluxComponent || 'data-flux-component');
    const id = root.getAttribute(liveflux.dataFluxComponentID || 'data-flux-component-id');
    if(!comp||!id) return;
    e.preventDefault();

    const submitter = e.submitter || root.querySelector(actionSelectorWithFallback);
    const selectAttr = submitter
      ? (liveflux.readSelectAttribute ? liveflux.readSelectAttribute(submitter) : '')
      : (liveflux.readSelectAttribute ? liveflux.readSelectAttribute(form) : '');
    const action = (submitter && (submitter.getAttribute(dataFluxAction) || submitter.getAttribute('flux-action'))) || form.getAttribute(dataFluxAction) || form.getAttribute('flux-action') || 'submit';

    // Use collectAllFields to support data-flux-include and data-flux-exclude on submitter
    const fields = submitter 
      ? liveflux.collectAllFields(submitter, root, form)
      : liveflux.serializeElement(form);

    const params = Object.assign({}, fields, { liveflux_component_kind: comp, liveflux_component_id: id, liveflux_action: action });
    const indicatorEls = liveflux.startRequestIndicators(submitter || form, root);

    liveflux.post(params).then((result)=>{
      const rawHtml = result.html || result;
      
      // Check if response contains target templates
      if(liveflux.hasTargetTemplates && liveflux.hasTargetTemplates(rawHtml)){
        const fallback = liveflux.applyTargets(rawHtml, root);
        if(fallback){
          // Targets failed, do full replacement with fallback HTML
          const tmp = document.createElement('div');
          tmp.innerHTML = fallback;
          const newNode = tmp.firstElementChild;
          if(newNode){
            root.replaceWith(newNode);
            liveflux.executeScripts(newNode);
          }
        }
        if(liveflux.initWire) liveflux.initWire();
        return;
      }
      
      // Traditional flow with data-flux-select support
      const html = liveflux.extractSelectedFragment ? liveflux.extractSelectedFragment(rawHtml, selectAttr) : rawHtml;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      let newNode = tmp.firstElementChild;
      const selectors = parseSelectors(selectAttr);
      if(newNode && selectors.length && !isComponentRootNode(newNode)){
        const applied = applySelectedFragment(root, selectors, newNode);
        if(applied){
          liveflux.executeScripts(newNode);
          if(liveflux.initWire) liveflux.initWire();
          return;
        }
        newNode = tmp.firstElementChild;
      }
      if(newNode){
        root.replaceWith(newNode);
        liveflux.executeScripts(newNode);
        if(liveflux.initWire) liveflux.initWire();
      }
    }).catch((err)=>{ console.error('form submit', err); })
      .finally(()=>{
        liveflux.endRequestIndicators(indicatorEls);
      });
  }

  // Expose
  liveflux.handleActionClick = handleActionClick;
  liveflux.handleFormSubmit = handleFormSubmit;
})();
