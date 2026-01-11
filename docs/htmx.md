### Install htmx via CDN (minified)

Source: https://htmx.org/docs/index

Provides the script tag for including the minified version of htmx from jsDelivr CDN. This is the quickest way to start using htmx.

```html
<script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js" integrity="sha384-/TgkGk7p307TH7EXJDuUlgG3Ce1UVolAOFopFekQkkXihi5u/6OCvVKyz1W+idaz" crossorigin="anonymous"></script>
```

--------------------------------

### htmx Swap Example with Options

Source: https://htmx.org/docs/index

Demonstrates how to use the `hx-swap` attribute with options to control swapping behavior. This example specifically shows how to replace the `outerHTML` of the target element and ignore any title tag found in the new content. It highlights the syntax for applying modifiers after the swap style, separated by colons.

```html
<button hx-post="/like" hx-swap="outerHTML ignoreTitle:true">Like</button>
```

--------------------------------

### Example: Dynamic Content with Mocked Response

Source: https://htmx.org/docs/index

This HTML example demonstrates using the htmx demo script to create a dynamic counter. It includes a button that triggers a POST request, a template tag defining a mocked response for the '/foo' URL with a delay, and JavaScript to manage a global counter variable.

```html
<!-- load demo environment -->
<script src="https://demo.htmx.org"></script>

<!-- post to /foo -->
<button hx-post="/foo" hx-target="#result">
    Count Up
</button>
<output id="result"></output>

<!-- respond to /foo with some dynamic content in a template tag -->
<script>
    globalInt = 0;
</script>
<template url="/foo" delay="500"> <!-- note the url and delay attributes -->
    ${globalInt++}
</template>

```

--------------------------------

### Install htmx Locally

Source: https://htmx.org/docs/index

Demonstrates how to include a locally downloaded htmx.min.js file in your HTML document using a script tag.

```html
<script src="/path/to/htmx.min.js"></script>
```

--------------------------------

### Install Htmx Extensions via CDN

Source: https://htmx.org/docs/index

Installs htmx extensions by loading them from a CDN. Ensure the core htmx library is included before any extensions. The extension name needs to be specified in the script tag.

```html
<head>
    <script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js" integrity="sha384-/TgkGk7p307TH7EXJDuUlgG3Ce1UVolAOFopFekQkkXihi5u/6OCvVKyz1W+idaz" crossorigin="anonymous"></script>
    <script src="https://cdn.jsdelivr.net/npm/htmx-ext-response-targets@2.0.4" integrity="sha384-T41oglUPvXLGBVyRdZsVRxNWnOOqCynaPubjUVjxhsjFTKrFJGEMm3/0KGmNQ+Pg" crossorigin="anonymous"></script>
</head>
<body hx-ext="extension-name">
    ...


```

--------------------------------

### HTMX Load Polling Example

Source: https://htmx.org/docs/index

Demonstrates how to use hx-trigger with 'load' and a delay to create a polling effect. The element replaces itself with the response from the GET request, effectively polling the server every second. This is useful for endpoints that terminate polling once a condition is met.

```html
<div hx-get="/messages"
    hx-trigger="load delay:1s"
    hx-swap="outerHTML">
</div>
```

--------------------------------

### HTMX Target for Live Search Results

Source: https://htmx.org/docs/index

Demonstrates using hx-target to specify where the response of an AJAX request should be loaded. In this live search example, keypress events on the input trigger a GET request, and the results are loaded into the '#search-results' div instead of the input element itself.

```html
<input type="text" name="q"
    hx-get="/trigger_delay"
    hx-trigger="keyup delay:500ms changed"
    hx-target="#search-results"
    placeholder="Search...">
<div id="search-results"></div>
```

--------------------------------

### Install htmx via npm

Source: https://htmx.org/docs/index

Command to install the htmx.org package using npm. This is suitable for projects using npm-based build systems.

```bash
npm install htmx.org@2.0.8
```

--------------------------------

### Install Htmx Extensions via npm

Source: https://htmx.org/docs/index

Installs htmx extensions using npm, typically for use with module bundlers like Webpack or Rollup. After installation, the extension needs to be imported into your main JavaScript file.

```bash
npm install htmx-ext-extension-name

```

```javascript
import `htmx.org`;
import `htmx-ext-extension-name`; // replace `extension-name` with the name of the extension 

```

--------------------------------

### Install htmx via CDN (unminified)

Source: https://htmx.org/docs/index

Provides the script tag for including the unminified version of htmx from jsDelivr CDN. Useful for debugging or development.

```html
<script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.js" integrity="sha384-ezjq8118wdwdRMj+nX4bevEi+cDLTbhLAeFF688VK8tPDGeLUe0WoY2MZtSla72F" crossorigin="anonymous"></script>
```

--------------------------------

### Processing Dynamically Loaded Htmx Content with Fetch API

Source: https://htmx.org/docs/index

This example demonstrates how to fetch HTML content using the Fetch API and then process it with htmx. After setting the `innerHTML` of a target div, `htmx.process()` is called on that div to initialize any htmx attributes within the newly added content.

```javascript
let myDiv = document.getElementById('my-div')
fetch('http://example.com/movies.json')
    .then(response => response.text())
    .then(data => { myDiv.innerHTML = data; htmx.process(myDiv); } );
```

--------------------------------

### htmx Button Example: POST Request

Source: https://htmx.org/docs/index

Illustrates how an htmx-enabled button triggers an HTTP POST request on click, replacing a target element with the response content using outerHTML swap.

```html
<button hx-post="/clicked"
    hx-trigger="click"
    hx-target="#parent-div"
    hx-swap="outerHTML">
    Click Me!
</button>
```

--------------------------------

### htmx: Polling with 'every' Trigger

Source: https://htmx.org/docs/index

Shows how to implement polling using the 'every' syntax in hx-trigger. The div will issue a GET request to '/news' every 2 seconds, loading the response into itself.

```html
<div hx-get="/news" hx-trigger="every 2s"></div>

```

--------------------------------

### htmx Data Attribute Usage

Source: https://htmx.org/docs/index

Shows an alternative way to use htmx attributes by prefixing them with 'data-'. This example demonstrates a POST request triggered by a click.

