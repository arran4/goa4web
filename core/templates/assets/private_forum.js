document.addEventListener('DOMContentLoaded', () => {
    const input = document.getElementById('participant-input');
    const addBtn = document.getElementById('add-participant');
    const list = document.getElementById('participants');
    const field = document.getElementById('participants-field');

    function updateField() {
        const names = Array.from(list.querySelectorAll('li')).map(li => li.textContent);
        field.value = names.join(',');
    }

    addBtn?.addEventListener('click', (e) => {
        e.preventDefault();
        const name = input.value.trim();
        if (!name) return;
        const li = document.createElement('li');
        li.textContent = name;
        li.addEventListener('click', () => {
            list.removeChild(li);
            updateField();
        });
        list.appendChild(li);
        input.value = '';
        updateField();
    });
});
