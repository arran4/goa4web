(function() {
    let pendingPaste = null;

    document.addEventListener('paste', function(e) {
        if (e.target.tagName !== 'TEXTAREA') return;
        if (e.shiftKey) return; // Ctrl+Shift+V

        const clipboardData = e.clipboardData || window.clipboardData;
        const text = clipboardData.getData('text/plain');
        if (!text) return;

        // Check if it's a URL
        const urlRegex = /^(https?:\/\/[^\s]+)$/i;
        const match = text.trim().match(urlRegex);
        if (!match) return;

        const url = match[1];

        const selectionStart = e.target.selectionStart;
        const selectionEnd = e.target.selectionEnd;

        // Predict where it will be
        pendingPaste = {
            url: url,
            start: selectionStart,
            end: selectionStart + text.length,
            originalText: text, // in case user modifies it, we might want to verify
            target: e.target
        };

        // We rely on browser insertion.
    }, true);

    document.addEventListener('input', function(e) {
        if (!pendingPaste) return;
        if (e.target !== pendingPaste.target) return;

        // Check if the input was a whitespace (space or newline)
        const val = e.target.value;
        const currentPos = e.target.selectionEnd;

        if (currentPos > pendingPaste.end) {
            const charBefore = val.substring(currentPos - 1, currentPos);
            if (/\s/.test(charBefore)) {
                // Trigger!
                triggerEnhancement(pendingPaste);
                pendingPaste = null;
            }
        }
    }, true);

    document.addEventListener('blur', function(e) {
        if (!pendingPaste) return;
        if (e.target !== pendingPaste.target) return;

        // Trigger on blur
        triggerEnhancement(pendingPaste);
        pendingPaste = null;
    }, true);

    function triggerEnhancement(pasteInfo) {
        const textarea = pasteInfo.target;
        const url = pasteInfo.url;

        // Verify the URL is still there (user didn't delete it)
        const currentVal = textarea.value;
        const actualText = currentVal.substring(pasteInfo.start, pasteInfo.end);

        // Loose check: contains the URL?
        if (!actualText.includes(url.trim())) {
           // User modified it
        }

        // Determine Context
        // Case 1: Inline/After Paragraph
        // Case 2: New Paragraph (Two new lines BEFORE the link)
        // Case 3: Only Content (The link is the only thing)

        // Check content BEFORE the paste position
        const textBefore = currentVal.substring(0, pasteInfo.start);
        const textAfter = currentVal.substring(pasteInfo.end);

        const isOnlyContent = textBefore.trim() === '' && textAfter.trim() === '';

        // Check for two new lines before or start of file
        const endsWithDoubleNewline = /\n\s*\n\s*$/.test(textBefore);
        const isStart = textBefore.trim() === '';

        const isNewParagraph = isOnlyContent || isStart || endsWithDoubleNewline;

        // Case distinction
        let mode = 1; // Inline
        if (isOnlyContent) {
            mode = 3;
        } else if (isNewParagraph) {
            mode = 2;
        }

        // Fetch metadata
        fetchMetadata(url).then(meta => {
            if (!meta) return;

            let replacement = '';

            let title = meta.title || url;
            let description = meta.description || '';
            let imageRef = meta.image_ref;

            if (mode === 1) {
                // Truncate title
                if (title.length > 255) title = title.substring(0, 252) + '...';
                replacement = `[url=${url}]${title}[/url]`;
            } else if (mode === 2) {
                // New Paragraph
                if (title.length > 255) title = title.substring(0, 252) + '...';
                replacement = `[url=${url}]${title}[/url]`;
                if (imageRef) {
                    replacement += `\n[img ${imageRef}]`;
                }
            } else if (mode === 3) {
                // Only Content
                replacement = `[url=${url}]${title}[/url]`;
                if (imageRef) {
                    replacement += `\n[img ${imageRef}]`;
                }
                if (description) {
                    replacement += `\n${description}`;
                }
            }

            textarea.setRangeText(replacement, pasteInfo.start, pasteInfo.end, 'end');
            textarea.dispatchEvent(new Event('input', { bubbles: true }));

        }).catch(err => {
            console.error("Metadata fetch failed", err);
        });
    }

    function fetchMetadata(url) {
        // Call our new API
        const apiEndpoint = '/api/metadata?url=' + encodeURIComponent(url);
        return fetch(apiEndpoint)
            .then(res => {
                if (!res.ok) throw new Error(res.statusText);
                return res.json();
            });
    }

})();
