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
        }
    });
});

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
    if (node.nodeType === Node.TEXT_NODE) {
        // Look for parent with data-start-pos
        const parent = node.parentElement;
        if (parent && parent.hasAttribute('data-start-pos')) {
            const baseStart = parseInt(parent.getAttribute('data-start-pos'), 10);
            const textContent = node.textContent;
            const prefix = textContent.substring(0, offset);
            const byteLen = new TextEncoder().encode(prefix).length;
            return baseStart + byteLen;
        }
    } else if (node.nodeType === Node.ELEMENT_NODE) {
        // If offset points to a child, try to find start pos of that child
        if (offset < node.childNodes.length) {
            const child = node.childNodes[offset];
            if (child.nodeType === Node.ELEMENT_NODE && child.hasAttribute('data-start-pos')) {
                return parseInt(child.getAttribute('data-start-pos'), 10);
            } else if (child.nodeType === Node.TEXT_NODE) {
                 return calculateSourceOffset(child, 0);
            }
        } else {
             // Offset at end.
             if (node.hasAttribute('data-end-pos')) {
                 return parseInt(node.getAttribute('data-end-pos'), 10);
             }
        }
    }
    // Fallback: try to find nearest ancestor with data-start-pos
    let current = node;
    while (current) {
        if (current.nodeType === Node.ELEMENT_NODE && current.hasAttribute('data-start-pos')) {
             return parseInt(current.getAttribute('data-start-pos'), 10);
        }
        current = current.parentNode;
    }
    return -1;
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
