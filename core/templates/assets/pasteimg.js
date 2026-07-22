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
        const scrollableParents = [];
        let parent = el.parentNode;
        while (parent && parent instanceof HTMLElement && parent !== document.body && parent !== document.documentElement) {
            if (parent.scrollHeight > parent.clientHeight || parent.scrollWidth > parent.clientWidth) {
                scrollableParents.push({
                    el: parent,
                    top: parent.scrollTop,
                    left: parent.scrollLeft
                });
            }
            parent = parent.parentNode;
        }
        const scrollX = window.scrollX;
        const scrollY = window.scrollY;
        el.style.height = 'auto';
        el.style.height = el.scrollHeight + 'px';
        window.scrollTo(scrollX, scrollY);
        scrollableParents.forEach(p => {
            p.el.scrollTop = p.top;
            p.el.scrollLeft = p.left;
        });
    }
    function handleImagePaste(e, item) {
        e.preventDefault();
        const file = item.getAsFile();
        const id = uuidv4();
        const placeholder = '[img uploading:'+id+']';
        const pos = insertAtCaret(e.target, placeholder);
        autoSize(e.target);
        const fd = new FormData();
        fd.append('image', file);
        fd.append('id', id);
        fd.append('task', 'Upload image');
        const xhr = new XMLHttpRequest();
        xhr.open('POST', uploadUrl);
        const csrf = document.querySelector("input[name='csrf_token']");
        if (csrf) {
            xhr.setRequestHeader('X-CSRF-Token', csrf.value);
        }
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
            } else if (xhr.status === 403) {
                let reason = xhr.responseText;
                if (reason) {
                    reason = reason.replace(/<[^>]*>?/gm, '').trim();
                }
                if (!reason || reason === '') {
                    reason = 'Permission Denied';
                }
                console.error('Image upload forbidden:', reason);
                const failedText = '[img upload denied: ' + reason.substring(0, 30) + ']';
                const v = e.target.value;
                e.target.value = v.substring(0,pos) + v.substring(pos).replace(placeholder, failedText);
                e.target.setSelectionRange(pos+failedText.length, pos+failedText.length);
                autoSize(e.target);
            } else {
                console.error('Image upload failed:', xhr.status, xhr.statusText, xhr.responseText);
                const failedText = '[img upload failed]';
                const v = e.target.value;
                e.target.value = v.substring(0,pos) + v.substring(pos).replace(placeholder, failedText);
                e.target.setSelectionRange(pos+failedText.length, pos+failedText.length);
                autoSize(e.target);
            }
        };
        xhr.onerror = function(){
            console.error('Image upload failed: network error');
            const failedText = '[img upload failed]';
            const v = e.target.value;
            e.target.value = v.substring(0,pos) + v.substring(pos).replace(placeholder, failedText);
            e.target.setSelectionRange(pos+failedText.length, pos+failedText.length);
            autoSize(e.target);
        };
        xhr.send(fd);
    }

    function handleUrlPaste(e) {
        const pastedText = e.clipboardData.getData('text');
        if (!pastedText || pastedText.trim() === '') {
            return false;
        }
        const urlStr = pastedText.trim();
        if (!/^https?:\/\/[^\s]+$/.test(urlStr)) {
            return false;
        }
        try {
            new URL(urlStr);
        } catch (err) {
            return false;
        }

        e.preventDefault();

        const start = e.target.selectionStart;
        const end = e.target.selectionEnd;
        const selectedText = e.target.value.substring(start, end);
        const cleanSelected = selectedText.trim();
        const hasNewline = cleanSelected.includes('\n') || cleanSelected.includes('\r');

        let replacement = '';
        if (cleanSelected.length > 0 && !hasNewline) {
            replacement = `[link ${urlStr} ${cleanSelected}]`;
        } else {
            replacement = `[link ${urlStr}]`;
        }

        e.target.setRangeText(replacement, start, end, 'end');
        e.target.dispatchEvent(new Event('input', { bubbles: true }));
        return true;
    }

    function handlePaste(e){
        if (e.target.readOnly || e.target.disabled) {
            return;
        }
        if (e.shiftKey) {
            return;
        }

        const items = e.clipboardData && e.clipboardData.items;
        if(!items) return;

        let hasImage = false;
        for(let i=0;i<items.length;i++){
            const item = items[i];
            if(item.kind === 'file' && item.type.startsWith('image/')){
                hasImage = true;
                handleImagePaste(e, item);
            }
        }

        if (!hasImage) {
            handleUrlPaste(e);
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
