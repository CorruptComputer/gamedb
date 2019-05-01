if ($('#settings-page').length > 0) {

    // Password
    $('input:password').pwstrength({
        ui: {
            showPopover: true,
            showErrors: true,
        },
        common: {
            usernameField: '#email'
        }
    });

    // Browser alert permissions
    // const $checkbox = $('#browser-alerts');
    //
    // $checkbox.on('click', function () {
    //     if ($(this).is(':checked')) {
    //
    //         Push.Permission.request(
    //             function () {
    //             },
    //             function () {
    //                 alert('You have denied notification access in your browser.');
    //                 $(this).prop("checked", false);
    //             }
    //         );
    //     }
    // });

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (to.attr('href') === '#events') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadEvents();
            }
        }

        // On any tab
        $.each(dataTables, function (index, value) {
            value.fixedHeader.adjust();
        });
    });

    function loadEvents() {

        const table = $('#events table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
            "ajax": function (data, callback, settings) {

                delete data.columns;
                delete data.length;
                delete data.search;

                $.ajax({
                    url: $(this).attr('data-path'),
                    data: data,
                    success: callback,
                    dataType: 'json',
                    cache: $(this).attr('data-cache') !== "false"
                });
            },
            "order": [[0, 'desc']],
            "columnDefs": [
                // Time
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[1] + '" data-livestamp="' + row[0] + '">' + row[1] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
                // Type
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<i class="fas ' + row[7] + '"></i> ' + row[2];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
                // IP
                {
                    "targets": 2,
                    "render": function (data, type, row) {

                        if (row[3] === row[6]) {
                            return '<span class="font-weight-bold" data-toggle="tooltip" data-placement="left" title="Your current IP">' + row[3] + '</span>';
                        }
                        return row[3];
                    },
                    "orderable": false
                },
                // User Agent
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        // return row[4];
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '">' + row[5] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                }
            ]
        }));

        dataTables.push(table);
    }
}
