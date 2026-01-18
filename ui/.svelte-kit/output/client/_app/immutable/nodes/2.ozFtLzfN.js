var zs=Object.defineProperty;var Vs=Object.getPrototypeOf;var Us=Reflect.get;var Qi=e=>{throw TypeError(e)};var Gs=(e,t,r)=>t in e?zs(e,t,{enumerable:!0,configurable:!0,writable:!0,value:r}):e[t]=r;var Q=(e,t,r)=>Gs(e,typeof t!="symbol"?t+"":t,r),ui=(e,t,r)=>t.has(e)||Qi("Cannot "+r);var h=(e,t,r)=>(ui(e,t,"read from private field"),r?r.call(e):t.get(e)),k=(e,t,r)=>t.has(e)?Qi("Cannot add the same private member more than once"):t instanceof WeakSet?t.add(e):t.set(e,r),kt=(e,t,r,n)=>(ui(e,t,"write to private field"),n?n.call(e,r):t.set(e,r),r),Yt=(e,t,r)=>(ui(e,t,"access private method"),r);var Ji=(e,t,r)=>Us(Vs(e),r,t);import{b as $i,e as Ys,c as B,a as x,f as F,t as Xs}from"../chunks/DR9TkfTq.js";import{o as qs,a as Zs}from"../chunks/tPXKr78t.js";import{T as yt,h as pi,j as to,O as Qs,bd as Js,i as eo,C as $s,d as ta,ap as ea,aj as ra,e as na,g as Sn,be as ia,aq as hr,aP as Vo,w as ni,I as oa,bf as sa,p as $,f as I,n as lt,a as tt,X as V,ay as Me,m as w,o as l,aQ as Y,bg as ro,q as aa,_ as vr,u as la,v as St,bh as ca,bi as no,bc as ua,bj as Uo,W as E,U as g,V as v,Y as It,bk as da}from"../chunks/BipdXalV.js";import{c as fa,s as _t}from"../chunks/DD1utUw7.js";import{i as H}from"../chunks/Do83GyS9.js";import{s as J}from"../chunks/BznXBT2o.js";import{o as Nt,d as ii,g as ha}from"../chunks/Bw1CFWHr.js";import{I as At,a as di,b as We,r as Go,e as he,i as ve,s as zt,c as Ae,d as mi,f as Yo}from"../chunks/Bn5TybZz.js";import{c as Zt,b as va}from"../chunks/BRHnqgSc.js";import{a as L,f as ga,t as Pe,M as io}from"../chunks/D-VKx0uZ.js";import{e as pa,c as ie,C as Ti,b as Tt,d as ma}from"../chunks/Ot-C942e.js";import{s as vt,r as dt,p as N}from"../chunks/DTnRGBIY.js";import"../chunks/DT8pZ-Is.js";import{i as ba}from"../chunks/d4bZMR06.js";import{X as bi}from"../chunks/i3ajiGCp.js";import{L as ya}from"../chunks/uRZL1eWH.js";import{G as oo}from"../chunks/Dm_xWJ-l.js";function fi(e,t,r=!1,n=!1,i=!1){var s=e,o="";yt(()=>{var a=Qs;if(o===(o=t()??"")){pi&&to();return}if(a.nodes!==null&&(Js(a.nodes.start,a.nodes.end),a.nodes=null),o!==""){if(pi){eo.data;for(var c=to(),u=c;c!==null&&(c.nodeType!==$s||c.data!=="");)u=c,c=ta(c);if(c===null)throw ea(),ra;$i(eo,u),s=na(c);return}var d=o+"";r?d=`<svg>${d}</svg>`:n&&(d=`<math>${d}</math>`);var f=Ys(d);if((r||n)&&(f=Sn(f)),$i(Sn(f),f.lastChild),r||n)for(;Sn(f);)s.before(Sn(f));else s.before(f)}})}function yi(e,t,r=t){var n=new WeakSet;ia(e,"input",async i=>{var s=i?e.defaultValue:e.value;if(s=hi(e)?vi(s):s,r(s),hr!==null&&n.add(hr),await Vo(),s!==(s=t())){var o=e.selectionStart,a=e.selectionEnd,c=e.value.length;if(e.value=s??"",a!==null){var u=e.value.length;o===a&&a===c&&u>c?(e.selectionStart=u,e.selectionEnd=u):(e.selectionStart=o,e.selectionEnd=Math.min(a,u))}}}),(pi&&e.defaultValue!==e.value||ni(t)==null&&e.value)&&(r(hi(e)?vi(e.value):e.value),hr!==null&&n.add(hr)),oa(()=>{var i=t();if(e===document.activeElement){var s=sa??hr;if(n.has(s))return}hi(e)&&i===vi(e.value)||e.type==="date"&&!i&&!e.value||i!==e.value&&(e.value=i??"")})}function hi(e){var t=e.type;return t==="number"||t==="range"}function vi(e){return e===""?null:+e}function xa(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83z"}],["path",{d:"M2 12a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 12"}],["path",{d:"M2 17a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 17"}]];At(e,vt({name:"layers"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Xo(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"m6 9 6 6 6-6"}]];At(e,vt({name:"chevron-down"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function wa(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915"}],["circle",{cx:"12",cy:"12",r:"3"}]];At(e,vt({name:"settings"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function _a(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M12 15V3"}],["path",{d:"M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"}],["path",{d:"m7 10 5 5 5-5"}]];At(e,vt({name:"download"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Sa(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"m16 17 5-5-5-5"}],["path",{d:"M21 12H9"}],["path",{d:"M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"}]];At(e,vt({name:"log-out"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function xi(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"m21 21-4.34-4.34"}],["circle",{cx:"11",cy:"11",r:"8"}]];At(e,vt({name:"search"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Aa(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M5 12h14"}],["path",{d:"M12 5v14"}]];At(e,vt({name:"plus"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Pa(e){return typeof e=="function"}function Ea(e){return e!==null&&typeof e=="object"}const Oa=["string","number","bigint","boolean"];function wi(e){return e==null||Oa.includes(typeof e)?!0:Array.isArray(e)?e.every(t=>wi(t)):typeof e=="object"?Object.getPrototypeOf(e)===Object.prototype:!1}const wr=Symbol("box"),Ci=Symbol("is-writable");function Ta(e){return Ea(e)&&wr in e}function Ca(e){return O.isBox(e)&&Ci in e}function O(e){let t=V(Me(e));return{[wr]:!0,[Ci]:!0,get current(){return l(t)},set current(r){w(t,r,!0)}}}function Ma(e,t){const r=Y(e);return t?{[wr]:!0,[Ci]:!0,get current(){return l(r)},set current(n){t(n)}}:{[wr]:!0,get current(){return e()}}}function ka(e){return O.isBox(e)?e:Pa(e)?O.with(e):O(e)}function Ia(e){return Object.entries(e).reduce((t,[r,n])=>O.isBox(n)?(O.isWritableBox(n)?Object.defineProperty(t,r,{get(){return n.current},set(i){n.current=i}}):Object.defineProperty(t,r,{get(){return n.current}}),t):Object.assign(t,{[r]:n}),{})}function Na(e){return O.isWritableBox(e)?{[wr]:!0,get current(){return e.current}}:e}O.from=ka;O.with=Ma;O.flatten=Ia;O.readonly=Na;O.isBox=Ta;O.isWritableBox=Ca;function qo(...e){return function(t){var r;for(const n of e)if(n){if(t.defaultPrevented)return;typeof n=="function"?n.call(this,t):(r=n.current)==null||r.call(this,t)}}}var so=/\/\*[^*]*\*+([^/*][^*]*\*+)*\//g,Fa=/\n/g,Ra=/^\s*/,Da=/^(\*?[-#/*\\\w]+(\[[0-9a-z_-]+\])?)\s*/,La=/^:\s*/,ja=/^((?:'(?:\\'|.)*?'|"(?:\\"|.)*?"|\([^)]*?\)|[^};])+)/,Ba=/^[;\s]*/,Wa=/^\s+|\s+$/g,Ha=`
`,ao="/",lo="*",Ge="",Ka="comment",za="declaration";function Va(e,t){if(typeof e!="string")throw new TypeError("First argument must be a string");if(!e)return[];t=t||{};var r=1,n=1;function i(b){var y=b.match(Fa);y&&(r+=y.length);var A=b.lastIndexOf(Ha);n=~A?b.length-A:n+b.length}function s(){var b={line:r,column:n};return function(y){return y.position=new o(b),u(),y}}function o(b){this.start=b,this.end={line:r,column:n},this.source=t.source}o.prototype.content=e;function a(b){var y=new Error(t.source+":"+r+":"+n+": "+b);if(y.reason=b,y.filename=t.source,y.line=r,y.column=n,y.source=e,!t.silent)throw y}function c(b){var y=b.exec(e);if(y){var A=y[0];return i(A),e=e.slice(A.length),y}}function u(){c(Ra)}function d(b){var y;for(b=b||[];y=f();)y!==!1&&b.push(y);return b}function f(){var b=s();if(!(ao!=e.charAt(0)||lo!=e.charAt(1))){for(var y=2;Ge!=e.charAt(y)&&(lo!=e.charAt(y)||ao!=e.charAt(y+1));)++y;if(y+=2,Ge===e.charAt(y-1))return a("End of comment missing");var A=e.slice(2,y-2);return n+=2,i(A),e=e.slice(y),n+=2,b({type:Ka,comment:A})}}function p(){var b=s(),y=c(Da);if(y){if(f(),!c(La))return a("property missing ':'");var A=c(ja),_=b({type:za,property:co(y[0].replace(so,Ge)),value:A?co(A[0].replace(so,Ge)):Ge});return c(Ba),_}}function m(){var b=[];d(b);for(var y;y=p();)y!==!1&&(b.push(y),d(b));return b}return u(),m()}function co(e){return e?e.replace(Wa,Ge):Ge}function Ua(e,t){let r=null;if(!e||typeof e!="string")return r;const n=Va(e),i=typeof t=="function";return n.forEach(s=>{if(s.type!=="declaration")return;const{property:o,value:a}=s;i?t(o,a,s):a&&(r=r||{},r[o]=a)}),r}const Ga=/\d/,Ya=["-","_","/","."];function Xa(e=""){if(!Ga.test(e))return e!==e.toLowerCase()}function qa(e){const t=[];let r="",n,i;for(const s of e){const o=Ya.includes(s);if(o===!0){t.push(r),r="",n=void 0;continue}const a=Xa(s);if(i===!1){if(n===!1&&a===!0){t.push(r),r=s,n=a;continue}if(n===!0&&a===!1&&r.length>1){const c=r.at(-1);t.push(r.slice(0,Math.max(0,r.length-1))),r=c+s,n=a;continue}}r+=s,n=a,i=o}return t.push(r),t}function Zo(e){return e?qa(e).map(t=>Qa(t)).join(""):""}function Za(e){return Ja(Zo(e||""))}function Qa(e){return e?e[0].toUpperCase()+e.slice(1):""}function Ja(e){return e?e[0].toLowerCase()+e.slice(1):""}function br(e){if(!e)return{};const t={};function r(n,i){if(n.startsWith("-moz-")||n.startsWith("-webkit-")||n.startsWith("-ms-")||n.startsWith("-o-")){t[Zo(n)]=i;return}if(n.startsWith("--")){t[n]=i;return}t[Za(n)]=i}return Ua(e,r),t}function ke(...e){return(...t)=>{for(const r of e)typeof r=="function"&&r(...t)}}function $a(e,t){const r=RegExp(e,"g");return n=>{if(typeof n!="string")throw new TypeError(`expected an argument of type string, but got ${typeof n}`);return n.match(r)?n.replace(r,t):n}}const tl=$a(/[A-Z]/,e=>`-${e.toLowerCase()}`);function el(e){if(!e||typeof e!="object"||Array.isArray(e))throw new TypeError(`expected an argument of type object, but got ${typeof e}`);return Object.keys(e).map(t=>`${tl(t)}: ${e[t]};`).join(`
`)}function Mi(e={}){return el(e).replace(`
`," ")}const rl={position:"absolute",width:"1px",height:"1px",padding:"0",margin:"-1px",overflow:"hidden",clip:"rect(0, 0, 0, 0)",whiteSpace:"nowrap",borderWidth:"0",transform:"translateX(-100%)"};Mi(rl);function nl(e){var t;return e.length>2&&e.startsWith("on")&&e[2]===((t=e[2])==null?void 0:t.toLowerCase())}function pe(...e){const t={...e[0]};for(let r=1;r<e.length;r++){const n=e[r];for(const i in n){const s=t[i],o=n[i],a=typeof s=="function",c=typeof o=="function";if(a&&nl(i)){const u=s,d=o;t[i]=qo(u,d)}else if(a&&c)t[i]=ke(s,o);else if(i==="class"){const u=wi(s),d=wi(o);u&&d?t[i]=di(s,o):u?t[i]=di(s):d&&(t[i]=di(o))}else if(i==="style"){const u=typeof s=="object",d=typeof o=="object",f=typeof s=="string",p=typeof o=="string";if(u&&d)t[i]={...s,...o};else if(u&&p){const m=br(o);t[i]={...s,...m}}else if(f&&d){const m=br(s);t[i]={...m,...o}}else if(f&&p){const m=br(s),b=br(o);t[i]={...m,...b}}else u?t[i]=s:d?t[i]=o:f?t[i]=s:p&&(t[i]=o)}else t[i]=o!==void 0?o:s}}return typeof t.style=="object"&&(t.style=Mi(t.style).replaceAll(`
`," ")),t.hidden!==!0&&(t.hidden=void 0,delete t.hidden),t.disabled!==!0&&(t.disabled=void 0,delete t.disabled),t}const Qo=typeof window<"u"?window:void 0;function il(e){let t=e.activeElement;for(;t!=null&&t.shadowRoot;){const r=t.shadowRoot.activeElement;if(r===t)break;t=r}return t}var de,fe,Ee,Bn,se,yr,On;const Xi=class Xi extends Map{constructor(r){super();k(this,se);k(this,de,new Map);k(this,fe,V(0));k(this,Ee,V(0));k(this,Bn,ro||-1);if(r){for(var[n,i]of r)super.set(n,i);h(this,Ee).v=super.size}}has(r){var n=h(this,de),i=n.get(r);if(i===void 0){var s=super.get(r);if(s!==void 0)i=Yt(this,se,yr).call(this,0),n.set(r,i);else return l(h(this,fe)),!1}return l(i),!0}forEach(r,n){Yt(this,se,On).call(this),super.forEach(r,n)}get(r){var n=h(this,de),i=n.get(r);if(i===void 0){var s=super.get(r);if(s!==void 0)i=Yt(this,se,yr).call(this,0),n.set(r,i);else{l(h(this,fe));return}}return l(i),super.get(r)}set(r,n){var f;var i=h(this,de),s=i.get(r),o=super.get(r),a=super.set(r,n),c=h(this,fe);if(s===void 0)s=Yt(this,se,yr).call(this,0),i.set(r,s),w(h(this,Ee),super.size),vr(c);else if(o!==n){vr(s);var u=c.reactions===null?null:new Set(c.reactions),d=u===null||!((f=s.reactions)!=null&&f.every(p=>u.has(p)));d&&vr(c)}return a}delete(r){var n=h(this,de),i=n.get(r),s=super.delete(r);return i!==void 0&&(n.delete(r),w(h(this,Ee),super.size),w(i,-1),vr(h(this,fe))),s}clear(){if(super.size!==0){super.clear();var r=h(this,de);w(h(this,Ee),0);for(var n of r.values())w(n,-1);vr(h(this,fe)),r.clear()}}keys(){return l(h(this,fe)),super.keys()}values(){return Yt(this,se,On).call(this),super.values()}entries(){return Yt(this,se,On).call(this),super.entries()}[Symbol.iterator](){return this.entries()}get size(){return l(h(this,Ee)),super.size}};de=new WeakMap,fe=new WeakMap,Ee=new WeakMap,Bn=new WeakMap,se=new WeakSet,yr=function(r){return ro===h(this,Bn)?V(r):aa(r)},On=function(){l(h(this,fe));var r=h(this,de);if(h(this,Ee).v!==r.size){for(var n of Ji(Xi.prototype,this,"keys").call(this))if(!r.has(n)){var i=Yt(this,se,yr).call(this,0);r.set(n,i)}}for([,i]of h(this,de))l(i)};let _i=Xi;var er,Pr;class ol{constructor(t={}){k(this,er);k(this,Pr);const{window:r=Qo,document:n=r==null?void 0:r.document}=t;r!==void 0&&(kt(this,er,n),kt(this,Pr,fa(i=>{const s=Nt(r,"focusin",i),o=Nt(r,"focusout",i);return()=>{s(),o()}})))}get current(){var t;return(t=h(this,Pr))==null||t.call(this),h(this,er)?il(h(this,er)):null}}er=new WeakMap,Pr=new WeakMap;new ol;function sl(e){return typeof e=="function"}function al(e,t){switch(e){case"post":St(t);break;case"pre":la(t);break}}function Jo(e,t,r,n={}){const{lazy:i=!1}=n;let s=!i,o=Array.isArray(e)?[]:void 0;al(t,()=>{const a=Array.isArray(e)?e.map(u=>u()):e();if(!s){s=!0,o=a;return}const c=ni(()=>r(a,o));return o=a,c})}function Ht(e,t,r){Jo(e,"post",t,r)}function ll(e,t,r){Jo(e,"pre",t,r)}Ht.pre=ll;function cl(e){return sl(e)?e():e}var Oe;class ul{constructor(t,r={box:"border-box"}){k(this,Oe,V(Me({width:0,height:0})));var i,s;const n=r.window??Qo;w(h(this,Oe),{width:((i=r.initialSize)==null?void 0:i.width)??0,height:((s=r.initialSize)==null?void 0:s.height)??0},!0),St(()=>{if(!n)return;const o=cl(t);if(!o)return;const a=new n.ResizeObserver(c=>{for(const u of c){const d=r.box==="content-box"?u.contentBoxSize:u.borderBoxSize,f=Array.isArray(d)?d:[d];l(h(this,Oe)).width=f.reduce((p,m)=>Math.max(p,m.inlineSize),0),l(h(this,Oe)).height=f.reduce((p,m)=>Math.max(p,m.blockSize),0)}});return a.observe(o),()=>{a.disconnect()}})}get current(){return l(h(this,Oe))}get width(){return l(h(this,Oe)).width}get height(){return l(h(this,Oe)).height}}Oe=new WeakMap;var Er,Or;class dl{constructor(t){k(this,Er,V(void 0));k(this,Or);St(()=>{w(h(this,Er),h(this,Or),!0),kt(this,Or,t())})}get current(){return l(h(this,Er))}}Er=new WeakMap,Or=new WeakMap;var Tr,Te;class Ze{constructor(t){k(this,Tr);k(this,Te);kt(this,Tr,t),kt(this,Te,Symbol(t))}get key(){return h(this,Te)}exists(){return ca(h(this,Te))}get(){const t=no(h(this,Te));if(t===void 0)throw new Error(`Context "${h(this,Tr)}" not found`);return t}getOr(t){const r=no(h(this,Te));return r===void 0?t:r}set(t){return ua(h(this,Te),t)}}Tr=new WeakMap,Te=new WeakMap;function $o(e){St(()=>()=>{e()})}function ae({id:e,ref:t,deps:r=()=>!0,onRefChange:n,getRootNode:i}){Ht([()=>e.current,r],([s])=>{const o=(i==null?void 0:i())??document,a=o==null?void 0:o.getElementById(s);a?t.current=a:t.current=null,n==null||n(t.current)}),$o(()=>{t.current=null,n==null||n(null)})}function ki(e,t){return setTimeout(t,e)}function oe(e){Vo().then(e)}function fl(e){St(()=>ni(()=>e()))}function ts(e){return e?"open":"closed"}function hl(e){return e?"true":"false"}function vl(e){return e?"true":"false"}function es(e){return e?"":void 0}const _r="ArrowDown",Ii="ArrowLeft",Ni="ArrowRight",Cn="ArrowUp",rs="End",ns="Enter",gl="Escape",is="Home",pl="PageDown",ml="PageUp",Fi=" ",os="Tab";function bl(e){return window.getComputedStyle(e).getPropertyValue("direction")}function yl(e="ltr",t="horizontal"){return{horizontal:e==="rtl"?Ii:Ni,vertical:_r}[t]}function xl(e="ltr",t="horizontal"){return{horizontal:e==="rtl"?Ni:Ii,vertical:Cn}[t]}function wl(e="ltr",t="horizontal"){return["ltr","rtl"].includes(e)||(e="ltr"),["horizontal","vertical"].includes(t)||(t="horizontal"),{nextKey:yl(e,t),prevKey:xl(e,t)}}const Ri=typeof document<"u",uo=_l();function _l(){var e,t;return Ri&&((e=window==null?void 0:window.navigator)==null?void 0:e.userAgent)&&(/iP(ad|hone|od)/.test(window.navigator.userAgent)||((t=window==null?void 0:window.navigator)==null?void 0:t.maxTouchPoints)>2&&/iPad|Macintosh/.test(window==null?void 0:window.navigator.userAgent))}function Ie(e){return e instanceof HTMLElement}function Mn(e){return e instanceof Element}function Sl(e){return e instanceof Element||e instanceof SVGElement}function Al(e){return e!==null}function Pl(e){return e instanceof HTMLInputElement&&"select"in e}function El(e,t){if(getComputedStyle(e).visibility==="hidden")return!0;for(;e;){if(t!==void 0&&e===t)return!1;if(getComputedStyle(e).display==="none")return!0;e=e.parentElement}return!1}function Ol(e){const t=O(null);function r(){if(!Ri)return[];const o=document.getElementById(e.rootNodeId.current);return o?e.candidateSelector?Array.from(o.querySelectorAll(e.candidateSelector)):e.candidateAttr?Array.from(o.querySelectorAll(`[${e.candidateAttr}]:not([data-disabled])`)):[]:[]}function n(){var a;const o=r();o.length&&((a=o[0])==null||a.focus())}function i(o,a,c=!1){var P;const u=document.getElementById(e.rootNodeId.current);if(!u||!o)return;const d=r();if(!d.length)return;const f=d.indexOf(o),p=bl(u),{nextKey:m,prevKey:b}=wl(p,e.orientation.current),y=e.loop.current,A={[m]:f+1,[b]:f-1,[is]:0,[rs]:d.length-1};if(c){const C=m===_r?Ni:_r,M=b===Cn?Ii:Cn;A[C]=f+1,A[M]=f-1}let _=A[a.key];if(_===void 0)return;a.preventDefault(),_<0&&y?_=d.length-1:_===d.length&&y&&(_=0);const S=d[_];if(S)return S.focus(),t.current=S.id,(P=e.onCandidateFocus)==null||P.call(e,S),S}function s(o){const a=r(),c=t.current!==null;return o&&!c&&a[0]===o?(t.current=o.id,0):(o==null?void 0:o.id)===t.current?0:-1}return{setCurrentTabStopId(o){t.current=o},getTabIndex:s,handleKeydown:i,focusFirstCandidate:n,currentTabStopId:t}}globalThis.bitsIdCounter??(globalThis.bitsIdCounter={current:0});function Re(e="bits"){return globalThis.bitsIdCounter.current++,`${e}-${globalThis.bitsIdCounter.current}`}function Ft(){}function Tl(e,t){const r=O(e);function n(s){return t[r.current][s]??r.current}return{state:r,dispatch:s=>{r.current=n(s)}}}function Cl(e,t){let r=V(Me({})),n=V("none");const i=e.current?"mounted":"unmounted";let s=V(null);const o=new dl(()=>e.current);Ht([()=>t.current,()=>e.current],([p,m])=>{!p||!m||oe(()=>{w(s,document.getElementById(p),!0)})});const{state:a,dispatch:c}=Tl(i,{mounted:{UNMOUNT:"unmounted",ANIMATION_OUT:"unmountSuspended"},unmountSuspended:{MOUNT:"mounted",ANIMATION_END:"unmounted"},unmounted:{MOUNT:"mounted"}});Ht(()=>e.current,p=>{if(l(s)||w(s,document.getElementById(t.current),!0),!l(s)||!(p!==o.current))return;const b=l(n),y=An(l(s));p?c("MOUNT"):y==="none"||l(r).display==="none"?c("UNMOUNT"):c(o&&b!==y?"ANIMATION_OUT":"UNMOUNT")});function u(p){if(l(s)||w(s,document.getElementById(t.current),!0),!l(s))return;const m=An(l(s)),b=m.includes(p.animationName)||m==="none";p.target===l(s)&&b&&c("ANIMATION_END")}function d(p){l(s)||w(s,document.getElementById(t.current),!0),l(s)&&p.target===l(s)&&w(n,An(l(s)),!0)}Ht(()=>a.current,()=>{if(l(s)||w(s,document.getElementById(t.current),!0),!l(s))return;const p=An(l(s));w(n,a.current==="mounted"?p:"none",!0)}),Ht(()=>l(s),p=>{if(p)return w(r,getComputedStyle(p),!0),ke(Nt(p,"animationstart",d),Nt(p,"animationcancel",u),Nt(p,"animationend",u))});const f=Y(()=>["mounted","unmountSuspended"].includes(a.current));return{get current(){return l(f)}}}function An(e){return e&&getComputedStyle(e).animationName||"none"}function Ml(e,t){$(t,!0);const r=Cl(O.with(()=>t.present),O.with(()=>t.id));var n=B(),i=I(n);{var s=o=>{var a=B(),c=I(a);J(c,()=>t.presence??lt,()=>({present:r})),x(o,a)};H(i,o=>{(t.forceMount||t.present||r.current)&&o(s)})}x(e,n),tt()}function ss(e,t,r,n){const i=Array.isArray(t)?t:[t];return i.forEach(s=>e.addEventListener(s,r,n)),()=>{i.forEach(s=>e.removeEventListener(s,r,n))}}class Di{constructor(t,r={bubbles:!0,cancelable:!0}){Q(this,"eventName");Q(this,"options");this.eventName=t,this.options=r}createEvent(t){return new CustomEvent(this.eventName,{...this.options,detail:t})}dispatch(t,r){const n=this.createEvent(r);return t.dispatchEvent(n),n}listen(t,r,n){const i=s=>{r(s)};return Nt(t,this.eventName,i,n)}}function fo(e,t=500){let r=null;const n=(...i)=>{r!==null&&clearTimeout(r),r=setTimeout(()=>{e(...i)},t)};return n.destroy=()=>{r!==null&&(clearTimeout(r),r=null)},n}function Li(e,t){return e===t||e.contains(t)}function as(e){return(e==null?void 0:e.ownerDocument)??document}function ls(e){return(e==null?void 0:e.ownerDocument)??document}function kl(e,t){const{clientX:r,clientY:n}=e,i=t.getBoundingClientRect();return r<i.left||r>i.right||n<i.top||n>i.bottom}globalThis.bitsDismissableLayers??(globalThis.bitsDismissableLayers=new Map);var rr,Ye,je,nr,ir,Be,Cr,Mr,Ce,Wn,cr,cs,Hn,or,Kn,zn,Vn,Un,kr,us,Gn,Yn;class Il{constructor(t){k(this,cr);Q(this,"opts");k(this,rr);k(this,Ye);k(this,je,{pointerdown:!1});k(this,nr,!1);k(this,ir,!1);Q(this,"node",O(null));k(this,Be);k(this,Cr);k(this,Mr,V(null));k(this,Ce,Ft);k(this,Wn,t=>{t.defaultPrevented||this.currNode&&oe(()=>{var r,n;!this.currNode||h(this,Un).call(this,t.target)||t.target&&!h(this,ir)&&((n=(r=h(this,Cr)).current)==null||n.call(r,t))})});k(this,Hn,t=>{let r=t;r.defaultPrevented&&(r=ho(t)),h(this,rr).current(t)});k(this,or,fo(t=>{if(!this.currNode){h(this,Ce).call(this);return}const r=this.opts.isValidEvent.current(t,this.currNode)||Dl(t,this.currNode);if(!h(this,nr)||Yt(this,cr,us).call(this)||!r){h(this,Ce).call(this);return}let n=t;if(n.defaultPrevented&&(n=ho(n)),h(this,Ye).current!=="close"&&h(this,Ye).current!=="defer-otherwise-close"){h(this,Ce).call(this);return}t.pointerType==="touch"?(h(this,Ce).call(this),kt(this,Ce,ss(h(this,Be),"click",h(this,Hn),{once:!0}))):h(this,rr).current(n)},10));k(this,Kn,t=>{h(this,je)[t.type]=!0});k(this,zn,t=>{h(this,je)[t.type]=!1});k(this,Vn,()=>{this.node.current&&kt(this,nr,Rl(this.node.current))});k(this,Un,t=>this.node.current?Li(this.node.current,t):!1);k(this,kr,fo(()=>{for(const t in h(this,je))h(this,je)[t]=!1;kt(this,nr,!1)},20));k(this,Gn,()=>{kt(this,ir,!0)});k(this,Yn,()=>{kt(this,ir,!1)});Q(this,"props",{onfocuscapture:h(this,Gn),onblurcapture:h(this,Yn)});this.opts=t,ae({id:t.id,ref:this.node,deps:()=>t.enabled.current,onRefChange:i=>{this.currNode=i}}),kt(this,Ye,t.interactOutsideBehavior),kt(this,rr,t.onInteractOutside),kt(this,Cr,t.onFocusOutside),St(()=>{kt(this,Be,as(this.currNode))});let r=Ft;const n=()=>{h(this,kr).call(this),globalThis.bitsDismissableLayers.delete(this),h(this,or).destroy(),r()};Ht([()=>this.opts.enabled.current,()=>this.currNode],([i,s])=>{if(!(!i||!s))return ki(1,()=>{this.currNode&&(globalThis.bitsDismissableLayers.set(this,h(this,Ye)),r(),r=Yt(this,cr,cs).call(this))}),n}),$o(()=>{h(this,kr).destroy(),globalThis.bitsDismissableLayers.delete(this),h(this,or).destroy(),h(this,Ce).call(this),r()})}get currNode(){return l(h(this,Mr))}set currNode(t){w(h(this,Mr),t,!0)}}rr=new WeakMap,Ye=new WeakMap,je=new WeakMap,nr=new WeakMap,ir=new WeakMap,Be=new WeakMap,Cr=new WeakMap,Mr=new WeakMap,Ce=new WeakMap,Wn=new WeakMap,cr=new WeakSet,cs=function(){return ke(Nt(h(this,Be),"pointerdown",ke(h(this,Kn),h(this,Vn)),{capture:!0}),Nt(h(this,Be),"pointerdown",ke(h(this,zn),h(this,or))),Nt(h(this,Be),"focusin",h(this,Wn)))},Hn=new WeakMap,or=new WeakMap,Kn=new WeakMap,zn=new WeakMap,Vn=new WeakMap,Un=new WeakMap,kr=new WeakMap,us=function(){return Object.values(h(this,je)).some(Boolean)},Gn=new WeakMap,Yn=new WeakMap;function Nl(e){return new Il(e)}function Fl(e){return e.findLast(([t,{current:r}])=>r==="close"||r==="ignore")}function Rl(e){const t=[...globalThis.bitsDismissableLayers],r=Fl(t);if(r)return r[0].node.current===e;const[n]=t[0];return n.node.current===e}function Dl(e,t){if("button"in e&&e.button>0)return!1;const r=e.target;return Mn(r)?as(r).documentElement.contains(r)&&!Li(t,r)&&kl(e,t):!1}function ho(e){const t=e.currentTarget,r=e.target;let n;e instanceof PointerEvent?n=new PointerEvent(e.type,e):n=new PointerEvent("pointerdown",e);let i=!1;return new Proxy(n,{get:(o,a)=>a==="currentTarget"?t:a==="target"?r:a==="preventDefault"?()=>{i=!0,typeof o.preventDefault=="function"&&o.preventDefault()}:a==="defaultPrevented"?i:a in o?o[a]:e[a]})}function Ll(e,t){$(t,!0);let r=N(t,"interactOutsideBehavior",3,"close"),n=N(t,"onInteractOutside",3,Ft),i=N(t,"onFocusOutside",3,Ft),s=N(t,"isValidEvent",3,()=>!1);const o=Nl({id:O.with(()=>t.id),interactOutsideBehavior:O.with(()=>r()),onInteractOutside:O.with(()=>n()),enabled:O.with(()=>t.enabled),onFocusOutside:O.with(()=>i()),isValidEvent:O.with(()=>s())});var a=B(),c=I(a);J(c,()=>t.children??lt,()=>({props:o.props})),x(e,a),tt()}globalThis.bitsEscapeLayers??(globalThis.bitsEscapeLayers=new Map);var Xn,qn;class jl{constructor(t){Q(this,"opts");k(this,Xn,()=>Nt(document,"keydown",h(this,qn),{passive:!1}));k(this,qn,t=>{if(t.key!==gl||!Wl(this))return;const r=new KeyboardEvent(t.type,t);t.preventDefault();const n=this.opts.escapeKeydownBehavior.current;n!=="close"&&n!=="defer-otherwise-close"||this.opts.onEscapeKeydown.current(r)});this.opts=t;let r=Ft;Ht(()=>t.enabled.current,n=>(n&&(globalThis.bitsEscapeLayers.set(this,t.escapeKeydownBehavior),r=h(this,Xn).call(this)),()=>{r(),globalThis.bitsEscapeLayers.delete(this)}))}}Xn=new WeakMap,qn=new WeakMap;function Bl(e){return new jl(e)}function Wl(e){const t=[...globalThis.bitsEscapeLayers],r=t.findLast(([i,{current:s}])=>s==="close"||s==="ignore");if(r)return r[0]===e;const[n]=t[0];return n===e}function Hl(e,t){$(t,!0);let r=N(t,"escapeKeydownBehavior",3,"close"),n=N(t,"onEscapeKeydown",3,Ft);Bl({escapeKeydownBehavior:O.with(()=>r()),onEscapeKeydown:O.with(()=>n()),enabled:O.with(()=>t.enabled)});var i=B(),s=I(i);J(s,()=>t.children??lt),x(e,i),tt()}const De=O([]);function Kl(){return{add(e){const t=De.current[0];t&&e.id!==t.id&&t.pause(),De.current=vo(De.current,e),De.current.unshift(e)},remove(e){var t;De.current=vo(De.current,e),(t=De.current[0])==null||t.resume()},get current(){return De.current}}}function zl(){let e=V(!1),t=V(!1);return{id:Re(),get paused(){return l(e)},get isHandlingFocus(){return l(t)},set isHandlingFocus(r){w(t,r,!0)},pause(){w(e,!0)},resume(){w(e,!1)}}}function vo(e,t){return[...e].filter(r=>r.id!==t.id)}function Vl(e){return e.filter(t=>t.tagName!=="A")}function Le(e,{select:t=!1}={}){if(!(e&&e.focus)||document.activeElement===e)return;const r=document.activeElement;e.focus({preventScroll:!0}),e!==r&&Pl(e)&&t&&e.select()}function ds(e,{select:t=!1}={}){const r=document.activeElement;for(const n of e)if(Le(n,{select:t}),document.activeElement!==r)return!0}function go(e,t){for(const r of e)if(!El(r,t))return r}function fs(e){const t=[],r=document.createTreeWalker(e,NodeFilter.SHOW_ELEMENT,{acceptNode:n=>{const i=n.tagName==="INPUT"&&n.type==="hidden";return n.disabled||n.hidden||i?NodeFilter.FILTER_SKIP:n.tabIndex>=0?NodeFilter.FILTER_ACCEPT:NodeFilter.FILTER_SKIP}});for(;r.nextNode();)t.push(r.currentNode);return t}function Ul(e){const t=fs(e),r=go(t,e),n=go(t.reverse(),e);return[r,n]}/*!
* tabbable 6.4.0
* @license MIT, https://github.com/focus-trap/tabbable/blob/master/LICENSE
*/var hs=["input:not([inert]):not([inert] *)","select:not([inert]):not([inert] *)","textarea:not([inert]):not([inert] *)","a[href]:not([inert]):not([inert] *)","button:not([inert]):not([inert] *)","[tabindex]:not(slot):not([inert]):not([inert] *)","audio[controls]:not([inert]):not([inert] *)","video[controls]:not([inert]):not([inert] *)",'[contenteditable]:not([contenteditable="false"]):not([inert]):not([inert] *)',"details>summary:first-of-type:not([inert]):not([inert] *)","details:not([inert]):not([inert] *)"],kn=hs.join(","),vs=typeof Element>"u",Xe=vs?function(){}:Element.prototype.matches||Element.prototype.msMatchesSelector||Element.prototype.webkitMatchesSelector,In=!vs&&Element.prototype.getRootNode?function(e){var t;return e==null||(t=e.getRootNode)===null||t===void 0?void 0:t.call(e)}:function(e){return e==null?void 0:e.ownerDocument},Nn=function(t,r){var n;r===void 0&&(r=!0);var i=t==null||(n=t.getAttribute)===null||n===void 0?void 0:n.call(t,"inert"),s=i===""||i==="true",o=s||r&&t&&(typeof t.closest=="function"?t.closest("[inert]"):Nn(t.parentNode));return o},Gl=function(t){var r,n=t==null||(r=t.getAttribute)===null||r===void 0?void 0:r.call(t,"contenteditable");return n===""||n==="true"},gs=function(t,r,n){if(Nn(t))return[];var i=Array.prototype.slice.apply(t.querySelectorAll(kn));return r&&Xe.call(t,kn)&&i.unshift(t),i=i.filter(n),i},Fn=function(t,r,n){for(var i=[],s=Array.from(t);s.length;){var o=s.shift();if(!Nn(o,!1))if(o.tagName==="SLOT"){var a=o.assignedElements(),c=a.length?a:o.children,u=Fn(c,!0,n);n.flatten?i.push.apply(i,u):i.push({scopeParent:o,candidates:u})}else{var d=Xe.call(o,kn);d&&n.filter(o)&&(r||!t.includes(o))&&i.push(o);var f=o.shadowRoot||typeof n.getShadowRoot=="function"&&n.getShadowRoot(o),p=!Nn(f,!1)&&(!n.shadowRootFilter||n.shadowRootFilter(o));if(f&&p){var m=Fn(f===!0?o.children:f.children,!0,n);n.flatten?i.push.apply(i,m):i.push({scopeParent:o,candidates:m})}else s.unshift.apply(s,o.children)}}return i},ps=function(t){return!isNaN(parseInt(t.getAttribute("tabindex"),10))},ms=function(t){if(!t)throw new Error("No node provided");return t.tabIndex<0&&(/^(AUDIO|VIDEO|DETAILS)$/.test(t.tagName)||Gl(t))&&!ps(t)?0:t.tabIndex},Yl=function(t,r){var n=ms(t);return n<0&&r&&!ps(t)?0:n},Xl=function(t,r){return t.tabIndex===r.tabIndex?t.documentOrder-r.documentOrder:t.tabIndex-r.tabIndex},bs=function(t){return t.tagName==="INPUT"},ql=function(t){return bs(t)&&t.type==="hidden"},Zl=function(t){var r=t.tagName==="DETAILS"&&Array.prototype.slice.apply(t.children).some(function(n){return n.tagName==="SUMMARY"});return r},Ql=function(t,r){for(var n=0;n<t.length;n++)if(t[n].checked&&t[n].form===r)return t[n]},Jl=function(t){if(!t.name)return!0;var r=t.form||In(t),n=function(a){return r.querySelectorAll('input[type="radio"][name="'+a+'"]')},i;if(typeof window<"u"&&typeof window.CSS<"u"&&typeof window.CSS.escape=="function")i=n(window.CSS.escape(t.name));else try{i=n(t.name)}catch(o){return console.error("Looks like you have a radio button with a name attribute containing invalid CSS selector characters and need the CSS.escape polyfill: %s",o.message),!1}var s=Ql(i,t.form);return!s||s===t},$l=function(t){return bs(t)&&t.type==="radio"},tc=function(t){return $l(t)&&!Jl(t)},ec=function(t){var r,n=t&&In(t),i=(r=n)===null||r===void 0?void 0:r.host,s=!1;if(n&&n!==t){var o,a,c;for(s=!!((o=i)!==null&&o!==void 0&&(a=o.ownerDocument)!==null&&a!==void 0&&a.contains(i)||t!=null&&(c=t.ownerDocument)!==null&&c!==void 0&&c.contains(t));!s&&i;){var u,d,f;n=In(i),i=(u=n)===null||u===void 0?void 0:u.host,s=!!((d=i)!==null&&d!==void 0&&(f=d.ownerDocument)!==null&&f!==void 0&&f.contains(i))}}return s},po=function(t){var r=t.getBoundingClientRect(),n=r.width,i=r.height;return n===0&&i===0},rc=function(t,r){var n=r.displayCheck,i=r.getShadowRoot;if(n==="full-native"&&"checkVisibility"in t){var s=t.checkVisibility({checkOpacity:!1,opacityProperty:!1,contentVisibilityAuto:!0,visibilityProperty:!0,checkVisibilityCSS:!0});return!s}if(getComputedStyle(t).visibility==="hidden")return!0;var o=Xe.call(t,"details>summary:first-of-type"),a=o?t.parentElement:t;if(Xe.call(a,"details:not([open]) *"))return!0;if(!n||n==="full"||n==="full-native"||n==="legacy-full"){if(typeof i=="function"){for(var c=t;t;){var u=t.parentElement,d=In(t);if(u&&!u.shadowRoot&&i(u)===!0)return po(t);t.assignedSlot?t=t.assignedSlot:!u&&d!==t.ownerDocument?t=d.host:t=u}t=c}if(ec(t))return!t.getClientRects().length;if(n!=="legacy-full")return!0}else if(n==="non-zero-area")return po(t);return!1},nc=function(t){if(/^(INPUT|BUTTON|SELECT|TEXTAREA)$/.test(t.tagName))for(var r=t.parentElement;r;){if(r.tagName==="FIELDSET"&&r.disabled){for(var n=0;n<r.children.length;n++){var i=r.children.item(n);if(i.tagName==="LEGEND")return Xe.call(r,"fieldset[disabled] *")?!0:!i.contains(t)}return!0}r=r.parentElement}return!1},Rn=function(t,r){return!(r.disabled||ql(r)||rc(r,t)||Zl(r)||nc(r))},Si=function(t,r){return!(tc(r)||ms(r)<0||!Rn(t,r))},ic=function(t){var r=parseInt(t.getAttribute("tabindex"),10);return!!(isNaN(r)||r>=0)},ys=function(t){var r=[],n=[];return t.forEach(function(i,s){var o=!!i.scopeParent,a=o?i.scopeParent:i,c=Yl(a,o),u=o?ys(i.candidates):a;c===0?o?r.push.apply(r,u):r.push(a):n.push({documentOrder:s,tabIndex:c,item:i,isScope:o,content:u})}),n.sort(Xl).reduce(function(i,s){return s.isScope?i.push.apply(i,s.content):i.push(s.content),i},[]).concat(r)},oc=function(t,r){r=r||{};var n;return r.getShadowRoot?n=Fn([t],r.includeContainer,{filter:Si.bind(null,r),flatten:!1,getShadowRoot:r.getShadowRoot,shadowRootFilter:ic}):n=gs(t,r.includeContainer,Si.bind(null,r)),ys(n)},sc=function(t,r){r=r||{};var n;return r.getShadowRoot?n=Fn([t],r.includeContainer,{filter:Rn.bind(null,r),flatten:!0,getShadowRoot:r.getShadowRoot}):n=gs(t,r.includeContainer,Rn.bind(null,r)),n},oi=function(t,r){if(r=r||{},!t)throw new Error("No node provided");return Xe.call(t,kn)===!1?!1:Si(r,t)},ac=hs.concat("iframe:not([inert]):not([inert] *)").join(","),lc=function(t,r){if(r=r||{},!t)throw new Error("No node provided");return Xe.call(t,ac)===!1?!1:Rn(r,t)};const cc=new Di("focusScope.autoFocusOnMount",{bubbles:!1,cancelable:!0}),uc=new Di("focusScope.autoFocusOnDestroy",{bubbles:!1,cancelable:!0}),xs=new Ze("FocusScope");function dc({id:e,loop:t,enabled:r,onOpenAutoFocus:n,onCloseAutoFocus:i,forceMount:s}){const o=Kl(),a=zl(),c=O(null),u=xs.getOr({ignoreCloseAutoFocus:!1});let d=null;ae({id:e,ref:c,deps:()=>r.current});function f(_){if(!(a.paused||!c.current||a.isHandlingFocus)){a.isHandlingFocus=!0;try{const S=_.target;if(!Ie(S))return;const P=c.current.contains(S);if(_.type==="focusin")if(P)d=S;else{if(u.ignoreCloseAutoFocus)return;Le(d,{select:!0})}else _.type==="focusout"&&!P&&!u.ignoreCloseAutoFocus&&Le(d,{select:!0})}finally{a.isHandlingFocus=!1}}}function p(_){if(!d||!c.current)return;let S=!1;for(const P of _){if(P.type==="childList"&&P.removedNodes.length>0)for(const C of P.removedNodes){if(C===d){S=!0;break}if(C.nodeType===Node.ELEMENT_NODE&&C.contains(d)){S=!0;break}}if(S)break}S&&c.current&&!c.current.contains(document.activeElement)&&Le(c.current)}Ht([()=>c.current,()=>r.current],([_,S])=>{if(!_||!S)return;const P=ke(Nt(document,"focusin",f),Nt(document,"focusout",f)),C=new MutationObserver(p);return C.observe(_,{childList:!0,subtree:!0,attributes:!1}),()=>{P(),C.disconnect()}}),Ht([()=>s.current,()=>c.current],([_,S])=>{if(_)return;const P=document.activeElement;return m(S,P),()=>{S&&b(P)}}),Ht([()=>s.current,()=>c.current,()=>r.current],([_,S])=>{if(!_)return;const P=document.activeElement;return m(S,P),()=>{S&&b(P)}});function m(_,S){if(_||(_=document.getElementById(e.current)),!_||!r.current)return;if(o.add(a),!_.contains(S)){const C=cc.createEvent();n.current(C),C.defaultPrevented||oe(()=>{if(!_)return;ds(Vl(fs(_)),{select:!0})||Le(_)})}}function b(_){var C;const S=uc.createEvent();(C=i.current)==null||C.call(i,S);const P=u.ignoreCloseAutoFocus;ki(0,()=>{!S.defaultPrevented&&_&&!P&&Le(oi(_)?_:document.body,{select:!0}),o.remove(a)})}function y(_){if(!r.current||!t.current&&!r.current||a.paused)return;const S=_.key===os&&!_.ctrlKey&&!_.altKey&&!_.metaKey,P=document.activeElement;if(!(S&&P))return;const C=c.current;if(!C)return;const[M,K]=Ul(C);M&&K?!_.shiftKey&&P===K?(_.preventDefault(),t.current&&Le(M,{select:!0})):_.shiftKey&&P===M&&(_.preventDefault(),t.current&&Le(K,{select:!0})):P===C&&_.preventDefault()}const A=Y(()=>({id:e.current,tabindex:-1,onkeydown:y}));return{get props(){return l(A)}}}function fc(e,t){$(t,!0);let r=N(t,"trapFocus",3,!1),n=N(t,"loop",3,!1),i=N(t,"onCloseAutoFocus",3,Ft),s=N(t,"onOpenAutoFocus",3,Ft),o=N(t,"forceMount",3,!1);const a=dc({enabled:O.with(()=>r()),loop:O.with(()=>n()),onCloseAutoFocus:O.with(()=>i()),onOpenAutoFocus:O.with(()=>s()),id:O.with(()=>t.id),forceMount:O.with(()=>o())});var c=B(),u=I(c);J(u,()=>t.focusScope??lt,()=>({props:a.props})),x(e,c),tt()}globalThis.bitsTextSelectionLayers??(globalThis.bitsTextSelectionLayers=new Map);var sr,Ir,Zn,ws,Qn,Nr;class hc{constructor(t){k(this,Zn);Q(this,"opts");k(this,sr,Ft);k(this,Ir,O(null));k(this,Qn,t=>{const r=h(this,Ir).current,n=t.target;!Ie(r)||!Ie(n)||!this.opts.enabled.current||!pc(this)||!Li(r,n)||(this.opts.onPointerDown.current(t),!t.defaultPrevented&&kt(this,sr,gc(r)))});k(this,Nr,()=>{h(this,sr).call(this),kt(this,sr,Ft)});this.opts=t,ae({id:t.id,ref:h(this,Ir),deps:()=>this.opts.enabled.current});let r=Ft;Ht(()=>this.opts.enabled.current,n=>(n&&(globalThis.bitsTextSelectionLayers.set(this,this.opts.enabled),r(),r=Yt(this,Zn,ws).call(this)),()=>{r(),h(this,Nr).call(this),globalThis.bitsTextSelectionLayers.delete(this)}))}}sr=new WeakMap,Ir=new WeakMap,Zn=new WeakSet,ws=function(){return ke(Nt(document,"pointerdown",h(this,Qn)),Nt(document,"pointerup",qo(h(this,Nr),this.opts.onPointerUp.current)))},Qn=new WeakMap,Nr=new WeakMap;function vc(e){return new hc(e)}const mo=e=>e.style.userSelect||e.style.webkitUserSelect;function gc(e){const t=document.body,r=mo(t),n=mo(e);return Pn(t,"none"),Pn(e,"text"),()=>{Pn(t,r),Pn(e,n)}}function Pn(e,t){e.style.userSelect=t,e.style.webkitUserSelect=t}function pc(e){const t=[...globalThis.bitsTextSelectionLayers];if(!t.length)return!1;const r=t.at(-1);return r?r[0]===e:!1}function mc(e,t){$(t,!0);let r=N(t,"preventOverflowTextSelection",3,!0),n=N(t,"onPointerDown",3,Ft),i=N(t,"onPointerUp",3,Ft);vc({id:O.with(()=>t.id),onPointerDown:O.with(()=>n()),onPointerUp:O.with(()=>i()),enabled:O.with(()=>t.enabled&&r())});var s=B(),o=I(s);J(o,()=>t.children??lt),x(e,s),tt()}function bc(e){let t=0,r=V(void 0),n;function i(){t-=1,n&&t<=0&&(n(),w(r,void 0),n=void 0)}return(...s)=>(t+=1,l(r)===void 0&&(n=Uo(()=>{w(r,e(...s),!0)})),St(()=>()=>{i()}),l(r))}const yc=bc(()=>{const e=new _i,t=Y(()=>{for(const s of e.values())if(s)return!0;return!1});let r=V(null),n=null;function i(){Ri&&(document.body.setAttribute("style",l(r)??""),document.body.style.removeProperty("--scrollbar-width"),uo&&(n==null||n()))}return St(()=>{const s=l(t);return ni(()=>{if(!s)return;w(r,document.body.getAttribute("style"),!0);const o=getComputedStyle(document.body),a=window.innerWidth-document.documentElement.clientWidth,u={padding:Number.parseInt(o.paddingRight??"0",10)+a,margin:Number.parseInt(o.marginRight??"0",10)};a>0&&(document.body.style.paddingRight=`${u.padding}px`,document.body.style.marginRight=`${u.margin}px`,document.body.style.setProperty("--scrollbar-width",`${a}px`),document.body.style.overflow="hidden"),uo&&(n=ss(document,"touchmove",d=>{d.target===document.documentElement&&(d.touches.length>1||d.preventDefault())},{passive:!1})),oe(()=>{document.body.style.pointerEvents="none",document.body.style.overflow="hidden"})})}),St(()=>()=>{n==null||n()}),{get map(){return e},resetBodyStyle:i}});function xc(e,t=()=>null){const r=Re(),n=yc();if(!n)return;const i=Y(t);n.map.set(r,e??!1);const s=O.with(()=>n.map.get(r)??!1,o=>n.map.set(r,o));return St(()=>()=>{n.map.delete(r),!wc(n.map)&&(l(i)===null?requestAnimationFrame(()=>n.resetBodyStyle()):ki(l(i),()=>n.resetBodyStyle()))}),s}function wc(e){for(const[t,r]of e)if(r)return!0;return!1}function bo(e,t){$(t,!0);let r=N(t,"preventScroll",3,!0),n=N(t,"restoreScrollDelay",3,null);xc(r(),()=>n()),tt()}function _s(e,t,r){const n=t.toLowerCase();if(n.endsWith(" ")){const f=n.slice(0,-1);if(e.filter(y=>y.toLowerCase().startsWith(f)).length<=1)return _s(e,f,r);const m=r==null?void 0:r.toLowerCase();if(m&&m.startsWith(f)&&m.charAt(f.length)===" "&&t.trim()===f)return r;const b=e.filter(y=>y.toLowerCase().startsWith(n));if(b.length>0){const y=r?e.indexOf(r):-1;return yo(b,Math.max(y,0)).find(S=>S!==r)||r}}const s=t.length>1&&Array.from(t).every(f=>f===t[0])?t[0]:t,o=s.toLowerCase(),a=r?e.indexOf(r):-1;let c=yo(e,Math.max(a,0));s.length===1&&(c=c.filter(f=>f!==r));const d=c.find(f=>f==null?void 0:f.toLowerCase().startsWith(o));return d!==r?d:void 0}function yo(e,t){return e.map((r,n)=>e[(t+n)%e.length])}const _c=["top","right","bottom","left"],Ke=Math.min,Jt=Math.max,Dn=Math.round,En=Math.floor,me=e=>({x:e,y:e}),Sc={left:"right",right:"left",bottom:"top",top:"bottom"},Ac={start:"end",end:"start"};function Ai(e,t,r){return Jt(e,Ke(t,r))}function Ne(e,t){return typeof e=="function"?e(t):e}function Fe(e){return e.split("-")[0]}function ur(e){return e.split("-")[1]}function ji(e){return e==="x"?"y":"x"}function Bi(e){return e==="y"?"height":"width"}const Pc=new Set(["top","bottom"]);function ge(e){return Pc.has(Fe(e))?"y":"x"}function Wi(e){return ji(ge(e))}function Ec(e,t,r){r===void 0&&(r=!1);const n=ur(e),i=Wi(e),s=Bi(i);let o=i==="x"?n===(r?"end":"start")?"right":"left":n==="start"?"bottom":"top";return t.reference[s]>t.floating[s]&&(o=Ln(o)),[o,Ln(o)]}function Oc(e){const t=Ln(e);return[Pi(e),t,Pi(t)]}function Pi(e){return e.replace(/start|end/g,t=>Ac[t])}const xo=["left","right"],wo=["right","left"],Tc=["top","bottom"],Cc=["bottom","top"];function Mc(e,t,r){switch(e){case"top":case"bottom":return r?t?wo:xo:t?xo:wo;case"left":case"right":return t?Tc:Cc;default:return[]}}function kc(e,t,r,n){const i=ur(e);let s=Mc(Fe(e),r==="start",n);return i&&(s=s.map(o=>o+"-"+i),t&&(s=s.concat(s.map(Pi)))),s}function Ln(e){return e.replace(/left|right|bottom|top/g,t=>Sc[t])}function Ic(e){return{top:0,right:0,bottom:0,left:0,...e}}function Ss(e){return typeof e!="number"?Ic(e):{top:e,right:e,bottom:e,left:e}}function jn(e){const{x:t,y:r,width:n,height:i}=e;return{width:n,height:i,top:r,left:t,right:t+n,bottom:r+i,x:t,y:r}}function _o(e,t,r){let{reference:n,floating:i}=e;const s=ge(t),o=Wi(t),a=Bi(o),c=Fe(t),u=s==="y",d=n.x+n.width/2-i.width/2,f=n.y+n.height/2-i.height/2,p=n[a]/2-i[a]/2;let m;switch(c){case"top":m={x:d,y:n.y-i.height};break;case"bottom":m={x:d,y:n.y+n.height};break;case"right":m={x:n.x+n.width,y:f};break;case"left":m={x:n.x-i.width,y:f};break;default:m={x:n.x,y:n.y}}switch(ur(t)){case"start":m[o]-=p*(r&&u?-1:1);break;case"end":m[o]+=p*(r&&u?-1:1);break}return m}const Nc=async(e,t,r)=>{const{placement:n="bottom",strategy:i="absolute",middleware:s=[],platform:o}=r,a=s.filter(Boolean),c=await(o.isRTL==null?void 0:o.isRTL(t));let u=await o.getElementRects({reference:e,floating:t,strategy:i}),{x:d,y:f}=_o(u,n,c),p=n,m={},b=0;for(let y=0;y<a.length;y++){const{name:A,fn:_}=a[y],{x:S,y:P,data:C,reset:M}=await _({x:d,y:f,initialPlacement:n,placement:p,strategy:i,middlewareData:m,rects:u,platform:o,elements:{reference:e,floating:t}});d=S??d,f=P??f,m={...m,[A]:{...m[A],...C}},M&&b<=50&&(b++,typeof M=="object"&&(M.placement&&(p=M.placement),M.rects&&(u=M.rects===!0?await o.getElementRects({reference:e,floating:t,strategy:i}):M.rects),{x:d,y:f}=_o(u,p,c)),y=-1)}return{x:d,y:f,placement:p,strategy:i,middlewareData:m}};async function Sr(e,t){var r;t===void 0&&(t={});const{x:n,y:i,platform:s,rects:o,elements:a,strategy:c}=e,{boundary:u="clippingAncestors",rootBoundary:d="viewport",elementContext:f="floating",altBoundary:p=!1,padding:m=0}=Ne(t,e),b=Ss(m),A=a[p?f==="floating"?"reference":"floating":f],_=jn(await s.getClippingRect({element:(r=await(s.isElement==null?void 0:s.isElement(A)))==null||r?A:A.contextElement||await(s.getDocumentElement==null?void 0:s.getDocumentElement(a.floating)),boundary:u,rootBoundary:d,strategy:c})),S=f==="floating"?{x:n,y:i,width:o.floating.width,height:o.floating.height}:o.reference,P=await(s.getOffsetParent==null?void 0:s.getOffsetParent(a.floating)),C=await(s.isElement==null?void 0:s.isElement(P))?await(s.getScale==null?void 0:s.getScale(P))||{x:1,y:1}:{x:1,y:1},M=jn(s.convertOffsetParentRelativeRectToViewportRelativeRect?await s.convertOffsetParentRelativeRectToViewportRelativeRect({elements:a,rect:S,offsetParent:P,strategy:c}):S);return{top:(_.top-M.top+b.top)/C.y,bottom:(M.bottom-_.bottom+b.bottom)/C.y,left:(_.left-M.left+b.left)/C.x,right:(M.right-_.right+b.right)/C.x}}const Fc=e=>({name:"arrow",options:e,async fn(t){const{x:r,y:n,placement:i,rects:s,platform:o,elements:a,middlewareData:c}=t,{element:u,padding:d=0}=Ne(e,t)||{};if(u==null)return{};const f=Ss(d),p={x:r,y:n},m=Wi(i),b=Bi(m),y=await o.getDimensions(u),A=m==="y",_=A?"top":"left",S=A?"bottom":"right",P=A?"clientHeight":"clientWidth",C=s.reference[b]+s.reference[m]-p[m]-s.floating[b],M=p[m]-s.reference[m],K=await(o.getOffsetParent==null?void 0:o.getOffsetParent(u));let q=K?K[P]:0;(!q||!await(o.isElement==null?void 0:o.isElement(K)))&&(q=a.floating[P]||s.floating[b]);const nt=C/2-M/2,st=q/2-y[b]/2-1,rt=Ke(f[_],st),D=Ke(f[S],st),R=rt,j=q-y[b]-D,z=q/2-y[b]/2+nt,it=Ai(R,z,j),at=!c.arrow&&ur(i)!=null&&z!==it&&s.reference[b]/2-(z<R?rt:D)-y[b]/2<0,X=at?z<R?z-R:z-j:0;return{[m]:p[m]+X,data:{[m]:it,centerOffset:z-it-X,...at&&{alignmentOffset:X}},reset:at}}}),Rc=function(e){return e===void 0&&(e={}),{name:"flip",options:e,async fn(t){var r,n;const{placement:i,middlewareData:s,rects:o,initialPlacement:a,platform:c,elements:u}=t,{mainAxis:d=!0,crossAxis:f=!0,fallbackPlacements:p,fallbackStrategy:m="bestFit",fallbackAxisSideDirection:b="none",flipAlignment:y=!0,...A}=Ne(e,t);if((r=s.arrow)!=null&&r.alignmentOffset)return{};const _=Fe(i),S=ge(a),P=Fe(a)===a,C=await(c.isRTL==null?void 0:c.isRTL(u.floating)),M=p||(P||!y?[Ln(a)]:Oc(a)),K=b!=="none";!p&&K&&M.push(...kc(a,y,b,C));const q=[a,...M],nt=await Sr(t,A),st=[];let rt=((n=s.flip)==null?void 0:n.overflows)||[];if(d&&st.push(nt[_]),f){const z=Ec(i,o,C);st.push(nt[z[0]],nt[z[1]])}if(rt=[...rt,{placement:i,overflows:st}],!st.every(z=>z<=0)){var D,R;const z=(((D=s.flip)==null?void 0:D.index)||0)+1,it=q[z];if(it&&(!(f==="alignment"?S!==ge(it):!1)||rt.every(U=>ge(U.placement)===S?U.overflows[0]>0:!0)))return{data:{index:z,overflows:rt},reset:{placement:it}};let at=(R=rt.filter(X=>X.overflows[0]<=0).sort((X,U)=>X.overflows[1]-U.overflows[1])[0])==null?void 0:R.placement;if(!at)switch(m){case"bestFit":{var j;const X=(j=rt.filter(U=>{if(K){const et=ge(U.placement);return et===S||et==="y"}return!0}).map(U=>[U.placement,U.overflows.filter(et=>et>0).reduce((et,ht)=>et+ht,0)]).sort((U,et)=>U[1]-et[1])[0])==null?void 0:j[0];X&&(at=X);break}case"initialPlacement":at=a;break}if(i!==at)return{reset:{placement:at}}}return{}}}};function So(e,t){return{top:e.top-t.height,right:e.right-t.width,bottom:e.bottom-t.height,left:e.left-t.width}}function Ao(e){return _c.some(t=>e[t]>=0)}const Dc=function(e){return e===void 0&&(e={}),{name:"hide",options:e,async fn(t){const{rects:r}=t,{strategy:n="referenceHidden",...i}=Ne(e,t);switch(n){case"referenceHidden":{const s=await Sr(t,{...i,elementContext:"reference"}),o=So(s,r.reference);return{data:{referenceHiddenOffsets:o,referenceHidden:Ao(o)}}}case"escaped":{const s=await Sr(t,{...i,altBoundary:!0}),o=So(s,r.floating);return{data:{escapedOffsets:o,escaped:Ao(o)}}}default:return{}}}}},As=new Set(["left","top"]);async function Lc(e,t){const{placement:r,platform:n,elements:i}=e,s=await(n.isRTL==null?void 0:n.isRTL(i.floating)),o=Fe(r),a=ur(r),c=ge(r)==="y",u=As.has(o)?-1:1,d=s&&c?-1:1,f=Ne(t,e);let{mainAxis:p,crossAxis:m,alignmentAxis:b}=typeof f=="number"?{mainAxis:f,crossAxis:0,alignmentAxis:null}:{mainAxis:f.mainAxis||0,crossAxis:f.crossAxis||0,alignmentAxis:f.alignmentAxis};return a&&typeof b=="number"&&(m=a==="end"?b*-1:b),c?{x:m*d,y:p*u}:{x:p*u,y:m*d}}const jc=function(e){return e===void 0&&(e=0),{name:"offset",options:e,async fn(t){var r,n;const{x:i,y:s,placement:o,middlewareData:a}=t,c=await Lc(t,e);return o===((r=a.offset)==null?void 0:r.placement)&&(n=a.arrow)!=null&&n.alignmentOffset?{}:{x:i+c.x,y:s+c.y,data:{...c,placement:o}}}}},Bc=function(e){return e===void 0&&(e={}),{name:"shift",options:e,async fn(t){const{x:r,y:n,placement:i}=t,{mainAxis:s=!0,crossAxis:o=!1,limiter:a={fn:A=>{let{x:_,y:S}=A;return{x:_,y:S}}},...c}=Ne(e,t),u={x:r,y:n},d=await Sr(t,c),f=ge(Fe(i)),p=ji(f);let m=u[p],b=u[f];if(s){const A=p==="y"?"top":"left",_=p==="y"?"bottom":"right",S=m+d[A],P=m-d[_];m=Ai(S,m,P)}if(o){const A=f==="y"?"top":"left",_=f==="y"?"bottom":"right",S=b+d[A],P=b-d[_];b=Ai(S,b,P)}const y=a.fn({...t,[p]:m,[f]:b});return{...y,data:{x:y.x-r,y:y.y-n,enabled:{[p]:s,[f]:o}}}}}},Wc=function(e){return e===void 0&&(e={}),{options:e,fn(t){const{x:r,y:n,placement:i,rects:s,middlewareData:o}=t,{offset:a=0,mainAxis:c=!0,crossAxis:u=!0}=Ne(e,t),d={x:r,y:n},f=ge(i),p=ji(f);let m=d[p],b=d[f];const y=Ne(a,t),A=typeof y=="number"?{mainAxis:y,crossAxis:0}:{mainAxis:0,crossAxis:0,...y};if(c){const P=p==="y"?"height":"width",C=s.reference[p]-s.floating[P]+A.mainAxis,M=s.reference[p]+s.reference[P]-A.mainAxis;m<C?m=C:m>M&&(m=M)}if(u){var _,S;const P=p==="y"?"width":"height",C=As.has(Fe(i)),M=s.reference[f]-s.floating[P]+(C&&((_=o.offset)==null?void 0:_[f])||0)+(C?0:A.crossAxis),K=s.reference[f]+s.reference[P]+(C?0:((S=o.offset)==null?void 0:S[f])||0)-(C?A.crossAxis:0);b<M?b=M:b>K&&(b=K)}return{[p]:m,[f]:b}}}},Hc=function(e){return e===void 0&&(e={}),{name:"size",options:e,async fn(t){var r,n;const{placement:i,rects:s,platform:o,elements:a}=t,{apply:c=()=>{},...u}=Ne(e,t),d=await Sr(t,u),f=Fe(i),p=ur(i),m=ge(i)==="y",{width:b,height:y}=s.floating;let A,_;f==="top"||f==="bottom"?(A=f,_=p===(await(o.isRTL==null?void 0:o.isRTL(a.floating))?"start":"end")?"left":"right"):(_=f,A=p==="end"?"top":"bottom");const S=y-d.top-d.bottom,P=b-d.left-d.right,C=Ke(y-d[A],S),M=Ke(b-d[_],P),K=!t.middlewareData.shift;let q=C,nt=M;if((r=t.middlewareData.shift)!=null&&r.enabled.x&&(nt=P),(n=t.middlewareData.shift)!=null&&n.enabled.y&&(q=S),K&&!p){const rt=Jt(d.left,0),D=Jt(d.right,0),R=Jt(d.top,0),j=Jt(d.bottom,0);m?nt=b-2*(rt!==0||D!==0?rt+D:Jt(d.left,d.right)):q=y-2*(R!==0||j!==0?R+j:Jt(d.top,d.bottom))}await c({...t,availableWidth:nt,availableHeight:q});const st=await o.getDimensions(a.floating);return b!==st.width||y!==st.height?{reset:{rects:!0}}:{}}}};function si(){return typeof window<"u"}function dr(e){return Ps(e)?(e.nodeName||"").toLowerCase():"#document"}function $t(e){var t;return(e==null||(t=e.ownerDocument)==null?void 0:t.defaultView)||window}function ye(e){var t;return(t=(Ps(e)?e.ownerDocument:e.document)||window.document)==null?void 0:t.documentElement}function Ps(e){return si()?e instanceof Node||e instanceof $t(e).Node:!1}function le(e){return si()?e instanceof Element||e instanceof $t(e).Element:!1}function be(e){return si()?e instanceof HTMLElement||e instanceof $t(e).HTMLElement:!1}function Po(e){return!si()||typeof ShadowRoot>"u"?!1:e instanceof ShadowRoot||e instanceof $t(e).ShadowRoot}const Kc=new Set(["inline","contents"]);function xn(e){const{overflow:t,overflowX:r,overflowY:n,display:i}=ce(e);return/auto|scroll|overlay|hidden|clip/.test(t+n+r)&&!Kc.has(i)}const zc=new Set(["table","td","th"]);function Vc(e){return zc.has(dr(e))}const Uc=[":popover-open",":modal"];function ai(e){return Uc.some(t=>{try{return e.matches(t)}catch{return!1}})}const Gc=["transform","translate","scale","rotate","perspective"],Yc=["transform","translate","scale","rotate","perspective","filter"],Xc=["paint","layout","strict","content"];function Hi(e){const t=Ki(),r=le(e)?ce(e):e;return Gc.some(n=>r[n]?r[n]!=="none":!1)||(r.containerType?r.containerType!=="normal":!1)||!t&&(r.backdropFilter?r.backdropFilter!=="none":!1)||!t&&(r.filter?r.filter!=="none":!1)||Yc.some(n=>(r.willChange||"").includes(n))||Xc.some(n=>(r.contain||"").includes(n))}function qc(e){let t=ze(e);for(;be(t)&&!lr(t);){if(Hi(t))return t;if(ai(t))return null;t=ze(t)}return null}function Ki(){return typeof CSS>"u"||!CSS.supports?!1:CSS.supports("-webkit-backdrop-filter","none")}const Zc=new Set(["html","body","#document"]);function lr(e){return Zc.has(dr(e))}function ce(e){return $t(e).getComputedStyle(e)}function li(e){return le(e)?{scrollLeft:e.scrollLeft,scrollTop:e.scrollTop}:{scrollLeft:e.scrollX,scrollTop:e.scrollY}}function ze(e){if(dr(e)==="html")return e;const t=e.assignedSlot||e.parentNode||Po(e)&&e.host||ye(e);return Po(t)?t.host:t}function Es(e){const t=ze(e);return lr(t)?e.ownerDocument?e.ownerDocument.body:e.body:be(t)&&xn(t)?t:Es(t)}function Ar(e,t,r){var n;t===void 0&&(t=[]),r===void 0&&(r=!0);const i=Es(e),s=i===((n=e.ownerDocument)==null?void 0:n.body),o=$t(i);if(s){const a=Ei(o);return t.concat(o,o.visualViewport||[],xn(i)?i:[],a&&r?Ar(a):[])}return t.concat(i,Ar(i,[],r))}function Ei(e){return e.parent&&Object.getPrototypeOf(e.parent)?e.frameElement:null}function Os(e){const t=ce(e);let r=parseFloat(t.width)||0,n=parseFloat(t.height)||0;const i=be(e),s=i?e.offsetWidth:r,o=i?e.offsetHeight:n,a=Dn(r)!==s||Dn(n)!==o;return a&&(r=s,n=o),{width:r,height:n,$:a}}function zi(e){return le(e)?e:e.contextElement}function tr(e){const t=zi(e);if(!be(t))return me(1);const r=t.getBoundingClientRect(),{width:n,height:i,$:s}=Os(t);let o=(s?Dn(r.width):r.width)/n,a=(s?Dn(r.height):r.height)/i;return(!o||!Number.isFinite(o))&&(o=1),(!a||!Number.isFinite(a))&&(a=1),{x:o,y:a}}const Qc=me(0);function Ts(e){const t=$t(e);return!Ki()||!t.visualViewport?Qc:{x:t.visualViewport.offsetLeft,y:t.visualViewport.offsetTop}}function Jc(e,t,r){return t===void 0&&(t=!1),!r||t&&r!==$t(e)?!1:t}function qe(e,t,r,n){t===void 0&&(t=!1),r===void 0&&(r=!1);const i=e.getBoundingClientRect(),s=zi(e);let o=me(1);t&&(n?le(n)&&(o=tr(n)):o=tr(e));const a=Jc(s,r,n)?Ts(s):me(0);let c=(i.left+a.x)/o.x,u=(i.top+a.y)/o.y,d=i.width/o.x,f=i.height/o.y;if(s){const p=$t(s),m=n&&le(n)?$t(n):n;let b=p,y=Ei(b);for(;y&&n&&m!==b;){const A=tr(y),_=y.getBoundingClientRect(),S=ce(y),P=_.left+(y.clientLeft+parseFloat(S.paddingLeft))*A.x,C=_.top+(y.clientTop+parseFloat(S.paddingTop))*A.y;c*=A.x,u*=A.y,d*=A.x,f*=A.y,c+=P,u+=C,b=$t(y),y=Ei(b)}}return jn({width:d,height:f,x:c,y:u})}function ci(e,t){const r=li(e).scrollLeft;return t?t.left+r:qe(ye(e)).left+r}function Cs(e,t){const r=e.getBoundingClientRect(),n=r.left+t.scrollLeft-ci(e,r),i=r.top+t.scrollTop;return{x:n,y:i}}function $c(e){let{elements:t,rect:r,offsetParent:n,strategy:i}=e;const s=i==="fixed",o=ye(n),a=t?ai(t.floating):!1;if(n===o||a&&s)return r;let c={scrollLeft:0,scrollTop:0},u=me(1);const d=me(0),f=be(n);if((f||!f&&!s)&&((dr(n)!=="body"||xn(o))&&(c=li(n)),be(n))){const m=qe(n);u=tr(n),d.x=m.x+n.clientLeft,d.y=m.y+n.clientTop}const p=o&&!f&&!s?Cs(o,c):me(0);return{width:r.width*u.x,height:r.height*u.y,x:r.x*u.x-c.scrollLeft*u.x+d.x+p.x,y:r.y*u.y-c.scrollTop*u.y+d.y+p.y}}function tu(e){return Array.from(e.getClientRects())}function eu(e){const t=ye(e),r=li(e),n=e.ownerDocument.body,i=Jt(t.scrollWidth,t.clientWidth,n.scrollWidth,n.clientWidth),s=Jt(t.scrollHeight,t.clientHeight,n.scrollHeight,n.clientHeight);let o=-r.scrollLeft+ci(e);const a=-r.scrollTop;return ce(n).direction==="rtl"&&(o+=Jt(t.clientWidth,n.clientWidth)-i),{width:i,height:s,x:o,y:a}}const Eo=25;function ru(e,t){const r=$t(e),n=ye(e),i=r.visualViewport;let s=n.clientWidth,o=n.clientHeight,a=0,c=0;if(i){s=i.width,o=i.height;const d=Ki();(!d||d&&t==="fixed")&&(a=i.offsetLeft,c=i.offsetTop)}const u=ci(n);if(u<=0){const d=n.ownerDocument,f=d.body,p=getComputedStyle(f),m=d.compatMode==="CSS1Compat"&&parseFloat(p.marginLeft)+parseFloat(p.marginRight)||0,b=Math.abs(n.clientWidth-f.clientWidth-m);b<=Eo&&(s-=b)}else u<=Eo&&(s+=u);return{width:s,height:o,x:a,y:c}}const nu=new Set(["absolute","fixed"]);function iu(e,t){const r=qe(e,!0,t==="fixed"),n=r.top+e.clientTop,i=r.left+e.clientLeft,s=be(e)?tr(e):me(1),o=e.clientWidth*s.x,a=e.clientHeight*s.y,c=i*s.x,u=n*s.y;return{width:o,height:a,x:c,y:u}}function Oo(e,t,r){let n;if(t==="viewport")n=ru(e,r);else if(t==="document")n=eu(ye(e));else if(le(t))n=iu(t,r);else{const i=Ts(e);n={x:t.x-i.x,y:t.y-i.y,width:t.width,height:t.height}}return jn(n)}function Ms(e,t){const r=ze(e);return r===t||!le(r)||lr(r)?!1:ce(r).position==="fixed"||Ms(r,t)}function ou(e,t){const r=t.get(e);if(r)return r;let n=Ar(e,[],!1).filter(a=>le(a)&&dr(a)!=="body"),i=null;const s=ce(e).position==="fixed";let o=s?ze(e):e;for(;le(o)&&!lr(o);){const a=ce(o),c=Hi(o);!c&&a.position==="fixed"&&(i=null),(s?!c&&!i:!c&&a.position==="static"&&!!i&&nu.has(i.position)||xn(o)&&!c&&Ms(e,o))?n=n.filter(d=>d!==o):i=a,o=ze(o)}return t.set(e,n),n}function su(e){let{element:t,boundary:r,rootBoundary:n,strategy:i}=e;const o=[...r==="clippingAncestors"?ai(t)?[]:ou(t,this._c):[].concat(r),n],a=o[0],c=o.reduce((u,d)=>{const f=Oo(t,d,i);return u.top=Jt(f.top,u.top),u.right=Ke(f.right,u.right),u.bottom=Ke(f.bottom,u.bottom),u.left=Jt(f.left,u.left),u},Oo(t,a,i));return{width:c.right-c.left,height:c.bottom-c.top,x:c.left,y:c.top}}function au(e){const{width:t,height:r}=Os(e);return{width:t,height:r}}function lu(e,t,r){const n=be(t),i=ye(t),s=r==="fixed",o=qe(e,!0,s,t);let a={scrollLeft:0,scrollTop:0};const c=me(0);function u(){c.x=ci(i)}if(n||!n&&!s)if((dr(t)!=="body"||xn(i))&&(a=li(t)),n){const m=qe(t,!0,s,t);c.x=m.x+t.clientLeft,c.y=m.y+t.clientTop}else i&&u();s&&!n&&i&&u();const d=i&&!n&&!s?Cs(i,a):me(0),f=o.left+a.scrollLeft-c.x-d.x,p=o.top+a.scrollTop-c.y-d.y;return{x:f,y:p,width:o.width,height:o.height}}function gi(e){return ce(e).position==="static"}function To(e,t){if(!be(e)||ce(e).position==="fixed")return null;if(t)return t(e);let r=e.offsetParent;return ye(e)===r&&(r=r.ownerDocument.body),r}function ks(e,t){const r=$t(e);if(ai(e))return r;if(!be(e)){let i=ze(e);for(;i&&!lr(i);){if(le(i)&&!gi(i))return i;i=ze(i)}return r}let n=To(e,t);for(;n&&Vc(n)&&gi(n);)n=To(n,t);return n&&lr(n)&&gi(n)&&!Hi(n)?r:n||qc(e)||r}const cu=async function(e){const t=this.getOffsetParent||ks,r=this.getDimensions,n=await r(e.floating);return{reference:lu(e.reference,await t(e.floating),e.strategy),floating:{x:0,y:0,width:n.width,height:n.height}}};function uu(e){return ce(e).direction==="rtl"}const du={convertOffsetParentRelativeRectToViewportRelativeRect:$c,getDocumentElement:ye,getClippingRect:su,getOffsetParent:ks,getElementRects:cu,getClientRects:tu,getDimensions:au,getScale:tr,isElement:le,isRTL:uu};function Is(e,t){return e.x===t.x&&e.y===t.y&&e.width===t.width&&e.height===t.height}function fu(e,t){let r=null,n;const i=ye(e);function s(){var a;clearTimeout(n),(a=r)==null||a.disconnect(),r=null}function o(a,c){a===void 0&&(a=!1),c===void 0&&(c=1),s();const u=e.getBoundingClientRect(),{left:d,top:f,width:p,height:m}=u;if(a||t(),!p||!m)return;const b=En(f),y=En(i.clientWidth-(d+p)),A=En(i.clientHeight-(f+m)),_=En(d),P={rootMargin:-b+"px "+-y+"px "+-A+"px "+-_+"px",threshold:Jt(0,Ke(1,c))||1};let C=!0;function M(K){const q=K[0].intersectionRatio;if(q!==c){if(!C)return o();q?o(!1,q):n=setTimeout(()=>{o(!1,1e-7)},1e3)}q===1&&!Is(u,e.getBoundingClientRect())&&o(),C=!1}try{r=new IntersectionObserver(M,{...P,root:i.ownerDocument})}catch{r=new IntersectionObserver(M,P)}r.observe(e)}return o(!0),s}function hu(e,t,r,n){n===void 0&&(n={});const{ancestorScroll:i=!0,ancestorResize:s=!0,elementResize:o=typeof ResizeObserver=="function",layoutShift:a=typeof IntersectionObserver=="function",animationFrame:c=!1}=n,u=zi(e),d=i||s?[...u?Ar(u):[],...Ar(t)]:[];d.forEach(_=>{i&&_.addEventListener("scroll",r,{passive:!0}),s&&_.addEventListener("resize",r)});const f=u&&a?fu(u,r):null;let p=-1,m=null;o&&(m=new ResizeObserver(_=>{let[S]=_;S&&S.target===u&&m&&(m.unobserve(t),cancelAnimationFrame(p),p=requestAnimationFrame(()=>{var P;(P=m)==null||P.observe(t)})),r()}),u&&!c&&m.observe(u),m.observe(t));let b,y=c?qe(e):null;c&&A();function A(){const _=qe(e);y&&!Is(y,_)&&r(),y=_,b=requestAnimationFrame(A)}return r(),()=>{var _;d.forEach(S=>{i&&S.removeEventListener("scroll",r),s&&S.removeEventListener("resize",r)}),f==null||f(),(_=m)==null||_.disconnect(),m=null,c&&cancelAnimationFrame(b)}}const vu=jc,gu=Bc,pu=Rc,mu=Hc,bu=Dc,yu=Fc,xu=Wc,wu=(e,t,r)=>{const n=new Map,i={platform:du,...r},s={...i.platform,_c:n};return Nc(e,t,{...i,platform:s})};function gr(e){return typeof e=="function"?e():e}function Ns(e){return typeof window>"u"?1:(e.ownerDocument.defaultView||window).devicePixelRatio||1}function Co(e,t){const r=Ns(e);return Math.round(t*r)/r}function Mo(e){return{[`--bits-${e}-content-transform-origin`]:"var(--bits-floating-transform-origin)",[`--bits-${e}-content-available-width`]:"var(--bits-floating-available-width)",[`--bits-${e}-content-available-height`]:"var(--bits-floating-available-height)",[`--bits-${e}-anchor-width`]:"var(--bits-floating-anchor-width)",[`--bits-${e}-anchor-height`]:"var(--bits-floating-anchor-height)"}}function _u(e){const t=e.whileElementsMounted,r=Y(()=>gr(e.open)??!0),n=Y(()=>gr(e.middleware)),i=Y(()=>gr(e.transform)??!0),s=Y(()=>gr(e.placement)??"bottom"),o=Y(()=>gr(e.strategy)??"absolute"),a=e.reference;let c=V(0),u=V(0);const d=O(null);let f=V(Me(l(o))),p=V(Me(l(s))),m=V(Me({})),b=V(!1);const y=Y(()=>{const M={position:l(f),left:"0",top:"0"};if(!d.current)return M;const K=Co(d.current,l(c)),q=Co(d.current,l(u));return l(i)?{...M,transform:`translate(${K}px, ${q}px)`,...Ns(d.current)>=1.5&&{willChange:"transform"}}:{position:l(f),left:`${K}px`,top:`${q}px`}});let A;function _(){a.current===null||d.current===null||wu(a.current,d.current,{middleware:l(n),placement:l(s),strategy:l(o)}).then(M=>{w(c,M.x,!0),w(u,M.y,!0),w(f,M.strategy,!0),w(p,M.placement,!0),w(m,M.middlewareData,!0),w(b,!0)})}function S(){typeof A=="function"&&(A(),A=void 0)}function P(){if(S(),t===void 0){_();return}a.current===null||d.current===null||(A=t(a.current,d.current,_))}function C(){l(r)||w(b,!1)}return St(_),St(P),St(C),St(()=>S),{floating:d,reference:a,get strategy(){return l(f)},get placement(){return l(p)},get middlewareData(){return l(m)},get isPositioned(){return l(b)},get floatingStyles(){return l(y)},get update(){return _}}}const Su={top:"bottom",right:"left",bottom:"top",left:"right"};class Au{constructor(){Q(this,"anchorNode",O(null));Q(this,"customAnchorNode",O(null));Q(this,"triggerNode",O(null));St(()=>{this.customAnchorNode.current?typeof this.customAnchorNode.current=="string"?this.anchorNode.current=document.querySelector(this.customAnchorNode.current):this.anchorNode.current=this.customAnchorNode.current:this.anchorNode.current=this.triggerNode.current})}}var Fr,Jn,Rr,$n,Dr,ti,Lr,jr,Br,Wr,Hr,Kr,zr,Vr,Ur,Gr,Yr,Xr,qr,Zr,Qr,Jr,$r,tn;class Pu{constructor(t,r){Q(this,"opts");Q(this,"root");Q(this,"contentRef",O(null));Q(this,"wrapperRef",O(null));Q(this,"arrowRef",O(null));Q(this,"arrowId",O(Re()));k(this,Fr,Y(()=>{if(typeof this.opts.style=="string")return br(this.opts.style);if(!this.opts.style)return{}}));k(this,Jn);k(this,Rr,new ul(()=>this.arrowRef.current??void 0));k(this,$n,Y(()=>{var t;return((t=h(this,Rr))==null?void 0:t.width)??0}));k(this,Dr,Y(()=>{var t;return((t=h(this,Rr))==null?void 0:t.height)??0}));k(this,ti,Y(()=>{var t;return((t=this.opts.side)==null?void 0:t.current)+(this.opts.align.current!=="center"?`-${this.opts.align.current}`:"")}));k(this,Lr,Y(()=>Array.isArray(this.opts.collisionBoundary.current)?this.opts.collisionBoundary.current:[this.opts.collisionBoundary.current]));k(this,jr,Y(()=>l(h(this,Lr)).length>0));k(this,Br,Y(()=>({padding:this.opts.collisionPadding.current,boundary:l(h(this,Lr)).filter(Al),altBoundary:this.hasExplicitBoundaries})));k(this,Wr,V(void 0));k(this,Hr,V(void 0));k(this,Kr,V(void 0));k(this,zr,V(void 0));k(this,Vr,Y(()=>[vu({mainAxis:this.opts.sideOffset.current+l(h(this,Dr)),alignmentAxis:this.opts.alignOffset.current}),this.opts.avoidCollisions.current&&gu({mainAxis:!0,crossAxis:!1,limiter:this.opts.sticky.current==="partial"?xu():void 0,...this.detectOverflowOptions}),this.opts.avoidCollisions.current&&pu({...this.detectOverflowOptions}),mu({...this.detectOverflowOptions,apply:({rects:t,availableWidth:r,availableHeight:n})=>{const{width:i,height:s}=t.reference;w(h(this,Wr),r,!0),w(h(this,Hr),n,!0),w(h(this,Kr),i,!0),w(h(this,zr),s,!0)}}),this.arrowRef.current&&yu({element:this.arrowRef.current,padding:this.opts.arrowPadding.current}),ku({arrowWidth:l(h(this,$n)),arrowHeight:l(h(this,Dr))}),this.opts.hideWhenDetached.current&&bu({strategy:"referenceHidden",...this.detectOverflowOptions})].filter(Boolean)));Q(this,"floating");k(this,Ur,Y(()=>Iu(this.floating.placement)));k(this,Gr,Y(()=>Nu(this.floating.placement)));k(this,Yr,Y(()=>{var t;return((t=this.floating.middlewareData.arrow)==null?void 0:t.x)??0}));k(this,Xr,Y(()=>{var t;return((t=this.floating.middlewareData.arrow)==null?void 0:t.y)??0}));k(this,qr,Y(()=>{var t;return((t=this.floating.middlewareData.arrow)==null?void 0:t.centerOffset)!==0}));k(this,Zr,V());k(this,Qr,Y(()=>Su[this.placedSide]));k(this,Jr,Y(()=>{var t,r,n;return{id:this.opts.wrapperId.current,"data-bits-floating-content-wrapper":"",style:{...this.floating.floatingStyles,transform:this.floating.isPositioned?this.floating.floatingStyles.transform:"translate(0, -200%)",minWidth:"max-content",zIndex:this.contentZIndex,"--bits-floating-transform-origin":`${(t=this.floating.middlewareData.transformOrigin)==null?void 0:t.x} ${(r=this.floating.middlewareData.transformOrigin)==null?void 0:r.y}`,"--bits-floating-available-width":`${l(h(this,Wr))}px`,"--bits-floating-available-height":`${l(h(this,Hr))}px`,"--bits-floating-anchor-width":`${l(h(this,Kr))}px`,"--bits-floating-anchor-height":`${l(h(this,zr))}px`,...((n=this.floating.middlewareData.hide)==null?void 0:n.referenceHidden)&&{visibility:"hidden","pointer-events":"none"},...l(h(this,Fr))},dir:this.opts.dir.current}}));k(this,$r,Y(()=>({"data-side":this.placedSide,"data-align":this.placedAlign,style:Mi({...l(h(this,Fr))})})));k(this,tn,Y(()=>({position:"absolute",left:this.arrowX?`${this.arrowX}px`:void 0,top:this.arrowY?`${this.arrowY}px`:void 0,[this.arrowBaseSide]:0,"transform-origin":{top:"",right:"0 0",bottom:"center 0",left:"100% 0"}[this.placedSide],transform:{top:"translateY(100%)",right:"translateY(50%) rotate(90deg) translateX(-50%)",bottom:"rotate(180deg)",left:"translateY(50%) rotate(-90deg) translateX(50%)"}[this.placedSide],visibility:this.cannotCenterArrow?"hidden":void 0})));this.opts=t,this.root=r,t.customAnchor&&(this.root.customAnchorNode.current=t.customAnchor.current),Ht(()=>t.customAnchor.current,n=>{this.root.customAnchorNode.current=n}),ae({id:this.opts.wrapperId,ref:this.wrapperRef,deps:()=>this.opts.enabled.current}),ae({id:this.opts.id,ref:this.contentRef,deps:()=>this.opts.enabled.current}),this.floating=_u({strategy:()=>this.opts.strategy.current,placement:()=>l(h(this,ti)),middleware:()=>this.middleware,reference:this.root.anchorNode,whileElementsMounted:(...n)=>{var s;return hu(...n,{animationFrame:((s=h(this,Jn))==null?void 0:s.current)==="always"})},open:()=>this.opts.enabled.current}),St(()=>{var n;this.floating.isPositioned&&((n=this.opts.onPlaced)==null||n.current())}),Ht(()=>this.contentRef.current,n=>{n&&(this.contentZIndex=window.getComputedStyle(n).zIndex)}),St(()=>{this.floating.floating.current=this.wrapperRef.current})}get hasExplicitBoundaries(){return l(h(this,jr))}set hasExplicitBoundaries(t){w(h(this,jr),t)}get detectOverflowOptions(){return l(h(this,Br))}set detectOverflowOptions(t){w(h(this,Br),t)}get middleware(){return l(h(this,Vr))}set middleware(t){w(h(this,Vr),t)}get placedSide(){return l(h(this,Ur))}set placedSide(t){w(h(this,Ur),t)}get placedAlign(){return l(h(this,Gr))}set placedAlign(t){w(h(this,Gr),t)}get arrowX(){return l(h(this,Yr))}set arrowX(t){w(h(this,Yr),t)}get arrowY(){return l(h(this,Xr))}set arrowY(t){w(h(this,Xr),t)}get cannotCenterArrow(){return l(h(this,qr))}set cannotCenterArrow(t){w(h(this,qr),t)}get contentZIndex(){return l(h(this,Zr))}set contentZIndex(t){w(h(this,Zr),t,!0)}get arrowBaseSide(){return l(h(this,Qr))}set arrowBaseSide(t){w(h(this,Qr),t)}get wrapperProps(){return l(h(this,Jr))}set wrapperProps(t){w(h(this,Jr),t)}get props(){return l(h(this,$r))}set props(t){w(h(this,$r),t)}get arrowStyle(){return l(h(this,tn))}set arrowStyle(t){w(h(this,tn),t)}}Fr=new WeakMap,Jn=new WeakMap,Rr=new WeakMap,$n=new WeakMap,Dr=new WeakMap,ti=new WeakMap,Lr=new WeakMap,jr=new WeakMap,Br=new WeakMap,Wr=new WeakMap,Hr=new WeakMap,Kr=new WeakMap,zr=new WeakMap,Vr=new WeakMap,Ur=new WeakMap,Gr=new WeakMap,Yr=new WeakMap,Xr=new WeakMap,qr=new WeakMap,Zr=new WeakMap,Qr=new WeakMap,Jr=new WeakMap,$r=new WeakMap,tn=new WeakMap;class Eu{constructor(t,r){Q(this,"opts");Q(this,"root");Q(this,"ref",O(null));this.opts=t,this.root=r,t.virtualEl&&t.virtualEl.current?r.triggerNode=O.from(t.virtualEl.current):ae({id:t.id,ref:this.ref,onRefChange:n=>{r.triggerNode.current=n}})}}const Vi=new Ze("Floating.Root"),Ou=new Ze("Floating.Content");function Tu(){return Vi.set(new Au)}function Cu(e){return Ou.set(new Pu(e,Vi.get()))}function Mu(e){return new Eu(e,Vi.get())}function ku(e){return{name:"transformOrigin",options:e,fn(t){var A,_,S;const{placement:r,rects:n,middlewareData:i}=t,o=((A=i.arrow)==null?void 0:A.centerOffset)!==0,a=o?0:e.arrowWidth,c=o?0:e.arrowHeight,[u,d]=Ui(r),f={start:"0%",center:"50%",end:"100%"}[d],p=(((_=i.arrow)==null?void 0:_.x)??0)+a/2,m=(((S=i.arrow)==null?void 0:S.y)??0)+c/2;let b="",y="";return u==="bottom"?(b=o?f:`${p}px`,y=`${-c}px`):u==="top"?(b=o?f:`${p}px`,y=`${n.floating.height+c}px`):u==="right"?(b=`${-c}px`,y=o?f:`${m}px`):u==="left"&&(b=`${n.floating.width+c}px`,y=o?f:`${m}px`),{data:{x:b,y}}}}}function Ui(e){const[t,r="center"]=e.split("-");return[t,r]}function Iu(e){return Ui(e)[0]}function Nu(e){return Ui(e)[1]}function Fu(e,t){$(t,!0),Tu();var r=B(),n=I(r);J(n,()=>t.children??lt),x(e,r),tt()}function Fs(e,t=1e4,r=Ft){let n=null,i=V(Me(e));function s(){return window.setTimeout(()=>{w(i,e,!0),r(e)},t)}return St(()=>()=>{n&&clearTimeout(n)}),O.with(()=>l(i),o=>{w(i,o,!0),r(o),n&&clearTimeout(n),n=s()})}function Ru(e){const t=Fs("",1e3),r=(o=>o.focus()),n=(()=>document.activeElement);function i(o,a){var m,b;if(!a.length)return;t.current=t.current+o;const c=n(),u=((b=(m=a.find(y=>y===c))==null?void 0:m.textContent)==null?void 0:b.trim())??"",d=a.map(y=>{var A;return((A=y.textContent)==null?void 0:A.trim())??""}),f=_s(d,t.current,u),p=a.find(y=>{var A;return((A=y.textContent)==null?void 0:A.trim())===f});return p&&r(p),p}function s(){t.current=""}return{search:t,handleTypeaheadSearch:i,resetTypeahead:s}}function Du(e,t){$(t,!0),Mu({id:O.with(()=>t.id),virtualEl:O.with(()=>t.virtualEl)});var r=B(),n=I(r);J(n,()=>t.children??lt),x(e,r),tt()}function Lu(e,t){$(t,!0);let r=N(t,"side",3,"bottom"),n=N(t,"sideOffset",3,0),i=N(t,"align",3,"center"),s=N(t,"alignOffset",3,0),o=N(t,"arrowPadding",3,0),a=N(t,"avoidCollisions",3,!0),c=N(t,"collisionBoundary",19,()=>[]),u=N(t,"collisionPadding",3,0),d=N(t,"hideWhenDetached",3,!1),f=N(t,"onPlaced",3,()=>{}),p=N(t,"sticky",3,"partial"),m=N(t,"updatePositionStrategy",3,"optimized"),b=N(t,"strategy",3,"fixed"),y=N(t,"dir",3,"ltr"),A=N(t,"style",19,()=>({})),_=N(t,"wrapperId",19,Re),S=N(t,"customAnchor",3,null);const P=Cu({side:O.with(()=>r()),sideOffset:O.with(()=>n()),align:O.with(()=>i()),alignOffset:O.with(()=>s()),id:O.with(()=>t.id),arrowPadding:O.with(()=>o()),avoidCollisions:O.with(()=>a()),collisionBoundary:O.with(()=>c()),collisionPadding:O.with(()=>u()),hideWhenDetached:O.with(()=>d()),onPlaced:O.with(()=>f()),sticky:O.with(()=>p()),updatePositionStrategy:O.with(()=>m()),strategy:O.with(()=>b()),dir:O.with(()=>y()),style:O.with(()=>A()),enabled:O.with(()=>t.enabled),wrapperId:O.with(()=>_()),customAnchor:O.with(()=>S())}),C=Y(()=>pe(P.wrapperProps,{style:{pointerEvents:"auto"}}));var M=B(),K=I(M);J(K,()=>t.content??lt,()=>({props:P.props,wrapperProps:l(C)})),x(e,M),tt()}function ju(e,t){$(t,!0),qs(()=>{var i;(i=t.onPlaced)==null||i.call(t)});var r=B(),n=I(r);J(n,()=>t.content??lt,()=>({props:{},wrapperProps:{}})),x(e,r),tt()}function Bu(e,t){let r=N(t,"isStatic",3,!1),n=dt(t,["$$slots","$$events","$$legacy","content","isStatic","onPlaced"]);var i=B(),s=I(i);{var o=c=>{ju(c,{get content(){return t.content},get onPlaced(){return t.onPlaced}})},a=c=>{Lu(c,vt({get content(){return t.content},get onPlaced(){return t.onPlaced}},()=>n))};H(s,c=>{r()?c(o):c(a,!1)})}x(e,i)}var Wu=F("<!> <!>",1);function Rs(e,t){$(t,!0);let r=N(t,"interactOutsideBehavior",3,"close"),n=N(t,"trapFocus",3,!0),i=N(t,"isValidEvent",3,()=>!1),s=N(t,"customAnchor",3,null),o=N(t,"isStatic",3,!1),a=dt(t,["$$slots","$$events","$$legacy","popper","onEscapeKeydown","escapeKeydownBehavior","preventOverflowTextSelection","id","onPointerDown","onPointerUp","side","sideOffset","align","alignOffset","arrowPadding","avoidCollisions","collisionBoundary","collisionPadding","sticky","hideWhenDetached","updatePositionStrategy","strategy","dir","preventScroll","wrapperId","style","onPlaced","onInteractOutside","onCloseAutoFocus","onOpenAutoFocus","onFocusOutside","interactOutsideBehavior","loop","trapFocus","isValidEvent","customAnchor","isStatic","enabled"]);Bu(e,{get isStatic(){return o()},get id(){return t.id},get side(){return t.side},get sideOffset(){return t.sideOffset},get align(){return t.align},get alignOffset(){return t.alignOffset},get arrowPadding(){return t.arrowPadding},get avoidCollisions(){return t.avoidCollisions},get collisionBoundary(){return t.collisionBoundary},get collisionPadding(){return t.collisionPadding},get sticky(){return t.sticky},get hideWhenDetached(){return t.hideWhenDetached},get updatePositionStrategy(){return t.updatePositionStrategy},get strategy(){return t.strategy},get dir(){return t.dir},get wrapperId(){return t.wrapperId},get style(){return t.style},get onPlaced(){return t.onPlaced},get customAnchor(){return s()},get enabled(){return t.enabled},content:(u,d)=>{let f=()=>d==null?void 0:d().props,p=()=>d==null?void 0:d().wrapperProps;var m=Wu(),b=I(m);{var y=S=>{bo(S,{get preventScroll(){return t.preventScroll}})},A=S=>{var P=B(),C=I(P);{var M=K=>{bo(K,{get preventScroll(){return t.preventScroll}})};H(C,K=>{t.forceMount||K(M)},!0)}x(S,P)};H(b,S=>{t.forceMount&&t.enabled?S(y):S(A,!1)})}var _=E(b,2);{const S=(C,M)=>{let K=()=>M==null?void 0:M().props;Hl(C,{get onEscapeKeydown(){return t.onEscapeKeydown},get escapeKeydownBehavior(){return t.escapeKeydownBehavior},get enabled(){return t.enabled},children:(q,nt)=>{Ll(q,{get id(){return t.id},get onInteractOutside(){return t.onInteractOutside},get onFocusOutside(){return t.onFocusOutside},get interactOutsideBehavior(){return r()},get isValidEvent(){return i()},get enabled(){return t.enabled},children:(rt,D)=>{let R=()=>D==null?void 0:D().props;mc(rt,{get id(){return t.id},get preventOverflowTextSelection(){return t.preventOverflowTextSelection},get onPointerDown(){return t.onPointerDown},get onPointerUp(){return t.onPointerUp},get enabled(){return t.enabled},children:(j,z)=>{var it=B(),at=I(it);{let X=Y(()=>({props:pe(a,f(),R(),K(),{style:{pointerEvents:"auto"}}),wrapperProps:p()}));J(at,()=>t.popper??lt,()=>l(X))}x(j,it)},$$slots:{default:!0}})},$$slots:{default:!0}})},$$slots:{default:!0}})};let P=Y(()=>t.enabled&&n());fc(_,{get id(){return t.id},get onOpenAutoFocus(){return t.onOpenAutoFocus},get onCloseAutoFocus(){return t.onCloseAutoFocus},get loop(){return t.loop},get trapFocus(){return l(P)},get forceMount(){return t.forceMount},focusScope:S,$$slots:{focusScope:!0}})}x(u,m)},$$slots:{content:!0}}),tt()}function Hu(e,t){let r=N(t,"interactOutsideBehavior",3,"close"),n=N(t,"trapFocus",3,!0),i=N(t,"isValidEvent",3,()=>!1),s=N(t,"customAnchor",3,null),o=N(t,"isStatic",3,!1),a=dt(t,["$$slots","$$events","$$legacy","popper","present","onEscapeKeydown","escapeKeydownBehavior","preventOverflowTextSelection","id","onPointerDown","onPointerUp","side","sideOffset","align","alignOffset","arrowPadding","avoidCollisions","collisionBoundary","collisionPadding","sticky","hideWhenDetached","updatePositionStrategy","strategy","dir","preventScroll","wrapperId","style","onPlaced","onInteractOutside","onCloseAutoFocus","onOpenAutoFocus","onFocusOutside","interactOutsideBehavior","loop","trapFocus","isValidEvent","customAnchor","isStatic"]);Ml(e,vt({get id(){return t.id},get present(){return t.present}},()=>a,{presence:u=>{Rs(u,vt({get popper(){return t.popper},get onEscapeKeydown(){return t.onEscapeKeydown},get escapeKeydownBehavior(){return t.escapeKeydownBehavior},get preventOverflowTextSelection(){return t.preventOverflowTextSelection},get id(){return t.id},get onPointerDown(){return t.onPointerDown},get onPointerUp(){return t.onPointerUp},get side(){return t.side},get sideOffset(){return t.sideOffset},get align(){return t.align},get alignOffset(){return t.alignOffset},get arrowPadding(){return t.arrowPadding},get avoidCollisions(){return t.avoidCollisions},get collisionBoundary(){return t.collisionBoundary},get collisionPadding(){return t.collisionPadding},get sticky(){return t.sticky},get hideWhenDetached(){return t.hideWhenDetached},get updatePositionStrategy(){return t.updatePositionStrategy},get strategy(){return t.strategy},get dir(){return t.dir},get preventScroll(){return t.preventScroll},get wrapperId(){return t.wrapperId},get style(){return t.style},get onPlaced(){return t.onPlaced},get customAnchor(){return s()},get isStatic(){return o()},get enabled(){return t.present},get onInteractOutside(){return t.onInteractOutside},get onCloseAutoFocus(){return t.onCloseAutoFocus},get onOpenAutoFocus(){return t.onOpenAutoFocus},get interactOutsideBehavior(){return r()},get loop(){return t.loop},get trapFocus(){return n()},get isValidEvent(){return i()},get onFocusOutside(){return t.onFocusOutside},forceMount:!1},()=>a))},$$slots:{presence:!0}}))}function Ku(e,t){let r=N(t,"interactOutsideBehavior",3,"close"),n=N(t,"trapFocus",3,!0),i=N(t,"isValidEvent",3,()=>!1),s=N(t,"customAnchor",3,null),o=N(t,"isStatic",3,!1),a=dt(t,["$$slots","$$events","$$legacy","popper","onEscapeKeydown","escapeKeydownBehavior","preventOverflowTextSelection","id","onPointerDown","onPointerUp","side","sideOffset","align","alignOffset","arrowPadding","avoidCollisions","collisionBoundary","collisionPadding","sticky","hideWhenDetached","updatePositionStrategy","strategy","dir","preventScroll","wrapperId","style","onPlaced","onInteractOutside","onCloseAutoFocus","onOpenAutoFocus","onFocusOutside","interactOutsideBehavior","loop","trapFocus","isValidEvent","customAnchor","isStatic","enabled"]);Rs(e,vt({get popper(){return t.popper},get onEscapeKeydown(){return t.onEscapeKeydown},get escapeKeydownBehavior(){return t.escapeKeydownBehavior},get preventOverflowTextSelection(){return t.preventOverflowTextSelection},get id(){return t.id},get onPointerDown(){return t.onPointerDown},get onPointerUp(){return t.onPointerUp},get side(){return t.side},get sideOffset(){return t.sideOffset},get align(){return t.align},get alignOffset(){return t.alignOffset},get arrowPadding(){return t.arrowPadding},get avoidCollisions(){return t.avoidCollisions},get collisionBoundary(){return t.collisionBoundary},get collisionPadding(){return t.collisionPadding},get sticky(){return t.sticky},get hideWhenDetached(){return t.hideWhenDetached},get updatePositionStrategy(){return t.updatePositionStrategy},get strategy(){return t.strategy},get dir(){return t.dir},get preventScroll(){return t.preventScroll},get wrapperId(){return t.wrapperId},get style(){return t.style},get onPlaced(){return t.onPlaced},get customAnchor(){return s()},get isStatic(){return o()},get enabled(){return t.enabled},get onInteractOutside(){return t.onInteractOutside},get onCloseAutoFocus(){return t.onCloseAutoFocus},get onOpenAutoFocus(){return t.onOpenAutoFocus},get interactOutsideBehavior(){return r()},get loop(){return t.loop},get trapFocus(){return n()},get isValidEvent(){return i()},get onFocusOutside(){return t.onFocusOutside}},()=>a,{forceMount:!0}))}function ko(e,t){$(t,!0);let r=N(t,"mounted",15,!1),n=N(t,"onMountedChange",3,Ft);fl(()=>(r(!0),n()(!0),()=>{r(!1),n()(!1)})),tt()}const zu=[ns,Fi],Vu=[_r,ml,is],Ds=[Cn,pl,rs],Uu=[...Vu,...Ds];function Io(e){return e.pointerType==="mouse"}function Gu(e){const t=Y(()=>e.enabled()),r=Fs(!1,e.transitTimeout??300,o=>{var a;l(t)&&((a=e.setIsPointerInTransit)==null||a.call(e,o))});let n=V(null);function i(){w(n,null),r.current=!1}function s(o,a){const c=o.currentTarget;if(!Ie(c))return;const u={x:o.clientX,y:o.clientY},d=Yu(u,c.getBoundingClientRect()),f=Xu(u,d),p=qu(a.getBoundingClientRect()),m=Qu([...f,...p]);w(n,m,!0),r.current=!0}return Ht([e.triggerNode,e.contentNode,e.enabled],([o,a,c])=>{if(!o||!a||!c)return;const u=f=>{s(f,a)},d=f=>{s(f,o)};return ke(Nt(o,"pointerleave",u),Nt(a,"pointerleave",d))}),Ht(()=>l(n),()=>Nt(document,"pointermove",a=>{var p,m;if(!l(n))return;const c=a.target;if(!Mn(c))return;const u={x:a.clientX,y:a.clientY},d=((p=e.triggerNode())==null?void 0:p.contains(c))||((m=e.contentNode())==null?void 0:m.contains(c)),f=!Zu(u,l(n));d?i():f&&(i(),e.onPointerExit())})),{isPointerInTransit:r}}function Yu(e,t){const r=Math.abs(t.top-e.y),n=Math.abs(t.bottom-e.y),i=Math.abs(t.right-e.x),s=Math.abs(t.left-e.x);switch(Math.min(r,n,i,s)){case s:return"left";case i:return"right";case r:return"top";case n:return"bottom";default:throw new Error("unreachable")}}function Xu(e,t,r=5){const n=r*1.5;switch(t){case"top":return[{x:e.x-r,y:e.y+r},{x:e.x,y:e.y-n},{x:e.x+r,y:e.y+r}];case"bottom":return[{x:e.x-r,y:e.y-r},{x:e.x,y:e.y+n},{x:e.x+r,y:e.y-r}];case"left":return[{x:e.x+r,y:e.y-r},{x:e.x-n,y:e.y},{x:e.x+r,y:e.y+r}];case"right":return[{x:e.x-r,y:e.y-r},{x:e.x+n,y:e.y},{x:e.x-r,y:e.y+r}]}}function qu(e){const{top:t,right:r,bottom:n,left:i}=e;return[{x:i,y:t},{x:r,y:t},{x:r,y:n},{x:i,y:n}]}function Zu(e,t){const{x:r,y:n}=e;let i=!1;for(let s=0,o=t.length-1;s<t.length;o=s++){const a=t[s].x,c=t[s].y,u=t[o].x,d=t[o].y;c>n!=d>n&&r<(u-a)*(n-c)/(d-c)+a&&(i=!i)}return i}function Qu(e){const t=e.slice();return t.sort((r,n)=>r.x<n.x?-1:r.x>n.x?1:r.y<n.y?-1:r.y>n.y?1:0),Ju(t)}function Ju(e){if(e.length<=1)return e.slice();const t=[];for(let n=0;n<e.length;n++){const i=e[n];for(;t.length>=2;){const s=t[t.length-1],o=t[t.length-2];if((s.x-o.x)*(i.y-o.y)>=(s.y-o.y)*(i.x-o.x))t.pop();else break}t.push(i)}t.pop();const r=[];for(let n=e.length-1;n>=0;n--){const i=e[n];for(;r.length>=2;){const s=r[r.length-1],o=r[r.length-2];if((s.x-o.x)*(i.y-o.y)>=(s.y-o.y)*(i.x-o.x))r.pop();else break}r.push(i)}return r.pop(),t.length===1&&r.length===1&&t[0].x===r[0].x&&t[0].y===r[0].y?t:t.concat(r)}function xr(){return{getShadowRoot:!0,displayCheck:typeof ResizeObserver=="function"&&ResizeObserver.toString().includes("[native code]")?"full":"none"}}function $u(e,t){if(!oi(e,xr()))return td(e,t);const r=oc(ls(e).body,xr());t==="prev"&&r.reverse();const n=r.indexOf(e);return n===-1?document.body:r.slice(n+1)[0]}function td(e,t){if(!lc(e,xr()))return document.body;const r=sc(ls(e).body,xr());t==="prev"&&r.reverse();const n=r.indexOf(e);return n===-1?document.body:r.slice(n+1).find(s=>oi(s,xr()))??document.body}const Gi=new Ze("Menu.Root"),Yi=new Ze("Menu.Root | Menu.Sub"),Ls=new Ze("Menu.Content"),ed=new Ze("Menu.Group | Menu.RadioGroup"),rd=new Di("bitsmenuopen",{bubbles:!1,cancelable:!0});var en,rn;class nd{constructor(t){Q(this,"opts");Q(this,"isUsingKeyboard",new Oi);k(this,en,V(!1));k(this,rn,V(!1));this.opts=t}get ignoreCloseAutoFocus(){return l(h(this,en))}set ignoreCloseAutoFocus(t){w(h(this,en),t,!0)}get isPointerInTransit(){return l(h(this,rn))}set isPointerInTransit(t){w(h(this,rn),t,!0)}getAttr(t){return`data-${this.opts.variant.current}-${t}`}}en=new WeakMap,rn=new WeakMap;var nn,on;class id{constructor(t,r,n){Q(this,"opts");Q(this,"root");Q(this,"parentMenu");Q(this,"contentId",O.with(()=>""));k(this,nn,V(null));k(this,on,V(null));this.opts=t,this.root=r,this.parentMenu=n,n&&Ht(()=>n.opts.open.current,()=>{n.opts.open.current||(this.opts.open.current=!1)})}get contentNode(){return l(h(this,nn))}set contentNode(t){w(h(this,nn),t,!0)}get triggerNode(){return l(h(this,on))}set triggerNode(t){w(h(this,on),t,!0)}toggleOpen(){this.opts.open.current=!this.opts.open.current}onOpen(){this.opts.open.current=!0}onClose(){this.opts.open.current=!1}}nn=new WeakMap,on=new WeakMap;var sn,an,ln,cn,un,He,js,Tn,dn,fn;class od{constructor(t,r){k(this,He);Q(this,"opts");Q(this,"parentMenu");k(this,sn,V(""));k(this,an,0);k(this,ln);Q(this,"rovingFocusGroup");k(this,cn,V(!1));k(this,un);Q(this,"onCloseAutoFocus",t=>{this.opts.onCloseAutoFocus.current(t),!(t.defaultPrevented||h(this,un))&&this.parentMenu.triggerNode&&oi(this.parentMenu.triggerNode)&&this.parentMenu.triggerNode.focus()});Q(this,"onOpenAutoFocus",t=>{if(t.defaultPrevented)return;t.preventDefault();const r=this.parentMenu.contentNode;r==null||r.focus()});k(this,dn,Y(()=>({open:this.parentMenu.opts.open.current})));k(this,fn,Y(()=>({id:this.opts.id.current,role:"menu","aria-orientation":"vertical",[this.parentMenu.root.getAttr("content")]:"","data-state":ts(this.parentMenu.opts.open.current),onkeydown:this.onkeydown,onblur:this.onblur,onfocus:this.onfocus,dir:this.parentMenu.root.opts.dir.current,style:{pointerEvents:"auto"}})));Q(this,"popperProps",{onCloseAutoFocus:t=>this.onCloseAutoFocus(t)});this.opts=t,this.parentMenu=r,r.contentId=t.id,kt(this,un,t.isSub??!1),this.onkeydown=this.onkeydown.bind(this),this.onblur=this.onblur.bind(this),this.onfocus=this.onfocus.bind(this),this.handleInteractOutside=this.handleInteractOutside.bind(this),ae({...t,deps:()=>this.parentMenu.opts.open.current,onRefChange:n=>{this.parentMenu.contentNode!==n&&(this.parentMenu.contentNode=n)}}),Gu({contentNode:()=>this.parentMenu.contentNode,triggerNode:()=>this.parentMenu.triggerNode,enabled:()=>{var n;return this.parentMenu.opts.open.current&&!!((n=this.parentMenu.triggerNode)!=null&&n.hasAttribute(this.parentMenu.root.getAttr("sub-trigger")))},onPointerExit:()=>{this.parentMenu.opts.open.current=!1},setIsPointerInTransit:n=>{this.parentMenu.root.isPointerInTransit=n}}),kt(this,ln,Ru().handleTypeaheadSearch),this.rovingFocusGroup=Ol({rootNodeId:this.parentMenu.contentId,candidateAttr:this.parentMenu.root.getAttr("item"),loop:this.opts.loop,orientation:O.with(()=>"vertical")}),Ht(()=>this.parentMenu.contentNode,n=>{if(!n)return;const i=()=>{oe(()=>{this.parentMenu.root.isUsingKeyboard.current&&this.rovingFocusGroup.focusFirstCandidate()})};return rd.listen(n,i)}),St(()=>{this.parentMenu.opts.open.current||window.clearTimeout(h(this,an))})}get search(){return l(h(this,sn))}set search(t){w(h(this,sn),t,!0)}get mounted(){return l(h(this,cn))}set mounted(t){w(h(this,cn),t,!0)}handleTabKeyDown(t){let r=this.parentMenu;for(;r.parentMenu!==null;)r=r.parentMenu;if(!r.triggerNode)return;t.preventDefault();const n=$u(r.triggerNode,t.shiftKey?"prev":"next");n?(this.parentMenu.root.ignoreCloseAutoFocus=!0,r.onClose(),oe(()=>{n.focus(),oe(()=>{this.parentMenu.root.ignoreCloseAutoFocus=!1})})):document.body.focus()}onkeydown(t){var u,d;if(t.defaultPrevented)return;if(t.key===os){this.handleTabKeyDown(t);return}const r=t.target,n=t.currentTarget;if(!Ie(r)||!Ie(n))return;const i=((u=r.closest(`[${this.parentMenu.root.getAttr("content")}]`))==null?void 0:u.id)===this.parentMenu.contentId.current,s=t.ctrlKey||t.altKey||t.metaKey,o=t.key.length===1;if(this.rovingFocusGroup.handleKeydown(r,t)||t.code==="Space")return;const c=Yt(this,He,js).call(this);i&&!s&&o&&h(this,ln).call(this,t.key,c),((d=t.target)==null?void 0:d.id)===this.parentMenu.contentId.current&&Uu.includes(t.key)&&(t.preventDefault(),Ds.includes(t.key)&&c.reverse(),ds(c))}onblur(t){var r,n;Mn(t.currentTarget)&&Mn(t.target)&&((n=(r=t.currentTarget).contains)!=null&&n.call(r,t.target)||(window.clearTimeout(h(this,an)),this.search=""))}onfocus(t){this.parentMenu.root.isUsingKeyboard.current&&oe(()=>this.rovingFocusGroup.focusFirstCandidate())}onItemEnter(){return Yt(this,He,Tn).call(this)}onItemLeave(t){if(t.currentTarget.hasAttribute(this.parentMenu.root.getAttr("sub-trigger"))||Yt(this,He,Tn).call(this)||this.parentMenu.root.isUsingKeyboard.current)return;const r=this.parentMenu.contentNode;r==null||r.focus(),this.rovingFocusGroup.setCurrentTabStopId("")}onTriggerLeave(){return!!Yt(this,He,Tn).call(this)}handleInteractOutside(t){var n;if(!Sl(t.target))return;const r=(n=this.parentMenu.triggerNode)==null?void 0:n.id;if(t.target.id===r){t.preventDefault();return}t.target.closest(`#${r}`)&&t.preventDefault()}get snippetProps(){return l(h(this,dn))}set snippetProps(t){w(h(this,dn),t)}get props(){return l(h(this,fn))}set props(t){w(h(this,fn),t)}}sn=new WeakMap,an=new WeakMap,ln=new WeakMap,cn=new WeakMap,un=new WeakMap,He=new WeakSet,js=function(){const t=this.parentMenu.contentNode;return t?Array.from(t.querySelectorAll(`[${this.parentMenu.root.getAttr("item")}]:not([data-disabled])`)):[]},Tn=function(){return this.parentMenu.root.isPointerInTransit},dn=new WeakMap,fn=new WeakMap;var ar,hn;class sd{constructor(t,r){Q(this,"opts");Q(this,"content");k(this,ar,V(!1));k(this,hn,Y(()=>({id:this.opts.id.current,tabindex:-1,role:"menuitem","aria-disabled":hl(this.opts.disabled.current),"data-disabled":es(this.opts.disabled.current),"data-highlighted":l(h(this,ar))?"":void 0,[this.content.parentMenu.root.getAttr("item")]:"",onpointermove:this.onpointermove,onpointerleave:this.onpointerleave,onfocus:this.onfocus,onblur:this.onblur})));this.opts=t,this.content=r,this.onpointermove=this.onpointermove.bind(this),this.onpointerleave=this.onpointerleave.bind(this),this.onfocus=this.onfocus.bind(this),this.onblur=this.onblur.bind(this),ae({...t,deps:()=>this.content.mounted})}onpointermove(t){if(!t.defaultPrevented&&Io(t))if(this.opts.disabled.current)this.content.onItemLeave(t);else{if(this.content.onItemEnter())return;const n=t.currentTarget;if(!Ie(n))return;n.focus()}}onpointerleave(t){t.defaultPrevented||Io(t)&&this.content.onItemLeave(t)}onfocus(t){oe(()=>{t.defaultPrevented||this.opts.disabled.current||w(h(this,ar),!0)})}onblur(t){oe(()=>{t.defaultPrevented||w(h(this,ar),!1)})}get props(){return l(h(this,hn))}set props(t){w(h(this,hn),t)}}ar=new WeakMap,hn=new WeakMap;var vn,ei,Bs,gn;class ad{constructor(t,r){k(this,ei);Q(this,"opts");Q(this,"item");k(this,vn,!1);Q(this,"root");k(this,gn,Y(()=>pe(this.item.props,{onclick:this.onclick,onpointerdown:this.onpointerdown,onpointerup:this.onpointerup,onkeydown:this.onkeydown})));this.opts=t,this.item=r,this.root=r.content.parentMenu.root,this.onkeydown=this.onkeydown.bind(this),this.onclick=this.onclick.bind(this),this.onpointerdown=this.onpointerdown.bind(this),this.onpointerup=this.onpointerup.bind(this)}onkeydown(t){const r=this.item.content.search!=="";if(!(this.item.opts.disabled.current||r&&t.key===Fi)&&zu.includes(t.key)){if(!Ie(t.currentTarget))return;t.currentTarget.click(),t.preventDefault()}}onclick(t){this.item.opts.disabled.current||Yt(this,ei,Bs).call(this)}onpointerup(t){var r;if(!t.defaultPrevented&&!h(this,vn)){if(!Ie(t.currentTarget))return;(r=t.currentTarget)==null||r.click()}}onpointerdown(t){kt(this,vn,!0)}get props(){return l(h(this,gn))}set props(t){w(h(this,gn),t)}}vn=new WeakMap,ei=new WeakSet,Bs=function(){if(this.item.opts.disabled.current)return;const t=new CustomEvent("menuitemselect",{bubbles:!0,cancelable:!0});this.opts.onSelect.current(t),oe(()=>{if(t.defaultPrevented){this.item.content.parentMenu.root.isUsingKeyboard.current=!1;return}this.opts.closeOnSelect.current&&this.item.content.parentMenu.root.opts.onClose()})},gn=new WeakMap;var pn,mn;class ld{constructor(t,r){Q(this,"opts");Q(this,"root");k(this,pn,V(void 0));k(this,mn,Y(()=>({id:this.opts.id.current,role:"group","aria-labelledby":this.groupHeadingId,[this.root.getAttr("group")]:""})));this.opts=t,this.root=r,ae(this.opts)}get groupHeadingId(){return l(h(this,pn))}set groupHeadingId(t){w(h(this,pn),t,!0)}get props(){return l(h(this,mn))}set props(t){w(h(this,mn),t)}}pn=new WeakMap,mn=new WeakMap;var bn;class cd{constructor(t,r){Q(this,"opts");Q(this,"root");k(this,bn,Y(()=>({id:this.opts.id.current,role:"group",[this.root.getAttr("separator")]:""})));this.opts=t,this.root=r,ae(t)}get props(){return l(h(this,bn))}set props(t){w(h(this,bn),t)}}bn=new WeakMap;var ri,yn;class ud{constructor(t,r){Q(this,"opts");Q(this,"parentMenu");k(this,ri,Y(()=>{if(this.parentMenu.opts.open.current&&this.parentMenu.contentId.current)return this.parentMenu.contentId.current}));k(this,yn,Y(()=>({id:this.opts.id.current,disabled:this.opts.disabled.current,"aria-haspopup":"menu","aria-expanded":vl(this.parentMenu.opts.open.current),"aria-controls":l(h(this,ri)),"data-disabled":es(this.opts.disabled.current),"data-state":ts(this.parentMenu.opts.open.current),[this.parentMenu.root.getAttr("trigger")]:"",onpointerdown:this.onpointerdown,onpointerup:this.onpointerup,onkeydown:this.onkeydown})));this.opts=t,this.parentMenu=r,this.onpointerdown=this.onpointerdown.bind(this),this.onpointerup=this.onpointerup.bind(this),this.onkeydown=this.onkeydown.bind(this),ae({...t,onRefChange:n=>{this.parentMenu.triggerNode=n}})}onpointerdown(t){if(!this.opts.disabled.current){if(t.pointerType==="touch")return t.preventDefault();t.button===0&&t.ctrlKey===!1&&(this.parentMenu.toggleOpen(),this.parentMenu.opts.open.current||t.preventDefault())}}onpointerup(t){this.opts.disabled.current||t.pointerType==="touch"&&(t.preventDefault(),this.parentMenu.toggleOpen())}onkeydown(t){if(!this.opts.disabled.current){if(t.key===Fi||t.key===ns){this.parentMenu.toggleOpen(),t.preventDefault();return}t.key===_r&&(this.parentMenu.onOpen(),t.preventDefault())}}get props(){return l(h(this,yn))}set props(t){w(h(this,yn),t)}}ri=new WeakMap,yn=new WeakMap;function dd(e){const t=new nd(e);return xs.set({get ignoreCloseAutoFocus(){return t.ignoreCloseAutoFocus}}),Gi.set(t)}function fd(e,t){return Yi.set(new id(t,e,null))}function hd(e){return new ud(e,Yi.get())}function vd(e){return Ls.set(new od(e,Yi.get()))}function gd(e){const t=new sd(e,Ls.get());return new ad(e,t)}function pd(e){return ed.set(new ld(e,Gi.get()))}function md(e){return new cd(e,Gi.get())}var bd=F("<div><!></div>");function yd(e,t){$(t,!0);let r=N(t,"ref",15,null),n=N(t,"id",19,Re),i=N(t,"disabled",3,!1),s=N(t,"onSelect",3,Ft),o=N(t,"closeOnSelect",3,!0),a=dt(t,["$$slots","$$events","$$legacy","child","children","ref","id","disabled","onSelect","closeOnSelect"]);const c=gd({id:O.with(()=>n()),disabled:O.with(()=>i()),onSelect:O.with(()=>s()),ref:O.with(()=>r(),b=>r(b)),closeOnSelect:O.with(()=>o())}),u=Y(()=>pe(a,c.props));var d=B(),f=I(d);{var p=b=>{var y=B(),A=I(y);J(A,()=>t.child,()=>({props:l(u)})),x(b,y)},m=b=>{var y=bd();We(y,()=>({...l(u)}));var A=g(y);J(A,()=>t.children??lt),v(y),x(b,y)};H(f,b=>{t.child?b(p):b(m,!1)})}x(e,d),tt()}var xd=F("<div><!></div>");function wd(e,t){$(t,!0);let r=N(t,"ref",15,null),n=N(t,"id",19,Re),i=dt(t,["$$slots","$$events","$$legacy","children","child","ref","id"]);const s=pd({id:O.with(()=>n()),ref:O.with(()=>r(),f=>r(f))}),o=Y(()=>pe(i,s.props));var a=B(),c=I(a);{var u=f=>{var p=B(),m=I(p);J(m,()=>t.child,()=>({props:l(o)})),x(f,p)},d=f=>{var p=xd();We(p,()=>({...l(o)}));var m=g(p);J(m,()=>t.children??lt),v(p),x(f,p)};H(c,f=>{t.child?f(u):f(d,!1)})}x(e,a),tt()}var _d=F("<div><!></div>");function Sd(e,t){$(t,!0);let r=N(t,"ref",15,null),n=N(t,"id",19,Re),i=dt(t,["$$slots","$$events","$$legacy","ref","id","child","children"]);const s=md({id:O.with(()=>n()),ref:O.with(()=>r(),f=>r(f))}),o=Y(()=>pe(i,s.props));var a=B(),c=I(a);{var u=f=>{var p=B(),m=I(p);J(m,()=>t.child,()=>({props:l(o)})),x(f,p)},d=f=>{var p=_d();We(p,()=>({...l(o)}));var m=g(p);J(m,()=>t.children??lt),v(p),x(f,p)};H(c,f=>{t.child?f(u):f(d,!1)})}x(e,a),tt()}function Ad(e,t){$(t,!0);let r=N(t,"open",15,!1),n=N(t,"dir",3,"ltr"),i=N(t,"onOpenChange",3,Ft),s=N(t,"_internal_variant",3,"dropdown-menu");const o=dd({variant:O.with(()=>s()),dir:O.with(()=>n()),onClose:()=>{r(!1),i()(!1)}});fd(o,{open:O.with(()=>r(),a=>{r(a),i()(a)})}),Fu(e,{children:(a,c)=>{var u=B(),d=I(u);J(d,()=>t.children??lt),x(a,u)},$$slots:{default:!0}}),tt()}var Pd=F("<div><div><!></div></div>"),Ed=F("<!> <!>",1),Od=F("<div><div><!></div></div>"),Td=F("<!> <!>",1);function Cd(e,t){$(t,!0);let r=N(t,"id",19,Re),n=N(t,"ref",15,null),i=N(t,"loop",3,!0),s=N(t,"onInteractOutside",3,Ft),o=N(t,"onEscapeKeydown",3,Ft),a=N(t,"onCloseAutoFocus",3,Ft),c=N(t,"forceMount",3,!1),u=dt(t,["$$slots","$$events","$$legacy","id","child","children","ref","loop","onInteractOutside","onEscapeKeydown","onCloseAutoFocus","forceMount"]);const d=vd({id:O.with(()=>r()),loop:O.with(()=>i()),ref:O.with(()=>n(),S=>n(S)),onCloseAutoFocus:O.with(()=>a())}),f=Y(()=>pe(u,d.props));function p(S){d.handleInteractOutside(S),!S.defaultPrevented&&(s()(S),!S.defaultPrevented&&d.parentMenu.onClose())}function m(S){o()(S),!S.defaultPrevented&&d.parentMenu.onClose()}var b=B(),y=I(b);{var A=S=>{Ku(S,vt(()=>l(f),()=>d.popperProps,{get enabled(){return d.parentMenu.opts.open.current},onInteractOutside:p,onEscapeKeydown:m,trapFocus:!0,get loop(){return i()},forceMount:!0,get id(){return r()},popper:(C,M)=>{let K=()=>M==null?void 0:M().props,q=()=>M==null?void 0:M().wrapperProps;const nt=Y(()=>pe(K(),{style:Mo("dropdown-menu")}));var st=Ed(),rt=I(st);{var D=z=>{var it=B(),at=I(it);{let X=Y(()=>({props:l(nt),wrapperProps:q(),...d.snippetProps}));J(at,()=>t.child,()=>l(X))}x(z,it)},R=z=>{var it=Pd();We(it,()=>({...q()}));var at=g(it);We(at,()=>({...l(nt)}));var X=g(at);J(X,()=>t.children??lt),v(at),v(it),x(z,it)};H(rt,z=>{t.child?z(D):z(R,!1)})}var j=E(rt,2);ko(j,{get mounted(){return d.mounted},set mounted(z){d.mounted=z}}),x(C,st)},$$slots:{popper:!0}}))},_=S=>{var P=B(),C=I(P);{var M=K=>{Hu(K,vt(()=>l(f),()=>d.popperProps,{get present(){return d.parentMenu.opts.open.current},onInteractOutside:p,onEscapeKeydown:m,trapFocus:!0,get loop(){return i()},forceMount:!1,get id(){return r()},popper:(nt,st)=>{let rt=()=>st==null?void 0:st().props,D=()=>st==null?void 0:st().wrapperProps;const R=Y(()=>pe(rt(),{style:Mo("dropdown-menu")}));var j=Td(),z=I(j);{var it=U=>{var et=B(),ht=I(et);{let pt=Y(()=>({props:l(R),wrapperProps:D(),...d.snippetProps}));J(ht,()=>t.child,()=>l(pt))}x(U,et)},at=U=>{var et=Od();We(et,()=>({...D()}));var ht=g(et);We(ht,()=>({...l(R)}));var pt=g(ht);J(pt,()=>t.children??lt),v(ht),v(et),x(U,et)};H(z,U=>{t.child?U(it):U(at,!1)})}var X=E(z,2);ko(X,{get mounted(){return d.mounted},set mounted(U){d.mounted=U}}),x(nt,j)},$$slots:{popper:!0}}))};H(C,K=>{c()||K(M)},!0)}x(S,P)};H(y,S=>{c()?S(A):S(_,!1)})}x(e,b),tt()}var Md=F("<button><!></button>");function kd(e,t){$(t,!0);let r=N(t,"id",19,Re),n=N(t,"ref",15,null),i=N(t,"disabled",3,!1),s=N(t,"type",3,"button"),o=dt(t,["$$slots","$$events","$$legacy","id","ref","child","children","disabled","type"]);const a=hd({id:O.with(()=>r()),disabled:O.with(()=>i()??!1),ref:O.with(()=>n(),u=>n(u))}),c=Y(()=>pe(o,a.props,{type:s()}));Du(e,{get id(){return r()},children:(u,d)=>{var f=B(),p=I(f);{var m=y=>{var A=B(),_=I(A);J(_,()=>t.child,()=>({props:l(c)})),x(y,A)},b=y=>{var A=Md();We(A,()=>({...l(c)}));var _=g(A);J(_,()=>t.children??lt),v(A),x(y,A)};H(p,y=>{t.child?y(m):y(b,!1)})}x(u,f)},$$slots:{default:!0}}),tt()}let pr=V(!1);const re=class re{constructor(){St(()=>(re._refs===0&&(re._cleanup=Uo(()=>{const t=[],r=i=>{w(pr,!1)},n=i=>{w(pr,!0)};return t.push(Nt(document,"pointerdown",r,{capture:!0}),Nt(document,"pointermove",r,{capture:!0}),Nt(document,"keydown",n,{capture:!0})),ke(...t)})),re._refs++,()=>{var t;re._refs--,re._refs===0&&(w(pr,!1),(t=re._cleanup)==null||t.call(re))}))}get current(){return l(pr)}set current(t){w(pr,t,!0)}};Q(re,"_refs",0),Q(re,"_cleanup");let Oi=re;const No=Ad,Fo=kd,Ro=Cd,mr=yd,Id=Sd,Do=wd;var Nd=F('<div class="w-5 h-5 rounded bg-gray-800 flex items-center justify-center text-[10px] text-gray-400 border border-gray-700"><!></div> <span class="font-semibold text-sm tracking-tight text-gray-200">All Projects</span> <!>',1),Fd=F('<div class="w-6 h-6 rounded bg-gray-800 border border-gray-700 flex items-center justify-center shrink-0"><span><i></i></span></div> <div class="flex-1 min-w-0"><div class="text-sm text-gray-400 group-hover:text-white truncate transition"> </div></div> <!>',1),Rd=F('<div class="px-4 py-8 text-center text-gray-600 text-xs">No projects found.</div>'),Dd=F('<div class="p-3 border-b border-gray-700 bg-popover"><div class="relative"><!> <input type="text" placeholder="Filter repositories..." class="w-full bg-gray-900 border border-gray-700 rounded-lg pl-8 pr-3 py-1.5 text-xs text-white focus:outline-none focus:border-blue-500 placeholder-gray-600"/></div></div> <div class="max-h-[320px] overflow-y-auto py-1"><!> <!> <!></div> <div class="px-3 py-2 bg-gray-900/50 border-t border-gray-800 text-[10px] text-gray-500 flex justify-between"><span> </span> <button class="hover:text-blue-400 cursor-pointer flex items-center gap-1"><!> New</button></div>',1),Ld=F("<!> <!>",1),jd=F('<div class="h-2 w-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.4)]"></div> <div class="w-6 h-6 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center text-[10px] font-bold text-gray-300"> </div>',1),Bd=F("<!> Settings",1),Wd=F("<!> Install App",1),Hd=F("<!> <!>",1),Kd=F("<!> Sign Out",1),zd=F('<div class="px-4 py-3 border-b border-gray-800 mb-1"><p class="text-[10px] text-gray-500 uppercase tracking-wider font-bold">Signed in as</p> <p class="text-xs font-medium text-gray-200 mt-1 truncate"> </p></div> <!> <!> <!>',1),Vd=F("<!> <!>",1),Ud=F('<header class="h-14 border-b border-linear-border bg-gray-900/80 backdrop-blur-md flex items-center justify-between px-4 z-20 shrink-0 fixed top-0 left-0 right-0"><!> <!></header>');function Gd(e,t){$(t,!0);let r=V("");const n=Y(()=>L.projects.filter(c=>c.name.toLowerCase().includes(l(r).toLowerCase())));async function i(){await L.logout()}var s=Ud(),o=g(s);Zt(o,()=>No,(c,u)=>{u(c,{get open(){return L.projectMenuOpen},set open(d){L.projectMenuOpen=d},children:(d,f)=>{var p=Ld(),m=I(p);Zt(m,()=>Fo,(y,A)=>{A(y,{class:"flex items-center gap-2 cursor-pointer active:opacity-70 transition",children:(_,S)=>{var P=Nd(),C=I(P),M=g(C);xa(M,{class:"w-3 h-3"}),v(C);var K=E(C,4);Xo(K,{class:"w-2.5 h-2.5 text-gray-600"}),x(_,P)},$$slots:{default:!0}})});var b=E(m,2);Zt(b,()=>Ro,(y,A)=>{A(y,{class:"w-72 bg-popover border border-gray-700 rounded-xl shadow-[0_0_50px_rgba(0,0,0,0.5)] overflow-hidden flex flex-col mt-2",sideOffset:8,children:(_,S)=>{var P=Dd(),C=I(P),M=g(C),K=g(M);xi(K,{class:"absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 w-3 h-3"});var q=E(K,2);Go(q),v(M),v(C);var nt=E(C,2),st=g(nt);Zt(st,()=>mr,(U,et)=>{et(U,{class:"w-full px-4 py-2 hover:bg-white/5 cursor-pointer text-sm font-bold text-white border-b border-gray-800/50 mb-1 text-left focus:bg-white/5 outline-none",children:(ht,pt)=>{It();var ft=Xs("All Projects");x(ht,ft)},$$slots:{default:!0}})});var rt=E(st,2);he(rt,17,()=>l(n),ve,(U,et)=>{var ht=B(),pt=I(ht);{let ft=Y(()=>ie("w-full px-4 py-2 hover:bg-white/5 cursor-pointer flex items-center gap-3 group transition text-left focus:bg-white/5 outline-none",L.activeProjectId===l(et).id&&"bg-white/5"));Zt(pt,()=>mr,(Ct,Vt)=>{Vt(Ct,{onSelect:()=>L.setActiveProject(l(et).id,l(et).name),get class(){return l(ft)},children:(Kt,T)=>{var W=Fd(),ot=I(W),bt=g(ot),Pt=g(bt);v(bt),v(ot);var ut=E(ot,2),Et=g(ut),Dt=g(Et,!0);v(Et),v(ut);var Rt=E(ut,2);{var Lt=mt=>{Ti(mt,{class:"w-3 h-3 text-green-500"})};H(Rt,mt=>{L.activeProjectId===l(et).id&&mt(Lt)})}yt(()=>{zt(bt,1,`text-[10px] ${l(et).color??""}`),zt(Pt,1,`fas ${l(et).icon??""}`),_t(Dt,l(et).name)}),x(Kt,W)},$$slots:{default:!0}})})}x(U,ht)});var D=E(rt,2);{var R=U=>{var et=Rd();x(U,et)};H(D,U=>{l(n).length===0&&U(R)})}v(nt);var j=E(nt,2),z=g(j),it=g(z);v(z);var at=E(z,2),X=g(at);Aa(X,{class:"w-2.5 h-2.5"}),It(),v(at),v(j),yt(()=>_t(it,`${L.projects.length??""} Repositories`)),yi(q,()=>l(r),U=>w(r,U)),x(_,P)},$$slots:{default:!0}})}),x(d,p)},$$slots:{default:!0}})});var a=E(o,2);Zt(a,()=>No,(c,u)=>{u(c,{children:(d,f)=>{var p=Vd(),m=I(p);Zt(m,()=>Fo,(y,A)=>{A(y,{class:"flex items-center gap-3 cursor-pointer hover:opacity-80 transition p-1",children:(_,S)=>{var P=jd(),C=E(I(P),2),M=g(C,!0);v(C),yt(K=>_t(M,K),[()=>pa(L.userEmail)]),x(_,P)},$$slots:{default:!0}})});var b=E(m,2);Zt(b,()=>Ro,(y,A)=>{A(y,{class:"w-56 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden py-1",align:"end",sideOffset:8,children:(_,S)=>{var P=zd(),C=I(P),M=E(g(C),2),K=g(M,!0);v(M),v(C);var q=E(C,2);Zt(q,()=>Do,(rt,D)=>{D(rt,{class:"px-2",children:(R,j)=>{var z=Hd(),it=I(z);Zt(it,()=>mr,(U,et)=>{et(U,{onSelect:()=>L.settingsOpen=!0,class:"w-full px-2 py-1.5 hover:bg-white/5 rounded text-xs text-gray-400 flex items-center gap-2 transition-colors text-left cursor-pointer focus:bg-white/5 outline-none",children:(ht,pt)=>{var ft=Bd(),Ct=I(ft);wa(Ct,{class:"w-4 h-4"}),It(),x(ht,ft)},$$slots:{default:!0}})});var at=E(it,2);{var X=U=>{var et=B(),ht=I(et);Zt(ht,()=>mr,(pt,ft)=>{ft(pt,{onSelect:()=>L.installPWA(),class:"w-full px-2 py-1.5 hover:bg-purple-500/10 rounded text-xs text-purple-400 hover:text-purple-300 flex items-center gap-2 transition-colors text-left cursor-pointer focus:bg-purple-500/10 outline-none",children:(Ct,Vt)=>{var Kt=Wd(),T=I(Kt);_a(T,{class:"w-4 h-4"}),It(),x(Ct,Kt)},$$slots:{default:!0}})}),x(U,et)};H(at,U=>{L.canInstallPWA&&U(X)})}x(R,z)},$$slots:{default:!0}})});var nt=E(q,2);Zt(nt,()=>Id,(rt,D)=>{D(rt,{class:"h-px bg-gray-800 my-1 mx-2"})});var st=E(nt,2);Zt(st,()=>Do,(rt,D)=>{D(rt,{class:"px-2 pb-1",children:(R,j)=>{var z=B(),it=I(z);Zt(it,()=>mr,(at,X)=>{X(at,{onSelect:i,class:"w-full px-2 py-1.5 hover:bg-red-500/10 rounded text-xs text-red-400 hover:text-red-300 flex items-center gap-2 transition-colors text-left cursor-pointer focus:bg-red-500/10 outline-none",children:(U,et)=>{var ht=Kd(),pt=I(ht);Sa(pt,{class:"w-4 h-4"}),It(),x(U,ht)},$$slots:{default:!0}})}),x(R,z)},$$slots:{default:!0}})}),yt(()=>_t(K,L.userEmail)),x(_,P)},$$slots:{default:!0}})}),x(d,p)},$$slots:{default:!0}})}),v(s),x(e,s),tt()}function Yd(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M21.801 10A10 10 0 1 1 17 3.335"}],["path",{d:"m9 11 3 3L22 4"}]];At(e,vt({name:"circle-check-big"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}var Xd=F("<div><!> <span> </span></div>");function qd(e,t){$(t,!1),ba();var r=B(),n=I(r);{var i=s=>{var o=Xd();let a;var c=g(o);Yd(c,{class:"w-4 h-4 text-green-500"});var u=E(c,2),d=g(u,!0);v(u),v(o),yt(()=>{a=zt(o,1,"fixed top-6 left-1/2 -translate-x-1/2 z-[60] bg-gray-900 border border-gray-700/50 text-white px-4 py-2 rounded-full shadow-2xl flex items-center gap-3 text-sm font-medium transition-all duration-300",null,a,{"translate-y-0":L.toastOpen,"-translate-y-full":!L.toastOpen,"opacity-100":L.toastOpen,"opacity-0":!L.toastOpen}),_t(d,L.toastMsg)}),x(s,o)};H(n,s=>{L.toastOpen&&s(i)})}x(e,r),tt()}function Zd(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"}]];At(e,vt({name:"folder"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Lo(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"m16 6-8.414 8.586a2 2 0 0 0 2.829 2.829l8.414-8.586a4 4 0 1 0-5.657-5.657l-8.379 8.551a6 6 0 1 0 8.485 8.485l8.379-8.551"}]];At(e,vt({name:"paperclip"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function jo(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z"}]];At(e,vt({name:"zap"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Bo(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M12 19v3"}],["path",{d:"M19 10v2a7 7 0 0 1-14 0v-2"}],["rect",{x:"9",y:"2",width:"6",height:"13",rx:"3"}]];At(e,vt({name:"mic"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Qd(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"m5 12 7-7 7 7"}],["path",{d:"M12 19V5"}]];At(e,vt({name:"arrow-up"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Jd(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z"}],["path",{d:"M14 2v5a1 1 0 0 0 1 1h5"}]];At(e,vt({name:"file"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}var $d=F('<div class="w-1 bg-red-500 rounded-full transition-all duration-75"></div>'),tf=F('<div class="absolute inset-0 bg-secondary rounded-3xl flex items-center px-4 z-10"><button type="button" class="w-8 h-8 rounded-full bg-gray-800 hover:bg-gray-700 flex items-center justify-center text-gray-400 hover:text-white transition shrink-0"><!></button> <div class="flex-1 flex items-center justify-center gap-[3px] h-10 mx-4"></div> <div class="flex items-center gap-2 shrink-0"><div class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></div> <span class="text-sm text-gray-300 font-mono"> </span></div></div>'),ef=F('<div class="absolute inset-0 bg-secondary rounded-3xl flex items-center justify-center z-10"><!> <span class="text-gray-400 text-sm">Transcribing...</span></div>'),rf=F('<button type="button"><!> <span class="truncate"> </span></button>'),nf=F('<div class="absolute bottom-full left-0 mb-2 w-80 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden max-h-48 overflow-y-auto z-50"><div class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-gray-800 flex items-center justify-between"><span>Files</span> <span class="text-gray-600 font-normal normal-case">↑↓ to navigate, Enter to select</span></div> <!></div>'),of=F('<div class="absolute bottom-full left-0 mb-2 w-64 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden z-50"><div class="px-3 py-4 text-xs text-gray-500 text-center"><!> <div>No files found</div></div></div>'),sf=F('<button type="button"><div class="flex items-center gap-3 overflow-hidden"><div class="w-2 h-2 rounded-full bg-blue-500/40 group-hover:bg-blue-500 transition-colors"></div> <span class="text-xs font-medium truncate"> </span></div> <!></button>'),af=F('<div class="px-3 py-1 text-[10px] font-bold text-gray-500 uppercase tracking-widest mb-1">Active Projects</div> <!> <div class="h-px bg-gray-800 my-2 mx-2"></div>',1),lf=F('<button type="button" class="w-full px-3 py-2 rounded-xl flex items-center gap-3 text-gray-400 hover:bg-white/5 hover:text-white group transition-all duration-150"><div class="w-2 h-2 rounded-full bg-gray-700 group-hover:bg-gray-500 transition-colors"></div> <span class="text-xs font-medium truncate"> </span></button>'),cf=F('<div class="px-4 py-8 text-center text-gray-600 text-[11px] italic"> </div>'),uf=F('<div class="absolute bottom-full left-0 mb-3 w-72 bg-popover border border-gray-700 rounded-2xl shadow-[0_0_50px_rgba(0,0,0,0.5)] overflow-hidden flex flex-col z-50 animate-in fade-in slide-in-from-bottom-2 duration-200"><div class="p-3 border-b border-gray-800 bg-gray-900/50"><div class="relative"><!> <input type="text" placeholder="Search repositories..." class="w-full bg-black/40 border border-gray-700 rounded-xl pl-8 pr-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 placeholder-gray-600 transition-all font-medium"/></div></div> <div class="max-h-64 overflow-y-auto py-2 px-1 scrollbar-thin"><!> <div class="px-3 py-1 text-[10px] font-bold text-gray-500 uppercase tracking-widest mb-1">All Repositories</div> <!></div> <div class="px-4 py-2 border-t border-gray-800 bg-gray-900/30"><p class="text-[9px] text-gray-600 text-center font-medium"> </p></div></div>'),df=F('<div class="relative"><button type="button"><div><!></div> <span class="max-w-[120px] truncate"> </span> <!></button> <!></div>'),ff=F('<div class="w-1.5 h-1.5 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]"></div>'),hf=F('<button type="button"><span class="text-xs font-medium"> </span> <!></button>'),vf=F('<div class="absolute bottom-full left-0 mb-3 w-52 bg-popover/95 border border-white/10 rounded-xl shadow-2xl overflow-hidden py-1 z-50 backdrop-blur-xl"><div class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-white/5 mb-1">Select Model</div> <div class="p-1.5 space-y-0.5"></div></div>'),gf=F('<button type="button" aria-label="Close chat" class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"><!></button> <button type="button" aria-label="Attach file" class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"><!></button> <div class="relative"><button type="button" aria-label="Select model" class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"><!></button> <!></div>',1),pf=F('<div class="w-1.5 h-1.5 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]"></div>'),mf=F('<button type="button"><span class="text-xs font-medium"> </span> <!></button>'),bf=F('<div class="absolute bottom-full right-0 mb-3 w-52 bg-popover/95 border border-white/10 rounded-xl shadow-2xl overflow-hidden py-1 z-50 backdrop-blur-xl"><div class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-white/5 mb-1">Select Model</div> <div class="p-1.5 space-y-0.5"></div></div>'),yf=F('<button type="button" aria-label="Attach file" class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"><!></button> <div class="relative"><button type="button" aria-label="Select model" class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"><!></button> <!></div>',1),xf=F('<div class="relative"><div class="bg-secondary/60 border border-white/10 rounded-[24px] shadow-2xl relative backdrop-blur-xl transition-all duration-200 ring-1 ring-white/5 flex flex-col group focus-within:border-white/20 focus-within:ring-white/10"><!> <!> <div class="relative px-4 pt-4 pb-2"><!> <!> <textarea rows="1" class="bg-transparent border-none focus:ring-0 focus:outline-none text-white text-[15px] placeholder-gray-500 w-full resize-none font-medium p-0 leading-relaxed max-h-[40vh] min-h-[24px]"></textarea></div> <div class="flex items-center justify-between px-2 pb-2 mt-1"><div class="flex items-center gap-1"><!></div> <div class="flex items-center gap-1"><!> <div class="w-px h-4 bg-gray-700 mx-1"></div> <button type="button" aria-label="Send message or hold to record voice"><!></button></div></div></div></div>');function Ws(e,t){$(t,!0);let r=N(t,"placeholder",3,"What do you want to build?"),n=V(""),i=V(null),s=V(!1),o=V(!1),a=V(Me([])),c=V(0),u=V("");function d(){if(!l(i))return;l(i).style.height="auto";let T=l(i).scrollHeight;const W=window.innerHeight*(t.mode==="create"?.4:.35);T>W&&(T=W),l(i).style.height=`${T}px`}async function f(T){const W=l(n).match(/@([^ ]*)$/);if(W){w(o,!0);const ot=W[1]||"";await p(ot)}else w(o,!1),w(a,[],!0);T.key==="Escape"&&(w(o,!1),w(a,[],!0))}async function p(T){const W=L.activeProjectId;if(W)try{w(a,await ga.search(W,T),!0),w(c,0)}catch(ot){console.error("File search failed:",ot),w(a,[],!0)}}function m(T){!l(o)||!l(a)||l(a).length===0||(T.key==="ArrowDown"?(T.preventDefault(),w(c,Math.min(l(c)+1,l(a).length-1),!0)):T.key==="ArrowUp"?(T.preventDefault(),w(c,Math.max(l(c)-1,0),!0)):(T.key==="Enter"||T.key==="Tab")&&(T.preventDefault(),l(a)[l(c)]&&b(l(a)[l(c)])))}function b(T){var W;w(n,l(n).replace(/@[^ ]*$/,"")+T+" "),w(o,!1),w(a,[],!0),w(c,0),(W=l(i))==null||W.focus()}function y(T){l(o)&&l(a)&&l(a).length>0&&["ArrowUp","ArrowDown","Tab"].includes(T.key)||l(o)&&l(a)&&l(a).length>0&&T.key==="Enter"&&!T.shiftKey?m(T):!l(o)&&T.key==="Enter"&&!T.shiftKey&&(T.preventDefault(),A())}async function A(){var W;if(!l(n).trim())return;if(t.mode==="create"&&!L.activeProjectId){L.showToast("Select a project first");return}const T=l(n).trim();if(w(n,""),l(i)&&(l(i).style.height="auto"),t.onSubmit)t.onSubmit(T,L.activeModelId);else if(t.mode==="create")try{await Pe.create(T,L.activeProjectId,L.activeModelId)}catch(ot){console.error("Failed to create task:",ot),L.showToast("Failed to create task")}else if(t.mode==="chat"&&t.taskId)try{await Pe.chat(t.taskId,T,L.activeModelId)}catch(ot){console.error("Failed to send message:",ot),L.showToast("Failed to send message")}(W=navigator.vibrate)==null||W.call(navigator,30)}function _(){var T;l(n).length>0&&((T=navigator.vibrate)==null||T.call(navigator,30),A())}var S=xf(),P=g(S),C=g(P);{var M=T=>{var W=tf(),ot=g(W),bt=g(ot);bi(bt,{class:"w-4 h-4"}),v(ot);var Pt=E(ot,2);he(Pt,21,()=>L.audioLevels,ve,(Rt,Lt)=>{var mt=$d();yt(jt=>Yo(mt,`height: ${jt??""}px; opacity: ${.4+l(Lt)/100*.6}`),[()=>Math.max(4,l(Lt)*.4)]),x(Rt,mt)}),v(Pt);var ut=E(Pt,2),Et=E(g(ut),2),Dt=g(Et);v(Et),v(ut),v(W),yt((Rt,Lt)=>_t(Dt,`${Rt??""}:${Lt??""}`),[()=>Math.floor(L.recordedDuration/60),()=>String(L.recordedDuration%60).padStart(2,"0")]),x(T,W)};H(C,T=>{L.isRecording&&T(M)})}var K=E(C,2);{var q=T=>{var W=ef(),ot=g(W);ya(ot,{class:"w-4 h-4 text-blue-400 mr-2 animate-spin"}),It(2),v(W),x(T,W)};H(K,T=>{L.isTranscribing&&T(q)})}var nt=E(K,2),st=g(nt);{var rt=T=>{var W=nf(),ot=E(g(W),2);he(ot,17,()=>l(a),ve,(bt,Pt,ut)=>{var Et=rf();Et.__click=()=>b(l(Pt));var Dt=g(Et);Jd(Dt,{class:"w-2.5 h-2.5 opacity-40"});var Rt=E(Dt,2),Lt=g(Rt,!0);v(Rt),v(Et),yt(mt=>{zt(Et,1,mt),_t(Lt,l(Pt))},[()=>Ae(ie("w-full px-3 py-2 text-xs font-mono cursor-pointer transition flex items-center gap-2 text-left",ut===l(c)?"bg-blue-500/20 text-blue-300":"text-gray-300 hover:bg-white/5"))]),ha("mouseenter",Et,()=>w(c,ut,!0)),x(bt,Et)}),v(W),x(T,W)};H(st,T=>{l(o)&&l(a)&&l(a).length>0&&T(rt)})}var D=E(st,2);{var R=T=>{var W=of(),ot=g(W),bt=g(ot);xi(bt,{class:"w-4 h-4 mx-auto mb-2 opacity-50"}),It(2),v(ot),v(W),x(T,W)};H(D,T=>{l(o)&&l(a)&&l(a).length===0&&T(R)})}var j=E(D,2);da(j),j.__input=d,j.__keyup=f,j.__keydown=y,va(j,T=>w(i,T),()=>l(i)),v(nt);var z=E(nt,2),it=g(z),at=g(it);{var X=T=>{var W=df(),ot=g(W);ot.__click=()=>L.inputProjectMenuOpen=!L.inputProjectMenuOpen;var bt=g(ot),Pt=g(bt);{let mt=Y(()=>ie("w-2.5 h-2.5",L.activeProjectId?"text-blue-400":"text-gray-400"));Zd(Pt,{get class(){return l(mt)}})}v(bt);var ut=E(bt,2),Et=g(ut,!0);v(ut);var Dt=E(ut,2);Xo(Dt,{class:"w-2.5 h-2.5 opacity-50 ml-0.5"}),v(ot);var Rt=E(ot,2);{var Lt=mt=>{var jt=uf(),Ut=g(jt),te=g(Ut),G=g(te);xi(G,{class:"absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 w-3 h-3"});var Z=E(G,2);Go(Z),v(te),v(Ut);var ct=E(Ut,2),xt=g(ct);{var Qt=wt=>{var Bt=af(),Wt=E(I(Bt),2);he(Wt,17,()=>L.projects.filter(Gt=>Gt.name.toLowerCase().includes(l(u).toLowerCase())),ve,(Gt,qt)=>{var ue=sf();ue.__click=()=>L.setActiveProject(l(qt).id,l(qt).name);var Ve=g(ue),ee=E(g(Ve),2),ne=g(ee,!0);v(ee),v(Ve);var xe=E(Ve,2);{var Qe=we=>{Ti(we,{class:"w-3 h-3 text-blue-500"})};H(xe,we=>{L.activeProjectId===l(qt).id&&we(Qe)})}v(ue),yt(we=>{zt(ue,1,we),_t(ne,l(qt).name)},[()=>Ae(ie("w-full px-3 py-2 rounded-xl flex items-center justify-between group transition-all duration-150",L.activeProjectId===l(qt).id?"bg-blue-500/10 text-white":"text-gray-400 hover:bg-white/5 hover:text-white"))]),x(Gt,ue)}),It(2),x(wt,Bt)};H(xt,wt=>{L.projects.length>0&&wt(Qt)})}var Mt=E(xt,4);he(Mt,17,()=>L.repos.filter(wt=>wt.full_name.toLowerCase().includes(l(u).toLowerCase())&&!L.projects.some(Bt=>Bt.name===wt.full_name)),ve,(wt,Bt)=>{var Wt=lf();Wt.__click=()=>L.setActiveProject(l(Bt).id.toString(),l(Bt).full_name);var Gt=E(g(Wt),2),qt=g(Gt,!0);v(Gt),v(Wt),yt(()=>_t(qt,l(Bt).full_name)),x(wt,Wt)},wt=>{var Bt=cf(),Wt=g(Bt);v(Bt),yt(()=>_t(Wt,`No repositories found matching "${l(u)??""}"`)),x(wt,Bt)}),v(ct);var gt=E(ct,2),Ot=g(gt),Xt=g(Ot);v(Ot),v(gt),v(jt),yt(()=>_t(Xt,`${L.repos.length??""} REPOSITORIES SYNCED`)),yi(Z,()=>l(u),wt=>w(u,wt)),x(mt,jt)};H(Rt,mt=>{L.inputProjectMenuOpen&&mt(Lt)})}v(W),yt((mt,jt,Ut)=>{zt(ot,1,mt),zt(bt,1,jt),_t(Et,Ut)},[()=>Ae(ie("flex items-center gap-2 px-2.5 py-1.5 rounded-full text-xs font-semibold transition-all duration-200 border border-transparent",L.activeProjectId?"bg-blue-500/20 text-blue-400 hover:bg-blue-500/30":"bg-gray-800/50 text-gray-400 hover:bg-gray-800")),()=>Ae(ie("flex items-center justify-center w-3.5 h-3.5 rounded-full",L.activeProjectId?"bg-blue-400/20":"bg-gray-700/50")),()=>L.activeProjectName?L.activeProjectName.split("/").pop():"Select project"]),x(T,W)},U=T=>{var W=gf(),ot=I(W);ot.__click=function(...jt){var Ut;(Ut=t.onClose)==null||Ut.apply(this,jt)};var bt=g(ot);bi(bt,{class:"w-4 h-4"}),v(ot);var Pt=E(ot,2),ut=g(Pt);Lo(ut,{class:"w-4 h-4"}),v(Pt);var Et=E(Pt,2),Dt=g(Et);Dt.__click=()=>w(s,!l(s));var Rt=g(Dt);jo(Rt,{class:"w-4 h-4"}),v(Dt);var Lt=E(Dt,2);{var mt=jt=>{var Ut=vf(),te=E(g(Ut),2);he(te,21,()=>io,ve,(G,Z)=>{var ct=hf();ct.__click=()=>{L.setModel(l(Z).id),w(s,!1)};var xt=g(ct),Qt=g(xt,!0);v(xt);var Mt=E(xt,2);{var gt=Ot=>{var Xt=ff();x(Ot,Xt)};H(Mt,Ot=>{L.activeModelId===l(Z).id&&Ot(gt)})}v(ct),yt(Ot=>{zt(ct,1,Ot),_t(Qt,l(Z).name)},[()=>Ae(ie("w-full flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer transition text-left group",L.activeModelId===l(Z).id?"bg-blue-500/10 text-blue-400 border border-blue-500/20":"hover:bg-white/5 text-gray-400 hover:text-white border border-transparent"))]),x(G,ct)}),v(te),v(Ut),x(jt,Ut)};H(Lt,jt=>{l(s)&&jt(mt)})}v(Et),x(T,W)};H(at,T=>{t.mode==="create"?T(X):T(U,!1)})}v(it);var et=E(it,2),ht=g(et);{var pt=T=>{var W=yf(),ot=I(W),bt=g(ot);Lo(bt,{class:"w-4 h-4"}),v(ot);var Pt=E(ot,2),ut=g(Pt);ut.__click=()=>w(s,!l(s));var Et=g(ut);jo(Et,{class:"w-4 h-4"}),v(ut);var Dt=E(ut,2);{var Rt=Lt=>{var mt=bf(),jt=E(g(mt),2);he(jt,21,()=>io,ve,(Ut,te)=>{var G=mf();G.__click=()=>{L.setModel(l(te).id),w(s,!1)};var Z=g(G),ct=g(Z,!0);v(Z);var xt=E(Z,2);{var Qt=Mt=>{var gt=pf();x(Mt,gt)};H(xt,Mt=>{L.activeModelId===l(te).id&&Mt(Qt)})}v(G),yt(Mt=>{zt(G,1,Mt),_t(ct,l(te).name)},[()=>Ae(ie("w-full flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer transition text-left group",L.activeModelId===l(te).id?"bg-blue-500/10 text-blue-400 border border-blue-500/20":"hover:bg-white/5 text-gray-400 hover:text-white border border-transparent"))]),x(Ut,G)}),v(jt),v(mt),x(Lt,mt)};H(Dt,Lt=>{l(s)&&Lt(Rt)})}v(Pt),x(T,W)};H(ht,T=>{t.mode==="create"&&T(pt)})}var ft=E(ht,4);ft.__click=_;var Ct=g(ft);{var Vt=T=>{Qd(T,{class:"w-4 h-4"})},Kt=T=>{var W=B(),ot=I(W);{var bt=ut=>{Bo(ut,{class:"w-4 h-4 animate-pulse"})},Pt=ut=>{Bo(ut,{class:"w-4 h-4"})};H(ot,ut=>{L.isRecording?ut(bt):ut(Pt,!1)},!0)}x(T,W)};H(Ct,T=>{l(n).length>0?T(Vt):T(Kt,!1)})}v(ft),v(et),v(z),v(P),v(S),yt(T=>{mi(j,"placeholder",r()),mi(j,"aria-label",r()),zt(ft,1,T)},[()=>Ae(ie("w-9 h-9 rounded-xl flex items-center justify-center text-base transition-all duration-150 select-none",L.isRecording?"bg-red-500 shadow-[0_0_20px_rgba(239,68,68,0.5)] scale-110":l(n).length>0?"bg-blue-500 text-white shadow-[0_0_15px_rgba(59,130,246,0.4)]":"bg-white/10 text-gray-300 hover:bg-white/15"))]),yi(j,()=>l(n),T=>w(n,T)),x(e,S),tt()}ii(["click","input","keyup","keydown"]);function wf(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M13 5h8"}],["path",{d:"M13 12h8"}],["path",{d:"M13 19h8"}],["path",{d:"m3 17 2 2 4-4"}],["path",{d:"m3 7 2 2 4-4"}]];At(e,vt({name:"list-checks"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function _f(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["rect",{width:"8",height:"4",x:"8",y:"2",rx:"1",ry:"1"}],["path",{d:"M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"}],["path",{d:"M12 11h4"}],["path",{d:"M12 16h4"}],["path",{d:"M8 11h.01"}],["path",{d:"M8 16h.01"}]];At(e,vt({name:"clipboard-list"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}var Sf=F('<span class="text-[11px] text-gray-400 group-hover:text-gray-300 max-w-[120px] truncate transition-colors"> </span>'),Af=F('<span class="text-[11px] text-gray-500">Tasks</span>'),Pf=F('<span class="text-[11px] text-green-400">Done!</span>'),Ef=F('<div class="fixed bottom-24 right-4 z-30"><button class="group flex items-center gap-2 h-9 pl-3 pr-4 bg-popover hover:bg-card border border-white/[0.08] hover:border-purple-500/30 rounded-full shadow-xl shadow-black/40 transition-all"><div class="relative w-5 h-5"><svg class="w-5 h-5 -rotate-90" viewBox="0 0 20 20"><circle cx="10" cy="10" r="8" fill="none" stroke="currentColor" stroke-width="2" class="text-gray-700"></circle><circle cx="10" cy="10" r="8" fill="none" stroke="currentColor" stroke-width="2" class="text-green-500 transition-all duration-300"></circle></svg> <span class="absolute inset-0 flex items-center justify-center text-[8px] font-bold text-white"> </span></div> <!></button></div>'),Of=F('<div class="w-4 h-4 rounded-full bg-green-500/20 flex items-center justify-center"><!></div>'),Tf=F('<div class="w-4 h-4 rounded-full bg-purple-500/20 flex items-center justify-center relative"><div class="w-1.5 h-1.5 rounded-full bg-purple-400 animate-pulse"></div> <div class="absolute inset-0 rounded-full border border-purple-400/50 animate-ping" style="animation-duration: 2s;"></div></div>'),Cf=F('<div class="w-4 h-4 rounded-full border border-gray-600"></div>'),Mf=F('<div><div class="w-5 h-5 shrink-0 flex items-center justify-center mt-0.5"><!></div> <div class="flex-1 min-w-0"><p> </p></div></div>'),kf=F('<div class="py-8 text-center"><div class="w-12 h-12 rounded-full bg-gray-800 flex items-center justify-center mx-auto mb-3"><!></div> <p class="text-sm text-gray-500">No tasks yet</p> <p class="text-xs text-gray-600 mt-1">The agent will create tasks as needed</p></div>'),If=F('<div class="fixed inset-0 z-[200] flex items-start justify-center pt-[15vh] bg-black/60 backdrop-blur-sm"><div class="bg-popover border border-white/[0.08] rounded-2xl w-[380px] max-h-[60vh] shadow-2xl shadow-black/50 overflow-hidden flex flex-col"><div class="px-4 py-3 border-b border-white/[0.06] flex items-center justify-between shrink-0"><div class="flex items-center gap-2"><div class="w-6 h-6 rounded-lg bg-purple-500/10 flex items-center justify-center"><!></div> <h3 class="text-sm font-semibold text-white">Agent Tasks</h3></div> <div class="flex items-center gap-2"><span class="text-[10px] text-gray-500 font-mono"> </span> <button class="w-6 h-6 rounded-lg hover:bg-white/[0.06] flex items-center justify-center text-gray-500 hover:text-gray-300 transition-colors"><!></button></div></div> <div class="h-0.5 bg-gray-800"><div class="h-full bg-gradient-to-r from-purple-500 to-green-500 transition-all duration-300"></div></div> <div class="flex-1 overflow-y-auto p-2 space-y-1"><!> <!></div></div></div>'),Nf=F("<!> <!>",1);function Ff(e,t){$(t,!0);let r=V(!1);var n=Nf(),i=I(n);{var s=c=>{var u=Ef(),d=g(u);d.__click=()=>w(r,!0);var f=g(d),p=g(f),m=E(g(p));v(p);var b=E(p,2),y=g(b);v(b),v(f);var A=E(f,2);{var _=P=>{var C=Sf(),M=g(C,!0);v(C),yt(()=>_t(M,Tt.inProgressTask)),x(P,C)},S=P=>{var C=B(),M=I(C);{var K=nt=>{var st=Af();x(nt,st)},q=nt=>{var st=Pf();x(nt,st)};H(M,nt=>{Tt.completedCount<Tt.todos.length?nt(K):nt(q,!1)},!0)}x(P,C)};H(A,P=>{Tt.inProgressTask?P(_):P(S,!1)})}v(d),v(u),yt(()=>{mi(m,"stroke-dasharray",`${Tt.completedCount/Tt.todos.length*50.26} 50.26`),_t(y,`${Tt.completedCount??""}/${Tt.todos.length??""}`)}),x(c,u)};H(i,c=>{Tt.todos.length>0&&c(s)})}var o=E(i,2);{var a=c=>{var u=If();u.__click=rt=>rt.target===rt.currentTarget&&w(r,!1);var d=g(u),f=g(d),p=g(f),m=g(p),b=g(m);wf(b,{class:"w-3 h-3 text-purple-400"}),v(m),It(2),v(p);var y=E(p,2),A=g(y),_=g(A);v(A);var S=E(A,2);S.__click=()=>w(r,!1);var P=g(S);bi(P,{class:"w-3 h-3"}),v(S),v(y),v(f);var C=E(f,2),M=g(C);v(C);var K=E(C,2),q=g(K);he(q,17,()=>Tt.todos,ve,(rt,D)=>{var R=Mf(),j=g(R),z=g(j);{var it=pt=>{var ft=Of(),Ct=g(ft);Ti(Ct,{class:"w-2 h-2 text-green-400"}),v(ft),x(pt,ft)},at=pt=>{var ft=B(),Ct=I(ft);{var Vt=T=>{var W=Tf();x(T,W)},Kt=T=>{var W=Cf();x(T,W)};H(Ct,T=>{l(D).status==="in_progress"?T(Vt):T(Kt,!1)},!0)}x(pt,ft)};H(z,pt=>{l(D).status==="completed"?pt(it):pt(at,!1)})}v(j);var X=E(j,2),U=g(X);let et;var ht=g(U,!0);v(U),v(X),v(R),yt(pt=>{zt(R,1,pt),et=zt(U,1,"text-sm leading-tight",null,et,{"text-gray-400":l(D).status==="completed","line-through":l(D).status==="completed","text-white":l(D).status==="in_progress","text-gray-300":l(D).status==="pending"}),_t(ht,l(D).status==="in_progress"?l(D).activeForm||l(D).content:l(D).content)},[()=>Ae(ie("flex items-start gap-3 px-3 py-2 rounded-xl transition-colors",l(D).status==="completed"&&"bg-green-500/5",l(D).status==="in_progress"&&"bg-purple-500/10 border border-purple-500/20",l(D).status==="pending"&&"hover:bg-white/[0.02]"))]),x(rt,R)});var nt=E(q,2);{var st=rt=>{var D=kf(),R=g(D),j=g(R);_f(j,{class:"w-5 h-5 text-gray-600"}),v(R),It(4),v(D),x(rt,D)};H(nt,rt=>{Tt.todos.length===0&&rt(st)})}v(K),v(d),v(u),yt(()=>{_t(_,`${Tt.completedCount??""} / ${Tt.todos.length??""} done`),Yo(M,`width: ${Tt.todos.length?Tt.completedCount/Tt.todos.length*100:0}%`)}),x(c,u)};H(o,c=>{l(r)&&c(a)})}x(e,n),tt()}ii(["click"]);function Rf(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"m12 19-7-7 7-7"}],["path",{d:"M19 12H5"}]];At(e,vt({name:"arrow-left"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Wo(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"}],["path",{d:"M3 6h18"}],["path",{d:"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"}]];At(e,vt({name:"trash"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Ho(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"}],["path",{d:"M3 3v5h5"}]];At(e,vt({name:"rotate-ccw"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Ko(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M21 21H8a2 2 0 0 1-1.42-.587l-3.994-3.999a2 2 0 0 1 0-2.828l10-10a2 2 0 0 1 2.829 0l5.999 6a2 2 0 0 1 0 2.828L12.834 21"}],["path",{d:"m5.082 11.09 8.828 8.828"}]];At(e,vt({name:"eraser"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function zo(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["circle",{cx:"18",cy:"18",r:"3"}],["circle",{cx:"6",cy:"6",r:"3"}],["path",{d:"M6 21V9a9 9 0 0 0 9 9"}]];At(e,vt({name:"git-merge"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}function Df(e,t){$(t,!0);/**
 * @license @lucide/svelte v0.562.0 - ISC
 *
 * ISC License
 *
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * ---
 *
 * The MIT License (MIT) (for portions derived from Feather)
 *
 * Copyright (c) 2013-2023 Cole Bemis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */let r=dt(t,["$$slots","$$events","$$legacy"]);const n=[["path",{d:"M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z"}]];At(e,vt({name:"message-square"},()=>r,{get iconNode(){return n},children:(i,s)=>{var o=B(),a=I(o);J(a,()=>t.children??lt),x(i,o)},$$slots:{default:!0}})),tt()}var Lf=F("<button> </button>"),jf=F('<div class="w-1.5 h-1.5 rounded-full bg-orange-400" title="In Progress"></div>'),Bf=F('<div class="w-1.5 h-1.5 rounded-full bg-blue-400" title="Needs Review"></div>'),Wf=F('<div class="w-1.5 h-1.5 rounded-full bg-green-400" title="Done"></div>'),Hf=F('<div class="pb-32"><div id="agent-content" class="mt-1"><!></div></div>'),Kf=F('<div class="p-0 min-h-full pb-32"><div class="px-4 py-3 border-b border-gray-800 sticky top-0 bg-[#0D1117] z-10 flex justify-between"><span class="text-sm text-gray-400 font-mono">changes</span> <span class="text-[10px] text-green-500 font-mono">git diff</span></div> <div class="p-3 diff-container"><!></div></div>'),zf=F('<div class="ml-4 text-sm text-gray-500 italic">No activity yet</div>'),Vf=F("<div><!></div>"),Uf=F('<div class="p-5 pb-32 space-y-6"><div class="relative border-l border-gray-800 ml-2 space-y-6" id="log-content"><!> <!></div></div>'),Gf=F('<div class="absolute bottom-0 inset-x-0 z-20 pb-6 px-3"><div class="absolute inset-0 bg-gradient-to-t from-[#0D1117] via-[#0D1117]/95 to-transparent pointer-events-none"></div> <div class="relative mx-auto max-w-xl"><!></div></div>'),Yf=F('<div class="space-y-4"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-amber-500/10 flex items-center justify-center"><!></div> <div><h3 class="font-semibold text-white">Retry Task</h3> <p class="text-xs text-gray-400">Re-run with the same prompt</p></div></div> <p class="text-sm text-gray-300">This will retry the previous prompt and overwrite any existing changes.</p> <div class="flex gap-2 pt-2"><button class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors">Cancel</button> <button class="flex-1 h-9 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-sm font-medium transition-colors">Retry</button></div></div>'),Xf=F(`<div class="space-y-4"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center"><!></div> <div><h3 class="font-semibold text-white">Clear History</h3> <p class="text-xs text-gray-400">Reset memory and context</p></div></div> <p class="text-sm text-gray-300">This will clear the chat history and agent output. The task will start fresh without
							any prior context.</p> <div class="flex gap-2 pt-2"><button class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors">Cancel</button> <button class="flex-1 h-9 rounded-lg bg-red-600 hover:bg-red-500 text-white text-sm font-medium transition-colors">Clear</button></div></div>`),qf=F('<div class="space-y-4"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-purple-500/10 flex items-center justify-center"><!></div> <div><h3 class="font-semibold text-white">Create Pull Request</h3> <p class="text-xs text-gray-400">Push changes to GitHub</p></div></div> <p class="text-sm text-gray-300">This will create a new pull request on GitHub with all the changes from this task.</p> <div class="flex gap-2 pt-2"><button class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors">Cancel</button> <button class="flex-1 h-9 rounded-lg bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium transition-colors">Create PR</button></div></div>'),Zf=F(`<div class="space-y-4"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-green-500/10 flex items-center justify-center"><!></div> <div><h3 class="font-semibold text-white">Merge to Main</h3> <p class="text-xs text-gray-400">Apply changes directly</p></div></div> <p class="text-sm text-gray-300">This will merge all changes directly into the main branch without creating a pull
							request.</p> <div class="flex gap-2 pt-2"><button class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors">Cancel</button> <button class="flex-1 h-9 rounded-lg bg-green-600 hover:bg-green-500 text-white text-sm font-medium transition-colors">Merge</button></div></div>`),Qf=F('<div class="space-y-4"><div class="flex items-center gap-3"><div class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center"><!></div> <div><h3 class="font-semibold text-white">Discard Task</h3> <p class="text-xs text-gray-400">Permanently delete task</p></div></div> <p class="text-sm text-gray-300">This will permanently delete this task and all its data. This action cannot be undone.</p> <div class="flex gap-2 pt-2"><button class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors">Cancel</button> <button class="flex-1 h-9 rounded-lg bg-red-600 hover:bg-red-500 text-white text-sm font-medium transition-colors">Discard</button></div></div>'),Jf=F('<div class="fixed inset-0 z-[200] flex items-start justify-center pt-[25vh] bg-black/60 backdrop-blur-sm"><div class="bg-popover border border-gray-700/50 rounded-xl p-5 w-[320px] shadow-2xl"><!></div></div>'),$f=F('<div class="flex flex-col h-full"><div class="px-4 py-2 border-b border-white/5 flex items-center justify-between shrink-0 bg-popover"><div class="flex items-center gap-3"><button class="w-11 h-11 rounded-full hover:bg-white/5 flex items-center justify-center text-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500/50" aria-label="Go back"><!></button> <div><div class="flex items-center gap-2"><span><i></i> </span> <span class="text-[10px] text-gray-600 font-mono"> </span></div> <h2 class="text-sm font-bold text-gray-200 line-clamp-1 w-48"> </h2></div></div> <button class="w-11 h-11 flex items-center justify-center text-gray-600 hover:text-red-400 transition focus:outline-none focus:ring-2 focus:ring-red-500/50 rounded-lg" aria-label="Discard task"><!></button></div> <div class="flex items-center justify-between p-2 bg-popover shrink-0 border-b border-white/5"><div class="w-6"></div> <div class="flex bg-gray-900 rounded-lg p-0.5 border border-gray-700/50"></div> <div class="w-6 flex justify-end pr-2"><!></div></div> <!> <div class="flex-1 overflow-y-auto bg-[#0D1117] relative w-full" id="content-scroll"><!> <!> <!></div> <!> <div><div class="flex justify-center gap-1 space-y-3"><button class="h-11 px-4 rounded-full flex items-center gap-2 text-xs text-gray-500 hover:text-gray-300 hover:bg-white/[0.04] active:bg-white/[0.08] transition-all" aria-label="Retry task"><!> <span>Retry</span></button> <span class="text-gray-700 self-center">·</span> <button class="h-11 px-4 rounded-full flex items-center gap-2 text-xs text-gray-500 hover:text-gray-300 hover:bg-white/[0.04] active:bg-white/[0.08] transition-all" aria-label="Clear history"><!> <span>Clear</span></button> <span class="text-gray-700 self-center">·</span> <button class="h-11 px-4 rounded-full flex items-center gap-2 text-xs text-gray-500 hover:text-gray-300 hover:bg-white/[0.04] active:bg-white/[0.08] transition-all" aria-label="Create pull request"><!> <span>Create PR</span></button></div> <div class="flex gap-2.5 mt-3"><button class="flex-1 h-12 bg-card hover:bg-card/80 active:bg-card/60 border border-white/[0.08] hover:border-purple-500/30 text-white rounded-xl flex items-center justify-center gap-2 transition-all active:scale-[0.98] relative group"><!> <span class="text-sm font-medium">Chat</span></button> <button class="flex-1 h-12 bg-white hover:bg-gray-100 active:bg-gray-200 text-black rounded-xl flex items-center justify-center gap-2 transition-all active:scale-[0.98] shadow-lg shadow-white/10"><!> <span class="text-sm font-medium">Merge</span></button></div></div> <!></div>');function th(e,t){$(t,!0);let r=V("agent"),n=V(!1),i=V(null);async function s(G,Z){w(n,!1);try{await Pe.chat(t.task.id,G,Z)}catch(ct){console.error("Failed to send message:",ct)}}async function o(G){w(i,null);try{G==="retry"?await Pe.retry(t.task.id):G==="clear"?await Pe.clear(t.task.id):G==="pr"?await Pe.createPR(t.task.id):G==="merge"?await Pe.merge(t.task.id):G==="discard"&&(await Pe.discard(t.task.id),L.closeModal())}catch(Z){console.error(`Failed to ${G}:`,Z)}}var a=$f(),c=g(a),u=g(c),d=g(u);d.__click=()=>L.closeModal();var f=g(d);Rf(f,{class:"w-5 h-5"}),v(d);var p=E(d,2),m=g(p),b=g(m),y=g(b),A=E(y);v(b);var _=E(b,2),S=g(_);v(_),v(m);var P=E(m,2),C=g(P,!0);v(P),v(p),v(u);var M=E(u,2);M.__click=()=>w(i,"discard");var K=g(M);Wo(K,{class:"w-5 h-5"}),v(M),v(c);var q=E(c,2),nt=E(g(q),2);he(nt,20,()=>["agent","diff","activity"],ve,(G,Z)=>{var ct=Lf();ct.__click=()=>w(r,Z,!0);var xt=g(ct,!0);v(ct),yt(Qt=>{zt(ct,1,Qt),_t(xt,Z==="agent"?"Agent":Z==="diff"?"Diff":"Log")},[()=>Ae(ie("px-4 py-2 text-[11px] font-medium rounded-md transition-all focus:outline-none focus:ring-2 focus:ring-purple-500/50",l(r)===Z?"bg-gray-800 text-white shadow":"text-gray-500"))]),x(G,ct)}),v(nt);var st=E(nt,2),rt=g(st);{var D=G=>{var Z=jf();x(G,Z)},R=G=>{var Z=B(),ct=I(Z);{var xt=Mt=>{var gt=Bf();x(Mt,gt)},Qt=Mt=>{var gt=B(),Ot=I(gt);{var Xt=wt=>{var Bt=Wf();x(wt,Bt)};H(Ot,wt=>{t.task.status==="done"&&wt(Xt)},!0)}x(Mt,gt)};H(ct,Mt=>{t.task.status==="review"?Mt(xt):Mt(Qt,!1)},!0)}x(G,Z)};H(rt,G=>{t.task.status==="in_progress"?G(D):G(R,!1)})}v(st),v(q);var j=E(q,2);{var z=G=>{Ff(G,{})};H(j,G=>{Tt.todos.length>0&&G(z)})}var it=E(j,2),at=g(it);{var X=G=>{var Z=Hf(),ct=g(Z),xt=g(ct);fi(xt,()=>t.agentContent),v(ct),v(Z),x(G,Z)};H(at,G=>{l(r)==="agent"&&G(X)})}var U=E(at,2);{var et=G=>{var Z=Kf(),ct=E(g(Z),2),xt=g(ct);fi(xt,()=>t.diffContent),v(ct),v(Z),x(G,Z)};H(U,G=>{l(r)==="diff"&&G(et)})}var ht=E(U,2);{var pt=G=>{var Z=Uf(),ct=g(Z),xt=g(ct);{var Qt=gt=>{var Ot=zf();x(gt,Ot)};H(xt,gt=>{t.logContent.length===0&&gt(Qt)})}var Mt=E(xt,2);he(Mt,17,()=>t.logContent,ve,(gt,Ot)=>{var Xt=Vf(),wt=g(Xt);fi(wt,()=>l(Ot)),v(Xt),x(gt,Xt)}),v(ct),v(Z),x(G,Z)};H(ht,G=>{l(r)==="activity"&&G(pt)})}v(it);var ft=E(it,2);{var Ct=G=>{var Z=Gf(),ct=E(g(Z),2),xt=g(ct);Ws(xt,{mode:"chat",get taskId(){return t.task.id},placeholder:"Continue the conversation...",onSubmit:s,onClose:()=>w(n,!1)}),v(ct),v(Z),x(G,Z)};H(ft,G=>{l(n)&&G(Ct)})}var Vt=E(ft,2);let Kt;var T=g(Vt),W=g(T);W.__click=()=>w(i,"retry");var ot=g(W);Ho(ot,{class:"w-3 h-3"}),It(2),v(W);var bt=E(W,4);bt.__click=()=>w(i,"clear");var Pt=g(bt);Ko(Pt,{class:"w-3 h-3"}),It(2),v(bt);var ut=E(bt,4);ut.__click=()=>w(i,"pr");var Et=g(ut);oo(Et,{class:"w-3 h-3"}),It(2),v(ut),v(T);var Dt=E(T,2),Rt=g(Dt);Rt.__click=()=>w(n,!0);var Lt=g(Rt);Df(Lt,{class:"w-4 h-4 text-purple-400 group-hover:text-purple-300 transition-colors"}),It(2),v(Rt);var mt=E(Rt,2);mt.__click=()=>w(i,"merge");var jt=g(mt);zo(jt,{class:"w-4 h-4"}),It(2),v(mt),v(Dt),v(Vt);var Ut=E(Vt,2);{var te=G=>{var Z=Jf();Z.__click=gt=>gt.target===gt.currentTarget&&w(i,null);var ct=g(Z),xt=g(ct);{var Qt=gt=>{var Ot=Yf(),Xt=g(Ot),wt=g(Xt),Bt=g(wt);Ho(Bt,{class:"w-5 h-5 text-amber-500"}),v(wt),It(2),v(Xt);var Wt=E(Xt,4),Gt=g(Wt);Gt.__click=()=>w(i,null);var qt=E(Gt,2);qt.__click=()=>o("retry"),v(Wt),v(Ot),x(gt,Ot)},Mt=gt=>{var Ot=B(),Xt=I(Ot);{var wt=Wt=>{var Gt=Xf(),qt=g(Gt),ue=g(qt),Ve=g(ue);Ko(Ve,{class:"w-5 h-5 text-red-500"}),v(ue),It(2),v(qt);var ee=E(qt,4),ne=g(ee);ne.__click=()=>w(i,null);var xe=E(ne,2);xe.__click=()=>o("clear"),v(ee),v(Gt),x(Wt,Gt)},Bt=Wt=>{var Gt=B(),qt=I(Gt);{var ue=ee=>{var ne=qf(),xe=g(ne),Qe=g(xe),we=g(Qe);oo(we,{class:"w-5 h-5 text-purple-500"}),v(Qe),It(2),v(xe);var _e=E(xe,4),Se=g(_e);Se.__click=()=>w(i,null);var Ue=E(Se,2);Ue.__click=()=>o("pr"),v(_e),v(ne),x(ee,ne)},Ve=ee=>{var ne=B(),xe=I(ne);{var Qe=_e=>{var Se=Zf(),Ue=g(Se),wn=g(Ue),fr=g(wn);zo(fr,{class:"w-5 h-5 text-green-500"}),v(wn),It(2),v(Ue);var Je=E(Ue,4),$e=g(Je);$e.__click=()=>w(i,null);var _n=E($e,2);_n.__click=()=>o("merge"),v(Je),v(Se),x(_e,Se)},we=_e=>{var Se=B(),Ue=I(Se);{var wn=fr=>{var Je=Qf(),$e=g(Je),_n=g($e),Hs=g(_n);Wo(Hs,{class:"w-5 h-5 text-red-500"}),v(_n),It(2),v($e);var qi=E($e,4),Zi=g(qi);Zi.__click=()=>w(i,null);var Ks=E(Zi,2);Ks.__click=()=>o("discard"),v(qi),v(Je),x(fr,Je)};H(Ue,fr=>{l(i)==="discard"&&fr(wn)},!0)}x(_e,Se)};H(xe,_e=>{l(i)==="merge"?_e(Qe):_e(we,!1)},!0)}x(ee,ne)};H(qt,ee=>{l(i)==="pr"?ee(ue):ee(Ve,!1)},!0)}x(Wt,Gt)};H(Xt,Wt=>{l(i)==="clear"?Wt(wt):Wt(Bt,!1)},!0)}x(gt,Ot)};H(xt,gt=>{l(i)==="retry"?gt(Qt):gt(Mt,!1)})}v(ct),v(Z),x(G,Z)};H(Ut,G=>{l(i)&&G(te)})}v(a),yt(()=>{zt(b,1,`${t.project.color??""} text-[10px]`),zt(y,1,`fas ${t.project.icon??""}`),_t(A,` ${t.project.name??""}`),_t(S,`#${t.task.id??""}`),_t(C,t.task.description),Kt=zt(Vt,1,"shrink-0 px-4 py-3 border-t border-white/[0.06] bg-popover/95 backdrop-blur-sm pb-6",null,Kt,{invisible:l(n),"pointer-events-none":l(n)})}),x(e,a),tt()}ii(["click"]);var eh=F('<div class="flex items-center justify-center h-full"><div class="flex flex-col items-center gap-3"><div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center"><i class="fas fa-spinner fa-spin text-sm text-violet-400"></i></div> <p class="text-xs text-gray-500">Loading task...</p></div></div>'),rh=F('<div class="flex items-center justify-center h-full"><div class="text-center"><p class="text-sm text-red-400 mb-2"> </p> <button class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-xs text-violet-300 hover:bg-violet-500/30 transition-colors">Retry</button></div></div>'),nh=F("<div><!></div>"),ih=F('<div class="h-screen flex flex-col overflow-hidden bg-background"><!> <!> <main class="flex-1 overflow-y-auto bg-background relative pt-14" id="feed-container"><div class="px-3 pt-6 pb-40"><!></div></main> <div class="fixed bottom-6 left-4 right-4 z-20 mx-auto max-w-3xl"><!></div> <!></div>');function Sh(e,t){$(t,!0);let r=V(null),n=V(null),i=V(!1),s=V(null),o=V(""),a=V(""),c=V(Me([])),u=null;async function d(D){if(D)try{w(i,!0),w(s,null),w(o,""),w(a,""),w(c,[],!0),u&&(u.close(),u=null);const R=await Pe.get(D);w(r,R.task,!0),w(n,R.project,!0),Tt.currentTask=R.task,R.messages&&R.messages.length>0&&w(o,f(R.messages,R.task.status==="in_progress"),!0),R.task.status==="in_progress"?w(a,m(),!0):R.task.gitDiff?w(a,b(R.task.gitDiff),!0):w(a,'<div class="text-gray-500 italic">No changes made</div>'),R.logs&&R.logs.length>0&&w(c,R.logs.map(j=>y(j)),!0),u=ma(D,{onAgentUpdate:j=>{w(o,j,!0)},onDiffUpdate:j=>{w(a,j,!0)},onLog:j=>{w(c,[...l(c),j],!0)},onStatus:j=>{},onComplete:j=>{l(r)&&(l(r).status=j)},onError:j=>{console.error("Task SSE error:",j)}})}catch(R){w(s,R instanceof Error?R.message:"Failed to load task",!0),console.error("Task load error:",R)}finally{w(i,!1)}}function f(D,R){if(D.length===0)return'<div class="p-5 text-gray-500 italic text-xs">No agent output</div>';let j='<div class="space-y-0">';for(const z of D)j+=p(z);return j+="</div>",R&&(j+=`
				<div class="flex items-center gap-3 px-4 py-3">
					<div class="relative">
						<div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">
							<i class="fas fa-robot text-sm text-violet-400 pulse-glow"></i>
						</div>
						<div class="absolute inset-0 animate-spin" style="animation-duration: 3s;">
							<div class="absolute -top-0.5 left-1/2 -translate-x-1/2 w-1 h-1 bg-violet-400 rounded-full"></div>
						</div>
					</div>
					<div>
						<p class="text-xs font-medium shimmer">Agent is thinking...</p>
						<p class="text-[10px] text-gray-600">Analyzing code</p>
					</div>
				</div>
			`),j}function p(D){const R=D.role==="user",j=R?"bg-violet-500/10 border-violet-500/20":"bg-gray-800/50 border-gray-700/50",z=R?"fa-user":"fa-robot",it=R?"text-violet-400":"text-blue-400";let at="";for(const X of D.content)X.type==="text"&&X.text?at+=`<p class="text-sm text-gray-300 leading-normal">${A(X.text)}</p>`:X.type==="tool_use"&&X.toolName?at+=`
					<div class="flex items-center gap-2 my-2">
						<span class="px-2 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded text-[10px] font-mono text-blue-300">
							${A(X.toolName)}
						</span>
					</div>
				`:X.type==="tool_result"&&(at+=`<pre class="text-xs text-gray-400 font-mono whitespace-pre-wrap bg-gray-900/50 rounded p-2 my-2">${A(X.text||"")}</pre>`);return`
			<div class="flex gap-3 px-4 py-3 ${j} border-b border-white/5">
				<div class="w-8 h-8 rounded-full ${it} bg-white/5 flex items-center justify-center shrink-0">
					<i class="fas ${z} text-xs"></i>
				</div>
				<div class="flex-1 min-w-0">
					${at}
				</div>
			</div>
		`}function m(){return`
			<div class="flex flex-col items-center justify-center h-48 text-gray-500 space-y-4">
				<i class="fas fa-cog fa-spin text-3xl opacity-50"></i>
				<p class="text-xs font-mono">Generating changes...</p>
			</div>
		`}function b(D){if(!D)return'<div class="text-gray-500 italic">No changes made</div>';let R="";for(const j of D.split(`
`)){const z=A(j);j.startsWith("+")?R+=`<div class="px-3 py-1 bg-green-500/10 text-green-400 font-mono text-xs border-l-2 border-green-500/50">${z.substring(1)}</div>`:j.startsWith("-")?R+=`<div class="px-3 py-1 bg-red-500/10 text-red-400 font-mono text-xs border-l-2 border-red-500/50">${z.substring(1)}</div>`:j.startsWith("@@")?R+=`<div class="px-3 py-1 bg-gray-800 text-gray-500 font-mono text-xs">${z}</div>`:j.trim()!==""&&(R+=`<div class="px-3 py-1 text-gray-400 font-mono text-xs">${z}</div>`)}return R}function y(D){return`
			<div class="ml-4 relative">
				<div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] bg-blue-500"></div>
				<p class="text-xs text-gray-400">${A(D.message)}</p>
			</div>
		`}function A(D){return D.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;").replace(/'/g,"&#39;")}St(()=>{L.modalOpen&&L.modalTaskId?d(L.modalTaskId):(u&&(u.close(),u=null),w(r,null),w(n,null),w(o,""),w(a,""),w(c,[],!0))}),St(()=>{Tt.currentTask&&L.modalTaskId===Tt.currentTask.id&&w(r,{...l(r),...Tt.currentTask},!0)}),Zs(()=>{u&&u.close()});var _=ih(),S=g(_);qd(S,{});var P=E(S,2);Gd(P,{});var C=E(P,2),M=g(C),K=g(M);J(K,()=>t.children),v(M),v(C);var q=E(C,2),nt=g(q);Ws(nt,{mode:"create",placeholder:"What do you want to build?"}),v(q);var st=E(q,2);{var rt=D=>{var R=nh();let j;var z=g(R);{var it=X=>{var U=eh();x(X,U)},at=X=>{var U=B(),et=I(U);{var ht=ft=>{var Ct=rh(),Vt=g(Ct),Kt=g(Vt),T=g(Kt,!0);v(Kt);var W=E(Kt,2);W.__click=()=>d(L.modalTaskId),v(Vt),v(Ct),yt(()=>_t(T,l(s))),x(ft,Ct)},pt=ft=>{var Ct=B(),Vt=I(Ct);{var Kt=T=>{th(T,{get task(){return l(r)},get project(){return l(n)},get agentContent(){return l(o)},get diffContent(){return l(a)},get logContent(){return l(c)}})};H(Vt,T=>{l(r)&&l(n)&&T(Kt)},!0)}x(ft,Ct)};H(et,ft=>{l(s)?ft(ht):ft(pt,!1)},!0)}x(X,U)};H(z,X=>{l(i)?X(it):X(at,!1)})}v(R),yt(()=>j=zt(R,1,"fixed inset-0 z-50 bg-popover flex flex-col overflow-hidden transition-all duration-300",null,j,{"translate-x-0":L.modalOpen,"translate-x-full":!L.modalOpen})),x(D,R)};H(st,D=>{L.modalOpen&&L.modalTaskId&&D(rt)})}v(_),x(e,_),tt()}ii(["click"]);export{Sh as component};
