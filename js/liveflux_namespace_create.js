/**
 * Creates the liveflux namespace.
 * 
 * Notes:
 * - This file is loaded in the browser and is not used in the Go code.
 * - It is loaded first.
 * 
 * @namespace liveflux
 */
(function(){
    function livefluxNamespaceCreate(){
        console.log('[Liveflux] Creating namespace...');
        
        // Expose liveflux namespace (lowercase) for convenience
        if(!window.liveflux){
            window.liveflux = {};
        }
    }
    
    livefluxNamespaceCreate();
})();