```html
<a data-hx-post="/click">Click Me!</a>
```

--------------------------------

### Basic Anchor Tag Behavior

Source: https://htmx.org/docs/index

Demonstrates the standard behavior of an anchor tag, instructing the browser to issue an HTTP GET request and load the response content.

```html
<a href="/blog">Blog</a>
```

--------------------------------

### Progressive Enhancement for Search Input

Source: https://htmx.org/docs/index

This example demonstrates progressive enhancement for an active search input. By wrapping the htmx-enhanced input in a form, users without JavaScript can still perform a search via a standard form submission. JavaScript-enabled clients receive the enhanced AJAX UX.

```html
<form action="/search" method="POST">
    <input class="form-control" type="search"
        name="search" placeholder="Begin typing to search users..."
        hx-post="/search"
        hx-trigger="keyup changed delay:500ms, search"
        hx-target="#search-results"
        hx-indicator=".htmx-indicator">
</form>
```

--------------------------------

### Custom Validation with hx-on (HTML)

Source: https://htmx.org/docs/index

This example demonstrates how to use the `hx-on` attribute to intercept the `htmx:validation:validate` event. It allows for custom validation logic before a form submission, enabling specific input requirements and reporting validation errors.

```html
<form id="example-form" hx-post="/test">
    <input name="example"
           onkeyup="this.setCustomValidity('') // reset the validation on keyup"
           hx-on:htmx:validation:validate="if(this.value != 'foo') {
                    this.setCustomValidity('Please enter the value foo') // set the validation error
                    htmx.find('#example-form').reportValidity()          // report the issue
                }">
</form>
```

--------------------------------

### htmx: Active Search with Delayed Keyup Trigger

Source: https://htmx.org/docs/index

Implements an active search pattern using hx-trigger with 'keyup changed delay:500ms'. A GET request to '/trigger_delay' is sent 500ms after the input value changes, and results are loaded into '#search-results'.

```html
<input type="text" name="q"
    hx-get="/trigger_delay"
    hx-trigger="keyup changed delay:500ms"
    hx-target="#search-results"
    placeholder="Search...">
<div id="search-results"></div>

```

--------------------------------

### Add CSRF Token with hx-headers (HTML Tag)

Source: https://htmx.org/docs/index

This example demonstrates how to include a CSRF token in HTTP headers for every htmx request by using the 'hx-headers' attribute on the html tag. This is a common method for CSRF prevention when using htmx.

```html
<html lang="en" hx-headers='{"X-CSRF-TOKEN": "CSRF_TOKEN_INSERTED_HERE"}'>
    :
</html>
```

--------------------------------

### Add CSRF Token with hx-headers (Body Tag)

Source: https://htmx.org/docs/index

This example shows how to add a CSRF token to HTTP headers for all htmx requests by applying the 'hx-headers' attribute to the body tag. This approach ensures that CSRF tokens are sent with requests initiated from within the body of the document.

```html
<body hx-headers='{"X-CSRF-TOKEN": "CSRF_TOKEN_INSERTED_HERE"}'>
    :
</body>
```

--------------------------------

### htmx: Triggering Request on Control-Click

Source: https://htmx.org/docs/index

Demonstrates using a trigger filter with square brackets to specify conditions for an AJAX request. This GET request to '/clicked' will only trigger if the element is control-clicked.

```html
<div hx-get="/clicked" hx-trigger="click[ctrlKey]">
    Control Click Me
</div>

```

--------------------------------

### Attribute Inheritance in htmx

Source: https://htmx.org/docs/index

This example illustrates how htmx attributes like `hx-confirm` can be inherited by child elements. By placing the attribute on a parent `div`, all descendant htmx-triggered elements will inherit that attribute, reducing code duplication. This is a core feature for managing common configurations across multiple elements.

```html
<div hx-confirm="Are you sure?">
    <button hx-delete="/account">
        Delete My Account
    </button>
    <button hx-put="/account">
        Update My Account
    </button>
</div>
```

--------------------------------

### Initialize Library on Htmx Load (Helper)

Source: https://htmx.org/docs/index

A cleaner way to initialize third-party libraries when Htmx loads content using the `htmx.onLoad` helper function.

```javascript
htmx.onLoad(function(target) {
    myJavascriptLib.init(target);
});


```

--------------------------------

### Include htmx Demo Script for Development

Source: https://htmx.org/docs/index

This script tag loads the htmx demo environment, which includes htmx, hyperscript, and a request mocking library. It's designed for quickly creating and demonstrating htmx features in isolated environments like JSFiddle.

```html
<script src="https://demo.htmx.org"></script>

```

--------------------------------

### View Transitions

Source: https://htmx.org/docs/index

Utilize the experimental View Transitions API for animated DOM state changes, with HTMX providing a fallback mechanism for browsers that don't support the API.

```APIDOC
## View Transitions

### Description
The experimental View Transitions API allows for animated transitions between different DOM states. HTMX supports integration with this API, providing a fallback to standard swapping mechanisms if the View Transitions API is unavailable in a browser.

### Configuration
- **Global Enable**: Set `htmx.config.globalViewTransitions` to `true` to enable transitions for all swaps.
- **Attribute Enable**: Use the `transition:true` option within the `hx-swap` attribute.
- **Event Handling**: The `htmx:beforeTransition` event can be caught, and `preventDefault()` can be called to cancel a transition.

### CSS Configuration
View Transitions can be configured using CSS, as detailed in the Chrome documentation for the feature.
```

--------------------------------

### Generalized Event Handling with hx-on:click (htmx)

Source: https://htmx.org/docs/index

Shows how to use the htmx `hx-on:click` attribute to respond to a click event, mirroring the functionality of the standard `onclick` attribute but within the htmx framework. This preserves the Locality of Behaviour.

```html
<button hx-on:click="alert('You clicked me!')">
    Click Me!
</button>

```

--------------------------------

### File Uploads

Source: https://htmx.org/docs/index

Details how to upload files using HTMX by setting the `hx-encoding` attribute to `multipart/form-data` and how to monitor upload progress.

