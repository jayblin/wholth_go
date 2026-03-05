"use strict";

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

    for(const input of container.form.elements ?? []) {
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

    for(const input of form.elements) {
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

// Утилита для расчёта веса ингредиентов в порции.
const portionSize = {
    _inited: false,
    _numbers: [],
    _ingredientsMass: 0,

    init(ingredientsMass) {
        if (this.inited) {
            return;
        }

        this._inited = true;
        this._numbers = document.querySelectorAll(
            `#portion_items input[type="number"]`
        );
        this._ingredientsMass = Number(ingredientsMass);
    },

    handleInput(event) {
        const dialog = event.currentTarget;
        switch(event.target.name) {
            case "portion_mass_range": {
                const portionMass = this._ingredientsMass * event.target.value;
                dialog.children.portion_mass_number.value = portionMass.toFixed(0);
                this.updateNumbers(portionMass);
                return;
            }
            case "portion_mass_number": {
                const portionMass = Number(event.target.value); 
                dialog.children.portion_mass_range.value = portionMass / this._ingredientsMass;
                this.updateNumbers(portionMass);
                return;
            }
        }

        if (event.target.dataset.initial) {
            const portionMass = this._ingredientsMass * Number(event.target.value) / Number(event.target.dataset.initial);

            dialog.children.portion_mass_number.value = portionMass.toFixed(0);
            dialog.children.portion_mass_range.value = portionMass / this._ingredientsMass;
            this.updateNumbers(portionMass);
        }
    },

    updateNumbers(portionMass) {
        this._numbers.forEach(n => {
            n.value = (Number(n.dataset.initial)/this._ingredientsMass * portionMass)
                .toFixed(0);
        });
    },
};

window.addEventListener("pageshow", () => {
    loaderUninstall();
});

window.addEventListener("beforeunload", () => {
    loaderInstall();
});
