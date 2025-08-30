(function(){
  if(window.__lwInitDone) return; window.__lwInitDone = true;
  function init(){
    function executeScripts(root){
      if(!root) return;
      const scripts = root.querySelectorAll('script');
      scripts.forEach((old)=>{
        const s = document.createElement('script');
        // copy attributes
        for (const attr of old.attributes) s.setAttribute(attr.name, attr.value);
        s.text = old.textContent || '';
        old.replaceWith(s);
      });
    }

    async function post(params){
      const body = new URLSearchParams(params);
      const res = await fetch((window.__lw && window.__lw.endpoint) || '/liveflux',{
        method:'POST',
        headers:{'Content-Type':'application/x-www-form-urlencoded','Accept':'text/html'},
        body,
        credentials:'same-origin'
      });
      if(!res.ok) throw new Error(''+res.status);
      return await res.text();
    }

    function serializeElement(el){
      const params = {};
      if(!el) return params;
      const elements = el.querySelectorAll('input[name], select[name], textarea[name]');
      elements.forEach((field)=>{
        const name = field.name;
        if(!name) return;
        const type = (field.type||'').toLowerCase();
        if((type === 'checkbox' || type === 'radio') && !field.checked) return;
        if(field.tagName === 'SELECT' && field.multiple){
          const selected = Array.from(field.options).filter(o=>o.selected).map(o=>o.value);
          if(selected.length === 0) return;
          // For multiple, keep last value for simplicity; extend if array support is needed
          params[name] = selected[selected.length-1];
          return;
        }
        params[name] = field.value ?? '';
      });
      return params;
    }

    function readParams(el){
      const out = {};
      if(!el || !el.attributes) return out;
      for(const attr of el.attributes){
        if(attr.name && attr.name.startsWith('data-lw-param-')){
          out[attr.name.substring('data-lw-param-'.length)] = attr.value;
        }
      }
      return out;
    }

    function mountPlaceholders(){
      document.querySelectorAll('[data-lw-mount="1"]').forEach((el)=>{
        const component = el.getAttribute('data-lw-component');
        if(!component) return;
        const params = readParams(el);
        params.component = component;
        post(params).then((html)=>{
          const tmp = document.createElement('div');
          tmp.innerHTML = html;
          const newNode = tmp.firstElementChild;
          if(newNode){ el.replaceWith(newNode); executeScripts(newNode); }
        }).catch((err)=>{ console.error(component+' mount', err); });
      });
    }

    function handleActionClick(e){
      const btn = e.target.closest('[data-lw-action]');
      if(!btn) return;
      const root = btn.closest('[data-lw-root]');
      if(!root) return;
      const comp = root.querySelector('input[name="component"]');
      const id = root.querySelector('input[name="id"]');
      if(!comp||!id) return;
      const action = btn.getAttribute('data-lw-action');
      const form = btn.closest('form');
      const fields = form ? serializeElement(form) : serializeElement(root);
      // Also include button-provided params
      const btnParams = readParams(btn);
      if (btn.name) { btnParams[btn.name] = btn.value; }
      const params = Object.assign({}, fields, btnParams, { component: comp.value, id: id.value, action });
      post(params).then((html)=>{
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        const newNode = tmp.firstElementChild;
        if(newNode){ root.replaceWith(newNode); executeScripts(newNode); }
      }).catch((err)=>{ console.error('action', err); });
    }

    function handleFormSubmit(e){
      const form = e.target.closest('[data-lw-root] form, form');
      if(!form) return;
      const root = form.closest('[data-lw-root]');
      if(!root) return;
      const comp = root.querySelector('input[name="component"]');
      const id = root.querySelector('input[name="id"]');
      if(!comp||!id) return;
      e.preventDefault();
      const submitter = e.submitter || root.querySelector('[data-lw-action]');
      const action = (submitter && submitter.getAttribute('data-lw-action')) || form.getAttribute('data-lw-action') || 'submit';
      const fields = serializeElement(form);
      // Include submitter name/value and data-lw-param-* (e.g., page number)
      if (submitter) {
        const extra = readParams(submitter);
        if (submitter.name) { extra[submitter.name] = submitter.value; }
        Object.assign(fields, extra);
      }
      const params = Object.assign({}, fields, { component: comp.value, id: id.value, action });
      post(params).then((html)=>{
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        const newNode = tmp.firstElementChild;
        if(newNode){ root.replaceWith(newNode); executeScripts(newNode); }
      }).catch((err)=>{ console.error('form submit', err); });
    }

    document.addEventListener('click', handleActionClick);
    document.addEventListener('submit', handleFormSubmit);
    mountPlaceholders();
  }

  if(document.readyState === 'loading'){
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
