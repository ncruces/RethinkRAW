let app = Application.currentApplication();
app.includeStandardAdditions = true;

function run() {
    ObjC.import('stdlib');
    let argv = [];
    let args = $.NSProcessInfo.processInfo.arguments;
    try {
        let argc = args.count;
        for (let i = 1; i < argc; ++i) {
            argv.push(ObjC.unwrap(args.objectAtIndex(i)));
        }    
    } finally {
        delete args;
    }
    openDocuments(argv);
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