```APIDOC
## File Upload

To upload files via an HTMX request, set the `hx-encoding` attribute to `multipart/form-data`. This will use a `FormData` object to submit the request, properly including the file.

Note that server-side handling of `multipart/form-data` requests may differ significantly depending on your technology stack.

HTMX fires a `htmx:xhr:progress` event periodically during upload, based on the standard `progress` event, allowing you to monitor and display upload progress.
```

--------------------------------

### Initialize TomSelect with htmx.onLoad

Source: https://htmx.org/docs/index

Initializes TomSelect on elements with the class 'tomselect' when new content is loaded by HTMX. This ensures rich select elements are correctly set up after AJAX requests.

```javascript
htmx.onLoad(function (target) {
    // find all elements in the new content that should be
    // an editor and init w/ TomSelect
    var editors = target.querySelectorAll(".tomselect")
            .forEach(elt => new TomSelect(elt));
});
```

--------------------------------

### Perform Out-of-Band Swaps in htmx Responses

Source: https://htmx.org/docs/index

Demonstrates how to directly manipulate the DOM with out-of-band (OOB) swaps using the `hx-swap-oob` attribute in an htmx response. This allows specific elements in the response to be inserted into the DOM at locations matching their IDs, independent of the primary target. It's effective for 'piggy-backing' updates on existing requests.

```html
<div id="message" hx-swap-oob="true">Swap me directly!</div>
Additional Content
```

--------------------------------

### Request Parameters

Source: https://htmx.org/docs/index

Explains how HTMX includes element values in requests, how form elements are handled, and the use of `hx-include`, `hx-params`, and `htmx:configRequest` for parameter management.

```APIDOC
## Request Parameters

By default, an element that causes a request will include its value if it has one. If the element is a form, it will include the values of all inputs within it.

As with HTML forms, the `name` attribute of the input is used as the parameter name in the request that HTMX sends.

Additionally, if the element causes a non-`GET` request, the values of all the inputs of the associated form will be included (typically this is the nearest enclosing form, but could be different if e.g. `<button form="associated-form">` is used).

### Including Other Elements
Use the `hx-include` attribute with a CSS selector to include values from other elements in the request.

### Filtering Parameters
Use the `hx-params` attribute to filter out specific parameters from the request.

### Programmatic Modification
Use the `htmx:configRequest` event to programmatically modify parameters before the request is sent.
```

--------------------------------

### Morph Swaps

Source: https://htmx.org/docs/index

Morphing swaps, available via extensions, merge new content into the existing DOM by mutating nodes in-place, preserving state like focus and video.

```APIDOC
## Morph Swaps

### Description
Morphing swaps, enabled through extensions, aim to merge new content into the existing DOM rather than performing a full replacement. This approach helps preserve elements like focus and video playback state by mutating existing DOM nodes directly.

### Available Extensions
- **Idiomorph**: A morphing algorithm developed by the HTMX team.
- **Morphdom Swap**: Based on the original `morphdom` library.
- **Alpine-morph**: Integrates with Alpine.js using the Alpine morph plugin.
```

--------------------------------

### Configure HTMX Response Handling

Source: https://htmx.org/docs/index

Demonstrates how to configure HTMX's response handling by modifying the `htmx.config.responseHandling` array. This allows customization of how different HTTP status codes are processed, including swapping, error handling, and more.

```javascript
    responseHandling: [
        {code:"204", swap: false},   // 204 - No Content by default does nothing, but is not an error
        {code:"[23]..", swap: true}, // 200 & 300 responses are non-errors and are swapped
        {code:"[45]..", swap: false, error:true}, // 400 & 500 responses are not swapped and are errors
        {code:"...", swap: false}    // catch all for any other response code
    ]
```

--------------------------------

### Configure htmx to Swap All Responses

Source: https://htmx.org/docs/index

This meta tag configures htmx to swap the content for all HTTP responses, regardless of the status code. It uses a wildcard '.' to match any response code, ensuring that all incoming HTML fragments are processed and swapped into the target element. This simplifies response handling when the server always returns swappable content.

```html
<meta name="htmx-config" content='{"responseHandling": [{"code":".*", "swap": true}]}' />
```

--------------------------------

### Encapsulate Table Elements for Out-of-Band Swaps with Template

Source: https://htmx.org/docs/index

Addresses the challenge of using out-of-band swaps with table elements (like `<tr>` or `<td>`) that cannot stand alone in the DOM. By wrapping these elements within a `<template>` tag in the response, they can be correctly processed and swapped into the DOM. This ensures valid HTML structure during OOB operations.

```html
<template>
  <tr id="message" hx-swap-oob="true"><td>Joe</td><td>Smith</td></tr>
</template>
```

--------------------------------

### Select Content for Out-of-Band Swaps with hx-select-oob

Source: https://htmx.org/docs/index

Explains the use of `hx-select-oob` for picking specific elements from an htmx response for out-of-band swaps. This attribute takes a list of element IDs, allowing targeted direct manipulation of the DOM for these selected elements, in addition to the primary swap target. This enhances flexibility in updating multiple DOM locations from a single response.

```html
<!-- Example usage: Response contains #header and #footer, both to be swapped out-of-band -->
<div hx-get="/layout" hx-select-oob="#header,#footer">Update Layout</div>
```

--------------------------------

### Configure htmx Default Swap Style with Meta Tag

Source: https://htmx.org/docs/index

This snippet demonstrates how to configure htmx by setting the default swap style using a meta tag. This approach is useful for global configurations that affect all htmx requests on the page. No external JavaScript libraries are required for this meta tag configuration.

```html
<meta name="htmx-config" content='{"defaultSwapStyle":"outerHTML"}'>
```

--------------------------------

### Inline Scripting with onclick Attribute (HTML)

Source: https://htmx.org/docs/index

Demonstrates the standard HTML `onclick` attribute for embedding inline JavaScript to respond to a click event. This approach is limited to a fixed set of DOM events and lacks generalized event handling.

```html
<button onclick="alert('You clicked me!')">
    Click Me!
</button>

```

--------------------------------

### Set CSP with Meta Tag

Source: https://htmx.org/docs/index

This snippet shows how to configure a Content Security Policy using a meta tag. It restricts browser requests to the originating host, enhancing security by preventing connections to non-origin hosts and evaluating inline scripts.

