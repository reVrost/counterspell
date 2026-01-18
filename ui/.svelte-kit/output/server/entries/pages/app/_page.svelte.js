import "clsx";
function _page($$renderer) {
  window.location.href = "/dashboard";
  $$renderer.push(`<noscript><meta http-equiv="refresh" content="0;url=/dashboard"/> <a href="/dashboard">Click here if you are not redirected.</a></noscript>`);
}
export {
  _page as default
};
