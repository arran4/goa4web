document.addEventListener('DOMContentLoaded', function() {
    const data = window.grantAddData || {actions: {}, items: {}};
    const rowsContainer = document.getElementById('grant-rows');
    const addRowBtn = document.getElementById('grant-add-row');
    const form = document.getElementById('grant-add-form');
    const searchInput = document.getElementById('grant-subject-search');
    const userIDsInput = document.getElementById('grant-user-ids');
    const roleIDsInput = document.getElementById('grant-role-ids');
    const clearBtn = document.getElementById('grant-clear-search');
    const selectedCount = document.getElementById('grant-selected-count');

    function buildSectionOptions() {
        return Object.keys(data.actions || {}).sort();
    }

    function buildItemOptions(section) {
        const items = data.actions && data.actions[section];
        if (!items) {
            return [];
        }
        return Object.keys(items).sort();
    }

    function getDefinition(section, item) {
        if (!data.actions || !data.actions[section]) {
            return null;
        }
        return data.actions[section][item];
    }

    function createPill(name, target) {
        const pill = document.createElement('span');
        pill.className = 'pill';
        pill.textContent = name;
        pill.draggable = true;
        pill.dataset.default = target;
        pill.addEventListener('dragstart', function() {
            pill.classList.add('dragging');
        });
        pill.addEventListener('dragend', function() {
            pill.classList.remove('dragging');
        });
        return pill;
    }

    function updateActionInput(row) {
        const selected = row.querySelector('.grant-actions-selected');
        const input = row.querySelector('input[name="actions"]');
        const actions = Array.from(selected.querySelectorAll('.pill')).map(p => p.textContent.trim());
        input.value = actions.join(',');
    }

    function setupDropzones(row) {
        const zones = row.querySelectorAll('.grant-actions-available, .grant-actions-selected');
        zones.forEach(function(zone) {
            zone.addEventListener('dragover', function(e) {
                e.preventDefault();
            });
            zone.addEventListener('drop', function(e) {
                e.preventDefault();
                const pill = row.querySelector('.pill.dragging');
                if (!pill || pill.parentNode === zone) {
                    return;
                }
                zone.appendChild(pill);
                updateActionInput(row);
            });
        });
    }

    function updateItems(row) {
        const sectionSelect = row.querySelector('select[name="section"]');
        const itemSelect = row.querySelector('select[name="item"]');
        const section = sectionSelect.value;
        const items = buildItemOptions(section);
        itemSelect.innerHTML = '';
        items.forEach(function(item) {
            const opt = document.createElement('option');
            opt.value = item;
            opt.textContent = item === '' ? 'all' : item;
            itemSelect.appendChild(opt);
        });
        if (items.length === 0) {
            const opt = document.createElement('option');
            opt.value = '';
            opt.textContent = 'all';
            itemSelect.appendChild(opt);
        }
        updateActions(row);
        updateItemOptions(row);
    }

    function updateItemOptions(row) {
        const section = row.querySelector('select[name="section"]').value;
        const item = row.querySelector('select[name="item"]').value;
        const input = row.querySelector('input[name="item_id"]');
        const datalist = row.querySelector('datalist');
        const key = `${section}|${item}`;
        const options = (data.items && data.items[key]) || [];

        datalist.innerHTML = '';
        if (options.length === 0) {
            input.removeAttribute('list');
            input.placeholder = item ? 'Enter item ID' : 'All';
            return;
        }
        input.setAttribute('list', datalist.id);
        options.forEach(function(option) {
            const opt = document.createElement('option');
            opt.value = option.ID;
            opt.textContent = option.Label || option.ID;
            datalist.appendChild(opt);
        });
    }

    function updateActions(row) {
        const section = row.querySelector('select[name="section"]').value;
        const item = row.querySelector('select[name="item"]').value;
        const available = row.querySelector('.grant-actions-available');
        const selected = row.querySelector('.grant-actions-selected');
        const def = getDefinition(section, item);

        available.innerHTML = '';
        selected.innerHTML = '';
        if (!def || !def.Actions) {
            updateActionInput(row);
            return;
        }
        def.Actions.forEach(function(action) {
            available.appendChild(createPill(action, 'available'));
        });
        const input = row.querySelector('input[name="item_id"]');
        if (def.RequireItemID) {
            input.setAttribute('required', 'required');
        } else {
            input.removeAttribute('required');
        }
        updateActionInput(row);
    }

    function createRow(index) {
        const row = document.createElement('div');
        row.className = 'grant-row';
        row.dataset.index = index;
        row.innerHTML = `
            <div class="grant-row-main">
                <label>Section
                    <select name="section"></select>
                </label>
                <label>Item
                    <select name="item"></select>
                </label>
                <label>Item ID
                    <input type="text" name="item_id" placeholder="All">
                    <datalist id="item-options-${index}"></datalist>
                </label>
                <button type="button" class="grant-remove-row">Remove</button>
            </div>
            <div class="grant-row-actions">
                <div>
                    <div class="grant-actions-title">Available actions</div>
                    <div class="grant-actions-available"></div>
                </div>
                <div>
                    <div class="grant-actions-title">Selected actions</div>
                    <div class="grant-actions-selected"></div>
                </div>
            </div>
            <input type="hidden" name="actions" value="">
        `;
        const sectionSelect = row.querySelector('select[name="section"]');
        buildSectionOptions().forEach(function(section) {
            const opt = document.createElement('option');
            opt.value = section;
            opt.textContent = section;
            sectionSelect.appendChild(opt);
        });
        sectionSelect.addEventListener('change', function() {
            updateItems(row);
        });
        row.querySelector('select[name="item"]').addEventListener('change', function() {
            updateActions(row);
            updateItemOptions(row);
        });
        row.querySelector('.grant-remove-row').addEventListener('click', function() {
            row.remove();
        });
        setupDropzones(row);
        updateItems(row);
        return row;
    }

    function addRow() {
        const index = rowsContainer.children.length;
        rowsContainer.appendChild(createRow(index));
    }

    function filterSubjects() {
        if (!searchInput) {
            return;
        }
        const query = searchInput.value.toLowerCase().trim();
        document.querySelectorAll('.grant-subject').forEach(function(subject) {
            const haystack = (subject.dataset.search || '').toLowerCase();
            subject.style.display = haystack.includes(query) ? '' : 'none';
        });
    }

    function updateSelectedCount() {
        if (!selectedCount) {
            return;
        }
        const count = document.querySelectorAll('.grant-subject input[type="checkbox"]:checked').length;
        selectedCount.textContent = String(count);
    }

    function parseIDList(value) {
        return value
            .split(',')
            .map(v => v.trim())
            .filter(v => v !== '');
    }

    function appendDirectIDs(name, values) {
        values.forEach(function(value) {
            const input = document.createElement('input');
            input.type = 'hidden';
            input.name = name;
            input.value = value;
            form.appendChild(input);
        });
    }

    if (addRowBtn) {
        addRowBtn.addEventListener('click', addRow);
    }
    if (searchInput) {
        searchInput.addEventListener('input', filterSubjects);
    }
    if (clearBtn) {
        clearBtn.addEventListener('click', function() {
            if (searchInput) {
                searchInput.value = '';
            }
            if (userIDsInput) {
                userIDsInput.value = '';
            }
            if (roleIDsInput) {
                roleIDsInput.value = '';
            }
            filterSubjects();
        });
    }
    if (form) {
        form.addEventListener('submit', function() {
            rowsContainer.querySelectorAll('.grant-row').forEach(updateActionInput);
            if (userIDsInput) {
                appendDirectIDs('user_id', parseIDList(userIDsInput.value));
            }
            if (roleIDsInput) {
                appendDirectIDs('role_id', parseIDList(roleIDsInput.value));
            }
        });
    }
    document.querySelectorAll('.grant-subject input[type="checkbox"]').forEach(function(box) {
        box.addEventListener('change', updateSelectedCount);
    });
    addRow();
    updateSelectedCount();
});
