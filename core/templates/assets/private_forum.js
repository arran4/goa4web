document.addEventListener('DOMContentLoaded', () => {
    const input = document.getElementById('participant-input');
    const addBtn = document.getElementById('add-participant');
    const list = document.getElementById('participants');
    const field = document.getElementById('participants-field');
    const invalidField = document.getElementById('invalid-participants-field');
    const message = document.getElementById('message-field');
    const createBtn = document.getElementById('create-button');
    const topicDetails = document.getElementById('topic-details');
    const titleField = document.getElementById('title-field');

    let invalidUsers = [];
    if (invalidField && invalidField.value) {
        invalidUsers = invalidField.value.split(',');
    }

    // Populate list from existing participants
    if (field && field.value) {
        field.value.split(',').forEach(name => {
            name = name.trim();
            if (name) addListItem(name);
        });
    }

    function updateParticipants() {
        // Collect names from spans inside lis, ignoring the remove button
        const names = Array.from(list.querySelectorAll('li')).map(li => {
            const span = li.querySelector('.name');
            return span ? span.textContent : li.firstChild.textContent; // Fallback if no span
        });
        field.value = names.join(',');
        const show = names.length > 0;
        if (topicDetails) topicDetails.style.display = show ? '' : 'none';
        if (createBtn) createBtn.style.display = show ? '' : 'none';
        // Only update title if it's empty or starts with "Private chat with"
        if (titleField && (!titleField.value || titleField.value.startsWith("Private chat with"))) {
             titleField.value = "Private chat with " + names.join(", ");
        }
    }

    function addListItem(name) {
        const li = document.createElement('li');

        const nameSpan = document.createElement('span');
        nameSpan.textContent = name;
        nameSpan.className = 'name';
        if (invalidUsers.includes(name)) {
            nameSpan.style.textDecoration = 'line-through';
            nameSpan.style.color = 'red';
        }
        li.appendChild(nameSpan);

        const removeBtn = document.createElement('button');
        removeBtn.textContent = '[x]';
        removeBtn.className = 'remove-participant';
        removeBtn.style.marginLeft = '10px';
        removeBtn.style.cursor = 'pointer';
        removeBtn.type = 'button'; // Prevent form submission

        removeBtn.addEventListener('click', (e) => {
            e.stopPropagation(); // Prevent bubbling if li has click listener
            list.removeChild(li);
            updateParticipants();
        });

        li.appendChild(removeBtn);
        list.appendChild(li);
    }

    addBtn?.addEventListener('click', (e) => {
        e.preventDefault();
        const name = input.value.trim();
        if (!name) return;
        addListItem(name);
        input.value = '';
        updateParticipants();
    });

    // Initial update not needed if we populate via loop, but good for visibility check
    updateParticipants();
});
