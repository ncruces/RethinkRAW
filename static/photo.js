void function() {

let form = document.getElementById('settings');

document.body.onload = async() => {
    let settings = await jsonRequest('GET', `/photo/${template.Path}?settings`);

    if (settings.process == null || settings.process < 6.7 || settings > 11) {
        if (settings.process) window.alert('This file was processed with an incompatible version of Camera Raw.\nPrevious edits will not be faithfully reproduced.');
        settings.process = 11;
    }
    if (settings.process < 11 && window.confirm('This file was processed with an older version of Camera Raw.\nPrevious edits may not be faithfully reproduced.\n\nUpdate to the current Camera Raw process version?')) {
        settings.process = 11;
    }

    form.orientation.value = settings.orientation;
    form.process.value = settings.process;
    form.profile.value = settings.profile;
    form.lensProfile.checked = settings.lensProfile;
    form.autoLateralCA.checked = settings.autoLateralCA;

    treatmentChange(form.grayscale, settings.grayscale);
    temperatureInput(form.temperature, settings.temperature);
    whiteBalanceChange(form.whiteBalance, settings.whiteBalance);

    let tone = 'Default';
    for (let k of ['exposure', 'contrast', 'highlights', 'shadows', 'whites', 'blacks', 'vibrance', 'saturation']) {
        if (settings[k] !== 0) tone = 'Custom';
        numberInput(form[k], settings[k]);
    }
    for (let k of ['tint', 'clarity', 'dehaze', 'sharpness', 'luminanceNR', 'colorNR']) {
        numberInput(form[k], settings[k]);
    }

    if (settings.autoTone) tone = 'Auto';
    toneChange(form.tone, tone);

    for (let n of form.querySelectorAll('fieldset')) {
        n.disabled = false;
    }
}

/*window.onbeforeunload = function (event) {
    return 'Do you want to leave this page? Changes you made may not be saved.';
}*/

window.exportJpeg = function() {
    window.location = `/photo/${template.Path}?export&` + formQuery();
}

window.valueChange = function() {
    let img = document.getElementById('preview');
    let done = true;

    function delayed() {
        img.onload = loaded;
        img.src = `/photo/${template.Path}?preview&` + formQuery();
    }

    function loaded() {
        let query = formQuery();
        if (query === img.src) {
            done = true;
        } else {
            done = false;
            setTimeout(delayed, 0);
        }
    }

    return () => {
        if (done) {
            done = false;
            setTimeout(delayed, 0);
        }
    };
}();

window.orientationChange = function(op) {
    const table = {
        ccw: [8, 8, 5, 6, 7, 4, 1, 2, 3],
        cw:  [6, 6, 7, 8, 5, 2, 3, 4, 1],
        hz:  [2, 2, 1, 4, 3, 6, 5, 8, 7],
        vt:  [4, 4, 3, 2, 1, 8, 7, 6, 5],
    };

    let orient = table[op][form.orientation.value]
    if (orient === void 0) orient = table[op][0];
    form.orientation.value = orient;

    valueChange();
}

window.treatmentChange = function(e, v) {
    const profiles = [
        ['Adobe Standard'],
        ['Adobe Standard'],
    ];

    if (v !== void 0) e.value = v;
    let color = e.value === 'false';
    if (e.length === 2) e = e[0];

    let profile = e.form.profile;
    profile.innerHTML = '';
    for (let o of profiles[+color]) {
        profile.insertAdjacentHTML('beforeend', `<option>${o}</option>`);
    }
    for (let n of e.form.querySelectorAll('div.color')) {
        n.classList.toggle('disabled-color', !color);
        disableInputs(n);
    }

    valueChange();
}

window.whiteBalanceChange = function(e, v) {
    const presets = {
        Daylight:   { temperature: 5500, tint: 10 },
        Cloudy:     { temperature: 6500, tint: 10 },
        Shade:      { temperature: 7500, tint: 10 },
        Tungsten:   { temperature: 2850, tint:  0 },
        Fluorescent:{ temperature: 3800, tint: 20 },
        Flash:      { temperature: 5500, tint:  0 },
    }

    if (v !== void 0) e.value = v;

    let temp = e.form.temperature;
    let tint = e.form.tint;
    let auto = false;
    if (e.value in presets) {
        let k = presets[e.value].temperature;
        let t = presets[e.value].tint;
        tint[0].value = t;
        tint[1].value = t;
        temp[0].value = k;
        temp[1].value = Math.log(k);
    } else if (e.value !== 'Custom') {
        auto = true;
    }
    for (let n of e.form.querySelectorAll('div.manualWB')) {
        n.classList.toggle('disabled-wb', auto);
        disableInputs(n);
    }

    valueChange();
}

window.toneChange = function(e, v) {
    if (v !== void 0) e.value = v;

    if (e.value === 'Default') {
        for (let i of e.form.querySelectorAll('div.customTone input')) {
            i.value = 0;
            numberInput(i);
        }
    }
    let auto = e.value === 'Auto';
    for (let n of e.form.querySelectorAll('div.customTone')) {
        n.classList.toggle('disabled-tone', auto);
        disableInputs(n);
    }

    valueChange();
}

window.temperatureInput = function(e, v) {
    if (e.length === 2) e = e[1];
    if (v !== void 0) e.value = Math.log(v);

    let n = Math.exp(Number(e.value));

    let r = n < 4000 ? 50 :
            n < 8000 ? 100 :
            n < 20000 ? 200 :
            n < 40000 ? 500 : 1000;

    e.previousElementSibling.value = Math.round(n / r) * r;
}

window.numberInput = function(e, v) {
    if (e.length === 2) e = e[1];
    if (v !== void 0) e.value = v;

    let n = Number(e.value);
    let s = n.toFixed(2 * (Number(e.step) < 1));
    if (n > 0 && e.min < 0 && s !== '0') s = '+' + s;
    e.previousElementSibling.value = s;
}

window.setCustomWhiteBalance = () => form.whiteBalance.value = 'Custom';
window.setCustomTone = () => form.tone.value = 'Custom';

function disableInputs(n) {
    let disabled = n.className.includes('disabled');
    for (let i of n.querySelectorAll('input')) {
        i.disabled = disabled;
    }
}

function formQuery() {
    let query = 'autoTone=' + (form.tone.value === 'Auto');

    for (let k of ['orientation', 'process', 'grayscale', 'whiteBalance']) {
        query += '&' + k + '=' + encodeURIComponent(form[k].value);
    }
    for (let k of ['temperature', 'tint', 'exposure', 'contrast', 'highlights', 'shadows', 'whites', 'blacks', 'clarity', 'dehaze', 'vibrance', 'saturation', 'sharpness', 'luminanceNR', 'colorNR']) {
        query += '&' + k + '=' + encodeURIComponent(form[k][0].value);
    }
    for (let k of ['lensProfile', 'autoLateralCA']) {
        query += '&' + k + '=' + encodeURIComponent(form[k].checked);
    }

    return query;
}

function jsonRequest(method, url, body) {
    return new Promise((resolve, reject) => {
        if (body !== void 0) body = JSON.stringify(body);
        let xhr = new XMLHttpRequest();
        xhr.responseType = 'json';
        xhr.open(method, url);
        xhr.onload = () => {
            if (200 <= xhr.status && xhr.status < 300) {
                resolve(xhr.response);
            } else {
                reject({
                    status: xhr.status,
                    statusText: xhr.statusText
                });
            }
        };
        xhr.onerror = () => reject({
            status: xhr.status,
            statusText: xhr.statusText
        });
        xhr.send(body);
    });
}

}()