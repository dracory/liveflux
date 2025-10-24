// // DEPRECATED: Use liveflux_websocket.js. This file remains for backward compatibility and will be removed.
// // LiveFlux WebSocket Client
// class LiveFluxWS {
//   constructor(url, options = {}) {
//     try { console.log('[LFWS] init with url:', url, 'options:', options); } catch (_) {}
//     this.url = url;
//     this.ws = null;
//     this.connected = false;
//     this.reconnectAttempts = 0;
//     this.maxReconnectAttempts = options.maxReconnectAttempts || 5;
//     this.reconnectDelay = options.reconnectDelay || 1000;
//     this.componentID = options.componentID || null;
//     this.rootEl = options.rootEl || document; // limit event handling scope
    
//     // Event handlers
//     this.onOpen = options.onOpen || (() => {});
//     this.onMessage = options.onMessage || (() => {});
//     this.onClose = options.onClose || (() => {});
//     this.onError = options.onError || (() => {});
    
//     // Connect
//     this.connect();
    
//     // Set up form handling
//     this.setupFormHandling();
//   }
  
//   connect() {
//     try {
//       try { console.log('[LFWS] connecting to', this.url); } catch (_) {}
//       this.ws = new WebSocket(this.url);
//       this.ws.onopen = this.handleOpen.bind(this);
//       this.ws.onmessage = this.handleMessage.bind(this);
//       this.ws.onclose = this.handleClose.bind(this);
//       this.ws.onerror = this.handleError.bind(this);
//     } catch (e) {
//       console.error('WebSocket connection error:', e);
//       this.handleError(e);
//     }
//   }
  
//   handleOpen() {
//     this.connected = true;
//     this.reconnectAttempts = 0;
//     try { console.log('[LFWS] open', this.url); } catch (_) {}
    
//     // Send initial message with component ID if available
//     if (this.componentID) {
//       const initMsg = {
//         type: 'init',
//         componentID: this.componentID
//       };
//       try { console.log('[LFWS] send init', initMsg); } catch (_) {}
//       this.send(initMsg);
//     }
    
//     this.onOpen();
//   }
  
//   handleMessage(event) {
//     try {
//       const message = JSON.parse(event.data);
//       try { console.log('[LFWS] message', message); } catch (_) {}
//       this.onMessage(message);
      
//       // Handle special message types
//       if (message.type === 'update') {
//         this.handleUpdate(message);
//       } else if (message.type === 'redirect') {
//         window.location.href = message.url;
//       }
//     } catch (e) {
//       console.error('Error processing message:', e);
//     }
//   }
  
//   handleClose(event) {
//     this.connected = false;
//     try { console.warn('[LFWS] close', { code: event && event.code, reason: event && event.reason }); } catch (_) {}
//     this.onClose(event);
    
//     // Attempt to reconnect
//     if (this.reconnectAttempts < this.maxReconnectAttempts) {
//       this.reconnectAttempts++;
//       const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      
//       try { console.log(`[LFWS] reconnect in ${delay}ms... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`); } catch (_) {}
//       setTimeout(() => this.connect(), delay);
//     } else {
//       console.error('[LFWS] Max reconnection attempts reached');
//     }
//   }
  
//   handleError(error) {
//     console.error('[LFWS] error:', error);
//     this.onError(error);
//   }
  
//   send(message) {
//     if (this.connected && this.ws) {
//       try { console.debug('[LFWS] send', message); } catch (_) {}
//       this.ws.send(JSON.stringify(message));
//     } else {
//       console.warn('[LFWS] send aborted: WebSocket not connected');
//     }
//   }
  
//   sendAction(componentID, action, data = {}) {
//     const msg = {
//       type: 'action',
//       componentID,
//       action,
//       data
//     };
//     try { console.debug('[LFWS] sendAction', { componentID, action, data }); } catch (_) {}
//     this.send(msg);
//   }
  
//   setupFormHandling() {
//     // Handle form submissions (scoped)
//     this.rootEl.addEventListener('submit', (e) => {
//       const form = e.target.closest('form');
//       if (!form) return;
      
//       const componentID = form.dataset.fluxComponentId || this.componentID;
//       const action = form.dataset.fluxAction || 'submit';
      
//       if (this.connected && componentID) {
//         e.preventDefault();
        
//         // Collect form data
//         const formData = new FormData(form);
//         const data = {};
//         for (let [key, value] of formData.entries()) {
//           data[key] = value;
//         }
        
