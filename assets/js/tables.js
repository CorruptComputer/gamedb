// Local
const $dataTables = $('table.table-datatable');
const $dataTables2 = $('table.table-datatable2');
let dataTables = [];

const $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

$dataTables.each(function (i) {

    // Find
    const disabled = [];
    $(this).find('thead tr th[data-disabled]').each(function (i) {
        disabled.push($(this).index());
    });

    // Init
    const dt = $(this).DataTable({
        "pageLength": 100,
        "paging": true,
        "ordering": true,
        "fixedHeader": true,
        "info": false,
        "searching": true,
        "search": {
            "smart": true
        },
        "autoWidth": false,
        "lengthChange": false,
        "stateSave": false,
        "dom": '<"dt-pagination"p>t<"dt-pagination"p>',
        "columnDefs": [
            {
                "targets": disabled,
                "orderable": false
            }
        ],
        "drawCallback": function (settings, json) {

            const api = this.api();
            if (api.page.info().pages <= 1) {
                $(this).parent().find('.dt-pagination').hide();
            }
        },
        "initComplete": function (settings, json) {

            $('table.table-datatable').on('order.dt', function (e, settings, processing) {

                $('#live-badge').trigger('click');

            });
        }
    });

    dataTables.push(dt);

});

// Local search
const $searchField = $('input#search');
$searchField.on('keyup', function (e) {
    $dataTables.DataTable().search($(this).val()).draw();
});

$searchField.on('keyup', function (e) {
    if ($(this).val() && e.key === "Escape") {
        $(this).val('');
        $dataTables.DataTable().search($(this).val()).draw();
        $dataTables2.DataTable().search($(this).val()).draw();
    }
});

// Local events
$dataTables.on('page.dt', function (e, settings, processing) {

    let padding = 15;

    if ($('.fixedHeader-floating').length > 0) {
        padding = padding + 48;
    }

    $('html, body').animate({
        scrollTop: $(this).prev().offset().top - padding
    }, 200);
});

$dataTables.on('draw.dt', function (e, settings) {

    fixBrokenImages();
});

function getPagingType() {

    if (user.userLevel === "0") {
        return 'simple_numbers'
    } else {
        return 'simple_numbers'
    }
}

// Server side
const dtDefaultOptions = {
    "ajax": function (data, callback, settings) {

        delete data.columns;
        delete data.length;
        delete data.search.regex;

        $.ajax({
            url: $(this).attr('data-path'),
            data: data,
            success: callback,
            dataType: 'json',
            cache: $(this).attr('data-cache') !== "false"
        });
    },
    "processing": false,
    "serverSide": true,
    "pageLength": 100,
    "fixedHeader": true,
    "paging": true,
    "ordering": true,
    "info": false,
    "searching": true,
    "autoWidth": false,
    "lengthChange": false,
    "stateSave": false,
    "orderMulti": false,
    "pagingType": getPagingType(),
    "dom": '<"dt-pagination"p>t<"dt-pagination"p>',
    "language": {
        "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>'
    },
    "drawCallback": function (settings, json) {

        const api = this.api();
        if (api.page.info().pages <= 1) {
            $(this).parent().find('.dt-pagination').hide();
        }

        $(".paginate_button > a").on("focus", function () {
            $(this).blur(); // Fixes scrolling to pagination on every click
        });
    },
    "initComplete": function (settings, json) {

        $dataTables2.on('order.dt', function (e, settings, processing) {

            $('#live-badge').trigger('click');

        });
    }
};

// Server side events
$dataTables2.filter(':not(.table-no-fade)').on('page.dt search.dt', function (e, settings, processing) {

    $(this).fadeTo(500, 0.3);

}).on('draw.dt', function (e, settings, processing) {

    $(this).fadeTo(100, 1);
});

$dataTables2.on('page.dt', function (e, settings, processing) {

    $('html, body').animate({
        scrollTop: $(this).prev().offset().top - 15
    }, 200);
});

$dataTables2.on('draw.dt', function (e, settings, processing) {

    highLightOwnedGames();
    fixBrokenImages();

    // Donate link
    if (user.userLevel === '0' || user.userLevel === '') {

        const bold = $('li.paginate_button.page-item.next.disabled').length > 0 ? 'font-weight-bold' : '';
        $('div.dataTables_paginate ul.pagination').append(
            $('<li><small><a href="/donate"><i class="fas fa-heart text-danger"></i> <span class="' + bold + '">See more!</span></a></small></li>').addClass('ml-1 donate')
        );
        // $('li.paginate_button.page-item.next.disabled').unbind().removeClass('disabled').children().html('<i class="fas fa-heart text-danger"></i> See more!').attr('href', '/donate');
    }
});

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
                v.createdCell($td[0], null, data, null, null); // todo, this [0] may not be needed
            }

            $td.find('[data-livestamp]').html('a few seconds ago');

            $row.append($td);
        }
    }


    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();
}

//
$.extend($.fn.dataTableExt.oStdClasses, {
    'sPageEllipsis': 'paginate_ellipsis',
    'sPageNumber': 'paginate_number',
    'sPageNumbers': 'paginate_numbers'
});

