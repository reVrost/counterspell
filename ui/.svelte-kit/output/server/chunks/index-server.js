import { a as ssr_context } from "./context.js";
import "clsx";
function onDestroy(fn) {
  /** @type {SSRContext} */
  ssr_context.r.on_destroy(fn);
}
async function tick() {
}
export {
  onDestroy as o,
  tick as t
};
