void function () {

let images = document.querySelectorAll("img.lazy");

if ("IntersectionObserver" in window) {
    let observer = new IntersectionObserver(entries => {
        for (let entry of entries) {
            if (entry.isIntersecting) {
                let img = entry.target;
                img.src = img.dataset.src;
                img.classList.remove("lazy");
                observer.unobserve(img);
            }
        };
    });

    for (let img of images) {
        observer.observe(img);
    };
} else {
    for (let img of images) {
        img.src = img.dataset.src;
    };
}

}()