(function(){
    const proto = location.protocol === 'https:' ? 'wss://' : 'ws://';
    const url = proto + location.host + '/ws/notifications';
    const seen = new Set();
    let conn;

    let retryDelay = 1000;

    function connect(){
        conn = new WebSocket(url);
        conn.onopen = () => {
            retryDelay = 1000;
        };
        conn.onmessage = evt => {
            try{
                const msg = JSON.parse(evt.data);
                if(!msg || !msg.Data || !msg.Data.notification){
                    return;
                }
                const n = msg.Data.notification;
                if(seen.has(n.id)) return;
                seen.add(n.id);
                updateCount(1);
                addNotification(n);
            }catch(e){
                console.log('ws message error', e);
            }
        };
        conn.onclose = () => {
             setTimeout(connect, retryDelay);
             retryDelay = Math.min(retryDelay * 2, 60000);
        };
    }

    function updateCount(delta){
        const link = document.getElementById('notif-index');
        if(!link) return;
        let m = link.textContent.match(/\((\d+)\)/);
        let count = m ? parseInt(m[1],10) : 0;
        count += delta;
        if(m){
            link.textContent = link.textContent.replace(/\(\d+\)/, '('+count+')');
        } else {
            link.textContent += ' ('+count+')';
        }
    }

    function addNotification(n){
        const list = document.getElementById('notifications-list');
        if(!list) return;
        if(document.getElementById('notif-'+n.id)) return;
        const empty = document.getElementById('notifications-empty');
        if(empty) empty.remove();
        const div = document.createElement('div');
        div.className = 'notification';
        div.id = 'notif-'+n.id;
        let html = '';
        if(n.link){
            html += '<a href="'+n.link+'">'+n.message+'</a>';
        } else {
            html += n.message;
        }
        html += ' <form method="post" action="/usr/notifications/dismiss" class="inline-form">';
        html += '<input type="hidden" name="id" value="'+n.id+'">';
        html += '<input type="submit" name="task" value="Dismiss">';
        html += '</form>';
        div.innerHTML = html;
        list.prepend(div);
    }

    window.addEventListener('load', function(){
        document.querySelectorAll('[data-notification-id]').forEach(el => seen.add(parseInt(el.dataset.notificationId,10)));
        connect();
    });
})();
