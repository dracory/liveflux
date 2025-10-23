(function(){
  if(window.__lwInitDone) return; window.__lwInitDone = true;

  /**
   * Bootstraps the Liveflux client:
   * - Attaches global click/submit listeners
   * - Mounts placeholders on initial load
   * - Initializes $wire for components (after mounting)
   * Runs after DOMContentLoaded if needed.
   * @returns {void}
   */
  function init(){
    document.addEventListener('click', window.__lw.handleActionClick);
    document.addEventListener('submit', window.__lw.handleFormSubmit);
    
    // Mount placeholders first, then initialize $wire
    window.__lw.mountPlaceholders();
    
    // Initialize $wire for any components that are already on the page
    if(window.__lw.initWire) {
      // Wait a tick to ensure mount operations complete
      setTimeout(function(){
        window.__lw.initWire();
      }, 0);
    }
    
    // Dispatch livewire:init event for compatibility
    document.dispatchEvent(new CustomEvent('livewire:init'));
  }

  if(document.readyState === 'loading'){
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
