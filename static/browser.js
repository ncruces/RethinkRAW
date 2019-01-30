void function () {

var old;
try {
    new Function('async()=>{}');
} catch (e) {
    old = e instanceof SyntaxError;
}
if (old) {
    location.replace('/browser-upgrade.html');
} else if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
}

}()