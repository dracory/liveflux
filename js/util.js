(function(){
  const g = window;
  g.__lw = g.__lw || {};
  // Default endpoint; can be overridden by integrators before or via JSWithEndpoint
  if (!g.__lw.endpoint) {
    g.__lw.endpoint = '/liveflux';
  }

  /**
   * Executes any <script> tags within the provided root node by cloning them,
   * ensuring inline scripts run after DOM replacement.
   * @param {Element|DocumentFragment|null} root - Container to scan for script tags.
   * @returns {void}
   */
  g.__lw.executeScripts = function(root){
    if(!root) return;
    const scripts = root.querySelectorAll('script');
    scripts.forEach((old)=>{
      const s = document.createElement('script');
      // copy attributes
      for (const attr of old.attributes) s.setAttribute(attr.name, attr.value);
      s.text = old.textContent || '';
      old.replaceWith(s);
    });
  };

  /**
   * Serializes form-like fields inside an element into a flat key/value object.
   * - Ignores unchecked checkboxes and radios.
   * - For multi-select, returns the last selected value (kept for backward compatibility).
   * @param {Element|null} el - Root element containing input/select/textarea fields.
   * @returns {Record<string, string>} Key/value map of form data.
   */
  g.__lw.serializeElement = function(el){
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
  };

  /**
   * Reads params from data-flux-param-<name> and flux-param-<name> attributes.
   * @param {Element|null} el - Element to read params from.
   * @returns {Record<string, string>} Map of param name to value.
   */
  g.__lw.readParams = function(el){
    const out = {};
    if(!el || !el.attributes) return out;
    for(const attr of el.attributes){
      if(!attr.name) continue;
      if(attr.name.startsWith('data-flux-param-')){
        out[attr.name.substring('data-flux-param-'.length)] = attr.value;
      } else if (attr.name.startsWith('flux-param-')){
        out[attr.name.substring('flux-param-'.length)] = attr.value;
      }
    }
    return out;
  };
})();
