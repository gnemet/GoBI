// --- Settings & State ---
const SETTINGS_KEY = 'gobi_report_settings_v1';
let currentSort = [];

// --- Column Management ---

function setupColumnChooser() {
    $('#column-chooser-btn').off('click').on('click', (e) => {
        e.stopPropagation();
        $('#column-chooser-dropdown').toggleClass('hidden');
    });

    $(document).on('click', (e) => {
        if (!$(e.target).closest('.column-chooser-wrapper').length) {
            $('#column-chooser-dropdown').addClass('hidden');
        }
    });

    $(document).off('change', '#column-chooser-dropdown input').on('change', '#column-chooser-dropdown input', function () {
        const colClass = $(this).data('column');
        const visible = $(this).is(':checked');
        $(`.${colClass}`).toggleClass('hidden-col', !visible);
        saveViewSettings();
    });
}

function rebuildColumnChooser() {
    const $chooser = $('#column-chooser-dropdown');
    const $headers = $('.results-table th[data-field]');
    if (!$headers.length) return;

    $chooser.empty();
    $headers.each(function () {
        const field = $(this).data('field');
        const label = $(this).text().trim() || field;
        const colClass = `col-${field}`;
        const isChecked = !$(this).hasClass('hidden-col');

        $chooser.append(`
            <label>
                <input type="checkbox" data-column="${colClass}" ${isChecked ? 'checked' : ''}>
                ${label}
            </label>
        `);
    });
}

// --- Sorting ---

function updateSortIcons() {
    $('.results-table th[data-field]').each(function () {
        const $th = $(this);
        const field = $th.data('field');
        $th.find('.sort-icon, .sort-index').remove();

        const sortIdx = currentSort.findIndex(s => s.field === field);
        if (sortIdx !== -1) {
            const s = currentSort[sortIdx];
            const icon = s.dir === 'ASC' ? 'fa-sort-up' : 'fa-sort-down';
            let html = `<i class="fas ${icon} sort-icon ml-1"></i>`;
            if (currentSort.length > 1) {
                html += `<span class="sort-index">${sortIdx + 1}</span>`;
            }
            $th.append(html);
        }
    });
}

function handleSortClick(e, $th) {
    const field = $th.data('field');
    if (!field) return;

    if (e.ctrlKey) {
        // Multi-sort
        const idx = currentSort.findIndex(s => s.field === field);
        if (idx !== -1) {
            currentSort[idx].dir = currentSort[idx].dir === 'ASC' ? 'DESC' : 'ASC';
        } else {
            currentSort.push({ field: field, dir: 'ASC' });
        }
    } else {
        // Single sort
        if (currentSort.length === 1 && currentSort[0].field === field) {
            currentSort[0].dir = currentSort[0].dir === 'ASC' ? 'DESC' : 'ASC';
        } else {
            currentSort = [{ field: field, dir: 'ASC' }];
        }
    }

    updateSortIcons();

    // Trigger HTMX refresh with sort params
    const $container = $('#results-table-container');
    const url = new URL(window.location.href);
    const params = new URLSearchParams(url.search);

    // Remove existing sort params
    params.delete('sort');
    currentSort.forEach(s => params.append('sort', `${s.field}:${s.dir}`));

    htmx.ajax('GET', `${url.pathname}?${params.toString()}`, {
        target: '#results-table-container',
        swap: 'innerHTML'
    });
}

// --- Drag & Drop ---

let dragSrcEl = null;

function setupDragAndDrop() {
    $(document).off('dragstart', '.results-table th').on('dragstart', '.results-table th', function (e) {
        dragSrcEl = this;
        $(this).addClass('dragging');
    });

    $(document).off('dragover', '.results-table th').on('dragover', '.results-table th', function (e) {
        e.preventDefault();
        $(this).addClass('drag-over');
    });

    $(document).off('dragleave', '.results-table th').on('dragleave', '.results-table th', function () {
        $(this).removeClass('drag-over');
    });

    $(document).off('drop', '.results-table th').on('drop', '.results-table th', function (e) {
        if (dragSrcEl && dragSrcEl !== this) {
            const srcIdx = $(dragSrcEl).index();
            const targetIdx = $(this).index();

            const $table = $('.results-table');
            const $theadRow = $table.find('thead tr');

            // Reorder headers
            if (srcIdx < targetIdx) $(this).after(dragSrcEl);
            else $(this).before(dragSrcEl);

            // Reorder body cells
            $table.find('tbody tr').each(function () {
                const cells = $(this).children('td');
                if (srcIdx < targetIdx) cells.eq(targetIdx).after(cells.eq(srcIdx));
                else cells.eq(targetIdx).before(cells.eq(srcIdx));
            });

            saveViewSettings();
        }
        $(this).removeClass('drag-over');
        return false;
    });

    $(document).off('dragend', '.results-table th').on('dragend', '.results-table th', function () {
        $(this).removeClass('dragging drag-over');
        dragSrcEl = null;
    });
}

