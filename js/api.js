/**
 * Public API for Liveflux
 * Exposes the main methods for external scripts and inline usage.
 */
(function(){
  const g = window;
  
  // Expose Liveflux global object for external scripts
  if(!window.Liveflux){
    window.Liveflux = {
      on: g.__lw.on,
      dispatch: g.__lw.dispatch,
      dispatchToAliasAndId: g.__lw.dispatchToAliasAndId
    };
  }
  
  // Expose liveflux namespace (lowercase) for convenience
  if(!window.liveflux){
    window.liveflux = window.Liveflux;
  }
})();
