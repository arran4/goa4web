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
        const text = selection.toString();
        if (text) {
            fetch('/api/forum/quote/' + commentId + '?type=selected&selection=' + encodeURIComponent(text))
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
