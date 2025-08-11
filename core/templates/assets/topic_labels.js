(function() {
    document.addEventListener('click', function(e) {
        if (e.target.classList.contains('remove')) {
            var span = e.target.parentElement;
            span.parentElement.removeChild(span);
        }
    });
    document.querySelectorAll('.label-input').forEach(function(input) {
        input.addEventListener('keydown', function(e) {
            if (e.key === ' ') {
                e.preventDefault();
                var val = input.value.trim();
                if (!val) {
                    return;
                }
                var name = input.dataset.type;
                var exists = Array.from(document.querySelectorAll('input[name="' + name + '"]')).some(function(n) { return n.value === val; });
                if (exists) {
                    input.value = '';
                    return;
                }
                var span = document.createElement('span');
                span.className = 'label ' + name;
                span.textContent = val + ' ';
                var btn = document.createElement('button');
                btn.type = 'button';
                btn.className = 'remove';
                btn.textContent = 'x';
                span.appendChild(btn);
                var hidden = document.createElement('input');
                hidden.type = 'hidden';
                hidden.name = name;
                hidden.value = val;
                span.appendChild(hidden);
                input.parentElement.insertBefore(span, input);
                input.value = '';
            }
        });
    });
})();
