## ⚠️ CRITICAL: Nested x-data Scope Rules

**NEVER create nested `x-data` attributes.** This creates isolated scopes that break state access.

### The Problem

When you add `x-data` to a child element, it creates a NEW isolated scope. The child can NO longer access parent state:

```html
<!-- ❌ BROKEN: Nested x-data creates isolated scope -->
<div x-data="{ count: 0, name: 'parent' }">
  <div x-data="{ helper: 'child' }">
    <!-- count is UNDEFINED here! Parent state is inaccessible -->
    <span x-text="count"></span>  <!-- FAILS -->
  </div>
</div>
```

### The Solution

**Merge all state into the parent `x-data`:**

```html
<!-- ✅ CORRECT: Single x-data with all state -->
<div x-data="{ count: 0, name: 'parent', helper: 'child' }">
  <div>
    <span x-text="count"></span>  <!-- Works! -->
  </div>
</div>
```

### Common Mistake: Adding x-data for `app` Access

```html
<!-- ❌ WRONG: Adding x-data to access global app state -->
<div x-data="parentState()">
  <div x-data="{ get app() { return Alpine.$data(document.body); } }">
    <!-- parentState properties are now GONE -->
  </div>
</div>

<!-- ✅ CORRECT: Add app getter to parent state -->
<script>
function parentState() {
  return {
    count: 0,
    get app() { return Alpine.$data(document.body); }
  }
}
</script>
<div x-data="parentState()">
  <div>
    <!-- Both count and app accessible -->
  </div>
</div>
```

### Key Rules

1. **One x-data per component tree** - all state in the root
2. **Need to access global state?** Add a getter to the parent x-data, not a new x-data
3. **$root only goes to nearest x-data** - not to top-level, so avoid nested x-data entirely
4. **Use $dispatch for cross-component communication** instead of trying to access other scopes

### Symptoms of Nested x-data Bugs

- State "resets to 0" or initial values
- Variables are "undefined" 
- Changes don't persist
- Functions not found

**If you see these, search for nested x-data attributes.**

--------------------------------

## ⚠️ CRITICAL: HTMX + Alpine.js Integration

When HTMX swaps in new HTML content with `x-data`, Alpine.js does NOT automatically initialize it. You MUST call `Alpine.initTree()` on the swapped content.

### The Fix (already in app.js)

```javascript
document.addEventListener('htmx:afterSwap', (event) => {
  if (window.Alpine && event.detail.target) {
    Alpine.initTree(event.detail.target);
  }
});
```

### Symptoms Without This Fix

- `x-data` components loaded via HTMX don't work at all
- Variables are undefined
- Event handlers don't fire
- x-model doesn't bind

--------------------------------

### Alpine.js x-init for Initialization Code

Source: https://alpinejs.dev/docs/index

The `x-init` directive allows you to run JavaScript code when an Alpine.js component is initialized. This is useful for setting up initial states or performing one-time setup tasks.

```html
<div x-init="date = new Date()"></div>
```

--------------------------------

### Alpine.js x-on for Event Handling

Source: https://alpinejs.dev/docs/index

The `x-on` directive listens for browser events on an element and executes Alpine.js expressions in response. This example uses `x-on:click` to toggle a boolean variable.

```html
<button x-on:click="open = ! open">
    Toggle
</button>
```

--------------------------------

### Basic Alpine.js Component with x-data and x-show

Source: https://alpinejs.dev/docs/index

Demonstrates a fundamental Alpine.js component. The `x-data` attribute initializes a component's state, and `x-show` toggles the visibility of an element based on that state. This example creates a simple expand/collapse functionality.

```html
<div x-data="{ open: false }">
    <button @click="open = true">Expand</button>

    <span x-show="open">
        Content...
    </span>
</div>
```

--------------------------------

### Alpine.js x-for for Rendering Lists

Source: https://alpinejs.dev/docs/index

The `x-for` directive is used to repeat a block of HTML for each item in a collection. It's commonly used with `<template>` tags to render lists of data, like blog posts in this example.

```html
<template x-for="post in posts">
    <h2 x-text="post.title"></h2>
</template>
```

--------------------------------

### Alpine.js x-bind for Dynamic Attributes

Source: https://alpinejs.dev/docs/index

The `x-bind` directive allows for dynamic binding of HTML attributes to Alpine.js data. This example shows how to conditionally apply a CSS class ('hidden') based on the `open` state.

```html
<div x-bind:class="! open ? 'hidden' : ''">
    ...
</div>
```

--------------------------------

### Alpine.js x-text for Text Content

Source: https://alpinejs.dev/docs/index

The `x-text` directive sets the text content of an element. It's useful for displaying dynamic data, such as the current year, as shown in this example.

```html
<div>
    Copyright © 
    <span x-text="new Date().getFullYear()"></span>
</div>
```

--------------------------------

### Alpine.js x-model for Input Synchronization

Source: https://alpinejs.dev/docs/index

The `x-model` directive creates a two-way binding between a piece of data and an input element. Changes in the input update the data, and changes in the data update the input, as shown in this search input example.

```html
<div x-data="{ search: '' }">
    <input type="text" x-model="search">
    
    Searching for: <span x-text="search"></span>
</div>
```

--------------------------------

### Alpine.js Alpine.store for Global Stores

Source: https://alpinejs.dev/docs/index

