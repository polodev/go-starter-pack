// Dark mode follows the saved choice, else the system setting. The layouts load this file in <head> without
// defer, so it runs before the page is painted (no flash). It is a file, not an inline script, because the
// Content-Security-Policy allows scripts from this site only.
(function () {
  var root = document.documentElement;
  try {
    if (localStorage.theme === "dark" || (!localStorage.theme && matchMedia("(prefers-color-scheme: dark)").matches)) {
      root.classList.add("dark");
    }
  } catch (error) {}

  document.addEventListener("click", function (event) {
    var button = event.target.closest && event.target.closest("[data-theme-toggle]");
    if (!button) return;
    root.classList.toggle("dark");
    try {
      localStorage.theme = root.classList.contains("dark") ? "dark" : "light";
    } catch (error) {}
  });
})();
