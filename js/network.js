// (function(){
//   /**
//    * DEPRECATED: Use liveflux_network.js. This file remains for backward compatibility and will be removed.
//    */
//   const g = window; g.__lw = g.__lw || {};

//   // Header names are provided by Go via ClientOptions; fall back to defaults
//   const HDR_REDIRECT = g.__lw.redirectHeader || 'X-Liveflux-Redirect';
//   const HDR_REDIRECT_AFTER = g.__lw.redirectAfterHeader || 'X-Liveflux-Redirect-After';

//   /**
//    * Performs a POST request to the Liveflux endpoint and returns HTML.
//    * @param {Record<string, string>} params - Key/value pairs to send as form data.
//    * @returns {Promise<{html: string, response: Response}>} Resolves with response text (HTML) and response object.
//    */
//   g.__lw.post = async function(params){
//     const body = new URLSearchParams(params);
//     const endpoint = g.__lw.endpoint || '/liveflux';
//     const headers = Object.assign({
//       'Content-Type':'application/x-www-form-urlencoded',
//       'Accept':'text/html'
//     }, g.__lw.headers || {});
//     const credentials = g.__lw.credentials || 'same-origin';
//     const timeoutMs = g.__lw.timeoutMs || 0;

//     const controller = (timeoutMs > 0 && 'AbortController' in g) ? new AbortController() : null;
//     let timeoutId = null;
//     if (controller && timeoutMs > 0) {
//       timeoutId = setTimeout(()=>controller.abort(), timeoutMs);
//     }

//     const res = await fetch(endpoint,{
//       method:'POST',
//       headers,
//       body,
//       credentials,
//       signal: controller ? controller.signal : undefined,
//     }).finally(()=>{ if (timeoutId) clearTimeout(timeoutId); });
//     if(!res.ok) throw new Error(''+res.status);
    
//     // Process events from response
//     const componentId = params.liveflux_component_id || '';
//     const componentAlias = params.liveflux_component_type || '';
//     if(g.__lw.processEvents){
//       g.__lw.processEvents(res, componentId, componentAlias);
//     }
    
//     const redirect = res.headers.get(HDR_REDIRECT);
//     if (redirect) {
//       const after = res.headers.get(HDR_REDIRECT_AFTER);
//       const delayMs = after ? (parseInt(after,10) * 1000 || 0) : 0;
//       if (delayMs > 0) {
//         setTimeout(() => { window.location.href = redirect; }, delayMs);
//       } else {
//         window.location.href = redirect;
//       }
//       // Return empty HTML (caller should ignore rendered body when redirecting)
//       return {html: '', response: res};
//     }
//     const html = await res.text();
//     return {html: html, response: res};
//   };
// })();
