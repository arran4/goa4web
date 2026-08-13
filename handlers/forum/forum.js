document.addEventListener('DOMContentLoaded', () => {
    // Label filtering
    const labelFilter = document.querySelector('.label-filter');
    if (labelFilter) {
        labelFilter.addEventListener('input', () => {
            const filterValue = labelFilter.value.trim();
            const items = document.querySelectorAll('.topic-item, .thread');

            if (!filterValue) {
                items.forEach(item => item.style.display = '');
                return;
            }

            // Split by " OR " (explicit upper case OR)
            const orGroups = filterValue.split(/\s+OR\s+/).filter(g => g.trim().length > 0);

            items.forEach(item => {
                const labels = Array.from(item.querySelectorAll('.label')).map(el => el.textContent.toLowerCase());
                const participants = Array.from(item.querySelectorAll('.participant')).map(el => el.textContent.toLowerCase());

                // OR logic: if any of the groups match, show the item
                const anyGroupMatches = orGroups.some(group => {
                    // AND logic within each OR group
                    const andTerms = group.trim().split(/\s+/).filter(t => t.length > 0);

                    return andTerms.every(term => {
                        const termLower = term.toLowerCase();
                        if (termLower.startsWith('label:')) {
                            const val = termLower.substring(6);
                            return labels.some(l => l.includes(val));
                        } else if (termLower.startsWith('participant:')) {
                            const val = termLower.substring(12);
                            return participants.some(p => p.includes(val));
                        } else {
                            // Fallback: search both
                            return labels.some(l => l.includes(termLower)) || participants.some(p => p.includes(termLower));
                        }
                    });
                });

                if (anyGroupMatches) {
                    item.style.display = '';
                } else {
                    item.style.display = 'none';
                }
            });
        });

        // Allow clicking on labels/participants to add them to the filter
        document.addEventListener('click', (e) => {
            const targetEl = e.target.closest('.label, .participant');
            if (targetEl) {
                const text = targetEl.textContent.trim();
                const prefix = targetEl.matches('.label') ? 'label:' : 'participant:';
                const token = `${prefix}${text}`;

                if (text) {
                    const currentVal = labelFilter.value.trim();
                    if (currentVal) {
                        // Avoid adding duplicates (robust check)
                        const terms = currentVal.toLowerCase().split(/\s+/);
                        if (!terms.includes(token.toLowerCase())) {
                            labelFilter.value = currentVal + ' ' + token;
                        }
                    } else {
                        labelFilter.value = token;
                    }
                    // Trigger input event to re-filter
                    labelFilter.dispatchEvent(new Event('input'));
                }
            }
        });
    }

    // Sorting
    const sortButtons = document.querySelectorAll('.sort-button');
    sortButtons.forEach(button => {
        button.addEventListener('click', () => {
            const sortType = button.dataset.sort;
            let order = button.dataset.order;
            const list = document.querySelector('.topic-list, .thread-list');
            const items = Array.from(list.children);

            items.sort((a, b) => {
                let valA, valB;

                const camelCaseSortType = sortType.replace(/-(\w)/g, (_, letter) => letter.toUpperCase());
                if (sortType === 'name') {
                    valA = a.dataset.name.toLowerCase();
                    valB = b.dataset.name.toLowerCase();
                } else {
                    valA = parseInt(a.dataset[camelCaseSortType + 'Time'] || a.dataset[camelCaseSortType], 10);
                    valB = parseInt(b.dataset[camelCaseSortType + 'Time'] || b.dataset[camelCaseSortType], 10);
                }

                if (order === 'asc') {
                    if (valA < valB) return -1;
                    if (valA > valB) return 1;
                    return 0;
                } else {
                    if (valA > valB) return -1;
                    if (valA < valB) return 1;
                    return 0;
                }
            });

            // Re-append sorted items
            items.forEach(item => list.appendChild(item));

            // Toggle sort order
            button.dataset.order = order === 'asc' ? 'desc' : 'asc';
        });
    });

    // Message folding
    const FOLD_THRESHOLD = 192; // px; reasonable collapsed height
    const foldableContent = document.querySelectorAll('.foldable');
    foldableContent.forEach(content => {
        const fullHeight = content.scrollHeight;
        if (fullHeight > FOLD_THRESHOLD) {
            content.classList.add('folded');

            const expandButton = document.createElement('button');
            expandButton.textContent = 'Click to expand';
            expandButton.classList.add('expand-button');
            content.parentNode.insertBefore(expandButton, content.nextSibling);

            expandButton.addEventListener('click', () => {
                content.classList.toggle('folded');
                expandButton.textContent = content.classList.contains('folded') ? 'Click to expand' : 'Click to collapse';
            });
        }
    });
});
