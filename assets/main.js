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

}()