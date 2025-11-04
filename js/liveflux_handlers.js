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

  function handleActionClick(e){
    const btn = e.target.closest(actionSelectorWithFallback);
    if(!btn) return;

    // Resolve component metadata with fallback chain
    const metadata = liveflux.resolveComponentMetadata(btn, rootSelectorWithFallback);
    if(!metadata) return;

    const action = btn.getAttribute(dataFluxAction) || btn.getAttribute('flux-action');
    const formId = btn.getAttribute('form');
    const assocForm = btn.closest('form') || (formId ? document.getElementById(formId) : null);

    const isSubmitBtn = (btn.tagName === 'BUTTON' || btn.tagName === 'INPUT') && (btn.getAttribute('type') || '').toLowerCase() === 'submit';
    if (assocForm && isSubmitBtn) { return; }

    e.preventDefault();

    // Use collectAllFields to support data-flux-include and data-flux-exclude
    const fields = liveflux.collectAllFields(btn, metadata.root, assocForm);

    const params = Object.assign({}, fields, {
      liveflux_component_type: metadata.comp,
      liveflux_component_id: metadata.id,
      liveflux_action: action
    });

    liveflux.post(params).then((result)=>{
      const html = result.html || result;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode && metadata.root){
        metadata.root.replaceWith(newNode);
        liveflux.executeScripts(newNode);
        if(liveflux.initWire) liveflux.initWire();
      } else if(newNode && !metadata.root){
        // Button was outside root - try to find root by ID to replace
        const targetRoot = document.querySelector('[data-flux-component-id="' + metadata.id + '"]');
        if(targetRoot){
          targetRoot.replaceWith(newNode);
          liveflux.executeScripts(newNode);
          if(liveflux.initWire) liveflux.initWire();
        }
      }
    }).catch((err)=>{ console.error('action', err); });
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
    const action = (submitter && (submitter.getAttribute(dataFluxAction) || submitter.getAttribute('flux-action'))) || form.getAttribute(dataFluxAction) || form.getAttribute('flux-action') || 'submit';

    // Use collectAllFields to support data-flux-include and data-flux-exclude on submitter
    const fields = submitter 
      ? liveflux.collectAllFields(submitter, root, form)
      : liveflux.serializeElement(form);

    const params = Object.assign({}, fields, { liveflux_component_type: comp, liveflux_component_id: id, liveflux_action: action });
    liveflux.post(params).then((result)=>{
      const html = result.html || result;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode){
        root.replaceWith(newNode);
        liveflux.executeScripts(newNode);
        if(liveflux.initWire) liveflux.initWire();
      }
    }).catch((err)=>{ console.error('form submit', err); });
  }

  // Expose
  liveflux.handleActionClick = handleActionClick;
  liveflux.handleFormSubmit = handleFormSubmit;
})();
