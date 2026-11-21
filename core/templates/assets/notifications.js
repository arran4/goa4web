(function(root){
    // Compatibility shim: delegates legacy /websocket/notifications.js consumers to the
    // server-rendered HTML fragment swapper over /ws/notifications without client-side HTML assembly.
    if(root && root.Goa4WebLiveEvents && typeof root.Goa4WebLiveEvents.initLiveEvents === 'function'){
        root.Goa4WebLiveEvents.initLiveEvents('/ws/notifications');
        return;
    }

    const proto = root.location && root.location.protocol === 'https:' ? 'wss://' : 'ws://';
    const url = root.location ? proto + root.location.host + '/ws/notifications' : '/ws/notifications';
    let retryDelay = 1000;
    const currentRefresh = new WeakMap();

    async function refreshFragments(){
        const nodes = root.document.querySelectorAll('[data-live-event][data-live-url]');
        for(const el of nodes){
            const fragURL = el.getAttribute('data-live-url');
            if(typeof fragURL !== 'string' || fragURL.length < 2 || fragURL[0] !== '/' || fragURL[1] === '/') continue;
            const request = {};
            currentRefresh.set(el, request);
            const resp = await root.fetch(fragURL, {
                method: 'GET',
                credentials: 'same-origin',
                headers: { 'X-Live-Fragment': '1' }
            });
            if(!resp || !resp.ok || resp.redirected) continue;
            const html = await resp.text();
            if(currentRefresh.get(el) !== request) continue;
            const mode = (el.getAttribute('data-live-swap') || 'innerHTML').trim();
            if(mode === 'outerHTML' || mode === 'replace'){
                el.outerHTML = html;
            } else {
                el.innerHTML = html;
            }
        }
    }

    function connect(){
        if(typeof root.WebSocket !== 'function') return;
        const conn = new root.WebSocket(url);
        conn.onopen = () => {
            retryDelay = 1000;
            refreshFragments().catch(() => {});
        };
        conn.onmessage = () => {
            refreshFragments().catch(() => {});
        };
        conn.onclose = () => {
            root.setTimeout(connect, retryDelay);
            retryDelay = Math.min(retryDelay * 2, 60000);
        };
    }

    if(root.addEventListener){
        root.addEventListener('load', connect);
    }
})(typeof window !== 'undefined' ? window : globalThis);
