(function(){
  if(!window.liveflux){
    console.log('[Liveflux Bootstrap] liveflux namespace not found');
    return;
  }

  const liveflux = window.liveflux;

  if(liveflux.__bootstrapInitDone) return;
  liveflux.__bootstrapInitDone = true;

  function init(){
    document.addEventListener('click', liveflux.handleActionClick);
    document.addEventListener('submit', liveflux.handleFormSubmit);

    liveflux.mountPlaceholders();

    if(liveflux.initWire){
      setTimeout(function(){ liveflux.initWire(); }, 0);
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
  liveflux.bootstrapInit = init;
})();
