document.addEventListener('DOMContentLoaded', function() {
    foldLongQuotes(document);

    document.body.addEventListener('click', function(e) {
        if (e.target && e.target.classList.contains('quote-link')) {
            e.preventDefault();
            const type = e.target.getAttribute('data-quote-type');
            const commentId = e.target.getAttribute('data-comment-id');
            quote(type, commentId);
        } else if (e.target && e.target.classList.contains('quote-new-thread-link')) {
            e.preventDefault();
            const commentId = e.target.getAttribute('data-comment-id');
            const topicId = e.target.getAttribute('data-topic-id');
            quoteInNewThread(commentId, topicId, e);
        } else if (e.target && e.target.classList.contains('folded-toggle')) {
            e.preventDefault();
            const targetId = e.target.getAttribute('data-target');
            const targetElement = document.getElementById(targetId);
            if (targetElement) {
                targetElement.classList.toggle('hidden');
            }
        } else if (e.target && e.target.classList.contains('convert-markdown-to-a4code')) {
            e.preventDefault();
            const targetId = e.target.getAttribute('data-target');
            convertMarkdownToA4Code(targetId);
        } else if (e.target && e.target.classList.contains('convert-a4code-to-markdown')) {
            e.preventDefault();
            const targetId = e.target.getAttribute('data-target');
            convertA4CodeToMarkdown(targetId);
        } else if (e.target && e.target.classList.contains('preview-a4code')) {
            e.preventDefault();
            const targetId = e.target.getAttribute('data-target');
            const previewUrl = e.target.getAttribute('data-preview-url');
            const containerId = e.target.getAttribute('data-container');
            previewA4Code(targetId, previewUrl, containerId);
        } else if (e.target && e.target.classList.contains('share-button')) {
            e.preventDefault();
            const link = e.target.getAttribute('data-link');
            const module = e.target.getAttribute('data-module');
            share(link, module, e.target);
        } else if (e.target && e.target.classList.contains('copy-share-url-button')) {
            e.preventDefault();
            const container = e.target.closest('div');
            if (container) {
                const input = container.querySelector('.share-url-input');
                if (input) {
                    navigator.clipboard.writeText(input.value).then(() => {
                    }).catch(err => {
                        console.error('Failed to copy text: ', err);
                    });
                }
            }
        } else if (e.target && e.target.classList.contains('copy-config-command')) {
            e.preventDefault();
            const command = e.target.getAttribute('data-copy');
            if (command) {
                navigator.clipboard.writeText(command).then(() => {
                }).catch(err => {
                    console.error('Failed to copy text: ', err);
                });
            }
        } else if (e.target && e.target.closest('.a4code-btn')) {
            e.preventDefault();
            const btn = e.target.closest('.a4code-btn');
            const tag = btn.getAttribute('data-tag');
            const targetId = btn.getAttribute('data-target');
            insertA4CodeTag(targetId, tag);
        }
    });

    setupKeyboardShortcuts();
});

function insertA4CodeTag(targetId, tag) {
    const textarea = document.getElementById(targetId);
    if (!textarea) return;

    if (tag === 'quote' && quoteDocumentSelectionIntoEditor(targetId)) {
        return;
    }

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);
    let replacement = '';

    if (tag === 'b') {
        replacement = `[b ${selectedText}]`;
    } else if (tag === 'i') {
        replacement = `[i ${selectedText}]`;
    } else if (tag === 'a') {
        const url = prompt("Enter URL:", "https://");
        if (url) {
            replacement = `[a ${url} ${selectedText}]`;
        } else {
            return; // Cancelled
        }
    } else if (tag === 'img') {
        const url = prompt("Enter Image URL:", "https://");
        if (url) {
            replacement = `[img ${url}]`;
        } else {
            return; // Cancelled
        }
    } else if (tag === 'u') {
        replacement = `[u ${selectedText}]`;
    } else if (tag === 'sub') {
        replacement = `[sub ${selectedText}]`;
    } else if (tag === 'sup') {
        replacement = `[sup ${selectedText}]`;
    } else if (tag === 'spoiler') {
        replacement = `[spoiler ${selectedText}]`;
    } else if (tag === 'hr') {
        replacement = `[hr]`;
    } else if (tag === 'quote') {
        replacement = `[quote ${selectedText}]`;
    } else if (tag === 'code') {
        replacement = `[code ${selectedText}]`;
    }

    textarea.setRangeText(replacement, start, end, 'select');
}

