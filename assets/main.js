"use strict";

void function () {

if (typeof String.prototype.replaceAll !== 'function') {
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
    var host = location.host.replace('[::1]', 'localhost');
    window.heartbeat = new WebSocket('ws://' + host + '/ws');
    window.heartbeat.onclose = createSocket;
});

document.documentElement.addEventListener('keydown', function (evt) {
    if (navigator.platform.includes('Mac') && evt.metaKey && !(evt.altKey || evt.ctrlKey) ||
       !navigator.platform.includes('Mac') && evt.ctrlKey && !(evt.altKey || evt.metaKey)) {
        var minimalUI = !window.matchMedia('(display-mode: browser)').matches;

        switch (evt.key) {
            case 'n':
            case 't':
                if (minimalUI) {
                    evt.preventDefault();
                    if (evt.repeat) return;
                    window.open('/', void 0, 'location=off');
                }
                break;

            case 'o':
                if (minimalUI) {
                    evt.preventDefault();
                    if (evt.repeat) return;
                    location.href = evt.shiftKey ? '/dialog?gallery' : '/dialog?photo';
                }
                break;

            case 's':
                evt.preventDefault();
                if (evt.repeat) return;
                if (evt.shiftKey && window.saveFileAs) {
                    window.saveFileAs();
                } else if (window.saveFile) {
                    window.saveFile();
                }
                break;
        }
    }
});

document.documentElement.addEventListener('click', function (evt) {
    if (evt.altKey || evt.ctrlKey || evt.metaKey || evt.shiftKey || evt.button !== 0) evt.preventDefault();
});

document.documentElement.addEventListener('contextmenu', function (evt) {
    evt.preventDefault();
});

}();