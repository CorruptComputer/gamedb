if ($('#sales-page').length > 0) {

    (function ($, document) {
        'use strict';

        // Setup drop downs
        $('select.form-control-chosen').chosen({
            disable_search_threshold: 10,
            allow_single_deselect: true,
            rtl: false,
            max_selected_options: 10
        });

        // Price slider
        const priceElement = $('#price-slider');
        const priceSlider = noUiSlider.create(priceElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100,
            }
        });

        // Score slider
        const scoreElement = $('#score-slider');
        const scoreSlider = noUiSlider.create(scoreElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100
            }
        });

        // Discount slider
        const discountElement = $('#discount-slider');
        const discountSlider = noUiSlider.create(discountElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100
            }
        });

        //
        function updateLabels(e) {

            const prices = priceSlider.get();
            if (prices[0] === prices[1]) {
                $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ')');
            } else {
                $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ' - ' + user.userCurrencySymbol + Math.round(prices[1]) + ')');
            }

            const scores = scoreSlider.get();
            if (scores[0] === scores[1]) {
                $('label#score-label').html('Score (' + Math.round(scores[0]) + '%)');
            } else {
                $('label#score-label').html('Score (' + Math.round(scores[0]) + '% - ' + Math.round(scores[1]) + '%)');
            }

            const discounts = discountSlider.get();
            if (discounts[0] === discounts[1]) {
                $('label#score-label').html('Score (' + Math.round(discounts[0]) + '%)');
            } else {
                $('label#score-label').html('Score (' + Math.round(discounts[0]) + '% - ' + Math.round(discounts[1]) + '%)');
            }
        }

        $(document).ready(updateLabels);

        //
        const options = {
            "order": [[0, 'asc']],
            "pageLength": 100,
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[3]);
                $(row).attr('data-app-id', data[0]);
            },
            "columnDefs": [
                // Icon / Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Price
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        if (!row[4]) {
                            return '-';
                        }
                        return row[4];
                    },
                    "orderSequence": ["desc"],
                },
                // Discount
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Rating
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Date
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[7] + '" data-livestamp="' + row[7] + '"></span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
            ]
        };

        const searchFields = [
            $('#tags-in'),
            $('#tags-out'),
            $('#categories'),
            $('#platforms'),
            priceElement,
            scoreElement,
            discountElement,
        ];

        $('table.table').gdbTable({
            tableOptions: options,
            searchFields: searchFields
        });

    })(jQuery, document);
}
