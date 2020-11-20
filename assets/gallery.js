void function () {

let observer = new IntersectionObserver(entries => {
    for (let entry of entries) {
        if (entry.isIntersecting) {
            let img = entry.target;
            img.src = img.dataset.src;
            img.classList.remove('lazy');
            observer.unobserve(img);
        }
    };
});

for (let img of document.querySelectorAll('img.lazy')) {
    observer.observe(img);
}

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