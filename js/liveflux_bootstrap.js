(function(){
  if(window.__livefluxInitDone) return;

  if(!window.liveflux){
    console.log('[Liveflux Bootstrap] liveflux namespace not found');
    return;
  }

  window.__livefluxInitDone = true;

  function init(){
    document.addEventListener('click', window.liveflux.handleActionClick);
    document.addEventListener('submit', window.liveflux.handleFormSubmit);

    window.liveflux.mountPlaceholders();

    if(window.liveflux.initWire){
      setTimeout(function(){ window.liveflux.initWire(); }, 0);
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
  window.liveflux.bootstrapInit = init;
})();