function quoteDocumentSelectionIntoEditor(targetId) {
    const ranges = selectedCommentRanges();
    if (ranges.length === 0) {
        return false;
    }

    const textarea = document.getElementById(targetId);
    if (!textarea) {
        return false;
    }

    const headers = {
        'Content-Type': 'application/json',
    };
    const csrfToken = document.querySelector('input[name="csrf_token"]');
    if (csrfToken) {
        headers['X-CSRF-Token'] = csrfToken.value;
    }

    fetch('/api/forum/quote-selection', {
        method: 'POST',
        headers: headers,
        body: JSON.stringify({ ranges: ranges })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Quote selection request failed');
        }
        return response.json();
    })
    .then(data => {
        insertTextIntoEditor(textarea, data.text || '');
    })
    .catch(error => {
        console.error('Error:', error);
        alert('An error occurred while quoting the selection.');
    });

    return true;
}

function selectedCommentRanges() {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0 || selection.isCollapsed) {
        return [];
    }

    const result = [];
    for (let i = 0; i < selection.rangeCount; i++) {
        const selectionRange = selection.getRangeAt(i);
        const comments = document.querySelectorAll('section.body > div[id^="comment-"], .body > div[id^="comment-"]');
        comments.forEach(comment => {
            if (!rangeIntersectsNode(selectionRange, comment)) {
                return;
            }

            const commentID = parseInt(comment.id.replace('comment-', ''), 10);
            if (Number.isNaN(commentID)) {
                return;
            }

            const start = commentContainsNode(comment, selectionRange.startContainer)
                ? calculateSourceOffset(selectionRange.startContainer, selectionRange.startOffset)
                : calculateSourceOffset(comment, 0);
            const end = commentContainsNode(comment, selectionRange.endContainer)
                ? calculateSourceOffset(selectionRange.endContainer, selectionRange.endOffset)
                : calculateSourceOffset(comment, comment.childNodes.length);

            if (start !== -1 && end !== -1 && end > start) {
                result.push({ comment_id: commentID, start: start, end: end });
            }
        });
    }

    result.sort((a, b) => {
        const aNode = document.getElementById('comment-' + a.comment_id);
        const bNode = document.getElementById('comment-' + b.comment_id);
        if (!aNode || !bNode || aNode === bNode) {
            return a.start - b.start;
        }
        const pos = aNode.compareDocumentPosition(bNode);
        return pos & Node.DOCUMENT_POSITION_PRECEDING ? 1 : -1;
    });
    return result;
}

function rangeIntersectsNode(range, node) {
    if (typeof range.intersectsNode === 'function') {
        try {
            return range.intersectsNode(node);
        } catch (e) {
            return false;
        }
    }

    const nodeRange = document.createRange();
    nodeRange.selectNodeContents(node);
    return range.compareBoundaryPoints(Range.END_TO_START, nodeRange) < 0 &&
        range.compareBoundaryPoints(Range.START_TO_END, nodeRange) > 0;
}

function commentContainsNode(comment, node) {
    return node === comment || comment.contains(node);
}

function insertTextIntoEditor(textarea, text) {
    if (!text) {
        return;
    }
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const prefix = start > 0 && textarea.value.charAt(start - 1) !== '\n' ? '\n' : '';
    const suffix = end < textarea.value.length && !text.endsWith('\n') ? '\n' : '';
    textarea.setRangeText(prefix + text + suffix, start, end, 'end');
    textarea.focus();
}

