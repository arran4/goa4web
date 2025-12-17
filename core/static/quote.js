function quote(type, commentId) {
    if (type === 'selected') {
        const selection = window.getSelection();
        if (selection.rangeCount > 0) {
            const range = selection.getRangeAt(0);
            const start = range.startOffset;
            const end = range.endOffset;
            fetch('/api/forum/quote/' + commentId + '?type=selected&start=' + start + '&end=' + end)
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
