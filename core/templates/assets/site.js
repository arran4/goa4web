document.addEventListener('DOMContentLoaded', function() {
    document.body.addEventListener('click', function(e) {
        if (e.target && e.target.classList.contains('quote-link')) {
            e.preventDefault();
            const type = e.target.getAttribute('data-quote-type');
            const commentId = e.target.getAttribute('data-comment-id');
            quote(type, commentId);
        } else if (e.target && e.target.classList.contains('folded-toggle')) {
            e.preventDefault();
            const targetId = e.target.getAttribute('data-target');
            const targetElement = document.getElementById(targetId);
            if (targetElement) {
                targetElement.classList.toggle('hidden');
            }
        }
    });
});

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
