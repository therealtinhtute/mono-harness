/* OpenPI Blueprint Design System · js/main.js
 * Sticky masthead state, active nav indicators, and smooth anchor scrolling.
 */
(function () {
  "use strict";

  var reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  /* Sticky masthead: lift a hairline + subtle shadow once scrolled */
  var mast = document.querySelector(".mast");
  if (mast) {
    var onScroll = function () {
      mast.classList.toggle("is-scrolled", window.scrollY > 0);
    };
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
  }

  /* Mark the current page in navigation by matching the trailing filename */
  var path = location.pathname.replace(/\/+$/, "");
  var here = (path.split("/").pop() || "index.html") || "index.html";

  document.querySelectorAll(".mast-nav a").forEach(function (link) {
    var href = link.getAttribute("href") || "";
    var target = href.replace(/\/+$/, "").split("/").pop() || "index.html";
    if (target === here) {
      link.setAttribute("aria-current", "page");
      link.classList.add("nav-link-active");
    } else {
      link.removeAttribute("aria-current");
      link.classList.remove("nav-link-active");
    }
  });

  /* Smooth-scroll in-page anchors, respecting reduced motion */
  if (!reduceMotion) {
    document.querySelectorAll('a[href^="#"]').forEach(function (anchor) {
      anchor.addEventListener("click", function (event) {
        var id = anchor.getAttribute("href").slice(1);
        if (!id) return;
        var target = document.getElementById(id);
        if (target) {
          event.preventDefault();
          target.scrollIntoView({ behavior: "smooth", block: "start" });
          history.replaceState(null, "", "#" + id);
        }
      });
    });
  }
})();
