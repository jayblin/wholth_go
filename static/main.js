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

function enableFormElements(form) {
    for(const input of form?.elements ?? []) {
        if (!input.dataset.skip_enabler) {
            input.disabled = false;
        }

        if ("search" === input.type) {
            input.focus();
            break;
        }
    }
}

function htmzReplaceElements(replacement, container) {
    if (!container) {
        return;
    }

    const id = container.id;

    container.replaceWith(...replacement);

    if (!id) {
        return;
    }

    container = document.getElementById(id);

    if (!container || !container.form) {
        return;
    }

    enableFormElements(container.form);

    // // container.previousElementSibling?.scrollIntoView({
    // container.form.scrollIntoView({
    //     block:"start",
    //     container:"all",
    //     behavior:"smooth"
    // });
}

function notificationsUpdateCounter(byHowMuch) {
    const counter = document.getElementById("notifications-counter");

    if (!counter) {
        return;
    }

    if ("flex" !== counter.style.display) {
        counter.style.display = "flex";
    }

    const currentCount = Number(counter.innerText ?? 0);
    const newCount = currentCount + byHowMuch;
    if  (newCount >= 100) {
        counter.innerText = "∞";
    } else {
        counter.innerText = newCount;
    }
}

function htmzHandleResponse(event) {
    const iframe = event.target;

    // TODO сделать так, чтобы все страницы перешли на схему с выдачей
    // элементов по id.
    const h1 = iframe.contentDocument.querySelector("h1");
    
    if (h1) {
        // TODO if body contains only text then push that text as a notification
        for (const ch of iframe.contentDocument.body.childNodes) {
            const id = ch.getAttribute?.('id');

            if (!id) {
                continue;
            }

            if ("notifications" === id) {
                const notificationCount = ch?.children?.length;
                notificationsUpdateCounter(notificationCount);

                document.getElementById("notifications")?.prepend(...(ch.childNodes));
                document.getElementById("notifications-container")?.style?.removeProperty("display");
            } else {
                document.getElementById(id)?.replaceWith(ch);
                enableFormElements(document.getElementById(id)?.form);
            }
        }
        // window.top.history.replaceState(null, "", iframe.contentDocument.location.pathname);

        return;
    }

    // console.log(window.location, window.top.location);
    const id = iframe.contentWindow.location.hash || null;

    // TODO сделать так, чтобы заменялись все элементы с id.
    const notificationsReceived = iframe.contentDocument.getElementById("notifications");
    if (notificationsReceived) {
        const notificationCount = notificationsReceived?.children?.length;
        notificationsUpdateCounter(notificationCount);

        document.getElementById("notifications")?.prepend(...notificationsReceived.childNodes);
        document.getElementById("notifications-container")?.style?.removeProperty("display");
    } else if (!id) {
        const plainText = iframe.contentDocument.body.innerText;
        notificationsUpdateCounter(1);

        const li = document.createElement("li");
        li.innerHTML = `<li class="status error">${plainText}</li>`;
        document.getElementById("notifications")?.prepend(li);
        document.getElementById("notifications-container")?.style?.removeProperty("display");
    }

    // TODO check that origin is the same
    // const iframeLoc = iframe.contentDocument.location.toString();
    // console.log(iframeLoc, document.location.toString(), iframe.contentDocument.location.pathname, document.location.pathname);
    // if (iframeLoc && iframeLoc != "about:blank" && iframeLoc !== document.location.toString()) {
    //     // document.location = iframeLoc;
    //     // return;
    //     // window.history.pushState(null, "", iframe.contentDocument.location.toString());
    //     // document.body.replaceWith(iframe.contentDocument.body);
    //     // return;
    // }

    if(notificationsReceived) {
        // add other elements from response to the document. 
        const replacement = iframe.contentDocument.body.querySelectorAll("body > *:not(#notifications)");
        let container = document.querySelector(id);

        if (replacement && container) {
            htmzReplaceElements(replacement, container)
        }
    } else if (id) {
        let container = document.querySelector(id);

        if (!container) {
            return;
        }

        htmzReplaceElements(iframe.contentDocument.body.childNodes, container)
    }
}

/**
 * @see https://leanrada.com/htmz/
 */
function htmzOnLoad(event) {
    htmzHandleResponse(event);
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
    let parent = event.target.parentElement;
    let i = 0;

    while (undefined !== parent.dataset.toggleable_bubble && i < 10) {
        i++;
        parent = parent.parentElement;
    }

    parent.querySelectorAll("input,button,select").forEach(
        (e) => e !== event.target
            && (e.disabled = ("hidden" === e.type ? false : !event.target.checked))
    );
}

function searchOnSubmit(event) {
    event.preventDefault();

    loaderInstall(event);

    const form = event.target.form ?? event.target;

    form.submit();

    for(const input of form.elements) {
        if (!input.dataset.skip_disabler && "hidden" !== input.type) {
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
        loader.style.display = "flex";
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
