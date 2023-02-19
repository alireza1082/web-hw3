var path = window.location.pathname;
var page = path.split("/").pop();
var pagename = page.split(".");

let isDark = window.matchMedia("(prefers-color-scheme:dark)").matches

if (isDark) {
    window.location.replace(`${pagename[0]}_dark.html`);
}

