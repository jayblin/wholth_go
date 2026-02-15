function windowQueryParameterSet(key, value) {
    const query = new URLSearchParams(window.location.search);

    query.set(key, value);

    history.replaceState(null, "", `?${query}`);
}

function windowQueryParameterDelete(key, value) {
    const query = new URLSearchParams(window.location.search);

    query.delete(key);

    history.replaceState(null, "", `?${query}`);
}

function htmzReplaceElements(event) {
    const iframe = event.target;
    const hash = iframe.contentWindow.location.hash || null;

    if (!hash) {
        return;
    }

    let container = document.querySelector(hash);

    if (!container) {
        return;
    }

    container.replaceWith(...iframe.contentDocument.body.childNodes);

    container = document.querySelector(hash);

    if (!container || !container.form) {
        return;
    }

    for(input of container.form.elements ?? []) {
        if (!input.dataset.skip_enabler) {
            input.disabled = false;
        }

        if ("search" === input.type) {
            input.focus();
            break;
        }
    }

    // // container.previousElementSibling?.scrollIntoView({
    // container.form.scrollIntoView({
    //     block:'start',
    //     container:'all',
    //     behavior:'smooth'
    // });
}

/**
 * @see https://leanrada.com/htmz/
 */
function htmzOnLoad(event) {
    htmzReplaceElements(event);
    loaderUninstall(event);
}

let debounceId = null;

function debounce(callback) {
    if (debounceId) {
        clearTimeout(debounceId);
    }

    debounceId = setTimeout(callback, 300);
}

function paginate(event) {
    debounce(() => {
        const form = event.target.form;

        try {
            if (form.elements.affect_query) {
                windowQueryParameterSet(
                    "page_number",
                    form.elements.page_number.value
                );
            }
        } finally {
            form.onsubmit(event);
        }
    });
}

function paginateUp(event) {
    event.preventDefault();
    event.target.form.elements.page_number.stepUp();
    paginate(event);
}

function paginateDown(event) {
    event.preventDefault();
    event.target.form.elements.page_number.stepDown();
    paginate(event);
}

// function disableElements(event, ...names) {
//     for (const n of names) {
//         event.target.form.elements[n].disabled = !event.target.checked
//     }
// }

function disableSiblings(event) {
    event.target.parentElement.querySelectorAll("input").forEach(
        (e) => e !== event.target && (e.disabled = !event.target.checked)
    );
}

function searchOnSubmit(event) {
    event.preventDefault();

    loaderInstall(event);

    const form = event.target.form ?? event.target;

    form.submit();

    for(input of form.elements) {
        if (!input.dataset.skip_disabler) {
            input.disabled = true;
        }
    }
}

function searchOnInput(event) {
    debounce(() => 
        event.target.form.submit_btn.disabled = !(event.target.value.length > 0)
    );
}

function searchOnClick(event) {
    const form = event.target.form;

    if (form.elements.affect_query) {
        windowQueryParameterSet("q", form.elements.q.value);
    }
}

function searchOnClear(event) {
    const form = event.target.form;

    form.q.value = ``;
    form.elements.submit_btn.disabled=true;

    if (form.elements.affect_query) {
        windowQueryParameterDelete("q");
    }
}

function loaderInstall(event) {
    const loader = document.querySelector("#loader");
    if (loader) {
        loader.style.display = "block";
    }
}

function loaderUninstall(event) {
    const loader = document.querySelector("#loader");
    if (loader) {
        loader.style.display = "none";
    }
}

window.addEventListener("pageshow", () => {
    loaderUninstall();
});

window.addEventListener("beforeunload", () => {
    loaderInstall();
});
