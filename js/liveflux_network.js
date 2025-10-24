(function(){
  // Ensure liveflux namespace exists
  if(!window.liveflux){
    console.log('[Liveflux Network] liveflux namespace not found');
    window.liveflux = window.liveflux || {};
  }
  const g = window;

  // Config defaults on liveflux
  if(!g.liveflux.endpoint){ g.liveflux.endpoint = '/liveflux'; }

  /**
   * Performs a POST request to the Liveflux endpoint and returns HTML.
   * @param {Record<string, string>} params
   * @returns {Promise<{html: string, response: Response}>}
   */
  function post(params){
    const body = new URLSearchParams(params);
    const endpoint = g.liveflux.endpoint || '/liveflux';
    const headers = Object.assign({
      'Content-Type':'application/x-www-form-urlencoded',
      'Accept':'text/html'
    }, g.liveflux.headers || {});
    const credentials = g.liveflux.credentials || 'same-origin';
    const timeoutMs = g.liveflux.timeoutMs || 0;

    const controller = (timeoutMs > 0 && 'AbortController' in g) ? new AbortController() : null;
    let timeoutId = null;
    if (controller && timeoutMs > 0) timeoutId = setTimeout(()=>controller.abort(), timeoutMs);

    const HDR_REDIRECT = g.liveflux.redirectHeader || 'X-Liveflux-Redirect';
    const HDR_REDIRECT_AFTER = g.liveflux.redirectAfterHeader || 'X-Liveflux-Redirect-After';

    return fetch(endpoint,{
      method:'POST', headers, body, credentials,
      signal: controller ? controller.signal : undefined,
    }).finally(()=>{ if (timeoutId) clearTimeout(timeoutId); })
      .then(async (res)=>{
        if(!res.ok) throw new Error(''+res.status);

        // Process events from response
        const componentId = params.liveflux_component_id || '';
        const componentAlias = params.liveflux_component_type || '';
        if(g.liveflux.events && g.liveflux.events.processEvents){
          g.liveflux.events.processEvents(res, componentId, componentAlias);
        }

        const redirect = res.headers.get(HDR_REDIRECT);
        if (redirect) {
          const after = res.headers.get(HDR_REDIRECT_AFTER);
          const delayMs = after ? (parseInt(after,10) * 1000 || 0) : 0;
          if (delayMs > 0) setTimeout(()=>{ g.location.href = redirect; }, delayMs);
          else g.location.href = redirect;
          return { html:'', response: res };
        }
        const html = await res.text();
        return { html, response: res };
      });
  }

  // Expose on liveflux
  g.liveflux.post = post;

  // Back-compat bridge (__lw -> liveflux)
  g.__lw = g.__lw || {};
  g.__lw.post = post;
})();