$.fn.dataTableExt.oPagination.gamedb = {
    'oDefaults': {
        'iShowPages': 5
    },
    'fnClickHandler': function (e) {
        var fnCallbackDraw = e.data.fnCallbackDraw,
            oSettings = e.data.oSettings,
            sPage = e.data.sPage;

        if ($(this).is('[disabled]')) {
            return false;
        }

        oSettings.oApi._fnPageChange(oSettings, sPage);
        fnCallbackDraw(oSettings);

        return true;
    },
    // fnInit is called once for each instance of pager
    'fnInit': function (oSettings, nPager, fnCallbackDraw) {
        var oClasses = oSettings.oClasses,
            oLang = oSettings.oLanguage.oPaginate,
            that = this;

        var iShowPages = oSettings.oInit.iShowPages || this.oDefaults.iShowPages,
            iShowPagesHalf = Math.floor(iShowPages / 2);

        $.extend(oSettings, {
            _iShowPages: iShowPages,
            _iShowPagesHalf: iShowPagesHalf,
        });

        var oFirst = $('<a class="' + oClasses.sPageButton + ' ' + oClasses.sPageFirst + '">' + oLang.sFirst + '</a>'),
            oPrevious = $('<a class="' + oClasses.sPageButton + ' ' + oClasses.sPagePrevious + '">' + oLang.sPrevious + '</a>'),
            oNumbers = $('<span class="' + oClasses.sPageNumbers + '"></span>'),
            oNext = $('<a class="' + oClasses.sPageButton + ' ' + oClasses.sPageNext + '">' + oLang.sNext + '</a>'),
            oLast = $('<a class="' + oClasses.sPageButton + ' ' + oClasses.sPageLast + '">' + oLang.sLast + '</a>');

        oFirst.click({'fnCallbackDraw': fnCallbackDraw, 'oSettings': oSettings, 'sPage': 'first'}, that.fnClickHandler);
        oPrevious.click({'fnCallbackDraw': fnCallbackDraw, 'oSettings': oSettings, 'sPage': 'previous'}, that.fnClickHandler);
        oNext.click({'fnCallbackDraw': fnCallbackDraw, 'oSettings': oSettings, 'sPage': 'next'}, that.fnClickHandler);
        oLast.click({'fnCallbackDraw': fnCallbackDraw, 'oSettings': oSettings, 'sPage': 'last'}, that.fnClickHandler);

        // Draw
        $(nPager).append(oFirst, oPrevious, oNumbers, oNext, oLast);
    },
    // fnUpdate is only called once while table is rendered
    'fnUpdate': function (oSettings, fnCallbackDraw) {
        var oClasses = oSettings.oClasses,
            that = this;

        var tableWrapper = oSettings.nTableWrapper;

        // Update stateful properties
        this.fnUpdateState(oSettings);

        if (oSettings._iCurrentPage === 1) {
            $('.' + oClasses.sPageFirst, tableWrapper).attr('disabled', true);
            $('.' + oClasses.sPagePrevious, tableWrapper).attr('disabled', true);
        } else {
            $('.' + oClasses.sPageFirst, tableWrapper).removeAttr('disabled');
            $('.' + oClasses.sPagePrevious, tableWrapper).removeAttr('disabled');
        }

        if (oSettings._iTotalPages === 0 || oSettings._iCurrentPage === oSettings._iTotalPages) {
            $('.' + oClasses.sPageNext, tableWrapper).attr('disabled', true);
            $('.' + oClasses.sPageLast, tableWrapper).attr('disabled', true);
        } else {
            $('.' + oClasses.sPageNext, tableWrapper).removeAttr('disabled');
            $('.' + oClasses.sPageLast, tableWrapper).removeAttr('disabled');
        }

        var i, oNumber, oNumbers = $('.' + oClasses.sPageNumbers, tableWrapper);

        // Erase
        oNumbers.html('');

        for (i = oSettings._iFirstPage; i <= oSettings._iLastPage; i++) {
            oNumber = $('<a class="' + oClasses.sPageButton + ' ' + oClasses.sPageNumber + '">' + oSettings.fnFormatNumber(i) + '</a>');

            if (oSettings._iCurrentPage === i) {
                oNumber.attr('active', true).attr('disabled', true);
            } else {
                oNumber.click({'fnCallbackDraw': fnCallbackDraw, 'oSettings': oSettings, 'sPage': i - 1}, that.fnClickHandler);
            }

            // Draw
            oNumbers.append(oNumber);
        }

        // Add ellipses
        if (1 < oSettings._iFirstPage) {
            oNumbers.prepend('<span class="' + oClasses.sPageEllipsis + '">...</span>');
        }

        if (oSettings._iLastPage < oSettings._iTotalPages) {
            oNumbers.append('<span class="' + oClasses.sPageEllipsis + '">...</span>');
        }
    },
    // fnUpdateState used to be part of fnUpdate
    // The reason for moving is so we can access current state info before fnUpdate is called
    'fnUpdateState': function (oSettings) {
        var iCurrentPage = Math.ceil((oSettings._iDisplayStart + 1) / oSettings._iDisplayLength),
            iTotalPages = Math.ceil(oSettings.fnRecordsDisplay() / oSettings._iDisplayLength),
            iFirstPage = iCurrentPage - oSettings._iShowPagesHalf,
            iLastPage = iCurrentPage + oSettings._iShowPagesHalf;

        if (iTotalPages < oSettings._iShowPages) {
            iFirstPage = 1;
            iLastPage = iTotalPages;
        } else if (iFirstPage < 1) {
            iFirstPage = 1;
            iLastPage = oSettings._iShowPages;
        } else if (iLastPage > iTotalPages) {
            iFirstPage = (iTotalPages - oSettings._iShowPages) + 1;
            iLastPage = iTotalPages;
        }

        $.extend(oSettings, {
            _iCurrentPage: iCurrentPage,
            _iTotalPages: iTotalPages,
            _iFirstPage: iFirstPage,
            _iLastPage: iLastPage
        });
    }
};
