(function(){
  const g = window; g.__lw = g.__lw || {};

  /**
   * Performs a POST request to the Liveflux endpoint and returns HTML.
   * @param {Record<string, string>} params - Key/value pairs to send as form data.
   * @returns {Promise<string>} Resolves with response text (HTML) or rejects on HTTP error.
   */
  g.__lw.post = async function(params){
    const body = new URLSearchParams(params);
    const endpoint = g.__lw.endpoint || '/liveflux';
    const headers = Object.assign({
      'Content-Type':'application/x-www-form-urlencoded',
      'Accept':'text/html'
    }, g.__lw.headers || {});
    const credentials = g.__lw.credentials || 'same-origin';
    const timeoutMs = g.__lw.timeoutMs || 0;

    const controller = (timeoutMs > 0 && 'AbortController' in g) ? new AbortController() : null;
    let timeoutId = null;
    if (controller && timeoutMs > 0) {
      timeoutId = setTimeout(()=>controller.abort(), timeoutMs);
    }

    const res = await fetch(endpoint,{
      method:'POST',
      headers,
      body,
      credentials,
      signal: controller ? controller.signal : undefined,
    }).finally(()=>{ if (timeoutId) clearTimeout(timeoutId); });
    if(!res.ok) throw new Error(''+res.status);
    const redirect = res.headers.get('X-Liveflux-Redirect');
    if (redirect) {
      const after = res.headers.get('X-Liveflux-Redirect-After');
      const delayMs = after ? (parseInt(after,10) * 1000 || 0) : 0;
      if (delayMs > 0) {
        setTimeout(() => { window.location.href = redirect; }, delayMs);
      } else {
        window.location.href = redirect;
      }
      // Return empty HTML (caller should ignore rendered body when redirecting)
      return '';
    }
    return await res.text();
  };
})();
