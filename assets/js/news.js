if ($('#news-page').length > 0) {

    const $modal = $('#news-modal');

    // Add hash when clicking row
    $('table.table').on('click', '.article-title', function (e) {
        history.pushState(undefined, undefined, '#' + $(this).closest('tr').attr('data-id'));
        showArt();
    });

    // Remove hash when closing modal
    $modal.on('hidden.bs.modal', function (e) {
        history.pushState("", document.title, window.location.pathname + window.location.search);
        showArt();
    });

    // News modal
    $(window).on('hashchange', showArt);
    $(document).on('draw.dt', showArt);

    function showArt() {

        const hash = window.location.hash.replace('#', '');
        if (hash) {
            $modal.find('.modal-body').html($('tr[data-id=' + hash + ']').find('.d-none').html());
            $modal.modal('show');
        } else {
            $modal.modal('hide');
        }
    }

    // Data tables
    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
        },
        "columnDefs": [
            // Game
            {
                "targets": 0,
                "render": function (data, type, row) {

                    // Icon URL
                    if (row[8]) {
                        row[8] = 'https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/' + row[6] + '/' + row[8] + '.jpg';
                    } else {
                        row[8] = '/assets/img/no-app-image-square.jpg';
                    }

                    return '<img src="' + row[8] + '" class="rounded square" alt="' + row[7] + '" onError="this.onerror=null;this.src=\'/assets/img/no-app-image-square.jpg\';"><span data-app-id="' + row[6] + '">' + row[7] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('data-link', rowData[9]);
                },
                "orderable": false
            },
            // Title
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<div>' + row[1] + '</div><div class="d-none">' + row[5] + '</div>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('article-title');
                },
                "orderable": false
            },
            // Date
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            }
        ]
    }));
}
