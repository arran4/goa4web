(function(){
    const uploadUrl = '/images/upload/image';

    function finalizeUpload(form, error) {
        if (!form) return;

        let uploads = parseInt(form.dataset.activeUploads || '0', 10);
        uploads--;
        if (uploads < 0) uploads = 0;
        form.dataset.activeUploads = uploads;

        if (error) {
            form.dataset.uploadError = 'true';
        }

        if (uploads === 0 && form.dataset.pendingSubmit === 'true') {
            form.dataset.pendingSubmit = 'false';

            const btnId = form.dataset.submitButtonId;
            let btn = null;
            if (btnId) {
                btn = document.getElementById(btnId);
            }

            if (btn) {
                btn.disabled = false;
                if (btn.dataset.originalText) {
                    if (btn.tagName === 'INPUT') {
                        btn.value = btn.dataset.originalText;
                    } else {
                        btn.innerHTML = btn.dataset.originalText;
                    }
                }

                if (form.dataset.uploadError === 'true') {
                    form.dataset.uploadError = 'false';
                    alert('An image upload failed. Please review your post before submitting.');
                } else {
                    btn.click();
                }
            }
        }
    }

    function uuidv4(){
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c){
            const r = Math.random()*16|0, v = c=='x'?r:(r&0x3|0x8);
            return v.toString(16);
        });
    }
    function replacePlaceholder(el, placeholder, finalText) {
        const v = el.value;
        const idx = v.indexOf(placeholder);
        if (idx !== -1) {
            const currentStart = el.selectionStart;
            const currentEnd = el.selectionEnd;

            el.value = v.substring(0, idx) + finalText + v.substring(idx + placeholder.length);

            const diff = finalText.length - placeholder.length;
            let newStart = currentStart;
            let newEnd = currentEnd;

            if (currentStart >= idx + placeholder.length) {
                newStart += diff;
            } else if (currentStart > idx) {
                newStart = idx + finalText.length;
            }

            if (currentEnd >= idx + placeholder.length) {
                newEnd += diff;
            } else if (currentEnd > idx) {
                newEnd = idx + finalText.length;
            }

            el.setSelectionRange(newStart, newEnd);
        }
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

        let galleryTarget = null;
        if (e.target.dataset.galleryTarget) {
            galleryTarget = document.querySelector(e.target.dataset.galleryTarget);
        }

        let gallery = null;
        if (galleryTarget) {
            gallery = galleryTarget.querySelector('.image-paste-gallery');
            if (!gallery) {
                gallery = document.createElement('div');
                gallery.className = 'image-paste-gallery';
                galleryTarget.appendChild(gallery);
            }
        } else {
            gallery = e.target.previousElementSibling;
            if (!gallery || !gallery.classList.contains('image-paste-gallery')) {
                gallery = document.createElement('div');
                gallery.className = 'image-paste-gallery';
                e.target.parentNode.insertBefore(gallery, e.target);
            }
        }

        const reader = new FileReader();
        reader.onload = function(re) {
            const itemDiv = document.createElement('div');
            itemDiv.className = 'image-paste-gallery-item';
            itemDiv.id = 'gallery-item-' + id;

            const img = document.createElement('img');
            img.src = re.target.result;
            img.className = 'image-paste-thumb';

            const statusDiv = document.createElement('div');
            statusDiv.className = 'image-paste-status';
            statusDiv.innerText = 'Uploading...';

            const insertBtn = document.createElement('button');
            insertBtn.className = 'image-paste-insert-btn';
            insertBtn.innerText = 'click to insert';
            insertBtn.disabled = true;
            insertBtn.type = 'button';

            itemDiv.appendChild(img);
            itemDiv.appendChild(statusDiv);
            itemDiv.appendChild(insertBtn);
            gallery.appendChild(itemDiv);
        };
        reader.readAsDataURL(file);

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
            const galleryItem = document.getElementById('gallery-item-' + id);
            let statusDiv, insertBtn;
            if (galleryItem) {
                statusDiv = galleryItem.querySelector('.image-paste-status');
                insertBtn = galleryItem.querySelector('.image-paste-insert-btn');
            }

            if(xhr.status >= 200 && xhr.status < 300){
                const ref = xhr.responseText;
                const finalText = '[img '+ref+']';
                replacePlaceholder(e.target, placeholder, finalText);
                autoSize(e.target);

                if (statusDiv) {
                    statusDiv.innerText = ref;
                }
                if (insertBtn) {
                    insertBtn.disabled = false;
                    insertBtn.onclick = function() {
                        const curPos = insertAtCaret(e.target, '[img ' + ref + ']');
                        e.target.dispatchEvent(new Event('input', { bubbles: true }));
                        autoSize(e.target);
                    };
                }
                finalizeUpload(e.target.form, false);
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
                replacePlaceholder(e.target, placeholder, failedText);
                autoSize(e.target);

                if (statusDiv) {
                    statusDiv.innerText = 'Denied';
                }
                finalizeUpload(e.target.form, true);
            } else {
                console.error('Image upload failed:', xhr.status, xhr.statusText, xhr.responseText);
                const failedText = '[img upload failed]';
                replacePlaceholder(e.target, placeholder, failedText);
                autoSize(e.target);

                if (statusDiv) {
                    statusDiv.innerText = 'Failed';
                }
                finalizeUpload(e.target.form, true);
            }
        };
        xhr.onerror = function(){
            console.error('Image upload failed: network error');
            const failedText = '[img upload failed]';
            replacePlaceholder(e.target, placeholder, failedText);
            autoSize(e.target);

            const galleryItem = document.getElementById('gallery-item-' + id);
            if (galleryItem) {
                const statusDiv = galleryItem.querySelector('.image-paste-status');
                if (statusDiv) {
                    statusDiv.innerText = 'Error';
                }
            }
            finalizeUpload(e.target.form, true);
        };

        if (e.target.form) {
            let uploads = parseInt(e.target.form.dataset.activeUploads || '0', 10);
            e.target.form.dataset.activeUploads = uploads + 1;
        }

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
            if (t.form && !t.form.dataset.pasteimgSubmitBound) {
                t.form.dataset.pasteimgSubmitBound = 'true';
                t.form.addEventListener('submit', function(ev) {
                    let uploads = parseInt(ev.target.dataset.activeUploads || '0', 10);
                    if (uploads > 0) {
                        ev.preventDefault();
                        ev.target.dataset.pendingSubmit = 'true';

                        let btn = ev.submitter;
                        if (!btn) {
                            btn = ev.target.querySelector('button[type="submit"], input[type="submit"]');
                        }

                        if (btn) {
                            if (!btn.id) {
                                btn.id = 'submit-btn-' + uuidv4();
                            }
                            ev.target.dataset.submitButtonId = btn.id;

                            if (btn.tagName === 'INPUT') {
                                btn.dataset.originalText = btn.value;
                                btn.value = 'Waiting for uploads...';
                            } else {
                                btn.dataset.originalText = btn.innerHTML;
                                btn.innerHTML = 'Waiting for uploads...';
                            }
                            btn.disabled = true;
                        }
                    }
                });
            }
        });
    });
})();
