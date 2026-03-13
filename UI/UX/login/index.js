const steamIdInput = document.getElementById("steamID");
const loginForm = document.getElementById("loginForm");
const appTitle = document.querySelector(".app-title");

const setTitleState = (state) => {
    if (!appTitle) {
        return;
    }

    appTitle.classList.remove("app-title--success", "app-title--error");

    if (state === "success") {
        appTitle.classList.add("app-title--success");
    }

    if (state === "error") {
        appTitle.classList.add("app-title--error");
    }
};

if (steamIdInput) {
    const validateSteamId = () => {
        steamIdInput.value = steamIdInput.value.replace(/\D/g, "");

        if (steamIdInput.value.length < 64) {
            steamIdInput.setCustomValidity("Steam ID should contain at least 64 digits");
            return;
        }

        steamIdInput.setCustomValidity("");
    };

    steamIdInput.addEventListener("input", () => {
        validateSteamId();
        setTitleState("idle");
    });

    validateSteamId();
}

if (loginForm) {
    loginForm.addEventListener("submit", (event) => {
        if (steamIdInput) {
            steamIdInput.value = steamIdInput.value.replace(/\D/g, "");
        }

        if (!loginForm.checkValidity()) {
            event.preventDefault();
            loginForm.reportValidity();
            setTitleState("error");
            return;
        }

        event.preventDefault();
        setTitleState("success");
    });
}
