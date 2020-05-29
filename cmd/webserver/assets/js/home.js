const $homePage = $('#home-page');

if ($homePage.length > 0) {

    // Sales
    $homePage.on('click', '#sales span[data-sort]:not(.badge-success)', function (e) {
        loadSales($(this).attr('data-sort'));
    });

    // Players
    $homePage.on('click', '#players span[data-sort]:not(.badge-success)', function (e) {
        loadPlayers($(this).attr('data-sort'));
    });

    // Fix top panel heights
    let maxPanelHeight = 0;
    const $panels = $('#panels .card');
    $panels.each(function () {
        if ($(this).height() > maxPanelHeight) {
            maxPanelHeight = $(this).height();
        }
    });
    $panels.css('min-height', maxPanelHeight + 'px');

    // Last load tables
    const config = {rootMargin: '50px 0px 50px 0px', threshold: 0};

    const levelsCallback = function (entries, self) {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                loadPlayers('level');
                self.unobserve(entry.target);
            }
        });
    };
    new IntersectionObserver(levelsCallback, config).observe(document.getElementById("players"));

    const salesCallback = function (entries, self) {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                loadSales('top-rated');
                self.unobserve(entry.target);
            }
        });
    };
    new IntersectionObserver(salesCallback, config).observe(document.getElementById("sales"));

    //
    function loadSales(sort) {

        $.ajax({
            url: '/home/sales/' + sort + '.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                const $allCols = $('#sales span[data-sort]');
                $allCols.removeClass('badge-success');
                $allCols.addClass('cursor-pointer');

                const $thisCol = $('#sales span[data-sort="' + sort + '"]');
                $thisCol.addClass('badge-success');
                $thisCol.removeClass('cursor-pointer');

                $('#sales tbody tr').remove();
                $('#sales .change').html(sort);

                if (isIterable(data)) {

                    const $container = $('#sales tbody');

                    $container.json2html(
                        data,
                        {
                            '<>': 'tr', 'data-app-id': '${id}', 'data-link': '${link}', 'html': [
                                {
                                    '<>': 'td', 'class': 'img', 'html': [
                                        {
                                            '<>': 'div', 'class': 'icon-name', 'html': [
                                                {
                                                    '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${icon}', 'alt': '', 'data-lazy-alt': '${name}'}],
                                                },
                                                {
                                                    '<>': 'div', 'class': 'name', 'html': '${name}'
                                                }
                                            ]
                                        }
                                    ],
                                },
                                {
                                    '<>': 'td', 'html': '${price}', 'class': 'nowrap',
                                },
                                {
                                    '<>': 'td', 'html': '${rating}',
                                },
                                {
                                    '<>': 'td', 'nowrap': 'nowrap', 'class': 'nowrap', 'html': [
                                        {
                                            '<>': 'span', 'data-toggle': 'tooltip', 'data-placement': 'left', 'data-livestamp': '${ends}',
                                        }
                                    ],
                                },
                                {
                                    '<>': 'td', 'html': [
                                        {
                                            '<>': 'a', 'href': '${store_link}', 'target': '_blank', 'rel': 'noopener', 'html': [
                                                {
                                                    '<>': 'i', 'class': 'fas fa-link',
                                                }
                                            ],
                                        },
                                    ]
                                },
                            ]
                        },
                        {
                            prepend: false,
                        }
                    );

                    observeLazyImages($container.find('img[data-lazy]'));
                    highLightOwnedGames($('#sales'));
                }
            },
        });
    }

    function loadPlayers(sort) {

        $.ajax({
            url: '/home/players/' + sort + '.json',
            dataType: 'json',
            cache: true,
            success: function (data, textStatus, jqXHR) {

                const $allCols = $('#players span[data-sort]');
                $allCols.removeClass('badge-success');
                $allCols.addClass('cursor-pointer');

                const $thisCol = $('#players span[data-sort="' + sort + '"]');
                $thisCol.addClass('badge-success');
                $thisCol.removeClass('cursor-pointer');

                $('#players tbody tr').remove();

                if (isIterable(data)) {

                    const $container = $('#players tbody');

                    const tds = [
                        {
                            '<>': 'td', 'class': 'font-weight-bold', 'html': '${rank}'
                        },
                        {
                            '<>': 'td', 'class': 'img', 'html': [
                                {
                                    '<>': 'div', 'class': 'icon-name', 'html': [
                                        {
                                            '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${avatar}', 'alt': '', 'data-lazy-alt': '${name}'}],
                                        },
                                        {
                                            '<>': 'div', 'class': 'name', 'html': '${name}'
                                        }
                                    ]
                                }
                            ]
                        },
                    ];

                    const $change1 = $('#players .change1');
                    const $change2 = $('#players .change2');

                    switch (sort) {
                        case 'level':
                            tds.push({
                                '<>': 'td', 'class': 'img', 'html': '<div class="icon-name"><div class="icon"><div class="${class}"></div></div><div class="name min nowrap">${level}</div></div>',
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${badges}"
                            });
                            $change1.html('Level');
                            $change2.html('Badges');
                            break;
                        case 'games':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${games}"
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${playtime}"
                            });
                            $change1.html('Games');
                            $change2.html('Playtime');
                            break;
                        case 'bans':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${game_bans}"
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${vac_bans}"
                            });
                            $change1.html('Game Bans');
                            $change2.html('VAC Bans');
                            break;
                        case 'profile':
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${friends}"
                            });
                            tds.push({
                                '<>': 'td', 'nowrap': 'nowrap', 'html': "${comments}"
                            });
                            $change1.html('Friends');
                            $change2.html('Comments');
                            break;
                    }

                    $container.json2html(
                        data,
                        {
                            '<>': 'tr', 'data-link': '${link}', 'html': tds,
                        },
                        {
                            prepend: false,
                        }
                    );

                    observeLazyImages($container.find('img[data-lazy]'));
                }
            },
        });
    }
}
