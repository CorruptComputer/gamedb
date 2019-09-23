if ($('#admin-page').length > 0) {

    $('#actions a').on('click', function () {
        return confirm('Are you sure?');
    });

    const queuesForm = $('form#queues');
    queuesForm.on("submit", function (e) {
        e.preventDefault();
        $.ajax({
            type: 'post',
            url: queuesForm.attr('action'),
            data: $(this).serialize(),
            success: function (data, textStatus, jqXHR) {
                toast(true, 'Queued');
                queuesForm.trigger("reset");
            },
        });
    });

    websocketListener('admin', function (e) {

        const data = $.parseJSON(e.data);
        toast(true, data.Data.message, '', 0);
    });
}
