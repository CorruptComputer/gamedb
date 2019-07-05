function observeLazyImages(target) {

    // https://www.sitepoint.com/five-techniques-lazy-load-images-website-performance/

    if (!target) {
        return;
    }

    const config = {
        rootMargin: '0px 0px 50px 0px',
        threshold: 0
    };

    let observer = new IntersectionObserver(function (entries, self) {

        // iterate over each entry
        entries.forEach(entry => {
            if (entry.isIntersecting) {

                const $target = $(entry.target);
                const $alt = $target.attr('data-lazy-alt');

                $target.attr('src', $target.attr('data-lazy'))
                if ($alt) {
                    $target.attr('alt', $alt)
                }

                // the image is now in place, stop watching
                self.unobserve(entry.target);
            }
        });
    }, config);

    const imgs = document.querySelectorAll(target);
    imgs.forEach(img => {
        observer.observe(img);
    });
}

observeLazyImages('img[data-lazy]');

function fixBrokenImages() {

    $('img').one('error', function () {

        const url = $(this).attr('data-src');
        if (url) {
            this.src = url;
        }
    });

    $('img[src=""][data-src]').each(function (i, value) {
        this.src = $(this).attr('data-src');
    });
}
