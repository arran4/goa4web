(function() {
    function addLabel(input) {
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
        span.className = 'label pill ' + name + ' unsaved';
        span.textContent = val;
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

    document.addEventListener('click', function(e) {
        if (e.target.classList.contains('remove')) {
            var span = e.target.parentElement;
            var hidden = span.querySelector('input[type="hidden"]');
            if (span.classList.contains('unsaved') && !span.classList.contains('removed')) {
                span.parentElement.removeChild(span);
                return;
            }
            span.classList.toggle('removed');
            if (span.classList.contains('removed')) {
                span.classList.add('unsaved');
                if (hidden) { hidden.disabled = true; }
            } else {
                span.classList.remove('unsaved');
                if (hidden) { hidden.disabled = false; }
            }
        }
    });
    document.querySelectorAll('.label-input').forEach(function(input) {
        input.addEventListener('keydown', function(e) {
            if (e.key === ' ') {
                e.preventDefault();
                addLabel(input);
            }
        });
    });
    document.querySelectorAll('form').forEach(function(form) {
        form.addEventListener('submit', function() {
            form.querySelectorAll('.label-input').forEach(addLabel);
        });
    });
})();