function setupKeyboardShortcuts() {
    const overlay = document.getElementById('keyboard-shortcuts-overlay');
    const closeBtn = document.getElementById('close-shortcuts');
    const tabs = document.querySelectorAll('.tab-btn');
    const sections = document.querySelectorAll('.shortcut-section');

    if (!overlay) return;

    function toggleOverlay() {
        overlay.classList.toggle('hidden');
        if (!overlay.classList.contains('hidden')) {
            let activeTab = 'global';
            if (document.activeElement && document.activeElement.tagName === 'TEXTAREA') {
                activeTab = 'editor';
            } else if (window.location.pathname.includes('/forum') || window.location.pathname.includes('/topic')) {
                activeTab = 'forum';
            }
            switchTab(activeTab);
        }
    }

    function switchTab(tabName) {
        tabs.forEach(t => {
            if (t.getAttribute('data-tab') === tabName) {
                t.classList.add('active');
            } else {
                t.classList.remove('active');
            }
        });
        sections.forEach(s => {
            if (s.id === 'section-' + tabName) {
                s.classList.remove('hidden');
            } else {
                s.classList.add('hidden');
            }
        });
    }

    if (closeBtn) {
        closeBtn.addEventListener('click', function() {
            overlay.classList.add('hidden');
        });
    }

    overlay.addEventListener('click', function(e) {
        if (e.target === overlay) {
            overlay.classList.add('hidden');
        }
    });

    tabs.forEach(tab => {
        tab.addEventListener('click', function() {
            switchTab(this.getAttribute('data-tab'));
        });
    });

    document.addEventListener('keydown', function(e) {
        if (e.ctrlKey && (e.key === '?' || e.key === '/')) {
            e.preventDefault();
            toggleOverlay();
            return;
        }

        if (e.key === 'Escape' && !overlay.classList.contains('hidden')) {
            overlay.classList.add('hidden');
            return;
        }

        if (!overlay.classList.contains('hidden')) return;

        if (e.ctrlKey && e.key === 'Enter') {
            if (e.target.tagName === 'TEXTAREA') {
                const form = e.target.closest('form');
                if (form) {
                    e.preventDefault();
                    const submitBtn = form.querySelector('input[type="submit"], button[type="submit"]');
                    if (submitBtn) {
                        submitBtn.click();
                    } else {
                        form.submit();
                    }
                }
            }
        }

        if (e.altKey && e.key === 'Home') {
            e.preventDefault();
            window.location.href = '/';
        }

        const isTyping = e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.isContentEditable;
        if (!isTyping && !e.ctrlKey && !e.altKey && !e.metaKey) {
            if (e.key === 'j') {
                scrollComment(1);
            } else if (e.key === 'k') {
                scrollComment(-1);
            } else if (e.key === 'q') {
                 const selection = window.getSelection();
                 if (selection.rangeCount > 0 && !selection.isCollapsed) {
                     const anchor = selection.anchorNode;
                     const commentDiv = anchor.nodeType === Node.ELEMENT_NODE ? anchor.closest('.comment') : anchor.parentElement.closest('.comment');
                     if (commentDiv) {
                         const commentId = commentDiv.id.replace('comment-', '');
                         if (commentId) {
                             e.preventDefault();
                             quote('selected', commentId);
                         }
                     }
                 }
            }
        }
    });
}

let currentCommentIndex = -1;
function scrollComment(direction) {
    const comments = document.querySelectorAll('.comment');
    if (comments.length === 0) return;

    if (currentCommentIndex === -1) {
         for (let i = 0; i < comments.length; i++) {
             const rect = comments[i].getBoundingClientRect();
             if (rect.top >= 0) {
                 currentCommentIndex = i;
                 break;
             }
         }
         if (currentCommentIndex === -1) currentCommentIndex = 0;
    } else {
        currentCommentIndex += direction;
    }

    if (currentCommentIndex < 0) currentCommentIndex = 0;
    if (currentCommentIndex >= comments.length) currentCommentIndex = comments.length - 1;

    comments[currentCommentIndex].scrollIntoView({ behavior: 'smooth', block: 'center' });
    comments.forEach(c => c.style.outline = 'none');
    comments[currentCommentIndex].style.outline = '2px solid #800000';
    setTimeout(() => {
         if (comments[currentCommentIndex]) comments[currentCommentIndex].style.outline = 'none';
    }, 2000);
}

function convertMarkdownToA4Code(targetId) {
    const textarea = document.getElementById(targetId);
    if (!textarea) return;
    if (window.A4Code) {
        textarea.value = A4Code.markdownToA4Code(textarea.value);
    } else {
        alert("A4Code library not loaded");
    }
}

function convertA4CodeToMarkdown(targetId) {
    const textarea = document.getElementById(targetId);
    if (!textarea) return;
    if (window.A4Code) {
        textarea.value = A4Code.a4codeToMarkdown(textarea.value);
    } else {
        alert("A4Code library not loaded");
    }
}

