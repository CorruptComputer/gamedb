const $appPage = $('#app-page');

if ($appPage.length > 0) {

    const $modal = $('#news-modal');

    // Background
    const background = $('.container[data-bg]').attr('data-bg');
    if (background !== '') {
        $('body').css("background-image", 'url(' + background + ')');
    }

    // Fix links
    $('#news a').each(function () {

        const href = $(this).attr('href');
        if (href && !(href.startsWith('http'))) {
            $(this).attr('href', 'http://' + href);
        }
    });

    // Add hash when clicking row
    $('#news table.table').on('click', 'td', function (e) {
        history.pushState(undefined, undefined, '#news,' + $(this).closest('tr').attr('data-id'));
        showArt();
    });

    // Remove hash when closing modal
    $modal.on('hidden.bs.modal', function (e) {
        history.pushState("", document.title, "#news");
        showArt();
    });

    // News modal
    $(window).on('hashchange', showArt);
    $(document).on('draw.dt', showArt);

    // Detials image click
    const $detailsImage = $('#details img');

    $detailsImage.on('click', function () {
        $('.card-header-tabs a[href="#media"]').tab('show');
    });
    $detailsImage.on("error", function () {
        $(this).attr('src', '/assets/img/no-app-image-banner.jpg');
        $(this).hide();
    });

    function showArt() {

        const split = window.location.hash.split(',');

        // If the hash has a news ID
        if (split.length === 2 && (split[0] === 'news' || split[0] === '#news') && split[1]) {

            let $art = $('tr[data-id=' + split[1] + ']').find('.d-none').html();
            $art = $("<div />").html($art).text(); // Decode HTML
            $modal.find('.modal-body').html($art);
            $modal.modal('show');

        } else {
            $modal.modal('hide');
        }
    }

    // Media carousel
    const $carousel1 = $('#carousel1');
    const $carousel2 = $('#carousel2');

    $carousel1.slick({
        waitForAnimate: false,
        arrows: false,
        autoplay: false,
        dots: false,
        asNavFor: $carousel2,
        adaptiveHeight: true,
        lazyLoad: 'ondemand',
    });

    $carousel2.slick({
        waitForAnimate: false,
        arrows: false,
        slidesToShow: 15,
        autoplay: false,
        dots: false,
        variableWidth: true,
        asNavFor: $carousel1,
        focusOnSelect: true,
        centerMode: true,
        infinite: true,
    });

    $carousel1.on('afterChange', function (event, slick, currentSlide) {

        // Stop all videos
        resetVideos();

        // Auto play current video
        const $video = $carousel1.find('div[data-slick-index=' + currentSlide + '] video');
        if ($video.length > 0) {
            $video[0].play();
        }
    });

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (to.attr('href') === '#media') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                $carousel1.slick('setPosition');
                $carousel2.slick('setPosition');
            }
        }
        if (to.attr('href') === '#news') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                loadNews();
            }
        }
        if (to.attr('href') === '#prices') {
            if (!to.attr('loaded')) {
                to.attr('loaded', 1);
                loadPriceChart();
            }
        }

        // On leaving tab
        if (from.attr('href') === '#media') {
            resetVideos();
        }
    });

    $(document).on('keydown', function (e) {
        if ($('a.active[href="#media"]').length > 0) {
            if (e.keyCode === 37) {
                $carousel1.slick('slickPrev');
            }
            if (e.keyCode === 39) {
                $carousel1.slick('slickNext');
            }
        }
    });

    function resetVideos() {
        $('video').each(function (index) {
            $(this)[0].pause();
            $(this)[0].currentTime = 0;
        });
    }

    // Websockets
    websocketListener('app', function (e) {

        const data = $.parseJSON(e.data);
        if (data.Data.toString() === $appPage.attr('data-id')) {
            toast(true, 'Click to refresh', 'This app has been updated', -1, 'refresh');
        }

    });

    // News data table
    function loadNews() {

        console.log('news loading');

        $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-id', data[0]);
            },
            "columnDefs": [
                // Title
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div><i class="fas fa-newspaper"></i> ' + row[1] + '</div><div class="d-none">' + row[5] + '</div>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('article-title');
                    },
                    "orderable": false
                },
                // Author
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[2];
                    },
                    "orderable": false
                },
                // Date
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '" data-livestamp="' + row[3] + '"></span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                }
            ]
        }));
    }
}
