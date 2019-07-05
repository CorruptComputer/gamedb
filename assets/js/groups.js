if ($('#groups-page').length > 0) {

    const $groupsTable = $('table.table-datatable2');

    $('form').on('submit', function (e) {

        $groupsTable.DataTable().draw();
        return false;
    });

    $('#type, #errors').on('change', function (e) {

        $groupsTable.DataTable().draw();
        return false;
    });

    $groupsTable.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            data.search = {};
            data.search.search = $('#search').val();
            data.search.type = $('#type').val();
            data.search.errors = $('#errors').val();

            dtDefaultOptions.ajax(data, callback, settings, $(this));
        },
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
            if (data[7] === 'game' && !$('#type').val()) {
                $(row).addClass('table-primary');
            }
            if (data[9]) {
                $(row).addClass('table-danger');
            }
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img data-src="/assets/img/no-app-image-square.jpg" data-lazy="' + row[3] + '" class="rounded square" data-lazy-alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Headline
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('d-none d-lg-table-cell');
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderable": false,
            },
            // Link
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<a href="' + row[8] + '" target="_blank" rel="nofollow"><i class="fas fa-link" data-target="_blank"></i></a>';
                },
                "orderable": false,
            },
        ]
    }));
}
