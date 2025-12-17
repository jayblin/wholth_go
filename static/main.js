let debounceId = null;

function debounce(callback) {
    if (debounceId) {
        clearTimeout(debounceId);
    }

    debounceId = setTimeout(callback, 300);
}

function paginate(event) {
    debounce(() => {
        const target = event.target.form.querySelector('[formmethod=get]');
        
        if (target && typeof target.click === "function") {
            target.click();
        } else {
            event.target.form.submit();
        }
    });
}

function paginateUp(event) {
    event.preventDefault();
    event.target.parentElement.children[2].stepUp();
    paginate(event);
}

function paginateDown(event) {
    event.preventDefault();
    event.target.parentElement.children[2].stepDown();
    paginate(event);
}

window.addEventListener("load", () => {
});
