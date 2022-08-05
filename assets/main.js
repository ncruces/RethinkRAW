"use strict";

void function () {

if (typeof String.prototype.replaceAll !== 'function') {
    location.replace('/browser.html');
}

document.documentElement.addEventListener('keydown', function (evt) {
    if (navigator.platform.startsWith('Mac') && evt.metaKey && !(evt.altKey || evt.ctrlKey) ||
       !navigator.platform.startsWith('Mac') && evt.ctrlKey && !(evt.altKey || evt.metaKey)) {
        var minimalUI = !window.matchMedia('(display-mode: browser)').matches;

        switch (evt.key) {
            case 'n':
            case 't':
                if (minimalUI) {
                    evt.preventDefault();
                    if (evt.repeat) return;
                    window.open('/', void 0, 'location=no,scrollbars=yes');
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

            case 'p':
                evt.preventDefault();
                if (evt.repeat) return;
                if (window.printFile) {
                    window.printFile();
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

// Newline-delimited JSON.
JSON.parseLast = function (ndjson) {
    let end = ndjson.lastIndexOf('\n');
    if (end < 0) return void 0;
    let start = ndjson.lastIndexOf('\n', end - 1);
    return JSON.parse(ndjson.substring(start, end));
};

JSON.parseLines = function (ndjson) {
    return ndjson.trimEnd().split('\n').map(JSON.parse);
};

// Register dialogs with polyfill, add type=cancel buttons.
for (let d of document.querySelectorAll('dialog')) {
    dialogPolyfill.registerDialog(d);
    d.addEventListener('cancel', function () { return d.returnValue = '' });
    for (let b of d.querySelectorAll('form button[type=cancel]')) {
        b.type = 'button';
        b.addEventListener('click', function () {
            d.dispatchEvent(new Event('cancel'));
            d.close();
        });
    }
}

}();

function back() {
    if (document.referrer) {
        history.back();
        window.close();
    } else {
        location.replace('/');
    }
}

function sleep(ms) {
    return new Promise(function (resolve) {
        return setTimeout(resolve, ms);
    });
}