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

  function queryComponentRoots(){
    const selector = liveflux.getComponentRootSelector ? liveflux.getComponentRootSelector() : `[${dataFluxComponentKind}][${dataFluxComponentID}]`;
    return Array.from(document.querySelectorAll(selector));
  }

  /**
   * Returns a list of all component root elements. Optionally filter them.
   * @param {Function} [predicate] - Optional filter function receiving each element.
   * @returns {HTMLElement[]} - Array of component root elements.
   */
  function findAllComponents(predicate){
    const elements = queryComponentRoots();
    if(typeof predicate === 'function'){
      return elements.filter(predicate);
    }
    return elements;
  }

  /**
   * Finds a component by kind and ID.
   * @param {string} componentKind - Kind of the target component.
   * @param {string} componentId - ID of the target component instance.
   * @returns {HTMLElement|null} - The component element if found, otherwise null.
   */
  function findComponent(componentKind, componentId){
    if(!componentKind || !componentId){
      return null;
    }
    return findAllComponents(function(el){
      return el.getAttribute(dataFluxComponentKind) === componentKind && el.getAttribute(dataFluxComponentID) === componentId;
    })[0] || null;
  }
  
  // Add functions to liveflux namespace
  liveflux.findAllComponents = findAllComponents;
  liveflux.findComponent = findComponent;
})();
