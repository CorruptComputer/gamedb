const $achievementsPage = $('#achievements-page');

if ($achievementsPage.length > 0) {

    $('table.table').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[8] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc", "asc"],
            },
            // Average
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[6].toLocaleString() + '%';
                },
                "orderSequence": ["desc", "asc"],
            },
            // Price
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Icons
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return json2html.transform(row[7], {'<>': 'img', 'data-lazy': '${i}', 'data-lazy-alt': '${d}', 'class': 'mr-1', 'data-toggle': 'tooltip', 'data-placement': 'top', 'data-lazy-title': '${d}'});
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
        ]
    }));
}
