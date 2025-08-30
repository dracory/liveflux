(function(){
  const g = window; g.__lw = g.__lw || {};

  /**
   * Handles clicks on elements with data-flux-action, serializes data from the
   * nearest form (or component root), merges button params, and swaps the
   * component DOM with server-rendered HTML.
   * @param {MouseEvent} e - Click event.
   * @returns {void}
   */
  g.__lw.handleActionClick = function(e){
    const btn = e.target.closest('[data-flux-action], [flux-action]');
    if(!btn) return;
    const root = btn.closest('[data-flux-root], [flux-root]');
    if(!root) return;
    const comp = root.querySelector('input[name="component"]');
    const id = root.querySelector('input[name="id"]');
    if(!comp||!id) return;
    const action = btn.getAttribute('data-flux-action') || btn.getAttribute('flux-action');
    // Resolve the associated form: prefer closest form, otherwise HTML 'form' attribute
    const formId = btn.getAttribute('form');
    const assocForm = btn.closest('form') || (formId ? document.getElementById(formId) : null);

    // If this is a submit button tied to a form, let the submit handler take over to avoid double-execution
    const isSubmitBtn = (btn.tagName === 'BUTTON' || btn.tagName === 'INPUT') && (btn.getAttribute('type') || '').toLowerCase() === 'submit';
    if (assocForm && isSubmitBtn) {
      return; // submit event will be intercepted by handleFormSubmit
    }

    // Prevent default navigation (e.g., anchors) when we handle click ourselves
    e.preventDefault();

    const fields = assocForm ? g.__lw.serializeElement(assocForm) : g.__lw.serializeElement(root);
    const btnParams = g.__lw.readParams(btn);
    if (btn.name) { btnParams[btn.name] = btn.value; }
    const params = Object.assign({}, fields, btnParams, { component: comp.value, id: id.value, action });
    g.__lw.post(params).then((html)=>{
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode){ root.replaceWith(newNode); g.__lw.executeScripts(newNode); }
    }).catch((err)=>{ console.error('action', err); });
  };

  /**
   * Intercepts form submission within a Liveflux component, serializes form data,
   * augments with submitter data and data-flux-param-*, sends to server, and swaps
   * the component DOM with server-rendered HTML.
   * @param {SubmitEvent} e - Submit event.
   * @returns {void}
   */
  g.__lw.handleFormSubmit = function(e){
    const form = e.target.closest('[data-flux-root] form, [flux-root] form, form');
    if(!form) return;
    const root = form.closest('[data-flux-root], [flux-root]');
    if(!root) return;
    const comp = root.querySelector('input[name="component"]');
    const id = root.querySelector('input[name="id"]');
    if(!comp||!id) return;
    e.preventDefault();
    const submitter = e.submitter || root.querySelector('[data-flux-action], [flux-action]');
    const action = (submitter && (submitter.getAttribute('data-flux-action') || submitter.getAttribute('flux-action'))) || form.getAttribute('data-flux-action') || form.getAttribute('flux-action') || 'submit';
    const fields = g.__lw.serializeElement(form);
    if (submitter) {
      const extra = g.__lw.readParams(submitter);
      if (submitter.name) { extra[submitter.name] = submitter.value; }
      Object.assign(fields, extra);
    }
    const params = Object.assign({}, fields, { component: comp.value, id: id.value, action });
    g.__lw.post(params).then((html)=>{
      const tmp = document.createElement('div');
      tmp.innerHTML = html;
      const newNode = tmp.firstElementChild;
      if(newNode){ root.replaceWith(newNode); g.__lw.executeScripts(newNode); }
    }).catch((err)=>{ console.error('form submit', err); });
  };
})();

