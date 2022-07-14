let app = Application.currentApplication();
app.includeStandardAdditions = true;

function run() {
    openDocuments([]);
}

function openDocuments(docs) {
    function escape(arg) {
        return "'" + String(arg).replace(/\'+/g, `'"$&"'`) + "'";
    }

    let args = docs.map(escape).join(' ');
    let path = String(app.pathToResource('RethinkRAW.app'));
    let output = app.doShellScript('ps -xo command=').split('\r');
    let running = output.some(line => line.startsWith(path));
    if (running) {
        app.doShellScript(`${escape(path)}/Contents/MacOS/rethinkraw ` + args);
    } else {
        app.doShellScript(`open -a ${escape(path)} --args ` + args);
    }
}