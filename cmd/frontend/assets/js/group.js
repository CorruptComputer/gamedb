const $groupPage = $('#group-page');

if ($groupPage.length > 0) {

    // Websockets
    websocketListener('group', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data.toString() === $groupPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This group has been updated', 0, 'refresh');
        }
    });

    loadAjaxOnObserve({
        'group-chart': function () {
            loadGroupChart($groupPage);
        },
        'players': loadGroupPlayers,
    });

    function loadGroupPlayers() {

        const options = {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[3]);
                $(row).attr('data-app-id', data[0]);
            },
            "columnDefs": [
                // Flag
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        if (row[6]) {
                            const img = '<img data-lazy="' + row[4] + '" alt="" data-lazy-alt="' + row[6] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[6] + '">';
                            return '<a href="/players?country=' + row[6] + '">' + img + '</a>';
                        }
                        return '';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Icon / Player Name
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<a href="' + row[8] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Avatar 2 / Level
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><div class="' + row[7] + '"></div></div><div class="name min nowrap">' + row[5].toLocaleString() + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderSequence": ["desc", "asc"],
                },
                // Games
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[9].toLocaleString();
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderSequence": ["desc", "asc"],
                },
                // Link
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        if (row[2]) {
                            return '<a href="' + row[2] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        $('#players').gdbTable({
            tableOptions: options,
            searchFields: [
                $('#items-search'),
            ],
        });
    }
}

// keep in global namespace so app page can use it.
function loadGroupChart($page) {

    const $groupChart = $('#group-chart');
    if ($groupChart.length === 0) {
        return
    }

    let plotlines = [];
    if ($groupChart.attr('data-release') !== '') {
        plotlines.push({
            value: parseInt($groupChart.attr('data-release')) * 1000,
            color: 'red',
            width: 1,
            zIndex: 3,
            label: {
                formatter: function () {
                    return 'Steam Release';
                }
            }
        });
    }

    // Load chart
    $.ajax({
        type: "GET",
        url: '/groups/' + $page.attr('data-group-id') + '/members.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            if (data === null) {
                data = [];
            }

            Highcharts.chart('group-chart', $.extend(true, {}, defaultChartOptions, {
                xAxis: {
                    plotLines: plotlines,
                },
                yAxis: {
                    allowDecimals: false,
                    title: {
                        text: ''
                    },
                    labels: {
                        formatter: function () {
                            return this.value.toLocaleString();
                        },
                    },
                },
                tooltip: {
                    formatter: function () {
                        switch (this.series.name) {
                            case 'In Chat':
                                return this.y.toLocaleString() + ' members in chat on ' + moment(this.key).format("dddd DD MMM YYYY");
                            case 'In Game':
                                return this.y.toLocaleString() + ' members in game on ' + moment(this.key).format("dddd DD MMM YYYY");
                            case 'Online':
                                return this.y.toLocaleString() + ' members online ' + moment(this.key).format("dddd DD MMM YYYY");
                            case 'Members':
                                return this.y.toLocaleString() + ' members on ' + moment(this.key).format("dddd DD MMM YYYY");
                        }
                    },
                },
                series: [
                    {
                        name: 'In Chat',
                        color: '#007bff',
                        data: data['max_members_in_chat'],
                        marker: {symbol: 'circle'},
                        visible: false,
                    },
                    {
                        name: 'In Game',
                        color: '#e83e8c',
                        data: data['max_members_in_game'],
                        marker: {symbol: 'circle'},
                        visible: false,
                    },
                    // {
                    //     name: 'Online',
                    //     color: '#ffc107',
                    //     data: data['max_members_online'],
                    //     marker: {symbol: 'circle'},
                    //     visible: false,
                    // },
                    {
                        name: 'Members',
                        color: '#28a745',
                        data: data['max_members_count'],
                        marker: {symbol: 'circle'},
                    },
                ],
            }));
        },
    });
}
