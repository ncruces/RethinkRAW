void function () {

if (!('CSS' in window && CSS.supports('grid-area', 'auto'))) {
    location.replace('/browser.html');
}

if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
}

window.back = function () {
    if (document.referrer) {
        history.back();
        window.close();
    } else {
        location.replace('/');
    }
}

}()