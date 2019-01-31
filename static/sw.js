addEventListener('fetch', evt => {
    let req = evt.request;
    if (req.method === 'GET' || req.method === 'HEAD') {
        let url = req.url.replace(/\?$/, '');
        if (url !== req.url) {
            return evt.respondWith(new Response(null, { status: 301, headers: { location: url } }));
        }
    }
});