document.addEventListener('DOMContentLoaded', () => {
    const input = document.getElementById('participant-input');
    const addBtn = document.getElementById('add-participant');
    const list = document.getElementById('participants');
    const field = document.getElementById('participants-field');
    const message = document.getElementById('message-field');
    const createBtn = document.getElementById('create-button');
    const topicDetails = document.getElementById('topic-details');
    const titleField = document.getElementById('title-field');

    function updateParticipants() {
        const names = Array.from(list.querySelectorAll('li')).map(li => li.textContent);
        field.value = names.join(',');
        const show = names.length > 0;
        if (topicDetails) topicDetails.style.display = show ? '' : 'none';
        if (createBtn) createBtn.style.display = show ? '' : 'none';
        if (titleField) titleField.value = "Private chat with " + names.join(", ");
    }

    addBtn?.addEventListener('click', (e) => {
        e.preventDefault();
        const name = input.value.trim();
        if (!name) return;
        const li = document.createElement('li');
        li.textContent = name;
        li.addEventListener('click', () => {
            list.removeChild(li);
            updateParticipants();
        });
        list.appendChild(li);
        input.value = '';
        updateParticipants();
    });

    updateParticipants();
});
