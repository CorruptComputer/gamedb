if ($('#admin-websockets-page').length > 0) {

    setTimeout(function () {
        window.location.reload(1);
    }, 5000);
}

if ($('#admin-tasks-page').length > 0) {

    $('#actions tbody tr').on('click', function () {
        if (confirm('Are you sure?')) {
            $.ajax({
                type: 'get',
                url: $(this).attr('data-action'),
                // success: function (data, textStatus, jqXHR) {
                //     toast(true, 'Triggered');
                // },
                error: function (jqXHR, textStatus, errorThrown) {
                    toast(false, errorThrown);
                },
            });
        }
        return false;
    });

    websocketListener('admin', function (e) {

        const data = JSON.parse(e.data);

        const taskID = data.Data.task_id;
        const action = data.Data.action;

        if (taskID && action) {

            const $row = $('tr[data-id="' + taskID + '"]');
            if (action === 'started') {
                $row.addClass('table-warning');
                $row.removeClass('table-danger');
                // toast(true, taskID + ' started', '', 0);
            } else if (action === 'finished') {
                $row.removeClass('table-warning');
                $row.removeClass('table-danger');
                $row.find('.prev').livestamp();
                $row.find('.next').livestamp(new Date(data.Data.time * 1000));
                toast(true, taskID + ' finished', '', 0);
            }
        }
    });
}

if ($('#admin-queues-page').length > 0) {

    const queuesForm = $('form#queues');
    queuesForm.on("submit", function (e) {
        e.preventDefault();
        $.ajax({
            type: 'post',
            url: queuesForm.attr('action'),
            data: $(this).serialize(),
            success: function (data, textStatus, jqXHR) {
                toast(true, 'Queued');
                queuesForm.trigger("reset");
            },
            error: function (jqXHR, textStatus, errorThrown) {
                toast(false, errorThrown);
            },
        });
    });
}

if ($('#admin-users-page').length > 0) {

    const options = {
        "order": [[1, 'desc']],
        "columnDefs": [
            // Email
            {
                'targets': 0,
                'render': function (data, type, row) {
                    if (row[2]) {
                        return '<i class="fas fa-check text-success fa-fw"></i> ' + row[1];
                    } else {
                        return '<i class="fas fa-times text-danger fa-fw"></i> ' + row[1];
                    }
                },
                'orderable': false,
            },
            // Signed Up
            {
                'targets': 1,
                'render': function (data, type, row) {
                    return row[0];
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Logged In
            {
                'targets': 2,
                'render': function (data, type, row) {
                    if (row[5] === '1970-01-01 00:00:00' || row[5] === '0001-01-01 00:00:00') {
                        return '';
                    }
                    return row[5];
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Profile
            {
                'targets': 3,
                'render': function (data, type, row) {
                    if (row[3] > 0) {
                        return '<a href="/players/' + row[3] + '">' + row[3] + '</a>';
                    }
                    return '';
                },
                'orderable': false,
            },
            // level
            {
                'targets': 4,
                'render': function (data, type, row) {
                    if (row[4] > 1) {
                        return row[4];
                    }
                    return '';
                },
                'orderSequence': ['desc'],
            },
        ]
    };

    const $table = $('table.table');
    const dt = $table.gdbTable({
        tableOptions: options,
    });
}

if ($('#admin-consumers-page').length > 0) {

    const options = {
        "order": [[0, 'desc']],
        "drawCallback": function (settings) {
            const api = this.api();
            if (api.order()[0] && api.order()[0][0] === 0) {
                const rows = api.rows({page: 'current'}).nodes();

                let last = null;
                api.rows().every(function (rowIdx, tableLoop, rowLoop) {
                    let group = this.data()[6];
                    if (last !== group) {
                        if (group) {
                            $(rows).eq(rowIdx).before('<tr class="table-success"><td colspan="6">Active</td></tr>');
                        } else {
                            $(rows).eq(rowIdx).before('<tr class="table-danger"><td colspan="6">Expired</td></tr>');
                        }
                        last = group;
                    }
                });
            }
        },
        "columnDefs": [
            // Expires
            {
                'targets': 0,
                'render': function (data, type, row) {
                    return row[0];
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Owner
            {
                'targets': 1,
                'render': function (data, type, row) {
                    return row[1];
                },
                'orderable': false,
            },
            // Environment
            {
                'targets': 2,
                'render': function (data, type, row) {
                    return row[2];
                },
                'orderable': false,
            },
            // Version
            {
                'targets': 3,
                'render': function (data, type, row) {
                    return row[3];
                },
                'orderable': false,
            },
            // Commits
            {
                'targets': 4,
                'render': function (data, type, row) {
                    return parseInt(row[4]).toLocaleString();
                },
                'orderSequence': ['desc', 'asc'],
            },
            // IP
            {
                'targets': 5,
                'render': function (data, type, row) {
                    return row[5];
                },
                'orderable': false,
            },
        ]
    };

    const $table = $('table.table');
    const dt = $table.gdbTable({
        tableOptions: options,
    });
}

if ($('#admin-patreon-page').length > 0) {

    const options = {
        "order": [[0, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).addClass('cursor-pointer');
        },
        "columnDefs": [
            // Date
            {
                'targets': 0,
                'render': function (data, type, row) {
                    return row[0];
                },
                'orderable': false,
            },
            // Event
            {
                'targets': 1,
                'render': function (data, type, row) {
                    return row[1];
                },
                'orderable': false,
            },
            // User
            {
                'targets': 2,
                'render': function (data, type, row) {
                    return row[2]['data']['relationships']['user']['data']['id'];
                },
                'orderable': false,
            },
        ]
    };

    const $table = $('table.table');
    const dt = $table.gdbTable({tableOptions: options});

    $table.on('click', 'tbody tr[role=row]', function () {

            const row = dt.row($(this));

            // noinspection JSUnresolvedFunction
            if (row.child.isShown()) {

                row.child.hide();
                $(this).removeClass('shown');

            } else {

                const rowx = row.data()[2];

                row.child('<pre>' + JSON.stringify(rowx, null, '  ') + '</pre>').show();
                $(this).addClass('shown');
            }
        }
    );
}

if ($('#admin-delays-page').length > 0) {

    const options = {
        "order": [[0, 'desc']],
        "columnDefs": [
            // First Seen
            // {
            //     'targets': 0,
            //     'render': function (data, type, row) {
            //         return row[1];
            //     },
            //     "createdCell": function (td, cellData, rowData, row, col) {
            //         $(td).attr('nowrap', 'nowrap');
            //     },
            // },
            // Last Seen
            {
                'targets': 0,
                'render': function (data, type, row) {
                    return row[2];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderSequence": ["desc", "asc"],
            },
            // Queue
            {
                'targets': 1,
                'render': function (data, type, row) {
                    return row[3];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                'orderable': false,
            },
            // Attempt
            {
                'targets': 2,
                'render': function (data, type, row) {
                    return row[4].toLocaleString();
                },
                "orderSequence": ["desc", "asc"],
            },
            // Message
            {
                'targets': 3,
                'render': function (data, type, row) {
                    return row[5].toLocaleString();
                },
                'orderable': false,
            },
        ]
    };

    $('table.table').gdbTable({
        tableOptions: options,
    });
}
