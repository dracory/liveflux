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
    // Inject constants when building the js from Go code
    // so that they are consistent with the Go code
    // !!! Do not modify the following lines !!!
    const dataFluxAction = "{{ DataFluxAction }}";
	const dataFluxDispatchTo = "{{ DataFluxDispatchTo }}";
	const dataFluxComponent = "{{ DataFluxComponent }}";
	const dataFluxComponentID = "{{ DataFluxComponentID }}";
	const dataFluxID = "{{ DataFluxID }}";
	const dataFluxMount = "{{ DataFluxMount }}";
	const dataFluxParam = "{{ DataFluxParam }}";
	const dataFluxRoot = "{{ DataFluxRoot }}";
	const dataFluxSubmit = "{{ DataFluxSubmit }}";
	const dataFluxWS = "{{ DataFluxWS }}";
	const dataFluxWSURL = "{{ DataFluxWSURL }}";
	const endpoint = "{{ Endpoint }}";

    function livefluxNamespaceCreate(){
        // Ensure liveflux namespace exists before other modules execute
        if(!window.liveflux){
            window.liveflux = {
                __bootstrapInitDone: false,
                endpoint,
                dataFluxAction,
                dataFluxDispatchTo,
                dataFluxComponent,
                dataFluxComponentID,
                dataFluxID,
                dataFluxMount,
                dataFluxParam,
                dataFluxRoot,
                dataFluxSubmit,
                dataFluxWS,
                dataFluxWSURL,
            };
            console.log('[Liveflux] Namespace created');
        }
    }
    
    livefluxNamespaceCreate();
})();