```html
<meta http-equiv="Content-Security-Policy" content="default-src 'self'">
```

--------------------------------

### Preserve Content Across htmx Swaps with hx-preserve

Source: https://htmx.org/docs/index

Demonstrates how to use the `hx-preserve` attribute to keep specific elements intact across htmx content swaps. When applied to an element, its content and state (e.g., video playback) will persist even if the parent container's content is replaced. This is crucial for maintaining user experience with long-lived interactive components.

```html
<div id="main-content">
  <video hx-preserve="true" src="/video.mp4" controls></video>
  <!-- Other content that might be swapped -->
</div>
```

--------------------------------

### Confirming Requests

Source: https://htmx.org/docs/index

Covers the `hx-confirm` attribute for simple JavaScript confirmation dialogs and the `htmx:confirm` event for more advanced, asynchronous confirmations using libraries like SweetAlert.

```APIDOC
## Confirming Requests

### `hx-confirm` Attribute
Often, you will want to confirm an action before issuing a request. HTMX supports the `hx-confirm` attribute, which allows you to confirm an action using a simple JavaScript dialog.

**Example:**
```html
<button hx-delete="/account" hx-confirm="Are you sure you wish to delete your account?">
    Delete My Account
</button>
```

### `htmx:confirm` Event
For more sophisticated confirmation dialogs, you can use events. The `htmx:confirm` event is fired on every request trigger and can be used for asynchronous confirmation.

**Example using SweetAlert:**
```javascript
document.body.addEventListener('htmx:confirm', function(evt) {
  if (evt.target.matches("[confirm-with-sweet-alert='true']")) {
    evt.preventDefault();
    swal({
      title: "Are you sure?",
      text: "Are you sure you are sure?",
      icon: "warning",
      buttons: true,
      dangerMode: true,
    }).then((confirmed) => {
      if (confirmed) {
        evt.detail.issueRequest();
      }
    });
  }
});
```
```

--------------------------------

### Apply CSS Transitions with Stable IDs in htmx

Source: https://htmx.org/docs/index

Explains how to leverage CSS transitions for elements that are replaced via htmx requests. By ensuring the target element maintains a stable 'id' across requests and applying CSS classes, htmx facilitates smooth visual transitions without JavaScript intervention. The swap & settle model in htmx handles attribute copying for this effect.

```html
<div id="div1">Original Content</div>
```

```html
<div id="div1" class="red">New Content</div>
```

```css
.red {
    color: red;
    transition: all ease-in 1s ;
}
```

--------------------------------

### HTMX Configuration Options

Source: https://htmx.org/docs/index

Htmx provides a comprehensive set of configuration variables that can be adjusted to modify its behavior. These settings control aspects like history management, animation delays, request handling, and more.

```APIDOC
## Configuring HTMX

Htmx offers a flexible configuration system that can be accessed either programmatically via `htmx.config` or declaratively through attributes. Below is a list of available configuration variables and their default settings:

### Configuration Variables

- **`htmx.config.historyEnabled`** (boolean) - Defaults to `true`. Primarily used for testing purposes.
- **`htmx.config.historyCacheSize`** (number) - Defaults to `10`. Determines the number of history entries to cache.
- **`htmx.config.refreshOnHistoryMiss`** (boolean) - Defaults to `false`. If `true`, HTMX will perform a full page refresh instead of an AJAX request when a history miss occurs.
- **`htmx.config.defaultSwapStyle`** (string) - Defaults to `innerHTML`. Specifies the default method for swapping content (e.g., `innerHTML`, `outerHTML`, `afterbegin`).
- **`htmx.config.defaultSwapDelay`** (number) - Defaults to `0`. The delay in milliseconds before swapping content.
- **`htmx.config.defaultSettleDelay`** (number) - Defaults to `20`. The delay in milliseconds after swapping content before settling animations begin.
- **`htmx.config.includeIndicatorStyles`** (boolean) - Defaults to `true`. Controls whether HTMX loads its default indicator styles.
- **`htmx.config.indicatorClass`** (string) - Defaults to `htmx-indicator`. The CSS class applied to loading indicators.
- **`htmx.config.requestClass`** (string) - Defaults to `htmx-request`. The CSS class added to elements during an AJAX request.
- **`htmx.config.addedClass`** (string) - Defaults to `htmx-added`. The CSS class added to newly inserted DOM elements.
- **`htmx.config.settlingClass`** (string) - Defaults to `htmx-settling`. The CSS class applied during the settling phase of a swap.
- **`htmx.config.swappingClass`** (string) - Defaults to `htmx-swapping`. The CSS class applied during the swapping phase of a swap.
- **`htmx.config.allowEval`** (boolean) - Defaults to `true`. Enables or disables HTMX's use of `eval` for features like trigger filters.
- **`htmx.config.allowScriptTags`** (boolean) - Defaults to `true`. Determines if HTMX processes `<script>` tags found in new content.
- **`htmx.config.inlineScriptNonce`** (string) - Defaults to `''`. A nonce to be added to inline scripts for security.
- **`htmx.config.attributesToSettle`** (array of strings) - Defaults to `["class", "style", "width", "height"]`. Attributes that HTMX will settle during the settling phase.
- **`htmx.config.inlineStyleNonce`** (string) - Defaults to `''`. A nonce to be added to inline styles for security.
- **`htmx.config.useTemplateFragments`** (boolean) - Defaults to `false`. If `true`, HTMX uses HTML template tags for parsing server content (not compatible with IE11).
- **`htmx.config.wsReconnectDelay`** (string) - Defaults to `full-jitter`. The strategy for WebSocket reconnect delays.
- **`htmx.config.wsBinaryType`** (string) - Defaults to `blob`. The data type for binary data received over WebSockets.
- **`htmx.config.disableSelector`** (string) - Defaults to `[hx-disable], [data-hx-disable]`. Elements matching this selector (or their parents) will be ignored by HTMX.
- **`htmx.config.withCredentials`** (boolean) - Defaults to `false`. Allows cross-site `Access-Control` requests with credentials (cookies, auth headers, etc.).
- **`htmx.config.timeout`** (number) - Defaults to `0`. The maximum time in milliseconds a request can take before being terminated.
- **`htmx.config.scrollBehavior`** (string) - Defaults to `instant`. Controls scroll behavior during `hx-swap` with the `show` modifier. Allowed values: `instant`, `smooth`, `auto`.
- **`htmx.config.defaultFocusScroll`** (boolean) - Defaults to `false`. Determines if the focused element is scrolled into view. Can be overridden by the `focus-scroll` swap modifier.
- **`htmx.config.getCacheBusterParam`** (boolean) - Defaults to `false`. If `true`, HTMX appends a cache-busting parameter (`org.htmx.cache-buster=targetElementId`) to GET requests.
- **`htmx.config.globalViewTransitions`** (boolean) - Defaults to `false`. If `true`, HTMX utilizes the View Transition API for content swaps.
- **`htmx.config.methodsThatUseUrlParams`** (array of strings) - Defaults to `["get", "delete"]`. HTTP methods that will format requests by encoding parameters in the URL.
- **`htmx.config.selfRequestsOnly`** (boolean) - Defaults to `true`. If `true`, HTMX only allows AJAX requests to the same domain as the current document.
- **`htmx.config.ignoreTitle`** (boolean) - Defaults to `false`. If `true`, HTMX will not update the document title from a `<title>` tag in new content.
- **`htmx.config.disableInheritance`** (boolean) - Defaults to `false`. Disables attribute inheritance, which can then be overridden by `hx-inherit`.
- **`htmx.config.scrollIntoViewOnBoost`** (boolean) - Defaults to `true`. Controls whether the target of a boosted element is scrolled into the viewport.
- **`htmx.config.triggerSpecsCache`** (object | null) - Defaults to `null`. A cache for evaluated trigger specifications to improve parsing performance.
- **`htmx.config.responseHandling`** (object) - Configures default response handling behavior for different status codes (swap or error).
- **`htmx.config.allowNestedOobSwaps`** (boolean) - Defaults to `true`. Determines if out-of-band (OOB) swaps are processed on nested elements.
- **`htmx.config.historyRestoreAsHxRequest`** (boolean) - Defaults to `true`. Treats history cache miss full page reload requests as `HX-Request` by returning the `HX-Request` header.

### Example Usage (Programmatic)

```javascript
htmx.config.historyEnabled = false;
htmx.config.defaultSwapStyle = 'outerHTML';
```

### Example Usage (Declarative - `hx-disable`)

```html
<div hx-disable>
  This element and its children will not trigger HTMX requests.
