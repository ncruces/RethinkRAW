let app = Application.currentApplication();
app.includeStandardAdditions = true;

function run() {
    let path = app.pathToResource('RethinkRAW.app').toString();
    app.doShellScript('open -a "' + path + '"');
}

function openDocuments(docs) {
    let path = app.pathToResource('RethinkRAW.app').toString();
    let args = [];
    for (let i = 0; i < docs.length; ++i) {
        args.push('"' + docs[i].toString() + '"');
    }
    app.doShellScript('open -a "' + path + '" --args ' + args.join(' '));
}