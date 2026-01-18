import * as universal from '../entries/pages/_page.ts.js';

export const index = 3;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/+page.ts";
export const imports = ["_app/immutable/nodes/3.CXQhTajo.js","_app/immutable/chunks/CN-VqVl8.js","_app/immutable/chunks/tPXKr78t.js","_app/immutable/chunks/BipdXalV.js","_app/immutable/chunks/B1iLSD93.js","_app/immutable/chunks/DR9TkfTq.js","_app/immutable/chunks/DD1utUw7.js","_app/immutable/chunks/Bw1CFWHr.js","_app/immutable/chunks/Do83GyS9.js","_app/immutable/chunks/DTnRGBIY.js","_app/immutable/chunks/c1eHLSmN.js","_app/immutable/chunks/D-VKx0uZ.js","_app/immutable/chunks/Dm_xWJ-l.js","_app/immutable/chunks/BznXBT2o.js","_app/immutable/chunks/Bn5TybZz.js","_app/immutable/chunks/i3ajiGCp.js"];
export const stylesheets = ["_app/immutable/assets/3.CoUJhNtc.css"];
export const fonts = [];
