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
        // Ensure liveflux namespace exists before other modules execute
        if(!window.liveflux){
            window.liveflux = {};
            console.log('[Liveflux] Namespace created');
        }
    }
    
    livefluxNamespaceCreate();
})();