</div>
```

```

--------------------------------

### Enable Htmx Extension

Source: https://htmx.org/docs/index

Enables an Htmx extension by adding the `hx-ext` attribute to an HTML element. The specified extension will be applied to all child elements of the element with the attribute.

```html
<body hx-ext="response-targets">
    ...
    <button hx-post="/register" hx-target="#response-div" hx-target-404="#not-found">
        Register!
    </button>
    <div id="response-div"></div>
    <div id="not-found"></div>
    ...
</body>

```

--------------------------------

### Import htmx in Webpack (global assignment)

Source: https://htmx.org/docs/index

Demonstrates a method to explicitly assign the imported htmx module to the window object for global access in a Webpack project. This requires a custom JS file.

```javascript
window.htmx = require('htmx.org');
```

--------------------------------

### HTMX Extended CSS Selectors for Targeting

Source: https://htmx.org/docs/index

Explains the extended CSS selector syntax supported by hx-target and other similar attributes. This includes using 'this', 'closest', 'next', 'previous', and 'find' to target elements relative to the requesting element, enabling more dynamic UI interactions without relying heavily on IDs.

```html
<!-- Example of 'closest' -->
<tr hx-get="/update" hx-target="closest tr">
    <td>Data 1</td>
    <td>Data 2</td>
</tr>

<!-- Example of 'find' -->
<div hx-get="/data" hx-target="find .content">
    <div class="header">Header</div>
    <div class="content">Content to be replaced</div>
</div>
```

--------------------------------

### Handle htmx:beforeSwap Event for Custom Swap Logic

Source: https://htmx.org/docs/index

This snippet demonstrates how to intercept the htmx:beforeSwap event to customize swap behavior based on HTTP response status codes. It allows for custom error handling (e.g., alerting on 404), enabling swaps for specific error codes like 422, and retargeting content for others like 418. It utilizes JavaScript event listeners.

```javascript
document.body.addEventListener('htmx:beforeSwap', function(evt) {
    if(evt.detail.xhr.status === 404){
        // alert the user when a 404 occurs (maybe use a nicer mechanism than alert())
        alert("Error: Could Not Find Resource");
    } else if(evt.detail.xhr.status === 422){
        // allow 422 responses to swap as we are using this as a signal that
        // a form was submitted with bad data and want to rerender with the
        // errors
        //
        // set isError to false to avoid error logging in console
        evt.detail.shouldSwap = true;
        evt.detail.isError = false;
    } else if(evt.detail.xhr.status === 418){
        // if the response code 418 (I'm a teapot) is returned, retarget the
        // content of the response to the element with the id `teapot`
        evt.detail.shouldSwap = true;
        evt.detail.target = htmx.find("#teapot");
    }
});

```

--------------------------------

### Extra Values

Source: https://htmx.org/docs/index

Explains how to include additional data in requests using `hx-vals` for static JSON values and `hx-vars` for dynamically computed values.

```APIDOC
## Extra Values

### `hx-vals` Attribute
Include extra values in a request using the `hx-vals` attribute. This attribute accepts name-expression pairs in JSON format.

### `hx-vars` Attribute
Include extra values that are dynamically computed using the `hx-vars` attribute. This attribute accepts comma-separated name-expression pairs.
```

--------------------------------

### Import htmx in Webpack (default)

Source: https://htmx.org/docs/index

Shows how to import htmx into your main JavaScript file when using Webpack. This makes htmx available globally.

```javascript
import 'htmx.org';
```

--------------------------------

### htmx: Basic AJAX Request with hx-put

Source: https://htmx.org/docs/index

Demonstrates a basic AJAX request using the hx-put attribute. When the button is clicked, it issues a PUT request to the '/messages' URL and loads the response into the button element.

