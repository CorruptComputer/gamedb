const $playerPage = $('#player-page');

if ($playerPage.length > 0) {

    // Add user ID to coop link
    if (user.isLoggedIn) {
        const $coop = $('#coop-link');
        $coop.attr('href', $coop.attr('href') + '&p=' + user.userID);
    }

    // Update link
    $('a[data-update-id]').on('click', function (e) {

        e.preventDefault();

        const $link = $(this);

        $('i', $link).addClass('fa-spin');

        $.ajax({
            url: '/players/' + $(this).attr('data-update-id') + '/ajax/update',
            dataType: 'json',
            cache: false,
            success: function (data, textStatus, jqXHR) {

                toast(data.success, data.toast);

                $('i', $link).removeClass('fa-spin');

                if (data.log) {
                    console.log(data.log);
                }
            },
        });
    });

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (to.attr('href') === '#charts') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadPlayerChart();
            }
        }
        if (to.attr('href') === '#games') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);

                loadPlayerGames();
            }
        }
    });

    // Websockets
    websocketListener('profile', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $playerPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This player has been updated', -1, 'refresh');
        }

    });

    function loadPlayerGames() {

        $('#games table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
                $(row).attr('data-link', data[7]);
            },
            "columnDefs": [
                // Icon / Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<img src="' + row[2] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img').attr('data-app-id', rowData[0]);
                    }
                },
                // Price
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    "orderable": false
                },
                // Time
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    }
                },
                // Price/Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "orderable": false
                }
            ]
        }));
    }

    if (typeof heatMapData !== 'undefined' && heatMapData.length > 0) {

        $('#heatmap').height(120);

        function keyToLabel(key) {
            return '$' + (key * 5) + '-' + ((key * 5) + 5);
        }

        Highcharts.chart('heatmap', {
            chart: {
                type: 'heatmap'
            },
            title: {
                text: ''
            },
            subtitle: {
                text: ''
            },
            xAxis: {
                categories: [''],
                title: {
                    text: ''
                },
                labels: {
                    formatter: function () {
                        return keyToLabel(this.value);
                    }
                }
            },
            yAxis: {
                categories: [''],
                title: {
                    text: ''
                }
            },
            credits: {
                enabled: false
            },
            colorAxis: {
                min: 0,
                minColor: '#FFFFFF',
                maxColor: '#28a745'
            },
            legend: {
                enabled: false
            },
            tooltip: {
                formatter: function () {
                    return this.point.value.toLocaleString() + ' apps cost ' + keyToLabel(this.point.value);
                }
            },
            series: [{
                name: '',
                borderWidth: 0,
                color: '#000',
                data: heatMapData,
                dataLabels: {
                    enabled: false,
                    color: '#000000'
                }
            }]
        });
    }

    function loadPlayerChart() {

        const defaultAppChartOptions = {};

        $.ajax({
            type: "GET",
            url: $('#ranks-chart').attr('data-ajax'),
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                Highcharts.chart('ranks-chart', $.extend(true, {}, defaultAppChartOptions, {
                    chart: {
                        type: 'line'
                    },
                    title: {
                        text: ''
                    },
                    subtitle: {
                        text: ''
                    },
                    credits: {
                        enabled: false
                    },
                    legend: {
                        enabled: false
                    },
                    xAxis: {
                        title: {
                            text: ''
                        },
                        type: 'datetime'
                    },
                    yAxis: {
                        allowDecimals: false,
                        title: {
                            text: ''
                        }
                    },
                    plotOptions: {},
                    tooltip: {
                        formatter: function () {
                            return this.y.toLocaleString() + ' apps released on ' + moment(this.key).format("dddd DD MMM YYYY");
                        },
                    },
                    series: [{
                        color: '#28a745',
                        data: data['mean_player_count']
                    }],
                }));

            },
        });

    }
}
