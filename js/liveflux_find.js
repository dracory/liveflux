/**
 * Liveflux Find
 * Exposes the main methods for finding components.
 * 
 * Notes:
 * - This file is loaded in the browser and is not used in the Go code.
 * - It is used to add find methods to the liveflux namespace.
 * - It depends on liveflux_namespace_create.js.
 * 
 * Dev notes:
 * - The functions are sorted alphabetically.
 */
(function(){
  // Check if liveflux namespace exists
  if(!window.liveflux){
    console.log('[Liveflux Find] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;
  const dataFluxComponentKind = liveflux.dataFluxComponentKind || 'data-flux-component-kind';
  const dataFluxComponentID = liveflux.dataFluxComponentID || 'data-flux-component-id';

  /**
   * Finds a component by kind and ID.
   * @param {string} componentKind - Kind of the target component.
   * @param {string} componentId - ID of the target component instance.
   * @returns {HTMLElement|null} - The component element if found, otherwise null.
   */
  function findComponent(componentKind, componentId){
    const selector = liveflux.getComponentRootSelector ? liveflux.getComponentRootSelector() : `[${dataFluxComponentKind}][${dataFluxComponentID}]`;
    const elements = document.querySelectorAll(selector);
    return Array.from(elements).find(function(el){
      return el.getAttribute(dataFluxComponentKind) === componentKind && el.getAttribute(dataFluxComponentID) === componentId;
    }) || null;
  }
  
  // Add functions to liveflux namespace
  liveflux.findComponent = findComponent;
})();
