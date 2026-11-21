(function(root){
    // A newer fragment request for the same target supersedes earlier requests.
    const currentRefresh = new WeakMap();
    // The gateway observes the same number of cards per connection.
    const maxLiveCards = 12;

    function buildWebSocketURL(endpoint){
        const proto = root.location.protocol === 'https:' ? 'wss://' : 'ws://';
        return proto + root.location.host + endpoint;
    }

    function isSafeFragmentURL(url){
        return typeof url === 'string' && url.length > 1 && url[0] === '/' && url[1] !== '/' && !/[\\\x00-\x20\x7f]/.test(url);
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
        const request = {};
        currentRefresh.set(el, request);
        const resp = await doFetch(url, {
            method: 'GET',
            credentials: 'same-origin',
            redirect: 'error',
            headers: { 'X-Live-Fragment': '1' }
        });
        if(!resp || !resp.ok || resp.redirected || resp.status === 204){
            return false;
        }
        const ct = resp.headers && typeof resp.headers.get === 'function' ? (resp.headers.get('content-type') || '') : 'text/html';
        if(ct && !ct.toLowerCase().includes('text/html')){
            return false;
        }
        const html = await resp.text();
        if(currentRefresh.get(el) !== request){
            return false;
        }
        swapFragment(el, html, el.getAttribute('data-live-swap'), doc);
        return true;
    }

    function liveTargets(doc){
        const targets = [];
        let cards = 0;
        for(const el of doc.querySelectorAll('[data-live-event][data-live-url]')){
            if(el.getAttribute('data-live-card') && ++cards > maxLiveCards) continue;
            targets.push(el);
        }
        return targets;
    }

    function matchingTargets(doc, kind){
        return liveTargets(doc).filter(el => {
            const events = (el.getAttribute('data-live-event') || '').trim().split(/\s+/);
            return events.includes(kind) || events.includes('*');
        });
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
        const nodes = liveTargets(targetDoc);
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
            const target = root.document.querySelector('[data-live-thread]');
            const thread = target && target.getAttribute('data-live-thread');
            const params = new URLSearchParams();
            if(/^[1-9][0-9]*$/.test(thread || '')){
                params.set('thread', thread);
                const keys = new Set();
                for(const card of liveTargets(root.document)){
                    const key = card.getAttribute('data-live-card');
                    if(/^[1-9][0-9]*-[a-f0-9]{64}$/.test(key || '')) keys.add(key);
                }
                for(const key of keys) params.append('card', key);
            }
            const scope = params.size ? '?' + params.toString() : '';
            conn = new root.WebSocket(buildWebSocketURL(endpoint + scope));
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