function previewA4Code(targetId, previewUrl, containerId) {
    const textarea = document.getElementById(targetId);
    if (!textarea) return;

    const text = textarea.value;
    let previewContainer = document.getElementById('preview-container');
    let previewContent = document.getElementById('preview-content');
    if (containerId) {
        previewContainer = document.getElementById(containerId);
        if (previewContainer) {
            previewContent = previewContainer.querySelector('.preview-box');
        }
    }

    const headers = {
        'Content-Type': 'text/plain',
    };
    const csrfToken = document.querySelector('input[name="csrf_token"]');
    if (csrfToken) {
        headers['X-CSRF-Token'] = csrfToken.value;
    }

    fetch(previewUrl, {
        method: 'POST',
        headers: headers,
        body: text
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.text();
    })
    .then(html => {
        previewContent.innerHTML = html;
        foldLongQuotes(previewContent);
        previewContainer.classList.remove('hidden');
    })
    .catch(error => {
        console.error('Error fetching preview:', error);
        alert('Failed to generate preview.');
    });
}

function quoteInNewThread(commentId, topicId, event) {
    const selection = window.getSelection();
    let url = '';

    // Determine base path based on current location (public or private forum)
    let basePath = '/forum';
    if (window.location.pathname.startsWith('/private')) {
        basePath = '/private';
    }

    if (selection.rangeCount > 0) {
        const range = selection.getRangeAt(0);
        const commentContainer = document.getElementById('comment-' + commentId);

        if (commentContainer && commentContainer.contains(range.commonAncestorContainer)) {
            // Calculate absolute offsets based on data attributes
            const start = calculateSourceOffset(range.startContainer, range.startOffset);
            const end = calculateSourceOffset(range.endContainer, range.endOffset);

            if (start !== -1 && end !== -1) {
                // Construct URL for selected text
                url = basePath + '/topic/' + topicId + '/thread/new?quote_comment_id=' + commentId + '&quote_type=selected&quote_start=' + start + '&quote_end=' + end;
            }
        }
    }

    // If no selection or invalid selection, maybe fallback to full quote?
    // Or just alert? The UI says "QUOTE SELECTED".
    // If invalid selection, we shouldn't proceed or just do nothing.
    if (!url) {
        alert("Please select text within the comment you are quoting.");
        return;
    }

    if (event.ctrlKey || event.metaKey || event.shiftKey) {
        window.open(url, '_blank');
    } else {
        window.location.href = url;
    }
}

function quote(type, commentId) {
    if (type === 'selected') {
        const selection = window.getSelection();
        if (selection.rangeCount > 0) {
            const range = selection.getRangeAt(0);
            const commentContainer = document.getElementById('comment-' + commentId);

            if (commentContainer && commentContainer.contains(range.commonAncestorContainer)) {
                // Calculate absolute offsets based on data attributes
                const start = calculateSourceOffset(range.startContainer, range.startOffset);
                const end = calculateSourceOffset(range.endContainer, range.endOffset);

                if (start !== -1 && end !== -1) {
                    // Construct URL
                    let url = '/api/forum/quote/' + commentId + '?type=selected&start=' + start + '&end=' + end;

                    fetch(url)
                        .then(response => response.json())
                        .then(data => {
                            let reply = document.getElementById('reply');
                            reply.value += data.text;
                            reply.focus();
                            reply.scrollIntoView();
                        })
                        .catch(error => {
                            console.error('Error:', error);
                            alert('An error occurred while quoting the comment.');
                        });
                } else {
                    console.error("Could not calculate source offset");
                    alert("Please select text within the comment you are quoting.");
                }
            } else {
                 console.error("Selection is not inside the expected comment container");
                 alert("Please select text within the comment you are quoting.");
            }
        }
    } else {
        fetch('/api/forum/quote/' + commentId + '?type=' + type)
            .then(response => response.json())
            .then(data => {
                insertQuote(data.text);
            })
            .catch(error => {
                console.error('Error:', error);
                alert('An error occurred while quoting the comment.');
            });
    }
}

// Helper to calculate absolute source offset based on data attributes
function calculateSourceOffset(node, offset) {
    const boundaryOffset = sourceBoundaryOffset(node, offset);
    if (boundaryOffset !== -1) {
        return boundaryOffset;
    }

    const annotated = nearestSourceAnnotatedElement(node);
    if (!annotated) {
        return -1;
    }

    const startAttr = sourceStartAttr(annotated);
    if (startAttr === null || startAttr === undefined) {
        return -1;
    }

    const relative = sourceOffsetWithin(annotated, node, offset);
    if (relative === -1) {
        return -1;
    }

    return parseInt(startAttr, 10) + relative;
}

