"use strict";

function onReady(fn) {
    if (document.readyState !== 'loading') {
        fn();
    } else {
        document.addEventListener('DOMContentLoaded', fn);
    }
}

onReady(() => {
    const draftsContainer = document.getElementById('drafts-container');
    if (!draftsContainer) {
        return;
    }

    const saveDraftButton = document.getElementById('save-draft');
    const draftsList = document.getElementById('drafts-list');
    const replyTextarea = document.getElementById('reply');
    const draftIdInput = document.querySelector('input[name="draft_id"]');
    const csrfToken = document.querySelector('input[name="csrf_token"]').value;

    const threadId = draftsContainer.dataset.threadId;

    function loadDrafts() {
        fetch(`/api/forum/thread/${threadId}/drafts`, {
            headers: {
                'X-CSRF-Token': csrfToken
            }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            draftsList.innerHTML = '';
            data.forEach(draft => {
                const li = document.createElement('li');
                li.innerHTML = `<a href="#" data-draft-id="${draft.id}">${draft.name}</a> <button class="delete-draft" data-draft-id="${draft.id}">X</button>`;
                draftsList.appendChild(li);
            });
        })
        .catch(error => {
            console.error('There has been a problem with your fetch operation:', error);
        });
    }

    saveDraftButton.addEventListener('click', () => {
        const content = replyTextarea.value;
        const draftId = draftIdInput.value;
        const draftName = prompt("Enter a name for your draft:", "Draft from " + new Date().toLocaleString());

        if (draftName === null) {
            return;
        }

        const formData = new FormData();
        formData.append('replytext', content);
        formData.append('draft_id', draftId);
        formData.append('draft_name', draftName);
        formData.append('csrf_token', csrfToken);

        fetch(`/api/forum/thread/${threadId}/drafts`, {
            method: 'POST',
            body: formData
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            if (data.success) {
                draftIdInput.value = data.draft_id;
                loadDrafts();
            }
        })
        .catch(error => {
            console.error('There has been a problem with your fetch operation:', error);
        });
    });

    draftsList.addEventListener('click', e => {
        if (e.target.tagName === 'A') {
            e.preventDefault();
            const draftId = e.target.dataset.draftId;
            fetch(`/api/forum/thread/${threadId}/drafts?id=${draftId}`, {
                headers: {
                    'X-CSRF-Token': csrfToken
                }
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(draft => {
                replyTextarea.value = draft.content;
                draftIdInput.value = draft.id;
            })
            .catch(error => {
                console.error('There has been a problem with your fetch operation:', error);
            });
        } else if (e.target.classList.contains('delete-draft')) {
            const draftId = e.target.dataset.draftId;
            fetch(`/api/forum/thread/${threadId}/drafts?id=${draftId}`, {
                method: 'DELETE',
                headers: {
                    'X-CSRF-Token': csrfToken
                }
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                if (data.success) {
                    loadDrafts();
                }
            })
            .catch(error => {
                console.error('There has been a problem with your fetch operation:', error);
            });
        }
    });

    loadDrafts();
});