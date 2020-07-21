const $playerPage = $('#player-page');

if ($playerPage.length > 0) {

    // Update link
    $('#update-button').on('click', function (e) {

        e.preventDefault();

        const $link = $(this);

        $('i, svg', $link).addClass('fa-spin');

        $.ajax({
            url: '/players/' + $playerPage.attr('data-id') + '/update.json',
            data: {
                'csrf': $(this).attr('data-csrf'),
            },
            dataType: 'json',
            cache: false,
            success: function (data, textStatus, jqXHR) {

                toast(data.success, data.toast);

                $('i, svg', $link).removeClass('fa-spin');

                $link.contents().last()[0].textContent = ' In Queue';
            },
        });
    });

    // Websockets
    websocketListener('profile', function (e) {

        const data = JSON.parse(e.data);
        if (data.Data['id'].toString() === $playerPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This player has been updated', 0, 'refresh');
        }
    });

    loadAjaxOnObserve({
        "all-games": loadPlayerLibraryTab,
        "recent-games": loadPlayerLibraryStatsTab,
        "details-charts": loadPlayerDetailsTab,
        "badges-table": loadPlayerBadgesTab,
        "friends-table": loadPlayerFriendsTab,
        "groups-table": loadPlayerGroupsTab,
        "wishlist-table": loadPlayerWishlistTab,
        "achievements-table": loadPlayerAchievementsTab,
    });

    //
    function loadPlayerLibraryTab() {

        const options = {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[7]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[7] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    }
                },
                // Price
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Time
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Price/Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Achievements
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        if (row[9] > 0) {
                            return row[8].toLocaleString() + ' / ' + row[9].toLocaleString();
                        }
                        return '-';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        rowData[10] = Math.ceil(rowData[10]);
                        $(td).css('background', 'linear-gradient(to right, rgba(0,0,0,.15) ' + rowData[10] + '%, transparent ' + rowData[10] + '%)');
                        $(td).addClass('thin');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
            ]
        };

        $('#all-games').gdbTable({
            tableOptions: options,
            searchFields: [
                $('#player-games-search'),
            ]
        });
    }

    function loadPlayerLibraryStatsTab() {

        const recentOptions = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[5]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[5] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[1] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    }
                },
                // Price
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[3].toLocaleString();
                    },
                },
                // Time
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    }
                },
            ]
        };

        $('#recent-games').gdbTable({
            tableOptions: recentOptions,
        });
    }

    function loadPlayerDetailsTab() {

        $.ajax({
            type: "GET",
            url: '/players/' + $playerPage.attr('data-id') + '/history.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                const yAxis = {
                    allowDecimals: false,
                    title: {
                        text: ''
                    },
                    labels: {
                        enabled: false
                    },
                };

                Highcharts.chart('history-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: [
                        yAxis,
                        yAxis,
                        yAxis,
                        yAxis,
                        yAxis,
                        yAxis,
                    ],
                    tooltip: {
                        formatter: function () {

                            switch (this.series.name) {
                                case 'Playtime':
                                    return this.y.toLocaleString() + ' minutes played on ' + moment(this.key).format("dddd DD MMM YYYY");
                                default:
                                    return this.y.toLocaleString() + ' ' + this.series.name.toLowerCase() + ' on ' + moment(this.key).format("dddd DD MMM YYYY");
                            }
                        },
                    },
                    series: [
                        {
                            name: 'Level',
                            data: data['max_level'],
                            marker: {symbol: 'circle'},
                            yAxis: 0,
                        },
                        {
                            name: 'Games',
                            data: data['max_games'],
                            marker: {symbol: 'circle'},
                            yAxis: 1,
                        },
                        {
                            name: 'Badges',
                            data: data['max_badges'],
                            marker: {symbol: 'circle'},
                            yAxis: 2,
                        },
                        {
                            name: 'Playtime',
                            data: data['max_playtime'],
                            marker: {symbol: 'circle'},
                            yAxis: 3,
                        },
                        {
                            name: 'Friends',
                            data: data['max_friends'],
                            marker: {symbol: 'circle'},
                            yAxis: 4,
                        },
                        {
                            name: 'Achievements',
                            data: data['max_achievements'],
                            marker: {symbol: 'circle'},
                            yAxis: 5,
                        },
                    ],
                }));

                Highcharts.chart('ranks-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: {
                        allowDecimals: false,
                        title: {
                            text: ''
                        },
                        reversed: true,
                        min: 1,
                    },
                    tooltip: {
                        formatter: function () {
                            return this.series.name + ' rank ' + this.y.toLocaleString() + ' on ' + moment(this.key).format("dddd DD MMM YYYY");
                        },
                    },
                    series: [
                        {
                            name: 'Level',
                            data: data['max_level_rank'],
                            marker: {symbol: 'circle'},
                        },
                        {
                            name: 'Games',
                            data: data['max_games_rank'],
                            marker: {symbol: 'circle'},
                        },
                        {
                            name: 'Badges',
                            data: data['max_badges_rank'],
                            marker: {symbol: 'circle'},
                        },
                        {
                            name: 'Playtime',
                            data: data['max_playtime_rank'],
                            marker: {symbol: 'circle'},
                        },
                        {
                            name: 'Friends',
                            data: data['max_friends_rank'],
                            marker: {symbol: 'circle'},
                        },
                        {
                            name: 'Achievements',
                            data: data['max_achievements_rank'],
                            marker: {symbol: 'circle'},
                        },
                    ],
                }));
            },
        });
    }

    function loadPlayerBadgesTab() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                if (data[0]) {
                    $(row).attr('data-app-id', data[0]);
                }
                $(row).attr('data-link', data[2]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {

                        let name = row[1];
                        if (row[9]) {
                            name += '<span class="badge badge-primary float-right ml-1">Special</span>';
                        }
                        if (row[10]) {
                            name += '<span class="badge badge-warning float-right ml-1">Event</span>';
                        }
                        if (row[4]) {
                            name += '<span class="badge badge-success float-right ml-1">Foil</span>';
                        }

                        return '<a href="' + row[2] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[5] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + name + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Level / XP
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[6].toLocaleString() + ' (' + row[8].toLocaleString() + 'xp)';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Scarcity
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[7].toLocaleString();
                    },
                    "orderSequence": ['asc', 'desc'],
                },
                // Completion Time
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[3].toLocaleString();
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
            ]
        };

        $('#badges-table').gdbTable({
            tableOptions: options,
            searchFields: [
                $('#player-badge-search'),
            ],
        });
    }

    function loadPlayerFriendsTab() {

        const options = {
            "order": [[1, 'desc'], [4, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[1]);
            },
            "columnDefs": [
                // Icon / Friend
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[1] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" data-src="/assets/img/no-player-image.jpg" alt="" data-lazy-alt="' + row[3] + '"></div><div class="name">' + row[3] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Level
                {
                    "targets": 1,
                    "render": function (data, type, row) {

                        if (row[4] === '' || row[4] === '-') {
                            $('#add-missing-friends').removeClass('d-none');
                        }

                        return row[4].toLocaleString();
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Games
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        if (!row[5]) {
                            return '-';
                        } else if (row[6] === 0) {
                            return $lockIcon;
                        } else {
                            return row[6].toLocaleString();
                        }
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Co-op
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        if (row[6] > 0) {
                            return '<a href="/games/coop/' + $playerPage.attr('data-id') + ',' + row[0] + '">Co-op</a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
                // Friend Since
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[7];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Link
                {
                    "targets": 5,
                    "render": function (data, type, row) {
                        if (row[8]) {
                            return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        $('#friends-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadPlayerGroupsTab() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[3]);
                $(row).attr('data-group-id', data[0]);
            },
            "columnDefs": [
                // Group
                {
                    "targets": 0,
                    "render": function (data, type, row) {

                        let badge = '';
                        if (row[7]) {
                            badge = '<span class="badge badge-success float-right">Primary</span>';
                        }

                        return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[4] + '" data-src="/assets/img/no-player-image.jpg" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + badge + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderSequence': ['asc'],
                },
                // Members
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[5].toLocaleString();
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Official
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "orderable": false,
                },
                // Link
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        if (row[8]) {
                            return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        $('#groups-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadPlayerWishlistTab() {

        const options = {
            "order": [[0, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[2]);
            },
            "columnDefs": [
                // Rank
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        if (row[4] === 0) {
                            return '-';
                        }
                        return ordinal(row[4]);
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('font-weight-bold')
                    },
                    'orderSequence': ['asc'],
                },
                // App Name
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<a href="' + row[2] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" data-src="/assets/img/no-player-image.jpg" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderSequence': ['asc'],
                },
                // Release State
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Release Date
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Price
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[7];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc', 'asc'],
                },
            ]
        };

        $('#wishlist-table').gdbTable({
            tableOptions: options,
        });
    }

    function loadPlayerAchievementsTab() {

        const recentOptions = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[0] + '#achievements');
            },
            "columnDefs": [
                // App / Achievement
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[0] + '#achievements" class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[4] + '" alt="" data-lazy-alt="' + row[3] + '"></div><div class="name">' + row[1] + ': ' + row[3] + '<br><small>' + row[5] + '</small></div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Date
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        if (row[6]) {
                            return '<span data-livestamp="' + row[6] + '"></span>';
                        } else {
                            return 'Unknown';
                        }
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc'],
                },
                // Completed
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[7] + '%';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        rowData[7] = Math.ceil(rowData[7]);
                        $(td).css('background', 'linear-gradient(to right, rgba(0,0,0,.15) ' + rowData[7] + '%, transparent ' + rowData[7] + '%)');
                        $(td).addClass('thin');
                    },
                    "orderSequence": ['asc'],
                },
            ]
        };

        $('#achievements-table').gdbTable({
            tableOptions: recentOptions,
        });
    }
}
