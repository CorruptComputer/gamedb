;(function ($, window, document, user, undefined) {

    "use strict";

    // Create the defaults once
    const pluginName = "gdbTable";
    const defaults = {
        fadeOnLoad: true,
        cache: true,
        searchFields: [],
        tableOptions: {
            "autoWidth": false,
            "dom": '<"dt-pagination"p>t<"dt-pagination"p>r',
            "fixedHeader": true,
            "info": false,
            "language": {
                "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>',
                "paginate": {
                    "next": '<i class="fas fa-chevron-right"></i>',
                    "previous": '<i class="fas fa-chevron-left"></i>',
                },
            },
            "lengthChange": false,
            "ordering": true,
            "pageLength": 100,
            "paging": true,
            "pagingType": 'simple_numbers',
            "searching": true,
            "stateSave": false,
        },
    };

    // The actual plugin constructor
    function Plugin(element, options) {

        if (options == null) {
            options = {}
        }

        if (options.tableOptions == null) {
            options.tableOptions = {};
        }

        options.isAjax = function () {
            return this.tableOptions.columnDefs != null
        }

        if (options.isAjax()) {

            defaults.tableOptions.processing = true;
            defaults.tableOptions.serverSide = true;
            defaults.tableOptions.orderMulti = false;

            const parentSettings = $.extend(true, {}, defaults, options);

            defaults.tableOptions.ajax = function (data, callback, settings) {

                delete data.columns;

                // Add search fields to ajax query
                for (const $field of parentSettings.searchFields) {
                    data.search[$field.attr('name')] = $field.val();
                }

                $.ajax({
                    url: function () {
                        const path = $(element).attr('data-path');
                        if (!path && user.log) {
                            console.log('Table data-path not set');
                        }
                        return path;
                    }(),
                    error: function (jqXHR, textStatus, errorThrown) {

                        data = {
                            "draw": "1",
                            "recordsTotal": "0",
                            "recordsFiltered": "0",
                            "data": [],
                            "limited": false
                        };

                        callback(data, textStatus, null);
                    },
                    data: data,
                    success: callback,
                    dataType: 'json',
                    cache: options.cache,
                });
            }
        } else {

            defaults.tableOptions.search = {
                "smart": true
            };

            defaults.tableOptions.columnDefs = [
                {
                    "orderable": false,
                    "targets": $(element).find('thead tr th[data-disabled]').map(function () {
                        return $(this).index();
                    }).get(),
                }
            ]
        }

        this.settings = $.extend(true, {}, defaults, options);

        // Override table settings with url values
        const params = new URL(window.location).searchParams;
        const page = params.get('page');
        const sort = params.get('sort');
        const order = params.get('order');
        if (page) {
            this.settings.tableOptions.displayStart = (page - 1) * this.settings.tableOptions.pageLength;
        }
        if (sort && order) {
            this.settings.tableOptions.sort = [[sort, order]];
        }

        // Set search field values from url values
        for (const $field of this.settings.searchFields) {
            const name = $field.attr('name');
            if (params.has(name)) {
                $field.val(params.get(name));
            }
        }

        //
        this.element = element;
        this.user = user;
        this._defaults = defaults;
        this._name = pluginName;
        this.init();
    }

    $.extend(Plugin.prototype, {
        init: function () {

            const dt = $(this.element).DataTable(this.settings.tableOptions);
            const parent = this;
            this.dt = dt;

            // On AJAX
            dt.on('xhr.dt', function (e, settings, json, xhr) {
                // Add donate button
                parent.limited = json.limited;
            });

            // On Draw
            dt.on('draw.dt', function (e, settings) {

                // Add donate button
                if (parent.limited) {
                    const bold = $('li.paginate_button.page-item.next.disabled').length > 0 ? 'font-weight-bold' : '';
                    const donate = $('<li class="donate"><small><a href="/donate"><i class="fas fa-heart text-danger"></i> <span class="' + bold + '">See more!</span></a></small></li>');
                    $(parent.element).parent().find('.dt-pagination ul.pagination').append(donate);
                }

                // Hide empty pagination
                const $pagination = $(parent.element).parent().find('.dt-pagination');
                (dt.page.info().pages <= 1)
                    ? $pagination.hide()
                    : $pagination.show()

                // Bold rows
                parent.highlightRows();

                // Lazy load images
                observeLazyImages($(parent.element).find('img[data-lazy]'));

                // Fix broken images
                fixBrokenImages();
            });

            // On page change
            dt.on('page.dt', function (e, settings, processing) {

                // Scroll on pagination click
                let padding = 15;
                if ($('.fixedHeader-floating').length > 0) {
                    padding = padding + 48;
                }
                $('html, body').animate({
                    scrollTop: $(this).prev().offset().top - padding
                }, 200);

                setUrlParam('page', dt.page.info().page + 1);
            });

            // On order
            dt.on('order.dt', function () {
                setUrlParam('sort', dt.order()[0][0]);
                setUrlParam('order', dt.order()[0][1]);
            });

            // Server side table events only
            if (this.settings.isAjax() && this.settings.fadeOnLoad) {

                dt.on('page.dt search.dt', function (e, settings) {
                    $(parent.element).fadeTo(500, 0.3);
                });

                dt.on('draw.dt', function (e, settings) {
                    $(parent.element).fadeTo(100, 1);
                });
            }

            // Fixes scrolling to pagination on every click
            $(this.element).parent().find(".paginate_button > a").one("focus", function () {
                $(this).blur();
            });

            // On tab change
            $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

                // Fixes hidden fixed header tables
                $.each(window.gdbTables, function (index, value) {
                    value.fixedHeader.adjust();
                });

                //
                clearUrlParams();
            });

            // Attach events to search fields
            if (this.settings.isAjax()) {
                for (const $field of this.settings.searchFields) {
                    $field.on('change search', function (e) {

                        dt.draw();

                        const name = $field.attr('name');
                        if (name != null) {
                            setUrlParam(name, $field.val());
                        }

                        return false;
                    });
                }
            } else {
                for (const $field of this.settings.searchFields) {
                    $field.on('change search', function (e) {
                        dt.search($(this).val());
                        dt.draw();
                    });
                }
            }

            // Keep track of tables
            if (window.gdbTables == null) {
                window.gdbTables = [];
            }
            window.gdbTables.push();
        },
        highlightRows: function () {

            if (this.user.isLoggedIn) {
                let games = localStorage.getItem('games');
                if (games != null) {
                    games = JSON.parse(games);
                    if (games != null) {
                        $('[data-app-id]').each(function () {
                            const id = $(this).attr('data-app-id');
                            if (games.indexOf(parseInt(id)) !== -1) {
                                $(this).addClass('font-weight-bold')
                            }
                        });
                    }
                }

                let groups = localStorage.getItem('groups');
                if (groups != null) {
                    groups = JSON.parse(groups);
                    if (groups != null) {
                        $('[data-group-id]').each(function () {
                            const id = $(this).attr('data-group-id');
                            if (groups.indexOf(id) !== -1) {
                                $(this).addClass('font-weight-bold')
                            }
                            const id64 = $(this).attr('data-group-id64');
                            if (groups.indexOf(id64) !== -1) {
                                $(this).addClass('font-weight-bold')
                            }
                        });
                    }
                }
            }
        }
    });

    $.fn[pluginName] = function (options) {
        return new Plugin(this, options).dt;
    };

    $('table.table.table-datatable').gdbTable();

})(jQuery, window, document, user);
