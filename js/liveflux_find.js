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

  /**
   * Finds a component by alias and ID.
   * @param {string} componentAlias - Alias of the target component.
   * @param {string} componentId - ID of the target component instance.
   * @returns {HTMLElement|null} - The component element if found, otherwise null.
   */
  function findComponent(componentAlias, componentId){
    return document.querySelector('[data-flux-root][data-flux-component="' + componentAlias + '"][data-flux-component-id="' + componentId + '"]');
  }
  
  // Add functions to liveflux namespace
  window.liveflux.findComponent = findComponent;
})();
