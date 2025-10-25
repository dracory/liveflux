(function(){
  if(!window.liveflux){
    console.log('[Liveflux WebSocket] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const {
    dataFluxAction,
    dataFluxComponentID,
    dataFluxWS,
    dataFluxWSURL,
  } = liveflux;

  const actionSelector = `[${dataFluxAction}]`;
  const actionSelectorWithFallback = `${actionSelector}`;
  const componentIdSelector = (id) => `[${dataFluxComponentID}="${id}"]`;
  const wsSelector = `[${dataFluxWS}]`;

  class LiveFluxWS {
    constructor(url, options = {}){
      this.url = url;
      this.ws = null;
      this.connected = false;
      this.reconnectAttempts = 0;
      this.maxReconnectAttempts = options.maxReconnectAttempts || 5;
      this.reconnectDelay = options.reconnectDelay || 1000;
      this.componentID = options.componentID || null;
      this.rootEl = options.rootEl || document;
      this.onOpen = options.onOpen || (()=>{});
      this.onMessage = options.onMessage || (()=>{});
      this.onClose = options.onClose || (()=>{});
      this.onError = options.onError || (()=>{});
      this.connect();
      this.setupFormHandling();
    }
    connect(){
      try {
        this.ws = new WebSocket(this.url);
        this.ws.onopen = this.handleOpen.bind(this);
        this.ws.onmessage = this.handleMessage.bind(this);
        this.ws.onclose = this.handleClose.bind(this);
        this.ws.onerror = this.handleError.bind(this);
      } catch (e){ console.error('[LFWS] connection error', e); this.handleError(e); }
    }
    handleOpen(){
      this.connected = true; this.reconnectAttempts = 0; this.onOpen();
      if(this.componentID){ this.send({ type:'init', componentID: this.componentID }); }
    }
    handleMessage(event){
      try { const message = JSON.parse(event.data); this.onMessage(message); if(message.type==='update') this.handleUpdate(message); if(message.type==='redirect') window.location.href = message.url; } catch(e){ console.error('[LFWS] message error', e); }
    }
    handleClose(event){
      this.connected = false; this.onClose(event);
      if (this.reconnectAttempts < this.maxReconnectAttempts){
        this.reconnectAttempts++; const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
        setTimeout(()=>this.connect(), delay);
      }
    }
    handleError(error){ console.error('[LFWS] error', error); this.onError(error); }
    send(message){ if(this.connected && this.ws){ this.ws.send(JSON.stringify(message)); } }
    sendAction(componentID, action, data={}){ this.send({ type:'action', componentID, action, data }); }
    setupFormHandling(){
      this.rootEl.addEventListener('submit', (e)=>{
        const form = e.target.closest('form'); if(!form) return;
        const componentID = form.getAttribute(dataFluxComponentID) || form.dataset.fluxComponentId || this.componentID;
        const action = form.getAttribute(dataFluxAction) || form.dataset.fluxAction || 'submit';
        if(this.connected && componentID){ e.preventDefault(); const fd = new FormData(form); const data = {}; for(const [k,v] of fd.entries()) data[k]=v; this.sendAction(componentID, action, data); }
      });
      this.rootEl.addEventListener('click', (e)=>{
        const el = e.target.closest(actionSelectorWithFallback); if(!el || !this.connected) return;
        if(el.tagName === 'FORM') return;
        const componentID = el.getAttribute(dataFluxComponentID) || el.dataset.fluxComponentId || this.componentID; const action = el.getAttribute(dataFluxAction) || el.dataset.fluxAction; if(!componentID||!action) return; e.preventDefault();
        const data = {}; for(const [key, value] of Object.entries(el.dataset)){ if(key.startsWith('fluxData')){ const k = key.replace(/^fluxData([A-Z])/, (_, p1) => p1.toLowerCase()); data[k]=value; } }
        this.sendAction(componentID, action, data);
      });
    }
    handleUpdate(message){
      const element = document.querySelector(componentIdSelector(message.componentID));
      if(element && message.data && message.data.html){
        element.outerHTML = message.data.html;
        if(this.connected){
          const refreshed = document.querySelector(componentIdSelector(message.componentID));
          if(refreshed){
            const status = refreshed.querySelector('.status'); if(status) status.textContent = 'Connected';
            this.rootEl = refreshed; this.setupFormHandling();
          }
        }
      }
    }
    close(){ if(this.ws){ this.ws.close(); } }
  }

  function autoInit(){
    const wsElements = document.querySelectorAll(wsSelector);
    wsElements.forEach(el => {
      const urlAttr = el.getAttribute(dataFluxWSURL) || el.dataset.fluxWsUrl;
      const url = urlAttr || (()=>{
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const cfg = window.liveflux || {};
        const wsPath = cfg.wsEndpoint || cfg.endpoint || '/liveflux';
        return `${protocol}//${window.location.host}${wsPath.startsWith('/') ? wsPath : ('/' + wsPath)}`;
      })();
      const componentID = el.getAttribute(dataFluxComponentID) || el.dataset.fluxComponentId || null;
      const client = new LiveFluxWS(url, {
        componentID,
        rootEl: el,
        onOpen: ()=>{ try{ el.dispatchEvent(new Event('flux-ws-open')); }catch(_){} const s = el.querySelector('.status'); if(s) s.textContent = 'Connected'; },
        onClose: ()=>{ try{ el.dispatchEvent(new Event('flux-ws-close')); }catch(_){} const s = el.querySelector('.status'); if(s) s.textContent = 'Disconnected'; },
        onError: (error)=>{ try{ const ev = new Event('flux-ws-error'); ev.error = error; el.dispatchEvent(ev); }catch(_){} const s = el.querySelector('.status'); if(s) s.textContent = 'Error'; },
        onMessage: (message)=>{ try{ const ev = new Event('flux-ws-message'); ev.data = message; el.dispatchEvent(ev); }catch(_){} }
      });
      try { el._lfws = client; } catch(_){}
    });
  }

  // Expose
  window.liveflux.LiveFluxWS = LiveFluxWS;
  try { window.LiveFluxWS = LiveFluxWS; } catch(_){}

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', autoInit); else autoInit();
})();
