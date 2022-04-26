let app = Application.currentApplication();
app.includeStandardAdditions = true;

function run() {
    openDocuments([]);
}

function openDocuments(docs) {
    let args = docs.map(doc => `"${doc}"`).join(' ');
    let path = app.pathToResource('RethinkRAW.app').toString();
    let running = app.doShellScript('ps do args').includes('--app=http://localhost:39639');
    if (running) {
        app.doShellScript(`"${path}/Contents/MacOS/rethinkraw" ` + args);
    } else {
        app.doShellScript(`open -a "${path}" --args ` + args);
    }
}