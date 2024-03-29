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
    var end = ndjson.lastIndexOf('\n');
    if (end < 0) return void 0;
    var start = ndjson.lastIndexOf('\n', end - 1);
    return JSON.parse(ndjson.substring(start, end));
};

JSON.parseLines = function (ndjson) {
    return ndjson.trimEnd().split('\n').map(JSON.parse);
};

// Register dialogs with polyfill, add type=cancel buttons.
var dialogs = document.querySelectorAll('dialog');
for (var i = 0; i < dialogs.length; ++i) {
    (function (dialog) {
        dialogPolyfill.registerDialog(dialog);
        dialog.addEventListener('cancel', function () { return dialog.returnValue = '' });
        var buttons = dialog.querySelectorAll('form button[type=cancel]');
        for (var i = 0; i < buttons.length; ++i) {
            (function (button) {
                button.type = 'button';
                button.addEventListener('click', function () {
                    dialog.dispatchEvent(new Event('cancel'));
                    dialog.close();
                });
            })(buttons[i]);
        }
    })(dialogs[i]);
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

function alertError(src, err) {
    console.log(err);
    var name = err && err.name || 'Error';
    var message = err && err.message;
    if (message) {
        var end = /\w$/.test(message) ? '.' : '';
        var sep = message.length > 25 ? '\n' : ' ';
        alert(name + '\n' + src + ' with:' + sep + message + end);
    } else {
        alert(name + '\n' + src + '.');
    }
}