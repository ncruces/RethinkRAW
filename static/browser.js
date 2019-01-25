void function () {

var old;
try {
    new Function('async()=>{}');
} catch (e) {
    old = e instanceof SyntaxError;
}
if (old) location.replace('/browser-upgrade.html');

}()