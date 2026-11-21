(function(root){
    function buildWebSocketURL(endpoint){
        const proto = root.location.protocol === 'https:' ? 'wss://' : 'ws://';
        return proto + root.location.host + endpoint;
    }

    function isSafeFragmentURL(url){
        return typeof url === 'string' && url.length > 1 && url[0] === '/' && url[1] !== '/';
    }

    function parseEventKind(raw){
        const msg = typeof raw === 'string' ? JSON.parse(raw) : raw;
        if(!msg || typeof msg !== 'object') return null;
        if(typeof msg.kind === 'string' && msg.kind){
            if(msg.kind === 'task' && msg.data && msg.data.notification){
                return 'notification';
            }
            return msg.kind;
        }
        const data = msg.data || msg.Data;
        if(data && data.notification){
            return 'notification';
        }
        return null;
    }

    function swapFragment(el, html, swapMode, doc){
        const mode = (swapMode || el.getAttribute('data-live-swap') || 'innerHTML').trim();
        if(mode === 'outerHTML' || mode === 'replace'){
            const ownerDoc = el.ownerDocument || doc || root.document;
            const tpl = ownerDoc && typeof ownerDoc.createElement === 'function' ? ownerDoc.createElement('template') : null;
            if(tpl){
                tpl.innerHTML = html.trim();
                if(tpl.content && tpl.content.firstElementChild && typeof el.replaceWith === 'function'){
                    el.replaceWith(tpl.content.firstElementChild);
                    return;
                }
            }
            el.outerHTML = html;
            return;
        }
        if(mode === 'prepend'){
            el.insertAdjacentHTML('afterbegin', html);
            return;
        }
        if(mode === 'append'){
            el.insertAdjacentHTML('beforeend', html);
            return;
        }
        el.innerHTML = html;
    }

    async function refreshTarget(el, fetchImpl, doc){
        const url = el.getAttribute('data-live-url');
        if(!isSafeFragmentURL(url)){
            return false;
        }
        const doFetch = fetchImpl || root.fetch;
        if(typeof doFetch !== 'function'){
            return false;
        }
        const resp = await doFetch(url, {
            method: 'GET',
            credentials: 'same-origin',
            headers: { 'X-Live-Fragment': '1' }
        });
        if(!resp || !resp.ok || resp.redirected){
            return false;
        }
        const ct = resp.headers && typeof resp.headers.get === 'function' ? (resp.headers.get('content-type') || '') : 'text/html';
        if(ct && !ct.toLowerCase().includes('text/html')){
            return false;
        }
        const html = await resp.text();
        swapFragment(el, html, el.getAttribute('data-live-swap'), doc);
        return true;
    }

    function matchingTargets(doc, kind){
        const nodes = doc.querySelectorAll('[data-live-event][data-live-url]');
        const matches = [];
        nodes.forEach(el => {
            const events = (el.getAttribute('data-live-event') || '').trim().split(/\s+/);
            if(events.includes(kind) || events.includes('*')){
                matches.push(el);
            }
        });
        return matches;
    }

    async function handleMessage(raw, doc, fetchImpl){
        const kind = parseEventKind(raw);
        if(!kind) return 0;
        const targetDoc = doc || root.document;
        const targets = matchingTargets(targetDoc, kind);
        let updated = 0;
        for(const el of targets){
            if(await refreshTarget(el, fetchImpl, targetDoc)){
                updated++;
            }
        }
        return updated;
    }

    async function refreshAllTargets(doc, fetchImpl){
        const targetDoc = doc || root.document;
        const nodes = targetDoc.querySelectorAll('[data-live-event][data-live-url]');
        let updated = 0;
        for(const el of nodes){
            if(await refreshTarget(el, fetchImpl, targetDoc)){
                updated++;
            }
        }
        return updated;
    }

    function initLiveEvents(endpoint){
        let conn;
        let retryDelay = 1000;

        function connect(){
            if(typeof root.WebSocket !== 'function') return;
            conn = new root.WebSocket(buildWebSocketURL(endpoint));
            conn.onopen = () => {
                retryDelay = 1000;
                refreshAllTargets(root.document).catch(e => console.log('ws reconcile error', e));
            };
            conn.onmessage = evt => {
                handleMessage(evt.data, root.document).catch(e => console.log('ws message error', e));
            };
            conn.onclose = () => {
                root.setTimeout(connect, retryDelay);
                retryDelay = Math.min(retryDelay * 2, 60000);
            };
        }

        root.addEventListener('load', function(){
            connect();
            const btnAll = root.document.getElementById('btn-select-all');
            if(btnAll) btnAll.addEventListener('click', () => root.document.querySelectorAll('#notifications-list input[type="checkbox"]').forEach(el => el.checked = true));
            const btnInvert = root.document.getElementById('btn-invert-selection');
            if(btnInvert) btnInvert.addEventListener('click', () => root.document.querySelectorAll('#notifications-list input[type="checkbox"]').forEach(el => el.checked = !el.checked));
        });
    }

    const api = {
        isSafeFragmentURL,
        parseEventKind,
        swapFragment,
        refreshTarget,
        matchingTargets,
        handleMessage,
        refreshAllTargets,
        initLiveEvents
    };

    if(typeof module !== 'undefined' && module.exports){
        module.exports = api;
    }
    if(root){
        root.Goa4WebLiveEvents = api;
        if(root.document && root.location && !root.__GOA4WEB_DISABLE_AUTO_LIVE_EVENTS__){
            initLiveEvents('/ws/events');
        }
    }
})(typeof window !== 'undefined' ? window : globalThis);