```html
<button hx-put="/messages">
    Put To Messages
</button>

```

--------------------------------

### Listen for Htmx Load Event

Source: https://htmx.org/docs/index

Registers a listener for the 'htmx:load' event, which fires when an element is loaded into the DOM by Htmx. This can be used to initialize third-party JavaScript libraries on dynamically loaded content.

```javascript
document.body.addEventListener('htmx:load', function(evt) {
    myJavascriptLib.init(evt.detail.elt);
});


```

```javascript
htmx.on("htmx:load", function(evt) {
    myJavascriptLib.init(evt.detail.elt);
});


```

--------------------------------

### SortableJS Integration with Htmx

Source: https://htmx.org/docs/index

This snippet shows how to set up SortableJS within an htmx form. The htmx attributes (`hx-post`, `hx-trigger`) are placed on the form, and SortableJS is initialized on the form element. The `htmx.onLoad` function ensures that SortableJS is initialized on newly loaded content as well.

```html
<form class="sortable" hx-post="/items" hx-trigger="end">
    <div class="htmx-indicator">Updating...</div>
    <div><input type='hidden' name='item' value='1'/>Item 1</div>
    <div><input type='hidden' name='item' value='2'/>Item 2</div>
    <div><input type='hidden' name='item' value='2'/>Item 3</div>
</form>
```

```javascript
htmx.onLoad(function(content) {
    var sortables = content.querySelectorAll(".sortable");
    for (var i = 0; i < sortables.length; i++) {
        var sortable = sortables[i];
        new Sortable(sortable, {
            animation: 150,
            ghostClass: 'blue-background-class'
        });
    }
})
```

--------------------------------

### Custom Event Handling with hx-on:htmx:config-request (htmx)

Source: https://htmx.org/docs/index

Illustrates using the `hx-on:htmx:config-request` attribute to modify request parameters before an htmx request is sent. This allows for dynamic adjustments to requests, such as adding custom parameters.

```html
<button hx-post="/example"
        hx-on:htmx:config-request="event.detail.parameters.example = 'Hello Scripting!'">
    Post Me!
</button>

```

--------------------------------

### Enable AJAX with hx-boost Attribute

Source: https://htmx.org/docs/index

The `hx-boost` attribute on a container element enhances all anchor tags and forms within it to use AJAX requests. By default, responses are swapped into the `body` tag. This attribute is a simple way to enable AJAX functionality across a section of your page.

```html
<div hx-boost="true">
    <a href="/blog">Blog</a>
</div>
```

--------------------------------

### Standard Swapping

Source: https://htmx.org/docs/index

HTMX provides several ways to swap HTML content into the DOM. The default behavior replaces the innerHTML of the target element.

```APIDOC
## Standard Swapping

### Description
HTMX offers several methods for swapping HTML content into the DOM. By default, the content replaces the `innerHTML` of the target element. This behavior can be modified using the `hx-swap` attribute.

### Method
N/A (Attribute-driven)

### Endpoint
N/A (Attribute-driven)

### Parameters
#### `hx-swap` Attribute Values
- **`innerHTML`** (string) - The default. Inserts the content inside the target element.
- **`outerHTML`** (string) - Replaces the entire target element with the returned content.
- **`afterbegin`** (string) - Prepends the content before the first child inside the target.
- **`beforebegin`** (string) - Prepends the content before the target element within its parent.
- **`beforeend`** (string) - Appends the content after the last child inside the target.
- **`afterend`** (string) - Appends the content after the target element within its parent.
- **`delete`** (string) - Deletes the target element, regardless of the response.
- **`none`** (string) - Does not append content from the response. Out-of-band swaps and response headers are still processed.
```

--------------------------------

### HTMX Request Indicator with Image

Source: https://htmx.org/docs/index

Shows how to use the 'htmx-indicator' class to display a loading spinner during an AJAX request. When a request is made (e.g., on button click), htmx adds the 'htmx-request' class, which makes the 'htmx-indicator' element visible. The indicator can be an image or any other element.

```html
<button hx-get="/click">
    Click Me!
    <img class="htmx-indicator" src="/spinner.gif" alt="Loading...">
</button>
```

--------------------------------

### Enable All htmx Event Logging

Source: https://htmx.org/docs/index

This JavaScript function call, htmx.logAll(), enables the logging of every single event that htmx triggers. This is a powerful debugging tool for observing the library's internal workings and tracking event sequences.

```javascript
htmx.logAll();

```

--------------------------------

### Synchronize Input Validation with Form Submission using hx-sync

Source: https://htmx.org/docs/index

Demonstrates coordinating requests between an input's validation and a form's submission. The hx-sync attribute on the input ensures its request is aborted if a form submission begins, preventing race conditions. This relies on htmx's ability to target parent elements with `closest form:abort`.

```html
<form hx-post="/store">
    <input id="title" name="title" type="text"
        hx-post="/validate"
        hx-trigger="change"
        hx-sync="closest form:abort">
    <button type="submit">Submit</button>
</form>
```

--------------------------------

### Selectively Swap Content with hx-select in htmx

Source: https://htmx.org/docs/index

Shows how to use the `hx-select` attribute to specify a subset of the response HTML for swapping. This attribute accepts a CSS selector, allowing only the matching elements from the response to be inserted into the target DOM element. This provides fine-grained control over content updates.

```html
<!-- Example usage: Assume response contains multiple divs, but only #specific-content is desired -->
<div hx-get="/content" hx-target="#container" hx-select="#specific-content">Load Specific Content</div>
```

--------------------------------

### htmx: Triggering POST Request on Mouse Enter

Source: https://htmx.org/docs/index

Shows how to trigger an AJAX POST request using the hx-trigger attribute. The request to '/mouse_entered' is sent when the mouse pointer enters the div element.

```html
<div hx-post="/mouse_entered" hx-trigger="mouseenter">
    [Here Mouse, Mouse!]
</div>

```

--------------------------------

### HTML Button with hx-confirm Attribute

Source: https://htmx.org/docs/index

This snippet demonstrates the basic usage of the `hx-confirm` attribute on an HTML button. When clicked, it will display a JavaScript confirmation dialog before proceeding with the delete request. No external libraries are required for this basic functionality.