// --- Settings ---

function saveViewSettings() {
    const settings = {
        columnOrder: [],
        hiddenColumns: [],
        docked: $('body').hasClass('sidebar-docked')
    };

    $('.results-table th').each(function () {
        const field = $(this).data('field');
        if (field) {
            settings.columnOrder.push(field);
            if ($(this).hasClass('hidden-col')) {
                settings.hiddenColumns.push(field);
            }
        }
    });

    localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings));
}

function applyViewSettings() {
    const raw = localStorage.getItem(SETTINGS_KEY);
    if (!raw) return;
    try {
        const settings = JSON.parse(raw);
        const $table = $('.results-table');
        if (!$table.length) return;

        // Apply Docked State
        if (settings.docked) {
            $('body').addClass('sidebar-docked');
            $('#dock-sidebar-btn').addClass('active');
        }

        // Apply Order
        if (settings.columnOrder && settings.columnOrder.length > 0) {
            const $theadRow = $table.find('thead tr');
            settings.columnOrder.forEach(field => {
                const $th = $theadRow.find(`th[data-field="${field}"]`);
                if ($th.length) $theadRow.append($th);
            });
            // Rebuild rows to match
            $table.find('tbody tr').each(function () {
                const $tr = $(this);
                settings.columnOrder.forEach(field => {
                    const $td = $tr.find(`.col-${field}`);
                    if ($td.length) $tr.append($td);
                });
            });
        }

        // Apply Visibility
        if (settings.hiddenColumns) {
            settings.hiddenColumns.forEach(field => {
                $(`.col-${field}`).addClass('hidden-col');
            });
        }
    } catch (e) {
        console.error("Failed to apply settings:", e);
    }
}

// --- Sidebar/Details ---

function setupSidebar() {
    $(document).off('click', '#close-sidebar-btn, #sidebar-overlay').on('click', '#close-sidebar-btn, #sidebar-overlay', function () {
        $('#detail-sidebar, #sidebar-overlay').removeClass('active');
        $('body').removeClass('sidebar-docked');
        saveViewSettings();
    });

    $(document).off('click', '#dock-sidebar-btn').on('click', '#dock-sidebar-btn', function () {
        $('body').toggleClass('sidebar-docked');
        $(this).toggleClass('active');
        saveViewSettings();
    });

    $(document).off('click', '.sidebar-section-title').on('click', '.sidebar-section-title', function () {
        $(this).toggleClass('is-collapsed');
        $(this).next('.collapsible-content').toggleClass('is-hidden');
    });
}

function showRowDetails(tr) {
    const $tr = $(tr);
    $('.clickable-row').removeClass('selected');
    $tr.addClass('selected');

    const rowData = {};
    const $headers = $('.results-table th[data-field]');

    $tr.find('td').each(function (idx) {
        const $th = $headers.eq(idx);
        if ($th.length) {
            const field = $th.data('field');
            const label = $th.text().trim() || field;
            rowData[label] = $(this).text().trim();
        }
    });

    const $section = $('#record-details-section');
    $section.empty();

    for (const [label, val] of Object.entries(rowData)) {
        if (label === 'Actions') continue;
        $section.append(`
            <div class="detail-row">
                <span class="detail-label">${label}</span>
                <span class="detail-value">${val}</span>
            </div>
        `);
    }

    $('#raw-data-section').text(JSON.stringify(rowData, null, 2));
    $('#detail-sidebar, #sidebar-overlay').addClass('active');
}

// --- Init ---

$(document).ready(() => {
    setupColumnChooser();
    setupDragAndDrop();
    setupSidebar();

    // Initial setup
    rebuildColumnChooser();
    updateSortIcons();
    applyViewSettings();
    $('.results-table th').attr('draggable', true);

    document.body.addEventListener('htmx:afterSwap', (evt) => {
        if (evt.target.id === 'results-table-container') {
            rebuildColumnChooser();
            updateSortIcons();
            applyViewSettings();
            $('.results-table th').attr('draggable', true);
        }
    });

    $(document).off('click', '.results-table th[data-field]').on('click', '.results-table th[data-field]', function (e) {
        if ($(e.target).hasClass('resizer')) return;
        handleSortClick(e, $(this));
    });

    $(document).off('click', '.clickable-row').on('click', '.clickable-row', function (e) {
        if ($(e.target).closest('a, button').length) return;
        showRowDetails(this);
    });
});
