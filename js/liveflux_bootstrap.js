(function(){
  if(window.__lwInitDone) return;
  window.__lwInitDone = true;

  const g = window; g.liveflux = g.liveflux || {}; g.__lw = g.__lw || {};

  function init(){
    document.addEventListener('click', g.liveflux.handleActionClick);
    document.addEventListener('submit', g.liveflux.handleFormSubmit);

    g.liveflux.mountPlaceholders();

    if(g.liveflux.initWire){
      setTimeout(function(){ g.liveflux.initWire(); }, 0);
    }

    // Compatibility event
    try { document.dispatchEvent(new CustomEvent('livewire:init')); } catch(_) {}
  }

  if(document.readyState === 'loading'){
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  // Expose for manual re-init if needed
  g.liveflux.bootstrapInit = init;
})();