```html
<button hx-delete="/account" hx-confirm="Are you sure you wish to delete your account?">
    Delete My Account
</button>
```

--------------------------------

### Swap Options Modifiers

Source: https://htmx.org/docs/index

The `hx-swap` attribute supports various options to fine-tune swapping behavior, such as ignoring the title, setting delays, and controlling scrolling.

```APIDOC
## Swap Options Modifiers

### Description
The `hx-swap` attribute accepts several options to customize the swapping process. These modifiers are appended to the swap style, separated by colons.

### Available Options
- **`transition`** (`true` or `false`): Enables or disables the use of the View Transitions API for the swap.
- **`swap: <delay>`** (e.g., `swap:100ms`): Sets the delay between clearing old content and inserting new content.
- **`settle: <delay>`** (e.g., `settle:100ms`): Sets the delay between inserting new content and considering the swap settled.
- **`ignoreTitle`** (`true`): Prevents HTMX from updating the document title with any title found in the new content.
- **`scroll`** (`top` or `bottom`): Scrolls the target element to its top or bottom after the swap.
- **`show`** (`top` or `bottom`): Scrolls the target element's top or bottom into view after the swap.

### Example
```html
<button hx-post="/like" hx-swap="outerHTML ignoreTitle:true">Like</button>
```
```

--------------------------------

### Attribute Inheritance

Source: https://htmx.org/docs/index

Explains how HTMX attributes are inherited by child elements and how to control this behavior using `hx-inherit`, `hx-disinherit`, and the `htmx.config.disableInheritance` setting.

```APIDOC
## Attribute Inheritance

Most attributes in HTMX are inherited: they apply to the element they are on as well as any children elements. This allows you to avoid code duplication by hoisting attributes up the DOM.

**Example of Inheritance:**
```html
<div hx-confirm="Are you sure?">
    <button hx-delete="/account">
        Delete My Account
    </button>
    <button hx-put="/account">
        Update My Account
    </button>
</div>
```
In this example, the `hx-confirm` attribute on the `div` applies to both buttons within it.

### Unsetting Inheritance
To prevent an attribute from being inherited or to override an inherited attribute, you can use the `"unset"` directive. For example, `hx-confirm="unset"` on a child element will disable the inherited confirmation.

**Example of Unsetting:**
```html
<div hx-confirm="Are you sure?">
    <button hx-delete="/account">
        Delete My Account
    </button>
    <button hx-put="/account">
        Update My Account
    </button>
    <button hx-confirm="unset" hx-get="/">
        Cancel
    </button>
</div>
```
In this example, the 'Cancel' button will not show a confirmation dialog.

### Disabling Inheritance

*   **`hx-disinherit` Attribute**: Use this attribute on an element to disable attribute inheritance for that element and its children.
*   **`htmx.config.disableInheritance`**: Set this configuration variable to `true` to disable attribute inheritance globally. You can then explicitly enable it on specific elements using the `hx-inherit` attribute.
```

--------------------------------

### Clean TomSelect Mutations Before History Save

Source: https://htmx.org/docs/index

Handles the 'htmx:beforeHistorySave' event to destroy TomSelect instances before HTMX takes a history snapshot. This prevents issues with reinitialized elements and ensures clean history restoration.

```javascript
htmx.on('htmx:beforeHistorySave', function() {
    // find all TomSelect elements
    document.querySelectorAll('.tomSelect')
            .forEach(elt => elt.tomselect.destroy()); // and call destroy() on them
})
```

--------------------------------

### Push URL to Browser History with hx-push-url

Source: https://htmx.org/docs/index

The `hx-push-url` attribute allows an element to push its request URL into the browser's navigation bar and add the current page state to the browser's history. This enables functionality similar to traditional navigation, including back button support.

```html
<a hx-get="/blog" hx-push-url="true">Blog</a>
```

--------------------------------

### Configure htmx Response Handling for Validation Errors

Source: https://htmx.org/docs/index

This configuration allows htmx to swap content for specific response codes like 2xx, 3xx, and 422, while treating other 4xx and 5xx codes as errors. It explicitly sets '204 No Content' to not swap and '422' to swap, overriding the default behavior for these codes. This is useful for handling server-side validation failures without interrupting the user experience.

```html
<meta
	name="htmx-config"
	content='{
        "responseHandling":[
            {"code":"204", "swap": false},
            {"code":"[23]..", "swap": true},
            {"code":"422", "swap": true},
            {"code":"[45]..", "swap": false, "error":true},
            {"code":"...", "swap": true}
        ]
    }'
/>
```

--------------------------------

### Validate URLs with htmx:validateUrl Event Listener

Source: https://htmx.org/docs/index

The `htmx:validateUrl` event allows for custom validation of request URLs. By listening to this event on the `document.body`, you can inspect the `evt.detail.url` and `evt.detail.sameHost` properties. Invoking `evt.preventDefault()` within the listener will stop the htmx request if the URL does not meet your defined criteria, offering flexible control over cross-domain or specific host requests.

```javascript
document.body.addEventListener('htmx:validateUrl', function (evt) {
  // only allow requests to the current server as well as myserver.com
  if (!evt.detail.sameHost && evt.detail.url.hostname !== "myserver.com") {
    evt.preventDefault();
  }
});

```

--------------------------------

### Alpine.js Conditional Content with Htmx

Source: https://htmx.org/docs/index

This snippet illustrates integrating htmx with Alpine.js for conditionally rendering content. The `x-watch` directive in Alpine.js monitors a boolean variable, and when it changes to true, `htmx.process()` is called on a specific element to initialize htmx attributes within that content.

```html
<div x-data="{show_new: false}"
    x-init="$watch('show_new', value => {
        if (show_new) {
            htmx.process(document.querySelector('#new_content'))
        }
    })">
    <button @click = "show_new = !show_new">Toggle New Content</button>
    <template x-if="show_new">
        <div id="new_content">
            <a hx-get="/server/newstuff" href="#">New Clickable</a>
        </div>
    </template>
</div>
```

--------------------------------

### Monitor Events on a Specific DOM Element

Source: https://htmx.org/docs/index