function sourceBoundaryOffset(node, offset) {
    if (node.nodeType !== Node.ELEMENT_NODE) {
        return -1;
    }

    if (offset < node.childNodes.length) {
        const child = node.childNodes[offset];
        if (child.nodeType === Node.ELEMENT_NODE) {
            const startAttr = sourceStartAttr(child);
            if (startAttr !== null && startAttr !== undefined) {
                return parseInt(startAttr, 10);
            }
        }
    }

    if (offset > 0 && offset <= node.childNodes.length) {
        const previous = node.childNodes[offset - 1];
        if (previous.nodeType === Node.ELEMENT_NODE) {
            const endAttr = previous.getAttribute('data-end-pos');
            if (endAttr !== null && endAttr !== undefined) {
                return parseInt(endAttr, 10);
            }
        }
    }

    if (offset === node.childNodes.length && node.hasAttribute('data-end-pos')) {
        return parseInt(node.getAttribute('data-end-pos'), 10);
    }

    return -1;
}

function nearestSourceAnnotatedElement(node) {
    let current = node.nodeType === Node.ELEMENT_NODE ? node : node.parentElement;
    while (current) {
        if (current.nodeType === Node.ELEMENT_NODE &&
            sourceStartAttr(current) !== null) {
            return current;
        }
        current = current.parentElement || current.parentNode;
    }
    return null;
}

function sourceOffsetWithin(root, targetNode, targetOffset) {
    let total = 0;
    let found = false;

    function walk(node) {
        if (found) {
            return;
        }

        if (node === targetNode) {
            if (node.nodeType === Node.TEXT_NODE) {
                total += byteLength(node.textContent.substring(0, targetOffset));
            } else if (node.nodeType === Node.ELEMENT_NODE) {
                const limit = Math.min(targetOffset, node.childNodes.length);
                for (let i = 0; i < limit; i++) {
                    total += sourceLength(node.childNodes[i]);
                }
            }
            found = true;
            return;
        }

        total += sourceLength(node);
    }

    if (root === targetNode) {
        walk(root);
    } else {
        for (let i = 0; i < root.childNodes.length && !found; i++) {
            walkUntilTarget(root.childNodes[i]);
        }
    }

    return found ? total : -1;

    function walkUntilTarget(node) {
        if (found) {
            return;
        }
        if (node === targetNode) {
            walk(node);
            return;
        }
        if (node.nodeType === Node.ELEMENT_NODE) {
            if (!nodeContainsTarget(node, targetNode)) {
                total += sourceLength(node);
                return;
            }
            for (let i = 0; i < node.childNodes.length && !found; i++) {
                walkUntilTarget(node.childNodes[i]);
            }
            if (!found) {
                total += elementOwnSourceLength(node);
            }
            return;
        }
        total += sourceLength(node);
    }
}

function nodeContainsTarget(node, targetNode) {
    if (node === targetNode) {
        return true;
    }
    for (let i = 0; i < node.childNodes.length; i++) {
        if (nodeContainsTarget(node.childNodes[i], targetNode)) {
            return true;
        }
    }
    return false;
}

function sourceLength(node) {
    if (node.nodeType === Node.TEXT_NODE) {
        return byteLength(node.textContent);
    }
    if (node.nodeType === Node.ELEMENT_NODE) {
        if (isLineBreak(node)) {
            return 1;
        }
        const annotatedLength = sourceAnnotatedLength(node);
        if (annotatedLength !== -1) {
            return annotatedLength;
        }
        let total = elementOwnSourceLength(node);
        for (let i = 0; i < node.childNodes.length; i++) {
            total += sourceLength(node.childNodes[i]);
        }
        return total;
    }
    return 0;
}

function elementOwnSourceLength(node) {
    if (isLineBreak(node)) {
        return 1;
    }
    return 0;
}

function isLineBreak(node) {
    return node.nodeType === Node.ELEMENT_NODE && node.tagName && node.tagName.toLowerCase() === 'br';
}

