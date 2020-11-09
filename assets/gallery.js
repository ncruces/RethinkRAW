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

}();