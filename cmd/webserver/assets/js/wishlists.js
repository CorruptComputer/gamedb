if ($('#wishlists-page').length > 0) {

    const appsOptions = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Icon / App
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[4] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                }
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3].toLocaleString();
                },
            },
        ]
    };

    $('#apps table.table').gdbTable({tableOptions: appsOptions});

    //
    const tagsOptions = {
        "pageLength": 1000,
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Tag
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<i class="fas fa-tag"></i> ' + row[1];
                },
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3].toLocaleString();
                },
            },
        ]
    };

    $('#tags table.table').gdbTable({tableOptions: tagsOptions});
}
