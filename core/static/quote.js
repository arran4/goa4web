function quote(type, commentId, username) {
    let text = '';
    if (type === 'selected') {
        text = window.getSelection().toString();
        if (text) {
            document.getElementById('reply').value += '[quoteof "' + username + '"]' + text + '[/quoteof]\n';
        }
    } else {
        fetch('/api/forum/quote/' + commentId + '?type=' + type)
            .then(response => response.json())
            .then(data => {
                document.getElementById('reply').value += data.text;
            })
            .catch(error => {
                console.error('Error:', error);
            });
    }
}
