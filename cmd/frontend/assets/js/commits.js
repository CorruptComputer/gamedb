if ($('#commits-page').length > 0) {

    const $table = $('table.table');
    let page = null;

    const options = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[3]);
            $(row).attr('data-target', '_blank');
        },
        "columnDefs": [
            // Message
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<a href="' + row[3] + '" target="_blank" class="icon-name"><div class="name">' + row[0] + '</div></a>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('id', rowData[4]);
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Time
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[2] + '" data-livestamp="' + row[1] + '"></span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Hash
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false,
            },
            // Live
            {
                "targets": 3,
                "render": function (data, type, row) {

                    if (page === null) {
                        page = $table.DataTable().page.info().page;
                    }

                    if (row[5] || page > 0) {
                        return '<i class="fas fa-check text-success"></i>';
                    } else {
                        return '<i class="fas fa-times text-danger"></i>';
                    }
                },
                "orderable": false,
            }
        ]
    };

    const dt = $table.gdbTable({tableOptions: options});

    dt.on('draw.dt', function (e, settings) {
        page = null;
    });
}
