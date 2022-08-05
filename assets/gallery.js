void function () {

if (!template.Upload) return;

let drop = document.getElementById('drop-target');

async function walkdir(directory) {
    function readEntries(reader) {
        return new Promise((resolve, reject) => reader.readEntries(resolve, reject));
    }

    async function readAll(reader) {
        let files = [];
        let entries;
        do {
            entries = await readEntries(reader);
            for (let entry of entries) {
                if (entry.isFile) {
                    files.push(entry);
                }
                if (entry.isDirectory) {
                    files.push(...await walkdir(entry))
                }        
            }
        } while (entries.length > 0);
        return files;
    }

    return await readAll(directory.createReader())
}

drop.addEventListener('dragover', evt => evt.preventDefault());
drop.addEventListener('drop', async evt => {
    evt.preventDefault();
    let files = [];
    let directories = [];
    for (let i of evt.dataTransfer.items) {
        let entry = i.webkitGetAsEntry()
        if (entry.isFile) {
            files.push(entry);
        }
        if (entry.isDirectory) {
            directories.push(entry);
        }
    }
    for (let d of directories) {
        files.push(...await walkdir(d));
    }
    for (let f of files) {
        if (ext(f.name).toUpperCase() in template.Upload.Exts) {
            console.log(f.fullPath);
        }
    }
});

function ext(name) {
    let slash = name.lastIndexOf('/');
    let dot = name.lastIndexOf('.');
    if (dot >= 0 && dot > slash) {
        return name.substring(dot);
    }
    return '';
}

}();