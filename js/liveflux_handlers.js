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
    const root = btn.closest(rootSelectorWithFallback);
    if(!root) return;
    const comp = root.querySelector('input[name="liveflux_component_type"]');
    const id = root.querySelector('input[name="liveflux_component_id"]');
    if(!comp||!id) return;

    const action = btn.getAttribute(dataFluxAction) || btn.getAttribute('flux-action');
    const formId = btn.getAttribute('form');
    const assocForm = btn.closest('form') || (formId ? document.getElementById(formId) : null);

    const isSubmitBtn = (btn.tagName === 'BUTTON' || btn.tagName === 'INPUT') && (btn.getAttribute('type') || '').toLowerCase() === 'submit';
    if (assocForm && isSubmitBtn) { return; }

    e.preventDefault();

    const fields = assocForm ? liveflux.serializeElement(assocForm) : liveflux.serializeElement(root);
    const btnParams = liveflux.readParams(btn);
    if (btn.name) { btnParams[btn.name] = btn.value; }

    const params = Object.assign({}, fields, btnParams, {
      liveflux_component_type: comp.value,
      liveflux_component_id: id.value,
      liveflux_action: action
    });

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
    }).catch((err)=>{ console.error('action', err); });
  }

  function handleFormSubmit(e){
    const form = e.target.closest(`${rootSelector} form, [flux-root] form, form`);
    if(!form) return;
    const root = form.closest(rootSelectorWithFallback);
    if(!root) return;
    const comp = root.querySelector('input[name="liveflux_component_type"]');
    const id = root.querySelector('input[name="liveflux_component_id"]');
    if(!comp||!id) return;
    e.preventDefault();

    const submitter = e.submitter || root.querySelector(actionSelectorWithFallback);
    const action = (submitter && (submitter.getAttribute(dataFluxAction) || submitter.getAttribute('flux-action'))) || form.getAttribute(dataFluxAction) || form.getAttribute('flux-action') || 'submit';

    const fields = liveflux.serializeElement(form);
    if (submitter) {
      const extra = liveflux.readParams(submitter);
      if (submitter.name) { extra[submitter.name] = submitter.value; }
      Object.assign(fields, extra);
    }

    const params = Object.assign({}, fields, { liveflux_component_type: comp.value, liveflux_component_id: id.value, liveflux_action: action });
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
