/**
 * Public API for Liveflux
 * Exposes the main methods for external scripts and inline usage.
 */
(function(){
  const g = window;

  function findComponent(alias, id){
    return document.querySelector('[data-flux-root][data-flux-component="' + alias + '"][data-flux-component-id="' + id + '"]');
  }
  
  // Expose Liveflux global object for external scripts
  if(!window.Liveflux){
    window.Liveflux = {
      dispatch: g.__lw.dispatch,
      dispatchToAliasAndId: g.__lw.dispatchToAliasAndId,
      findComponent: findComponent,
      on: g.__lw.on,
    };
  }
  
  // Expose liveflux namespace (lowercase) for convenience
  if(!window.liveflux){
    window.liveflux = window.Liveflux;
  }
})();