This command, monitorEvents(htmx.find("#theElement")), is used in the browser's developer console to log all events occurring on a specified DOM element. It helps identify which events a particular element is firing or receiving, aiding in debugging interactions.

```javascript
monitorEvents(htmx.find("#theElement"));

```

--------------------------------

### Configure htmx Security Options with JavaScript

Source: https://htmx.org/docs/index

htmx provides several global configuration options to enhance security. These include restricting requests to the same domain (`selfRequestsOnly`), disabling script tag processing (`allowScriptTags`), controlling history cache size (`historyCacheSize`), and disabling features that rely on `eval()` (`allowEval`). These settings are configured via `htmx.config` object.

```javascript
htmx.config.selfRequestsOnly = true;
htmx.config.allowScriptTags = false;
htmx.config.historyCacheSize = 0;
htmx.config.allowEval = false;

```

--------------------------------

### JavaScript for SweetAlert Confirmation with htmx

Source: https://htmx.org/docs/index

This JavaScript code snippet shows how to integrate the sweetalert2 library for more sophisticated confirmation dialogs. It listens for the `htmx:confirm` event and, if the target element has the `confirm-with-sweet-alert='true'` attribute, it presents a custom confirmation modal. The request is only issued if the user confirms.

```javascript
document.body.addEventListener('htmx:confirm', function(evt) {
  if (evt.target.matches("[confirm-with-sweet-alert='true']")) {
    evt.preventDefault();
    swal({
      title: "Are you sure?",
      text: "Are you sure you are sure?",
      icon: "warning",
      buttons: true,
      dangerMode: true,
    }).then((confirmed) => {
      if (confirmed) {
        evt.detail.issueRequest();
      }
    });
  }
});
```

--------------------------------

### Configure Htmx Request with Events

Source: https://htmx.org/docs/index

Modifies an AJAX request before it is sent by listening to the 'htmx:configRequest' event. This allows adding custom headers or parameters to the request.

```javascript
document.body.addEventListener('htmx:configRequest', function(evt) {
    evt.detail.parameters['auth_token'] = getAuthToken(); // add a new parameter into the request
    evt.detail.headers['Authentication-Token'] = getAuthToken(); // add a new header into the request
});


```

--------------------------------

### HTMX Explicit Request Indicator Targeting

Source: https://htmx.org/docs/index

Illustrates how to specify a different element as the request indicator using the hx-indicator attribute. This allows the indicator to be placed anywhere in the DOM, not just as a child of the requesting element. The attribute takes a CSS selector to identify the indicator.

```html
<div>
    <button hx-get="/click" hx-indicator="#indicator">
        Click Me!
    </button>
    <img id="indicator" class="htmx-indicator" src="/spinner.gif" alt="Loading..."/>
</div>
```

--------------------------------

### Programmatically Cancel htmx Requests with htmx:abort Event

Source: https://htmx.org/docs/index

Illustrates how to cancel an in-flight htmx request programmatically. By triggering the 'htmx:abort' event on a specific element, any active requests initiated by that element can be stopped. This is useful for user-initiated cancellations or complex state management.

```html
<button id="request-button" hx-post="/example">
    Issue Request
</button>
<button onclick="htmx.trigger('#request-button', 'htmx:abort')">
    Cancel Request
</button>
```

--------------------------------

### htmx: Triggering Request Once on Mouse Enter

Source: https://htmx.org/docs/index

Illustrates using the 'once' modifier with hx-trigger to ensure an AJAX request is sent only one time. The POST request to '/mouse_entered' will occur the first time the mouse enters the div.

```html
<div hx-post="/mouse_entered" hx-trigger="mouseenter once">
    [Here Mouse, Mouse!]
</div>

```

--------------------------------

### Configure htmx Logger for Event Logging

Source: https://htmx.org/docs/index

This code sets a custom logger function for htmx, which outputs all triggered events, the associated element, and any data to the browser's console. This is useful for debugging and understanding the flow of htmx events within an application.

```javascript
htmx.logger = function(elt, event, data) {
    if(console) {
        console.log(event, elt, data);
    }
}

```

--------------------------------

### HTMX Request Indicator CSS Transitions

Source: https://htmx.org/docs/index

Provides CSS rules to control the visibility of elements with the 'htmx-indicator' class during requests. By default, the indicator is hidden. When an element has the 'htmx-request' class, the indicator becomes visible, offering alternative display mechanisms beyond opacity.

```css
.htmx-indicator{
    display:none;
}
.htmx-request .htmx-indicator{
    display:inline;
}
.htmx-request.htmx-indicator{
    display:inline;
}
```

--------------------------------

### Disabling Attribute Inheritance with 'unset'

Source: https://htmx.org/docs/index

This snippet demonstrates how to prevent attribute inheritance for specific elements. By using `hx-confirm="unset"` on a button within a parent element that has `hx-confirm` defined, the confirmation dialog will not be displayed for that specific button's action. This allows for granular control over inherited attributes.

```html
<div hx-confirm="Are you sure?">
    <button hx-delete="/account">
        Delete My Account
    </button>
    <button hx-put="/account">
        Update My Account
    </button>
    <button hx-confirm="unset" hx-get="/">
        Cancel
    </button>
</div>
```

--------------------------------

### Prevent htmx History Caching with hx-history

Source: https://htmx.org/docs/index

The `hx-history` attribute, when set to `false`, prevents a specific page or section from being stored in the htmx history cache (`localStorage`). This is crucial for pages containing sensitive data that should not persist in the client-side cache.

```html
<div hx-history="false">
    Sensitive Content Here
</div>

```

--------------------------------

### Disable htmx Processing on Content with hx-disable

Source: https://htmx.org/docs/index

The `hx-disable` attribute prevents htmx from processing any of its related attributes or features on the element it's applied to, and all its descendants. This is useful when injecting untrusted HTML content to ensure htmx doesn't inadvertently activate features within that content. This attribute's effect is hierarchical and cannot be overridden by nested content.

```html
<div hx-disable>
    <%= raw(user_content) %>
</div>

```

=== COMPLETE CONTENT === This response contains all available snippets from this library. No additional content exists. Do not make further requests.
