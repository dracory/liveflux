(function(){
  const g = window; g.__lw = g.__lw || {};

  /**
   * Finds placeholders marked with data-lw-mount="1" and replaces them with
   * server-rendered component HTML, executing any scripts inside.
   * @returns {void}
   */
  g.__lw.mountPlaceholders = function(){
    document.querySelectorAll('[data-lw-mount="1"]').forEach((el)=>{
      const component = el.getAttribute('data-lw-component');
      if(!component) return;
      const params = g.__lw.readParams(el);
      params.component = component;
      g.__lw.post(params).then((html)=>{
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        const newNode = tmp.firstElementChild;
        if(newNode){ el.replaceWith(newNode); g.__lw.executeScripts(newNode); }
      }).catch((err)=>{ console.error(component+' mount', err); });
    });
  };
})();
