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

    // Recursively find files.
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

    // Filter files by wanted extensions.
    files = files.filter(f => ext(f.name).toUpperCase() in template.Upload.Exts);
    
    let dialog = document.getElementById('progress-dialog');
    let progress = dialog.querySelector('progress');
    progress.removeAttribute('value');
    progress.max = files.length;
    dialog.firstChild.textContent = 'Uploading…';
    dialog.showModal();

    // Upload files.
    try {
        let i = 0;
        for (let f of files) {
            await uploadRequest(f);
            progress.value = ++i;
        }
    } catch (err) {
        alertError('Upload failed', err);
    }

    dialog.close();
    location.reload();
});

function uploadRequest(entry) {
    return new Promise((resolve, reject) => {
        entry.file(file => {
            let data = new FormData();
            data.set('root', template.Upload.Path)
            data.set('path', entry.fullPath)
            data.set('file', file);

            let xhr = new XMLHttpRequest();
            xhr.open('POST', '/upload');
            xhr.onload = () => {
                if (xhr.status < 400) {
                    resolve(xhr.response);
                } else {
                    reject({
                        status: xhr.status,
                        name: xhr.statusText,
                        message: xhr.response,
                    });
                }
            };
            xhr.onerror = () => reject({
                status: xhr.status,
                name: xhr.statusText,
            });
            xhr.setRequestHeader('Accept', 'application/json');
            xhr.responseType = 'json';
            xhr.send(data);
        }, reject);
    });
}

function ext(name) {
    let slash = name.lastIndexOf('/');
    let dot = name.lastIndexOf('.');
    if (dot >= 0 && dot > slash) {
        return name.substring(dot);
    }
    return '';
}

}();