The `Alpine.store()` method defines a global, reactive data store that can be accessed from anywhere in the application using the `$store` magic property. This is ideal for managing shared application state.

```javascript
Alpine.store('notifications', {
  items: [],

  notify(message) {
    this.items.push(message)
  }
})
```

--------------------------------

### Alpine.js $store for Global State Management

Source: https://alpinejs.dev/docs/index

The `$store` magic property allows access to globally registered reactive data stores. These stores are defined using `Alpine.store()` and can be accessed from any component.

```html
<h1 x-text="$store.site.title"></h1>
```

--------------------------------

### Include Alpine.js Framework

Source: https://alpinejs.dev/docs/index

This snippet shows how to include the Alpine.js framework in your HTML document using a script tag from a CDN. This is the primary way to enable Alpine.js functionality on your page.

```html
<script src="//unpkg.com/alpinejs" defer></script>
```

--------------------------------

### Alpine.js Alpine.data for Reusable Components

Source: https://alpinejs.dev/docs/index

The `Alpine.data()` method allows you to define reusable data objects that can be referenced by `x-data`. This promotes code organization and reusability for common component patterns.

```javascript
Alpine.data('dropdown', () => ({
  open: false,

  toggle() {
    this.open = ! this.open
  }
}))
```

--------------------------------

### Alpine.js x-ignore for Skipping Initialization

Source: https://alpinejs.dev/docs/index

The `x-ignore` directive prevents a block of HTML from being initialized by Alpine.js. This is useful when you have complex, manually managed JavaScript within an Alpine.js application that you don't want Alpine to interfere with.

```html
<div x-ignore>
    ...
</div>
```

--------------------------------

### Alpine.js x-transition for Element Transitions

Source: https://alpinejs.dev/docs/index

The `x-transition` directive adds CSS transition effects to elements when they are shown or hidden by Alpine.js. This allows for smooth visual feedback during visibility changes.

```html
<div x-show="open" x-transition>
    ...
</div>
```

--------------------------------

### Alpine.js x-cloak for Initial Rendering

Source: https://alpinejs.dev/docs/index

The `x-cloak` attribute hides an element until Alpine.js has finished initializing its contents. This prevents FOUC (Flash of Unstyled Content) by ensuring that Alpine-processed content is only visible after it's ready.

```html
<div x-cloak>
    ...
</div>
```

--------------------------------

### Alpine.js x-effect for Reactive Effects

Source: https://alpinejs.dev/docs/index

The `x-effect` directive executes a script whenever any of its reactive dependencies change. This is powerful for logging side effects or triggering actions based on data mutations.

```html
<div x-effect="console.log('Count is '+count)"></div>
```

--------------------------------

### Alpine.js $nextTick for Next Browser Paint

Source: https://alpinejs.dev/docs/index

The `$nextTick` magic property defers the execution of a callback function until the next browser paint cycle. This is useful for ensuring that DOM updates have been rendered before acting upon them.

```html
<div
  x-text="count"
  x-text="$nextTick(() => {
    console.log('count is ' + $el.textContent)
  })">
...</div>
```

--------------------------------

### Alpine.js $dispatch for Custom Events

Source: https://alpinejs.dev/docs/index

The `$dispatch` magic property allows you to dispatch custom browser events from the current element. This enables communication between different Alpine.js components or with other parts of your application.

```html
<div x-on:notify="...">
    <button x-on:click="$dispatch('notify')">...</button>
</div>
```

--------------------------------

### Alpine.js x-html for Inner HTML

Source: https://alpinejs.dev/docs/index

The `x-html` directive sets the inner HTML of an element. This is useful for injecting dynamic HTML content, potentially fetched from an API, as demonstrated with an `axios` call.

```html
<div x-html="(await axios.get('/some/html/partial')).data">
    ...
</div>
```

--------------------------------

### Alpine.js x-ref and $refs for Element Referencing

Source: https://alpinejs.dev/docs/index

The `x-ref` attribute assigns a key to an element, which can then be accessed via the `$refs` magic property. This allows direct manipulation or access to specific DOM elements within Alpine.js logic.

```html
<input type="text" x-ref="content">
<button x-on:click="navigator.clipboard.writeText($refs.content.value)">
    Copy
</button>
```

--------------------------------

### Alpine.js x-if for Conditional Rendering

Source: https://alpinejs.dev/docs/index

The `x-if` directive conditionally adds or removes a block of HTML from the DOM. Unlike `x-show` which toggles CSS display, `x-if` completely manipulates the element's presence in the document.

```html
<template x-if="open">
    <div>...</div>
</template>
```

--------------------------------

### Alpine.js $watch for Data Observation

Source: https://alpinejs.dev/docs/index

The `$watch` magic property allows you to observe changes in a piece of Alpine.js data and execute a callback function when the data changes. This is useful for reacting to state updates.

```html
<div x-init="$watch('count', value => {
  console.log('count is ' + value)
})">...</div>
```

--------------------------------

### Alpine.js $el for Referencing Current Element

Source: https://alpinejs.dev/docs/index

The `$el` magic property provides a reference to the current DOM element the Alpine.js directive is attached to. This is useful for direct DOM manipulation or integrating with third-party libraries.

```html
<div x-init="new Pikaday($el)"></div>
```

=== COMPLETE CONTENT === This response contains all available snippets from this library. No additional content exists. Do not make further requests.
