(function(){
  const g = window; g.__lw = g.__lw || {};

  /**
   * Performs a POST request to the Liveflux endpoint and returns HTML.
   * @param {Record<string, string>} params - Key/value pairs to send as form data.
   * @returns {Promise<string>} Resolves with response text (HTML) or rejects on HTTP error.
   */
  g.__lw.post = async function(params){
    const body = new URLSearchParams(params);
    const res = await fetch(g.__lw.endpoint || '/liveflux',{
      method:'POST',
      headers:{'Content-Type':'application/x-www-form-urlencoded','Accept':'text/html'},
      body,
      credentials:'same-origin'
    });
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
