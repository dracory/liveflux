(function(){
  const g = window; g.liveflux = g.liveflux || {};

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

    const fields = assocForm ? g.liveflux.serializeElement(assocForm) : g.liveflux.serializeElement(root);
    const btnParams = g.liveflux.readParams(btn);
    if (btn.name) { btnParams[btn.name] = btn.value; }

    const params = Object.assign({}, fields, btnParams, {
      liveflux_component_type: comp.value,
      liveflux_component_id: id.value,
      liveflux_action: action
    });

    g.liveflux.post(params).then((result)=>{
      const html = result.html || result;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode){
        root.replaceWith(newNode);
        g.liveflux.executeScripts(newNode);
        if(g.liveflux.initWire) g.liveflux.initWire();
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

    const fields = g.liveflux.serializeElement(form);
    if (submitter) {
      const extra = g.liveflux.readParams(submitter);
      if (submitter.name) { extra[submitter.name] = submitter.value; }
      Object.assign(fields, extra);
    }

    const params = Object.assign({}, fields, { liveflux_component_type: comp.value, liveflux_component_id: id.value, liveflux_action: action });
    g.liveflux.post(params).then((result)=>{
      const html = result.html || result;
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode){
        root.replaceWith(newNode);
        g.liveflux.executeScripts(newNode);
        if(g.liveflux.initWire) g.liveflux.initWire();
      }
    }).catch((err)=>{ console.error('form submit', err); });
  }

  // Expose
  g.liveflux.handleActionClick = handleActionClick;
  g.liveflux.handleFormSubmit = handleFormSubmit;
})();