function sourceAnnotatedLength(node) {
    if (node.nodeType !== Node.ELEMENT_NODE) {
        return -1;
    }
    const startAttr = sourceStartAttr(node);
    const endAttr = node.getAttribute('data-end-pos');
    if (startAttr === null || startAttr === undefined || endAttr === null || endAttr === undefined) {
        return -1;
    }
    const start = parseInt(startAttr, 10);
    const end = parseInt(endAttr, 10);
    if (Number.isNaN(start) || Number.isNaN(end) || end < start) {
        return -1;
    }
    return end - start;
}

function sourceStartAttr(node) {
    const startAttr = node.getAttribute('data-start-pos');
    return startAttr === undefined ? null : startAttr;
}

function byteLength(text) {
    return new TextEncoder().encode(text).length;
}

function share(link, module, button) {
    const shareLinkInput = button.closest('div').querySelector('.share-url-input');
    const copyButton = button.closest('div').querySelector('.copy-share-url-button');
    fetch('/api/' + module + '/share?link=' + encodeURIComponent(link))
        .then(response => response.json())
        .then(data => {
            shareLinkInput.value = data.signed_url + window.location.hash;
            shareLinkInput.style.display = 'inline-block';
            copyButton.style.display = 'inline-block';
            button.style.display = 'none';
            shareLinkInput.select();
        })
        .catch(error => {
            console.error('Error:', error);
            alert('An error occurred while generating the share link.');
        });
}

function insertQuote(text) {
    let reply = document.getElementById('reply');
    if (!reply) return;

    let cursorPos = reply.selectionEnd;
    let textVal = reply.value;
    let nextNewLine = textVal.indexOf('\n', cursorPos);
    let insertPos = nextNewLine === -1 ? textVal.length : nextNewLine + 1;

    let textToInsert = text;
    // Ensure we start on a new line if not at the beginning
    if (insertPos > 0 && textVal.charAt(insertPos - 1) !== '\n') {
        textToInsert = '\n' + textToInsert;
    }

    // Ensure we end with a newline if we are not at the end of the text
    if (insertPos < textVal.length && !textToInsert.endsWith('\n')) {
        textToInsert += '\n';
    }

    const before = textVal.substring(0, insertPos);
    const after = textVal.substring(insertPos);

    reply.value = before + textToInsert + after;

    const newCursorPos = before.length + textToInsert.length;
    reply.setSelectionRange(newCursorPos, newCursorPos);

    reply.focus();
    reply.scrollIntoView();
}

function foldLongQuotes(container) {
    if (!container) return;
    const quotes = container.querySelectorAll('.quote-body');
    quotes.forEach(quote => {
        // Skip if already processed
        if (quote.nextElementSibling && quote.nextElementSibling.classList.contains('folded-toggle')) {
            return;
        }

        if (quote.scrollHeight > 250) {
            quote.classList.add('collapsed');

            const toggle = document.createElement('a');
            toggle.className = 'folded-toggle';
            toggle.innerText = 'Expand quote';
            toggle.href = '#';
            toggle.onclick = function(e) {
                e.preventDefault();
                quote.classList.toggle('collapsed');
                if (quote.classList.contains('collapsed')) {
                    toggle.innerText = 'Expand quote';
                } else {
                    toggle.innerText = 'Collapse quote';
                }
            };

            // Insert after the quote body
            if (quote.parentNode) {
                quote.parentNode.insertBefore(toggle, quote.nextSibling);
            }
        }
    });
}

function replaceContent(event, url, targetId) {
    if (event) {
        event.preventDefault();
        event.stopPropagation();
    }

    const separator = url.includes('?') ? '&' : '?';
    const fetchUrl = url + separator + 'ajax=1';

    fetch(fetchUrl)
        .then(response => {
            if (response.ok) {
                return response.text();
            }
            throw new Error('Network response was not ok');
        })
        .then(html => {
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, 'text/html');
            const newElement = doc.body.firstElementChild;
            const currentElement = document.getElementById(targetId);

            if (currentElement && newElement) {
                const currentTs = parseInt(currentElement.dataset.timestamp || '0', 10);
                const newTs = parseInt(newElement.dataset.timestamp || '0', 10);

                if (newTs > currentTs) {
                    currentElement.replaceWith(newElement);
                }
            }
        })
        .catch(error => {
            console.error('Error replacing content:', error);
        });
    return false;
}
