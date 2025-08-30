(function(){
  if(window.__lwInitDone) return; window.__lwInitDone = true;

  /**
   * Bootstraps the Livewire client:
   * - Attaches global click/submit listeners
   * - Mounts placeholders on initial load
   * Runs after DOMContentLoaded if needed.
   * @returns {void}
   */
  function init(){
    document.addEventListener('click', window.__lw.handleActionClick);
    document.addEventListener('submit', window.__lw.handleFormSubmit);
    window.__lw.mountPlaceholders();
  }

  if(document.readyState === 'loading'){
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
