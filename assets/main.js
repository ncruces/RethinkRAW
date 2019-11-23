void function () {

if (!('CSS' in window && CSS.supports('grid-area', 'auto'))) {
    location.replace('/browser.html');
}

window.back = function () {
    if (document.referrer) {
        history.back();
        window.close();
    } else {
        location.replace('/');
    }
}

// RadioNodeList polyfill (Edge)
if (typeof RadioNodeList === 'undefined') {
    Object.defineProperty(HTMLCollection.prototype, 'value', {
        get: function () {
            for (var i = 0; i < this.length; i++) {
                var el = this[i];
                if (el.type === 'radio' && el.checked) return el.value;
            }
        },
        set: function (value) {
            for (var i = 0; i < this.length; i++) {
                var el = this[i];
                if (el.type === 'radio') el.checked = el.value == String(value);
            }
        }
    });
}

}()