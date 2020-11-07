void function () {

if (!String.prototype.replaceAll) {
    location.replace('/browser.html');
}

window.back = function () {
    if (document.referrer) {
        history.back();
        window.close();
    } else {
        location.replace('/');
    }
};

window.addEventListener('pageshow', function createSocket() {
    let host = location.host.replace('[::1]', 'localhost');
    window.heartbeat = new WebSocket('ws://' + host + '/ws');
    window.heartbeat.onclose = createSocket;
});

window.addEventListener('keydown', function (evt) {
    if (navigator.platform.includes('Mac') && evt.metaKey && !(evt.altKey || evt.ctrlKey) ||
       !navigator.platform.includes('Mac') && evt.ctrlKey && !(evt.altKey || evt.metaKey)) {
        switch (evt.key) {
            case 'n':
            case 't':
                window.open('/', void 0, 'location=off');
                evt.preventDefault();

            case 's':
                if (exportFile && evt.shiftKey) exportFile('dialog');
                if (saveFile && !evt.shiftKey) saveFile();
                evt.preventDefault();
        }
    }
});

document.documentElement.addEventListener('click', function (evt) {
	if (evt.altKey || evt.ctrlKey || evt.metaKey || evt.shiftKey || evt.button !== 0) evt.preventDefault();
});

document.documentElement.addEventListener('contextmenu', function (evt) {
    evt.preventDefault();
});

}()