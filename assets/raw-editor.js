void function () {

let form = document.getElementById('settings');
let save = document.getElementById('save');
let edit = document.getElementById('edit') || {};
let zoom = document.getElementById('zoom');
let white = document.getElementById('white');
let photo = document.getElementById('photo');
let spinner = document.getElementById('spinner');

async function loadSettings() {
    if (form.hidden || !form.querySelector('fieldset').disabled) return;

    let settings;
    try {
        settings = await restRequest('GET', '?settings');
    } catch (e) {
        alertError('Load failed', e);
        spinner.hidden = true;
        return;
    }

    if (settings === '') return;

    let upgraded = false;
    if (settings.process == null || settings.process < 6.7 || settings > 11) {
        if (settings.process) alert('This file was processed with an incompatible version of Camera Raw.\nPrevious edits will not be faithfully reproduced.');
        settings.process = 11;
        upgraded = true;
    }
    if (settings.process < 11 && confirm('This file was processed with an older version of Camera Raw.\nPrevious edits may not be faithfully reproduced.\n\nUpdate to the current Camera Raw process version?')) {
        settings.process = 11;
        upgraded = true;
    }

    if (settings.orientation) form.orientation.value = settings.orientation;
    if (settings.process) form.process.value = settings.process.toFixed(1);
    if (settings.profiles) {
        let group = form.profile.lastElementChild;
        group.prepend(...settings.profiles.map(p => new Option(p.replace(/ v2$/, ''), p)));
    }
    form.toneCurve.value = settings.toneCurve;
    form.lensProfile.checked = settings.lensProfile;
    form.autoLateralCA.checked = settings.autoLateralCA;

    profileChange(form.profile, settings.profile);
    temperatureInput(form.temperature, settings.temperature);
    whiteBalanceChange(form.whiteBalance, settings.whiteBalance);

    let tone = 'Default';
    for (let k of ['exposure', 'contrast', 'highlights', 'shadows', 'whites', 'blacks', 'vibrance', 'saturation']) {
        if (settings[k] !== 0) tone = 'Custom';
        rangeInput(form[k], settings[k]);
    }
    for (let k of ['tint', 'texture', 'clarity', 'dehaze', 'sharpness', 'luminanceNR', 'colorNR']) {
        rangeInput(form[k], settings[k]);
    }

    if (settings.autoTone) tone = 'Auto';
    toneChange(form.tone, tone);

    save.disabled = !upgraded;
    edit.disabled = upgraded;
    for (let n of form.querySelectorAll('fieldset')) {
        n.disabled = false;
    }
    for (let n of form.querySelectorAll('select option[hidden]')) {
        n.remove();
    }

    await sleep();
    try {
        let wb = await restRequest('GET', '?wb');
        if (wb.temperature) {
            let restore = save.disabled;
            whiteBalancePresets['As Shot'] = wb;
            whiteBalanceChange(form.whiteBalance);
            temperatureInput(form.temperature);
            save.disabled = restore;
            edit.disabled = !restore;
        }
    } catch (e) {
    }
}

window.addEventListener('beforeunload', evt => {
    if (!save.disabled) {
        evt.returnValue = 'Leave this page? Changes that you made may not be saved.';
        evt.preventDefault();
    }
});

window.toggleEdit = () => {
    form.hidden = !form.hidden;
    loadSettings();
};

window.valueChange = () => {
    save.disabled = false;
    edit.disabled = true;
    updatePhoto();
};

window.orientationChange = (op) => {
    const table = {
        ccw: [8, 8, 5, 6, 7, 4, 1, 2, 3],
        cw:  [6, 6, 7, 8, 5, 2, 3, 4, 1],
        hz:  [2, 2, 1, 4, 3, 6, 5, 8, 7],
        vt:  [4, 4, 3, 2, 1, 8, 7, 6, 5],
    };

    let orient = table[op][form.orientation.value];
    if (orient === void 0) orient = table[op][0];
    form.orientation.value = orient;

    valueChange();
};

window.profileChange = (e, val) => {
    if (val !== void 0) e.value = val;
    let bw = e.value.includes('B&W') ||
        e.value.includes('Monochrome') ||
        e.value.includes('Monotone') ||
        e.value.includes('ACROS') ||
        e.value.includes('BW');

    for (let n of e.form.querySelectorAll('div.color')) {
        n.classList.toggle('disabled-color', bw);
        disableInputs(n);
    }

    valueChange();
};

const whiteBalancePresets = {
    Daylight:   { temperature: 5500, tint: 10 },
    Cloudy:     { temperature: 6500, tint: 10 },
    Shade:      { temperature: 7500, tint: 10 },
    Tungsten:   { temperature: 2850, tint:  0 },
    Fluorescent:{ temperature: 3800, tint: 20 },
    Flash:      { temperature: 5500, tint:  0 },
}

window.whiteBalanceChange = (e, val) => {
    if (val !== void 0) e.value = val;

    let temp = e.form.temperature;
    let tint = e.form.tint;
    let lock = false;
    if (e.value in whiteBalancePresets) {
        let k = whiteBalancePresets[e.value].temperature;
        let t = whiteBalancePresets[e.value].tint;
        tint[0].value = t;
        tint[1].value = t;
        temp[0].value = k;
        temp[1].value = Math.log(k);
    } else if (e.value !== 'Custom') {
        lock = true;
    }
    for (let n of e.form.querySelectorAll('div.manualWB')) {
        n.classList.toggle('disabled-wb', lock);
        disableInputs(n);
    }

    valueChange();
};

window.toneChange = (e, val) => {
    if (val !== void 0) e.value = val;

    if (e.value === 'Default') {
        for (let i of e.form.querySelectorAll('div.customTone input')) {
            i.value = 0;
            rangeInput(i);
        }
    }
    let auto = e.value === 'Auto';
    for (let n of e.form.querySelectorAll('div.customTone')) {
        n.classList.toggle('disabled-tone', auto);
        disableInputs(n);
    }

    valueChange();
};

window.temperatureInput = (e, val) => {
    if (e.length === 2) e = e[1];
    if (val !== void 0) e.value = Math.log(val);

    let n = Math.exp(Number(e.value));

    let r = n < 4000 ? 50 :
            n < 8000 ? 100 :
            n < 20000 ? 200 :
            n < 40000 ? 500 : 1000;

    e.previousElementSibling.value = Math.round(n / r) * r;
};

window.rangeInput = (e, val) => {
    if (e.length === 2) e = e[1];
    if (val !== void 0) e.value = val;

    let n = Number(e.value);
    let s = formatNumber(e.value, e.step);
    if (n > 0 && e.min < 0 && s !== '0') s = '+' + s;
    e.previousElementSibling.value = s;
};

window.setCustomWhiteBalance = () => form.whiteBalance.value = 'Custom';
window.setCustomTone = () => form.tone.value = 'Custom';

window.saveFileAs = () => exportFile('dialog');

window.saveFile = async () => {
    let query = formQuery();

    let dialog = document.getElementById('progress-dialog');
    let progress = dialog.querySelector('progress');
    progress.removeAttribute('value');
    dialog.firstChild.textContent = 'Saving…';
    dialog.showModal();
    try {
        await restRequest('POST', '?save&' + query, { progress: progress });
        save.disabled = true;
        edit.disabled = false;
    } catch (e) {
        alertError('Save failed', e);
    }
    dialog.close();

    if (template.Path) {
        pingRequest(`/thumb/${encodeURI(template.Path)}`);
    } else for (let photo of template.Photos) {
        pingRequest(`/thumb/${encodeURI(photo.Path)}`);
    }
};

window.exportFile = async (state) => {
    if (state === 'dialog') {
        exportChange(document.getElementById('export-form'));
        let dialog = document.getElementById('export-dialog');
        dialog.onclose = () => dialog.returnValue && exportFile(dialog.returnValue);
        dialog.showModal();
        return;
    }

    let query = formQuery();
    if (state === 'export') exportQuery(query);

    let dialog = document.getElementById('progress-dialog');
    let progress = dialog.querySelector('progress');
    progress.removeAttribute('value');
    dialog.firstChild.textContent = 'Exporting…';
    dialog.showModal();
    try {
        await restRequest('POST', '?export&' + query, { progress: progress });
    } catch (e) {
        alertError('Export failed', e);
    }
    dialog.close();
};

window.printFile = () => {
    if (!photo) return;

    let dialog = document.getElementById('progress-dialog');
    let progress = dialog.querySelector('progress');
    progress.removeAttribute('value');
    dialog.firstChild.textContent = 'Printing…';
    dialog.showModal();

    let frame = document.createElement("iframe");
    frame.style.display = "none";
    frame.onload = load;
    frame.src = '?print&' + formQuery().toString();
    document.body.appendChild(frame);

    function load() {
        let win = this.contentWindow;
        win.onbeforeunload = done;
        win.onafterprint = done;
        win.print();
    }

    function done() {
        document.body.removeChild(frame);
        dialog.close();
    }
};

window.toggleZoom = evt => {
    if (photo.style.cursor === 'crosshair') toggleWhite();
    let zoomed = photo.style.cursor === 'zoom-out';
    zoomed = !zoomed;

    photo.style.cursor = zoomed ? 'zoom-out' : 'unset';
    zoom.classList.toggle('pushed', zoomed);
    if (evt && evt.detail) zoom.blur();
    if (zoomed) updatePhoto();
    updateZoom();
};

window.toggleWhite = evt => {
    if (photo.style.cursor === 'zoom-out') toggleZoom();
    let picking = photo.style.cursor === 'crosshair';
    picking = !picking;

    photo.style.cursor = picking ? 'crosshair' : 'unset';
    white.classList.toggle('pushed', picking);
    if (evt && evt.detail) white.blur();
};

window.showMeta = async () => {
    let html = await htmlRequest('GET', '?meta');
    let dialog = document.getElementById('meta-dialog');
    dialog.onclick = () => dialog.close();
    dialog.innerHTML = html;
    dialog.showModal();
};

window.exportChange = (e) => {
    let form = e.tagName === 'FORM' ? e : e.form;

    document.getElementById('export-dng').hidden = form.format.value !== 'DNG';
    document.getElementById('export-jpeg').hidden = form.format.value !== 'JPEG';

    // density unit changed?
    let newden = form.denunit.value;
    let oldden = form.denunit.previousValue;
    if (oldden && oldden !== newden) {
        let min, max, val;
        if (newden === 'ppi') {
            min = 72;
            max = 600;
            val = (form.density.value * 2.5) || 300;
        } else {
            min = 28;
            max = 240;
            val = (form.density.value / 2.5) || 120;
        }
        form.density.min = min;
        form.density.max = max;
        if (val < min) val = min;
        if (val > max) val = max;
        form.density.value = Math.round(val);
    }
    form.denunit.previousValue = newden;

    // dimension unit changed?
    let newdim = form.dimunit.value;
    let olddim = form.dimunit.previousValue;
    if (olddim && olddim !== newdim) {
        let mul = 1;
        let ppi = Number(form.density.value) || 300;
        if (newden !== 'ppi') ppi *= 2.5;
        if (olddim === 'in') mul = ppi;
        if (olddim === 'cm') mul = ppi / 2.5;
        if (newdim === 'in') mul /= ppi;
        if (newdim === 'cm') mul /= ppi / 2.5;

        let min, max, step;
        switch (newdim) {
            case 'in':
                min = 1;
                max = 40;
                step = 0.01;
                break;
            case 'cm':
                min = 2.5;
                max = 100;
                step = 0.01;
                break;
            default:
                min = 80;
                max = 5120;
                step = 1;
                break;
        }

        for (let k of ['long', 'short', 'width', 'height']) {
            let e = form[k];
            let x = e.value * mul;

            e.min = min;
            e.max = max;
            e.step = step;
            if (x) {
                if (x < min) x = min;
                if (x > max) x = max;
                e.value = x;
            }
        }
    }
    form.dimunit.previousValue = newdim;

    // disable/hide/format/require
    let resample = form.resample.checked;
    let dims = form.fit.value === 'dims';
    let mpix = form.fit.value === 'mpix';
    let dens = form.dimunit.value !== 'px' && !mpix;

    for (let k of ['quality', 'fit', 'long', 'short', 'width', 'height', 'dimunit', 'density', 'denunit', 'mpixels']) {
        form[k].disabled = !resample;
    }

    if (resample) {
        form.density.disabled = !dens;
        form.denunit.disabled = !dens;
        form.dimunit.disabled = mpix;
        form.mpixels.disabled = !mpix;

        if (mpix) {
            form.long.disabled = true;
            form.short.disabled = true;
            form.width.disabled = true;
            form.height.disabled = true;
        } else {
            form.long.disabled = !dims;
            form.short.disabled = !dims;
            form.width.disabled = dims;
            form.height.disabled = dims;
        }
    }
    if (!mpix) {
        form.long.hidden = !dims;
        form.short.hidden = !dims;
        form.width.hidden = dims;
        form.height.hidden = dims;
    }

    formatElement(form.long);
    formatElement(form.short);
    formatElement(form.width);
    formatElement(form.height);
    formatElement(form.mpixels);
    form.long.required = form.short.value == '';
    form.short.required = form.long.value == '';
    form.width.required = form.height.value == '';
    form.height.required = form.width.value == '';
    form.density.required = true;
    form.mpixels.required = true;

    function formatElement(e) { if (e.value !== '') e.value = formatNumber(e.value, e.step); }
};

function disableInputs(n) {
    let disabled = n.className.includes('disabled');
    for (let i of n.querySelectorAll('input')) {
        i.disabled = disabled;
    }
}

function alertError(src, err) {
    console.log(err);
    let name = err && err.name || 'Error';
    let message = err && err.message;
    if (message) {
        let end = /\w$/.test(message) ? '.' : '';
        let sep = message.length > 25 ? '\n' : ' ';
        alert(name + '\n' + src + ' with:' + sep + message + end);
    } else {
        alert(name + '\n' + src + '.');
    }
}

let updatePhoto = function () {
    if (!photo) return () => {};
    let loading, query, size;

    function calcSize() {
        if (photo.style.cursor === 'zoom-out') return Infinity;
        return Math.ceil(Math.max(photo.width, photo.height) * devicePixelRatio);
    }

    function load() {
        if (loading) return;
        let newSize = calcSize();
        let newQuery = formQuery().toString();
        if (size >= newSize && query === newQuery) {
            spinner.hidden = true;
            loading = false;
            return;
        }
        spinner.hidden = false;
        loading = true;

        size = newSize;
        query = newQuery;
        photo.addEventListener('load', loaded, { passive: true, once: true });
        photo.addEventListener('error', loaded, { passive: true, once: true });
        photo.src = `?preview${Number.isSafeInteger(size) ? '=' + size : ''}&` + query;
    }

    function loaded() {
        photo.removeEventListener('load', loaded);
        photo.removeEventListener('error', loaded);
        spinner.hidden = true;
        loading = false;
        updateZoom();
        load();
    }

    let loadTimer;
    function loadDelayed(ms) {
        clearTimeout(loadTimer);
        loadTimer = setTimeout(load, ms);
    }

    window.addEventListener('resize', () => loadDelayed(500), { passive: true });
    return loadDelayed;
}();

function formQuery(query) {
    if (query === void 0) query = new URLSearchParams();
    if (form.hidden) return query;

    for (let k of ['orientation', 'process', 'profile', 'whiteBalance', 'toneCurve']) {
        if (form[k].value) query.set(k, form[k].value);
    }
    if (form.whiteBalance.value === 'Custom') {
        query.set('temperature', form.temperature[0].value);
        query.set('tint', form.tint[0].value);
    }
    if (form.tone.value === 'Auto') query.set('autoTone', '1');
    else for (let k of ['exposure', 'contrast', 'highlights', 'shadows', 'whites', 'blacks', 'vibrance', 'saturation']) {
        if (form[k][0].value == 0) continue;
        query.set(k, form[k][0].value);
    }
    for (let k of ['texture', 'clarity', 'dehaze', 'sharpness', 'luminanceNR', 'colorNR']) {
        if (form[k][0].value == 0) continue;
        query.set(k, form[k][0].value);
    }
    for (let k of ['lensProfile', 'autoLateralCA']) {
        if (form[k].checked) query.set(k, '1');
    }

    return query;
}

function exportQuery(query) {
    if (query === void 0) query = new URLSearchParams();

    let form = document.getElementById('export-form');
    if (form.format.value === 'DNG') {
        query.set('dng', '1');
        query.set('preview', form.preview.value);
        for (let k of ['lossy', 'embed']) {
            if (form[k].checked) query.set(k, '1');
        }
    } else if (form.resample.checked) {
        query.set('resample', '1');
        for (let k of ['quality', 'fit', 'long', 'short', 'width', 'height', 'dimunit', 'density', 'denunit', 'mpixels']) {
            if (form[k].value == 0) continue;
            query.set(k, form[k].value);
        }
    }

    return query;
}

function restRequest(method, url, { body, progress } = {}) {
    return new Promise((resolve, reject) => {
        let xhr = new XMLHttpRequest();
        xhr.open(method, url);
        xhr.onreadystatechange = () => {
            if (xhr.readyState === xhr.HEADERS_RECEIVED) {
                if (xhr.getResponseHeader('Content-Disposition') &&
                    xhr.getResponseHeader('Content-Disposition').startsWith('attachment')) {
                    xhr.responseType = 'blob';
                    return;
                }
                if (xhr.getResponseHeader('Content-Type') === 'application/json') {
                    xhr.responseType = 'json';
                    return;
                }
            }
        };
        xhr.onload = () => {
            if (xhr.responseType === 'blob') {
                if (xhr.status < 400) {
                    let a = document.createElement('a');
                    let disposition = xhr.getResponseHeader('Content-Disposition');
                    if (disposition) {
                        let match;
                        if (match = disposition.match(/\bfilename=([^,;]+)/)) a.download = match[1];
                        if (match = disposition.match(/\bfilename="([^"\\]+)"/)) a.download = match[1];
                        if (match = disposition.match(/\bfilename\*=UTF-8''([^,;]+)/)) a.download = decodeURIComponent(match[1]);
                    }
                    a.href = URL.createObjectURL(xhr.response);
                    a.click();
                    resolve();
                    URL.revokeObjectURL(xhr.response);
                } else {
                    var reader = new FileReader();
                    reader.onload = () => {
                        reject({
                            status: xhr.status,
                            name: xhr.statusText,
                            message: JSON.parse(reader.result),
                        });
                    };
                    reader.readAsText(xhr.response);
                }
            } else {
                if (xhr.status === 207 && xhr.getResponseHeader('Content-Type') === 'application/x-ndjson') {
                    let count = 0;
                    let lines = JSON.parseLines(xhr.responseText);
                    for (let status of lines) {
                        if (status.code >= 400) count++;
                    }
                    if (count === 0) {
                        resolve(xhr.response);
                    } else {
                        reject({
                            status: xhr.status,
                            name: xhr.statusText,
                            message: `${count} of ${lines.length} operations failed.`,
                        });
                    }
                } else if (xhr.status < 400) {
                    resolve(xhr.response);
                } else {
                    reject({
                        status: xhr.status,
                        name: xhr.statusText,
                        message: xhr.response,
                    });
                }
            }
        };
        xhr.onerror = () => reject({
            status: xhr.status,
            name: xhr.statusText,
        });
        if (progress !== void 0) {
            xhr.onprogress = evt => {
                if (xhr.status === 207 && xhr.getResponseHeader('Content-Type') === 'application/x-ndjson') {
                    let last = JSON.parseLast(xhr.responseText);
                    if (isFinite(last.done) && isFinite(last.total)) {
                        progress.value = last.done;
                        progress.max = last.total;
                    }
                    return;
                }
                if (evt.lengthComputable) {
                    progress.value = evt.loaded;
                    progress.max = evt.total;
                    return;
                }
            };
        }
        if (body !== void 0) {
            xhr.setRequestHeader('Content-Type', 'application/json');
            body = JSON.stringify(body);
        }
        xhr.setRequestHeader('Accept', 'application/json');
        xhr.send(body);
    });
}

function htmlRequest(method, url) {
    return new Promise((resolve, reject) => {
        let xhr = new XMLHttpRequest();
        xhr.open(method, url);
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
        xhr.setRequestHeader('Accept', 'text/html');
        xhr.send();
    });
}

function pingRequest(url) {
    let xhr = new XMLHttpRequest();
    xhr.open('HEAD', url);
    xhr.setRequestHeader('Cache-Control', 'max-age=0');
    xhr.send();
}

function formatNumber(val, step) {
    step = Number(step);
    if (!Number.isFinite(step)) return val.toString();

    let fmt = step.toString();
    if (fmt.includes('e')) return val.toString();

    let n = Number(val);
    let i = fmt.indexOf('.');
    if (i < 0) return n.toFixed(0);
    return n.toFixed(fmt.length - i - 1);
}

{
    if (!navigator.platform.includes('Mac')) {
        for (let n of document.querySelectorAll('.alt-off')) n.title = n.title.replace('⌥', 'alt');
    }
    function toggleAlt(evt) {
        let key = evt.altKey && !(evt.ctrlKey || evt.metaKey || evt.shiftKey);
        for (let n of document.querySelectorAll('.alt-off')) n.hidden = key;
        for (let n of document.querySelectorAll('.alt-on')) n.hidden = !key;
    }
    window.addEventListener('keydown', toggleAlt);
    window.addEventListener('keyup', toggleAlt);
    toggleAlt({});

    for (let n of document.querySelectorAll('fieldset legend')) {
        n.addEventListener('click', () => {
            for (let c of n.parentElement.children) {
                if (c !== n) c.hidden = !c.hidden;
            }
        });
        n.title = n.title.replace('⌥', 'alt');
    }
}

if (photo) {
    let mousePos;

    function updateZoom(evt) {
        if (evt) mousePos = evt.type === 'mouseleave' ? null : {x: evt.offsetX, y: evt.offsetY};
        if (mousePos && photo.style.cursor === 'zoom-out') {
            let width = photo.naturalWidth / photo.width / devicePixelRatio;
            let height = photo.naturalHeight / photo.height / devicePixelRatio;
            photo.style.transform = `scale(${Math.max(1.5, width, height)})`;
            photo.style.transformOrigin = `${mousePos.x}px ${mousePos.y}px`;
        } else {
            photo.style.transform = 'unset';
        }
    }

    window.addEventListener('keydown', evt => {
        if (evt.key === 'z' && photo.style.cursor !== 'zoom-out') {
            photo.style.cursor = 'zoom-in';
            zoom.classList.add('pushed');
            white.classList.remove('pushed');
        }
        if (evt.key === 'w' && photo.style.cursor !== 'zoom-out') {
            photo.style.cursor = 'crosshair';
            white.classList.add('pushed');
            zoom.classList.remove('pushed');
        }
    });
    window.addEventListener('keyup', evt => {
        if (evt.key === 'z' && photo.style.cursor === 'zoom-in') {
            photo.style.cursor = 'unset';
            zoom.classList.remove('pushed');
        }
        if (evt.key === 'w' && photo.style.cursor === 'crosshair') {
            photo.style.cursor = 'unset';
            white.classList.remove('pushed');
        }
    });
    photo.addEventListener('mouseleave', updateZoom, { passive: true });
    photo.addEventListener('mousemove', updateZoom, { passive: true });

    photo.addEventListener('click', async (evt) => {
        switch (photo.style.cursor) {
            case 'zoom-in':
            case 'zoom-out':
                toggleZoom();
                break;

            case 'crosshair': {
                let wr = photo.width / photo.naturalWidth;
                let hr = photo.height / photo.naturalHeight;
                let ratio = Math.min(wr, hr);
                let width = photo.naturalWidth * ratio;
                let height = photo.naturalHeight * ratio;

                let posx = (evt.offsetX - (photo.width - width) / 2) / width;
                let posy = (evt.offsetY - (photo.height - height) / 2) / height;
                if (posx < 0 || posy < 0 || posx > 1 || posy > 1) break;

                switch (form.orientation.value) {
                    case '2': [posx, posy] = [1 - posx, posy/**/]; break;
                    case '3': [posx, posy] = [1 - posx, 1 - posy]; break;
                    case '4': [posx, posy] = [posx/**/, 1 - posy]; break;
                    case '5': [posx, posy] = [posy/**/, posx/**/]; break;
                    case '6': [posx, posy] = [posy/**/, 1 - posx]; break;
                    case '7': [posx, posy] = [1 - posy, 1 - posx]; break;
                    case '8': [posx, posy] = [1 - posy, posx/**/]; break;
                }

                let wb;
                try {
                    spinner.hidden = false;
                    wb = await restRequest('GET', `?wb=${posx},${posy}`);
                } catch (e) {
                    alertError('White balance failed', e);
                } finally {
                    spinner.hidden = true;
                }
                if (wb.temperature) { 
                    rangeInput(form.tint, wb.tint);
                    temperatureInput(form.temperature, wb.temperature);
                    whiteBalanceChange(form.whiteBalance, 'Custom');
                } else {
                    alertError('White balance failed');
                }
                break;
            }
        }
    });

    loadSettings();
}

JSON.parseLast = (ndjson) => {
    let end = ndjson.lastIndexOf('\n');
    if (end < 0) return void 0;
    let start = ndjson.lastIndexOf('\n', end - 1);
    return JSON.parse(ndjson.substring(start, end));
};

JSON.parseLines = (ndjson) => {
    return ndjson.trimEnd().split('\n').map(JSON.parse);
};

// dialog polyfill, add type=cancel buttons
for (let d of document.querySelectorAll('dialog')) {
    dialogPolyfill.registerDialog(d);
    d.addEventListener('cancel', () => d.returnValue = '');
    for (let b of d.querySelectorAll('form button[type=cancel]')) {
        b.type = 'button';
        b.addEventListener('click', () => {
            d.dispatchEvent(new Event('cancel'));
            d.close();
        });
    }
}

}();