void function () {

if (typeof template.Upload !== 'string') return;

let gallery = document.getElementById('gallery');

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

gallery.addEventListener('dragover', evt => evt.preventDefault());
gallery.addEventListener('drop', async evt => {
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
    for (let e of files) {
        console.log(e.fullPath);
    }
});

}();