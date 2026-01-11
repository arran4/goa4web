document.addEventListener('DOMContentLoaded', () => {
    // Label filtering
    const labelFilter = document.querySelector('.label-filter');
    if (labelFilter) {
        labelFilter.addEventListener('input', () => {
            const filterValue = labelFilter.value.toLowerCase();
            const items = document.querySelectorAll('.topic-item, .thread');
            items.forEach(item => {
                const labels = item.querySelectorAll('.label');
                let found = false;
                labels.forEach(label => {
                    if (label.textContent.toLowerCase().includes(filterValue)) {
                        found = true;
                    }
                });
                if (found || filterValue === '') {
                    item.style.display = '';
                } else {
                    item.style.display = 'none';
                }
            });
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
