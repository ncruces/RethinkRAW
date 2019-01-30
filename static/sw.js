addEventListener('fetch', event => {
    let req = event.request;

    if (req.cache === 'only-if-cached' && req.mode !== 'same-origin') {
        return;
    }

    if (req.method === "GET" || req.method === "HEAD") {
        let url = req.url.replace(/\?$/, '');
        if (url !== req.url) {
            return event.respondWith(new Response(null, { status: 301, headers: { location: url } }));
        }
    }
});