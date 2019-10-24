const $document = $(document);
const $body = $("body");

// Data links
let dataLinkDrag = false;
let dataLinkX = 0;
let dataLinkY = 0;

// On document for elements that are created with JS
$document.on('mousedown', '[data-link]', function (e) {
    dataLinkX = e.screenX;
    dataLinkY = e.screenY;
    dataLinkDrag = false;
});

$document.on('mousemove', '[data-link]', function handler(e) {
    if (!dataLinkDrag && (Math.abs(dataLinkX - e.screenX) > 5 || Math.abs(dataLinkY - e.screenY) > 5)) {
        dataLinkDrag = true;
    }
});

$(document).on('mouseup', '[data-link]', function (e) {

    e.stopPropagation();

    const link = $(this).attr('data-link');
    const target = $(this).attr('data-target');

    if (!link) {
        return true;
    }

    if (dataLinkDrag) {
        return true;
    }

    // Right click
    if (e.which === 3) {
        return true;
    }

    // Middle click
    if (e.ctrlKey || e.shiftKey || e.metaKey || e.which === 2 || target === '_blank') {
        if (!$(e.target).is("a")) {
            window.open(link, '_blank');
        }
        return true;
    }

    window.location.href = link;
    return true;
});

$(document).on('mouseup', '[data-link] a', function (e) {
    e.stopPropagation();
    return true;
});

$('.stop-prop').on('click', function (e) {
    e.stopPropagation();
});

// Auto dropdowns
$('.navbar .dropdown').hover(
    function () {
        $(this).addClass("show").find('.dropdown-menu').addClass("show");
    }, function () {
        $(this).removeClass("show").find('.dropdown-menu').removeClass("show");
    }
).click(function (e) {
    e.stopPropagation();
});

// Tooptips
$body.tooltip({
    selector: '[data-toggle="tooltip"]'
});

//
$('.json').each(function (i, value) {

    const json = $(this).text();

    if (isJson(json)) {
        const jsonObj = JSON.parse(json);
        $(this).text(JSON.stringify(jsonObj, null, '  '));
    }
});

// Tabs
(function ($, window) {
    'use strict';

    $(document).ready(function () {

        // Choose tab from URL
        const hash = window.location.hash;
        if (hash) {

            let fullHash = '';
            hash.split(/[,\-]/).map(function (hash) {

                fullHash = (fullHash === '') ? hash : fullHash + '-' + hash;

                $('.nav-link[href="' + fullHash + '"]').tab('show');
            });
        }

        // Set URL from tab
        $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
            const hash = $(e.target).attr('href');
            if (history.pushState) {
                history.pushState(null, null, hash);
            } else {
                location.hash = hash;
            }
        });
    });

})(jQuery, window);


// Scroll to top link
const $top = $("#top");

$(window).on('scroll', function (e) {

    if ($(window).scrollTop() >= 1000) {
        $top.addClass("show");
    } else {
        $top.removeClass("show");
    }
});

$top.click(function (e) {
    $('html, body').animate({scrollTop: 0}, 500);
});

// Toasts
if (isIterable(user.toasts)) {
    for (const v of user.toasts) {
        toast(v.success, v.message, v.title, v.timeout, v.link);
    }
}

// Fix URLs
$(document).ready(function () {
    const path = $('#app-page, #package-page, #player-page, #bundle-page, #group-page').attr('data-path');
    if (path && path !== window.location.pathname) {
        history.replaceState(null, null, path + window.location.hash);
    }
});

//
const $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

//
function addDataTablesRow(options, data, limit, $table) {

    let $row = $('<tr class="fade-green" />');
    options.createdRow($row[0], data, null);

    if (isIterable(options.columnDefs)) {
        for (const v of options.columnDefs) {

            let value = data[v];

            if ('render' in v) {
                value = v.render(null, null, data);
            }

            const $td = $('<td />').html(value);

            if ('createdCell' in v) {
                v.createdCell($td, null, data, null, null);
            }

            $td.find('[data-livestamp]').html('a few seconds ago');

            $row.append($td);
        }
    }


    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();

    observeLazyImages($row.find('img[data-lazy]'));
}

// Loading icon
(function () {

    // const originalXhr = new window.XMLHttpRequest();
    const originalXhr = $.ajaxSettings.xhr;
    $.ajaxSetup({
        xhr: function () {
            const xhr = originalXhr();
            if (xhr) {

                const $loadingBar = $('#loading');

                xhr.addEventListener('loadstart', function (e) {
                    $loadingBar.fadeTo(100, 1);
                });
                xhr.addEventListener('loadend', function (e) {
                    $loadingBar.fadeTo(100, 0);
                });
                xhr.addEventListener('error', function (e) {
                    console.log('XHR Error', e)
                });
                xhr.addEventListener('abort', function (e) {
                    console.log('XHR Aborted', e)
                });
            }
            return xhr;
        }
    });
})();

function setCookieFlag(key, value) {

    let cookie = Cookies.get('gamedb-session-2');
    if (cookie === undefined || cookie === '') {
        cookie = {};
    } else {
        cookie = JSON.parse(cookie);
    }

    cookie[key] = value;

    Cookies.set('gamedb-session-2', JSON.stringify(cookie));
}

$('.jumbotron button.close').on('click', function (e) {
    $(this).closest('.jumbotron').slideUp();
    setCookieFlag($(this).attr('data-id'), true);
});

const $darkMode = $('#dark-mode');
$darkMode.on('click', '.fa-sun, .fa-moon', function (e) {

    const $sun = $darkMode.find('.fa-sun');
    const $moon = $darkMode.find('.fa-moon');

    if ($sun.hasClass("d-none")) {

        $sun.removeClass('d-none');
        $moon.addClass('d-none');
        setCookieFlag('dark', true);

        $('body').removeClass('dark');

    } else {

        $sun.addClass('d-none');
        $moon.removeClass('d-none');
        $('body').addClass('dark');
        setCookieFlag('dark', false);
    }

    return false;
});
