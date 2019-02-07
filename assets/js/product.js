const $priceChart = $('#app-page #prices-chart, #package-page #prices-chart');

if ($priceChart.length > 0 && prices) {

    let chart, request;

    function upateChart(code) {

        // Cancel any current requests
        if (request) {
            request.abort();
        }

        // Update row styles
        $('tr[data-code]').removeClass('font-weight-bold').attr('data-link', '');
        $('tr[data-code=' + code + ']').addClass('font-weight-bold').removeAttr('data-link');

        // Show loading screen
        chart.showLoading();

        request = $.ajax({
            type: "GET",
            data: {
                code: code
            },
            url: $priceChart.attr('data-ajax'),
            success: function (data, textStatus, jqXHR) {

                if ('prices' in data) {
                    chart.series[0].setData(data.prices);
                    chart.yAxis[0].update({title: {text: 'Price (' + data.symbol + ')'}});
                    chart.hideLoading();
                }
            },
            dataType: 'json',
            cache: true
        });
    }

    $('#prices table tbody tr[data-code]').on('click', function (e) {

        if ($(this).hasClass('font-weight-bold')) {
            return
        }

        upateChart($(this).attr('data-code'));

    });

    chart = Highcharts.chart('prices-chart', {
        chart: {
            zoomType: 'x'
        },
        title: {
            text: ''
        },
        subtitle: {
            text: ''
        },
        xAxis: {
            title: {
                text: 'Date'
            },
            type: 'datetime',
            labels: {
                step: 1,
                formatter: function () {
                    return moment(this.value).format("Do MMM YY");
                },
            },
        },
        yAxis: {
            title: {
                text: 'Price ($)'
            },
            type: 'linear',
            min: 0,
            allowDecimals: true
        },
        legend: {
            enabled: false
        },
        credits: {
            enabled: false
        },
        series: [
            {
                type: 'line',
                name: 'Price',
                step: 'left',
                color: '#28a745'
            }
        ],
        annotations: [{
            labelOptions: {
                backgroundColor: 'rgba(255,255,255,0.5)',
                verticalAlign: 'top',
                y: 15
            },
            labels: [{
                point: {
                    xAxis: 0,
                    yAxis: 0,
                    x: 27.98,
                    y: 255
                },
                text: 'Arbois'
            }, {
                point: {
                    xAxis: 0,
                    yAxis: 0,
                    x: 45.5,
                    y: 611
                },
                text: 'Montrond'
            }, {
                point: {
                    xAxis: 0,
                    yAxis: 0,
                    x: 63,
                    y: 651
                },
                text: 'Mont-sur-Monnet'
            }]
        }]
    });

    upateChart(user.country);
}