//         // Send via WebSocket
//         this.sendAction(componentID, action, data);
//       }
//       // If not connected, let the form submit normally
//     });
    
//     // Handle click events on elements with data-flux-action (scoped)
//     this.rootEl.addEventListener('click', (e) => {
//       const closestActionEl = e.target.closest('[data-flux-action]');
//       if (!closestActionEl || !this.connected) return;

//       // If the closest action element is a FORM, ignore click here and let the submit handler handle it.
//       if (closestActionEl.tagName === 'FORM') {
//         return;
//       }

//       const componentID = closestActionEl.dataset.fluxComponentId || this.componentID;
//       const action = closestActionEl.dataset.fluxAction;
//       if (!componentID || !action) return;

//       e.preventDefault();

//       // Collect any data-* from the action element
//       const data = {};
//       for (const [key, value] of Object.entries(closestActionEl.dataset)) {
//         if (key.startsWith('fluxData')) {
//           const dataKey = key.replace(/^fluxData([A-Z])/, (_, p1) => p1.toLowerCase());
//           data[dataKey] = value;
//         }
//       }

//       this.sendAction(componentID, action, data);
//     });
//   }
  
//   handleUpdate(message) {
//     // Find the component element by ID and update its content
//     const element = document.querySelector(`[data-flux-component-id="${message.componentID}"]`);
//     if (element && message.data && message.data.html) {
//       try { console.debug('[LFWS] updating HTML for', message.componentID); } catch (_) {}
//       element.outerHTML = message.data.html;
//       // After replacement, update status text if still connected
//       if (this.connected) {
//         const refreshed = document.querySelector(`[data-flux-component-id="${message.componentID}"]`);
//         if (refreshed) {
//           const status = refreshed.querySelector('.status');
//           if (status) status.textContent = 'Connected';
//           // Rebind event handlers to the new root so subsequent clicks work
//           this.rootEl = refreshed;
//           this.setupFormHandling();
//         }
//       }
//     }
//   }
  
//   close() {
//     if (this.ws) {
//       this.ws.close();
//     }
//   }
// }

// // Expose constructor for debugging
// try { window.LiveFluxWS = LiveFluxWS; } catch (_) {}

// function lfwsAutoInit() {
//   const wsElements = document.querySelectorAll('[data-flux-ws]');
//   try { console.log('[LFWS] auto-init elements found:', wsElements.length); } catch (_) {}
//   wsElements.forEach(el => {
//     const url = el.dataset.fluxWsUrl || (() => {
//       const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
//       const g = window; const cfg = (g && g.__lw) ? g.__lw : {};
//       const wsPath = cfg.wsEndpoint || cfg.endpoint || '/liveflux';
//       return `${protocol}//${window.location.host}${wsPath.startsWith('/') ? wsPath : ('/' + wsPath)}`;
//     })();
//     const componentID = el.dataset.fluxComponentId || null;
//     try { console.log('[LFWS] creating client', { url, componentID }); } catch (_) {}

//     const client = new LiveFluxWS(url, {
//       componentID,
//       rootEl: el,
//       onOpen: () => {
//         try { console.log('[LFWS] onOpen dispatch'); } catch (_) {}
//         const evt = new Event('flux-ws-open');
//         el.dispatchEvent(evt);
//         // also set status text directly for visibility
//         const status = el.querySelector('.status');
//         if (status) status.textContent = 'Connected';
//       },
//       onClose: () => {
//         try { console.log('[LFWS] onClose dispatch'); } catch (_) {}
//         el.dispatchEvent(new Event('flux-ws-close'));
//         const status = el.querySelector('.status');
//         if (status) status.textContent = 'Disconnected';
//       },
//       onError: (error) => {
//         try { console.error('[LFWS] onError dispatch', error); } catch (_) {}
//         const event = new Event('flux-ws-error');
//         event.error = error;
//         el.dispatchEvent(event);
//         const status = el.querySelector('.status');
//         if (status) status.textContent = 'Error';
//       },
//       onMessage: (message) => {
//         try { console.log('[LFWS] onMessage dispatch', message); } catch (_) {}
//         const event = new Event('flux-ws-message');
//         event.data = message;
//         el.dispatchEvent(event);
//       }
//     });

//     // attach for debugging
//     try { el._lfws = client; } catch(_) {}
//   });
// }

// if (document.readyState === 'loading') {
//   document.addEventListener('DOMContentLoaded', lfwsAutoInit);
// } else {
//   // DOM is already ready; init immediately
//   lfwsAutoInit();
// }
