document.addEventListener('DOMContentLoaded', function() {
    const globalBtn = document.getElementById('commit-all');
    function updateGlobalBtn() {
        if (document.querySelectorAll('.pill.moved').length > 0) {
            globalBtn.classList.remove('hidden');
        } else {
            globalBtn.classList.add('hidden');
        }
    }
    function prepareForm(form) {
        const row = form.closest('tr');
        const have = row.querySelector('.have');
        const disabled = row.querySelector('.disabled');
        const acts = Array.from(have.querySelectorAll('.pill')).map(p => p.textContent.trim());
        const dis = Array.from(disabled.querySelectorAll('.pill')).map(p => p.textContent.trim());
        form.querySelector('input[name="actions"]').value = acts.join(',');
        form.querySelector('input[name="disabled_actions"]').value = dis.join(',');
    }
    document.querySelectorAll('tr[data-section]').forEach(function(row) {
        const have = row.querySelector('.have');
        const avail = row.querySelector('.available');
        const disabled = row.querySelector('.disabled');
        const form = row.querySelector('.commit-form');
        const btn = form.querySelector('input[type="submit"]');
        function updateBtn() {
            if (row.querySelectorAll('.pill.moved').length > 0) {
                btn.classList.remove('hidden');
            } else {
                btn.classList.add('hidden');
            }
            updateGlobalBtn();
        }
        function dropHandler(e) {
            e.preventDefault();
            const pill = document.querySelector('.pill.dragging');
            if (!pill || e.currentTarget === pill.parentNode) return;
            e.currentTarget.appendChild(pill);
            let tgt = 'available';
            if (e.currentTarget.classList.contains('have')) tgt = 'have';
            else if (e.currentTarget.classList.contains('disabled')) tgt = 'disabled';
            if (pill.dataset.default !== tgt) {
                pill.classList.add('moved');
            } else {
                pill.classList.remove('moved');
            }
            updateBtn();
        }
        [have, avail, disabled].forEach(function(col) {
            col.addEventListener('dragstart', function(e) {
                if (e.target.classList.contains('pill')) {
                    e.dataTransfer.setData('text/plain', e.target.textContent);
                    e.dataTransfer.effectAllowed = 'move';
                    e.target.classList.add('dragging');
                }
            }, true);
            col.addEventListener('dragend', function(e) {
                if (e.target.classList.contains('pill')) {
                    e.target.classList.remove('dragging');
                }
            }, true);
            col.addEventListener('dragover', function(e) { e.preventDefault(); });
            col.addEventListener('drop', dropHandler);
        });
        form.addEventListener('submit', function() {
            prepareForm(form);
        });
    });
    globalBtn.addEventListener('click', function() {
        const forms = Array.from(document.querySelectorAll('.commit-form')).filter(f => !f.querySelector('input[type="submit"]').classList.contains('hidden'));
        Promise.all(forms.map(f => {
            prepareForm(f);
            return fetch(f.action, {method: 'POST', body: new FormData(f)});
        })).then(() => window.location.reload());
    });
});
