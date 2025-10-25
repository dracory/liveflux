(function(){
  if(!window.liveflux){
    console.log('[Liveflux Handlers] liveflux namespace not found');
    return;
  }

  function handleActionClick(e){
    const btn = e.target.closest('[data-flux-action], [flux-action]');
    if(!btn) return;
    const root = btn.closest('[data-flux-root], [flux-root]');
    if(!root) return;
    const comp = root.querySelector('input[name="liveflux_component_type"]');
    const id = root.querySelector('input[name="liveflux_component_id"]');
    if(!comp||!id) return;

    const action = btn.getAttribute('data-flux-action') || btn.getAttribute('flux-action');
    const formId = btn.getAttribute('form');
    const assocForm = btn.closest('form') || (formId ? document.getElementById(formId) : null);

    const isSubmitBtn = (btn.tagName === 'BUTTON' || btn.tagName === 'INPUT') && (btn.getAttribute('type') || '').toLowerCase() === 'submit';
    if (assocForm && isSubmitBtn) { return; }

    e.preventDefault();

    const fields = assocForm ? window.liveflux.serializeElement(assocForm) : window.liveflux.serializeElement(root);
    const btnParams = window.liveflux.readParams(btn);
    if (btn.name) { btnParams[btn.name] = btn.value; }

    const params = Object.assign({}, fields, btnParams, {
      liveflux_component_type: comp.value,
      liveflux_component_id: id.value,
      liveflux_action: action
    });

    window.liveflux.post(params).then((result)=>{
      const html = result.html || result;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode){
        root.replaceWith(newNode);
        window.liveflux.executeScripts(newNode);
        if(window.liveflux.initWire) window.liveflux.initWire();
      }
    }).catch((err)=>{ console.error('action', err); });
  }

  function handleFormSubmit(e){
    const form = e.target.closest('[data-flux-root] form, [flux-root] form, form');
    if(!form) return;
    const root = form.closest('[data-flux-root], [flux-root]');
    if(!root) return;
    const comp = root.querySelector('input[name="liveflux_component_type"]');
    const id = root.querySelector('input[name="liveflux_component_id"]');
    if(!comp||!id) return;
    e.preventDefault();

    const submitter = e.submitter || root.querySelector('[data-flux-action], [flux-action]');
    const action = (submitter && (submitter.getAttribute('data-flux-action') || submitter.getAttribute('flux-action'))) || form.getAttribute('data-flux-action') || form.getAttribute('flux-action') || 'submit';

    const fields = window.liveflux.serializeElement(form);
    if (submitter) {
      const extra = window.liveflux.readParams(submitter);
      if (submitter.name) { extra[submitter.name] = submitter.value; }
      Object.assign(fields, extra);
    }

    const params = Object.assign({}, fields, { liveflux_component_type: comp.value, liveflux_component_id: id.value, liveflux_action: action });
    window.liveflux.post(params).then((result)=>{
      const html = result.html || result;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode){
        root.replaceWith(newNode);
        window.liveflux.executeScripts(newNode);
        if(window.liveflux.initWire) window.liveflux.initWire();
      }
    }).catch((err)=>{ console.error('form submit', err); });
  }

  // Expose
  window.liveflux.handleActionClick = handleActionClick;
  window.liveflux.handleFormSubmit = handleFormSubmit;
})();
