document.addEventListener('DOMContentLoaded', () => {
    // Label filtering
    const labelFilter = document.querySelector('.label-filter');
    if (labelFilter) {
        labelFilter.addEventListener('input', () => {
            const filterTerms = labelFilter.value.toLowerCase().split(/\s+/).filter(t => t.length > 0);
            const items = document.querySelectorAll('.topic-item, .thread');
            items.forEach(item => {
                const searchableElements = item.querySelectorAll('.label, .participant');

                let allTermsFound = true;
                if (filterTerms.length > 0) {
                    allTermsFound = filterTerms.every(term => {
                        let termFound = false;
                        searchableElements.forEach(el => {
                            if (el.textContent.toLowerCase().includes(term)) {
                                termFound = true;
                            }
                        });
                        return termFound;
                    });
                }

                if (allTermsFound) {
                    item.style.display = '';
                } else {
                    item.style.display = 'none';
                }
            });
        });

        // Allow clicking on labels/participants to add them to the filter
        document.addEventListener('click', (e) => {
            if (e.target.matches('.label') || e.target.matches('.participant')) {
                const text = e.target.textContent.trim();
                if (text) {
                    const currentVal = labelFilter.value.trim();
                    if (currentVal) {
                        // Avoid adding duplicates
                        const terms = currentVal.split(/\s+/);
                        if (!terms.includes(text)) {
                            labelFilter.value = currentVal + ' ' + text;
                        }
                    } else {
                        labelFilter.value = text;
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
