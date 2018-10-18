// Local datatable
$("table.table-datatable").each(function (i) {

    var order = [[0, 'asc']];
    var pageLength = 0;
    var paging = false;
    var dom = 't';

    // Limit
    var limit = $(this).attr('data-limit');
    if (limit > 0) {
        paging = true;
        pageLength = Number(limit);
        dom = '<"dt-pagination"p>t<"dt-pagination"p>';
    }

    // Find default column to sort by
    var $column = $(this).find('thead tr th[data-sort]');
    if ($column.length > 0) {

        var index = $column.index();
        var sort = $column.attr('data-sort');

        order = [[index, sort]];
    }

    // Find
    var disabled = [];
    $(this).find('thead tr th[data-disabled]').each(function (i) {
        disabled.push($(this).index());
    });

    // Init
    $(this).DataTable({
        "pageLength": pageLength,
        "order": order,
        "paging": paging,
        "ordering": true,
        "info": false,
        "searching": true,
        "search": {
            "smart": true
        },
        "autoWidth": false,
        "lengthChange": false,
        "stateSave": false,
        "dom": dom,
        "columnDefs": [
            {
                "targets": disabled,
                "orderable": false,
                "searchable": false
            }
        ]
    });

});

// Filter table on search box enter key
$('input#search').keypress(function (e) {
    if (e.which === 13) {
        var table = $('#DataTables_Table_0');
        if (table.length === 1) {
            table.DataTable().search($(this).val()).draw();
        }
    }
});

// Clear search box on escape and reset filter
$('input#search').on('keyup', function (e) {
    if ($(this).val() && e.key === "Escape") {

        $(this).val('');

        var table = $('#DataTables_Table_0');
        if (table.length) {
            table.DataTable().search($(this).val()).draw();
        }
    }
});


// Server side datatable events
$('table.table-datatable2').on('page.dt search.dt', function (e, settings, processing) {

    $(this).fadeTo(500, 0.3);

    if (e.type === 'page') {

        var top = $(this).prev().offset().top - 15;
        $('html, body').animate({scrollTop: top}, 500);
    }

}).on('draw.dt', function (e, settings, processing) {

    $(this).fadeTo(100, 1);

});

$('table.table-datatable').on('page.dt', function (e, settings, processing) {

    var top = $(this).prev().offset().top - 15;
    $('html, body').animate({scrollTop: top}, 200);

});

// Lock icon
var $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

// Server side defaults
var dtDefaultOptions = {
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
    "dom": '<"dt-pagination"p>t<"dt-pagination"p>',
    "language": {
        "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>'
    },
    "drawCallback": function (settings, json) {
        $(".paginate_button > a").on("focus", function () {
            $(this).blur(); // Fixes scrolling to pagination on every click
        });
    }
};

function addDataTablesRow(columnDefs, data, limit, $table) {

    var $row = $('<tr class="fade-green" />');

    for (var i in columnDefs) {
        if (columnDefs.hasOwnProperty(i)) {

            var value = data[i];

            if ('render' in columnDefs[i]) {
                value = columnDefs[i].render(null, null, data);
            }

            var $td = $('<td />').html(value);

            if ('createdCell' in columnDefs[i]) {
                columnDefs[i].createdCell($td[0], null, data, null, null); // todo, this [0] may not be needed
            }

            $row.append($td);
        }
    }

    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();
}
