// (function(){
//   /**
//    * DEPRECATED: Use liveflux_wire.js. This file remains for backward compatibility and will be removed.
//    */
//   const g = window; g.__lw = g.__lw || {};

//   /**
//    * Creates a $wire proxy object for a specific component.
//    * This provides Livewire-like syntax for component scripts.
//    * @param {string} componentId - The ID of the component.
//    * @param {string} componentAlias - The alias of the component.
//    * @returns {Object} - The $wire proxy object.
//    */
//   g.__lw.createWire = function(componentId, componentAlias, rootEl){
//     console.log('[Liveflux Wire] Creating $wire for component:', componentId, componentAlias);
    
//     return {
//       /**
//        * Registers an event listener for this component.
//        * @param {string} eventName - The name of the event to listen for.
//        * @param {Function} callback - The callback function to execute.
//        * @returns {Function} - A cleanup function to remove the listener.
//        */
//       on: function(eventName, callback){
//         console.log('[Liveflux Wire] $wire.on called for', eventName, 'on component', componentId);
//         return g.__lw.onComponent(componentId, eventName, callback);
//       },

//       /**
//        * Dispatches an event from this component.
//        * @param {string} eventName - The name of the event to dispatch.
//        * @param {Object} data - Optional data to pass with the event.
//        * @returns {void}
//        */
//       dispatch: function(eventName, data){
//         g.__lw.dispatch(eventName, data);
//       },

//       /**
//        * Dispatches an event only to this component (self-only).
//        * @param {string} eventName - The name of the event to dispatch.
//        * @param {Object} data - Optional data to pass with the event.
//        * @returns {void}
//        */
//       dispatchSelf: function(eventName, data){
//         const eventData = data || {};
//         eventData.__self = true;
//         g.__lw.dispatch(eventName, eventData);
//       },

//       /**
//        * Dispatches an event to a specific component type.
//        * @param {string} targetAlias - The alias of the target component.
//        * @param {string} eventName - The name of the event to dispatch.
//        * @param {Object} data - Optional data to pass with the event.
//        * @returns {void}
//        */
//       dispatchTo: function(targetAlias, eventName, data){
//         const eventData = data || {};
//         eventData.__target = targetAlias;
//         g.__lw.dispatch(eventName, eventData);
//       },

//       /**
//        * Calls a server action for this component and updates its DOM.
//        * @param {string} action - The action name to call.
//        * @param {Object} data - Additional form data to send.
//        * @returns {Promise<any>} Resolves with the result from g.__lw.post
//        */
//       call: function(action, data){
//         action = action || 'submit';
//         const params = Object.assign({}, data || {}, {
//           liveflux_component_type: componentAlias,
//           liveflux_component_id: componentId,
//           liveflux_action: action
//         });

//         console.log('[Liveflux Wire] $wire.call', action, params);
//         return g.__lw.post(params).then(function(result){
//           const html = result.html || result;
//           const tmp = document.createElement('div');
//           tmp.innerHTML = html;
//           const newNode = tmp.firstElementChild;
//           if(newNode && rootEl){
//             rootEl.replaceWith(newNode);
//             rootEl = newNode;
//             g.__lw.executeScripts(newNode);
//             if(g.__lw.initWire) g.__lw.initWire();
//           }
//           return result;
//         });
//       },

//       // Component metadata
//       id: componentId,
//       alias: componentAlias
//     };
//   };

//   /**
//    * Initializes $wire for all component roots on the page.
//    * This is called automatically by the bootstrap script.
//    * @returns {void}
//    */
//   g.__lw.initWire = function(){
//     console.log('[Liveflux Wire] Initializing $wire for all components');
//     const roots = document.querySelectorAll('[data-flux-root], [flux-root]');
//     console.log('[Liveflux Wire] Found', roots.length, 'component roots');
    
//     roots.forEach(function(root){
//       const comp = root.querySelector('input[name="liveflux_component_type"]');
//       const id = root.querySelector('input[name="liveflux_component_id"]');
//       if(!comp || !id) {
//         console.warn('[Liveflux Wire] Component root missing type or id inputs:', root);
//         return;
//       }

//       const componentId = id.value;
//       const componentAlias = comp.value;

//       // Attach $wire to the root element for script access
//       root.$wire = g.__lw.createWire(componentId, componentAlias, root);
//       console.log('[Liveflux Wire] Attached $wire to component:', componentAlias, componentId);
//     });
//   };
// })();
