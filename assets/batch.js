void function () {

window.popup = (elem, evt) => {
    if (evt.altKey || evt.ctrlKey || evt.metaKey || evt.shiftKey || evt.button !== 0) {
        return false;
    }

    let minimalUI = !window.matchMedia('(display-mode: browser)').matches;
    if (minimalUI) {
        return !window.open(elem.href, void 0, 'location=no,scrollbars=yes');
    }
    return !window.open(elem.href);
};

}();