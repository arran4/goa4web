function quote(type, commentId) {
    let text = '';
    if (type === 'selected') {
        text = window.getSelection().toString();
        if (text) {
            let reply = document.getElementById('reply');
            reply.value += '[quote]' + text + '[/quote]\n';
            reply.focus();
            reply.scrollIntoView();
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
