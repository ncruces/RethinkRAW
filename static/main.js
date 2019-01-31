void function () {

var old;
try {
    new Function('async()=>{}');
} catch (e) {
    old = e instanceof SyntaxError;
}
if (old) {
    location.replace('/browser.html');
} else if ('serviceWorker' in navigator) {
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