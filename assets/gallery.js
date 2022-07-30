void function () {

if (upload === '') return;

let gallery = document.getElementById('gallery');

gallery.addEventListener('dragover', evt => evt.preventDefault());
gallery.addEventListener('drop', evt => {
    evt.preventDefault();
    console.log(evt.dataTransfer.files);
});

}();