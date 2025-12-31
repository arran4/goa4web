document.addEventListener('DOMContentLoaded', function() {
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
            // Calculate absolute offsets relative to the comment container
            const start = calculateOffset(commentContainer, range.startContainer, range.startOffset);
            const end = calculateOffset(commentContainer, range.endContainer, range.endOffset);

            // Construct URL for selected text
            url = basePath + '/topic/' + topicId + '/thread/new?quote_comment_id=' + commentId + '&quote_type=selected&quote_start=' + start + '&quote_end=' + end;
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
                // Calculate absolute offsets relative to the comment container
                const start = calculateOffset(commentContainer, range.startContainer, range.startOffset);
                const end = calculateOffset(commentContainer, range.endContainer, range.endOffset);

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
                 console.error("Selection is not inside the expected comment container");
                 alert("Please select text within the comment you are quoting.");
            }
        }
    } else {
        fetch('/api/forum/quote/' + commentId + '?type=' + type)
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
    }
}

// Helper to calculate absolute character offset of (node, offset) relative to root
function calculateOffset(root, node, offset) {
    const range = document.createRange();
    range.setStart(root, 0);
    range.setEnd(node, offset);
    return range.toString().length;
}
