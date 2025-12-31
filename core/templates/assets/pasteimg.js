(function(){
    const uploadUrl = '/images/upload/image';
    function uuidv4(){
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c){
            const r = Math.random()*16|0, v = c=='x'?r:(r&0x3|0x8);
            return v.toString(16);
        });
    }
    function insertAtCaret(el, text){
        const start = el.selectionStart;
        const end = el.selectionEnd;
        el.setRangeText(text, start, end, 'end');
        return start;
    }
    function autoSize(el){
        el.style.height = 'auto';
        el.style.height = el.scrollHeight + 'px';
    }
    function handlePaste(e){
        const items = e.clipboardData && e.clipboardData.items;
        if(!items) return;
        for(let i=0;i<items.length;i++){
            const item = items[i];
            if(item.kind === 'file' && item.type.startsWith('image/')){
                e.preventDefault();
                const file = item.getAsFile();
                const id = uuidv4();
                const placeholder = '[img uploading:'+id+']';
                const pos = insertAtCaret(e.target, placeholder);
                autoSize(e.target);
                const fd = new FormData();
                fd.append('image', file);
                fd.append('id', id);
                const xhr = new XMLHttpRequest();
                xhr.open('POST', uploadUrl);
                let last = 0;
                xhr.upload.addEventListener('progress', ev => {
                    if(ev.lengthComputable){
                        const pct = Math.floor((ev.loaded/ev.total)*100);
                        if(pct - last >= 10){
                            last = pct - pct%10;
                            console.log('upload '+last+'%');
                        }
                    }
                });
                xhr.onload = function(){
                    if(xhr.status >= 200 && xhr.status < 300){
                        const ref = xhr.responseText;
                        const finalText = '[img '+ref+']';
                        const v = e.target.value;
                        e.target.value = v.substring(0,pos) + v.substring(pos).replace(placeholder, finalText);
                        e.target.setSelectionRange(pos+finalText.length, pos+finalText.length);
                        autoSize(e.target);
                    } else {
                        e.target.value = e.target.value.replace(placeholder, '');
                        autoSize(e.target);
                    }
                };
                xhr.onerror = function(){
                    e.target.value = e.target.value.replace(placeholder, '');
                    autoSize(e.target);
                };
                xhr.send(fd);
            }
        }
    }
    window.addEventListener('load', function(){
        document.querySelectorAll('textarea').forEach(function(t){
            t.addEventListener('paste', handlePaste);
            autoSize(t);
            t.addEventListener('input', function(){
                autoSize(this);
            });
        });
    });
})();
