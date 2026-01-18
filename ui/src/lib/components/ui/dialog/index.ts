import Root from './dialog.svelte';
import Overlay from './overlay.svelte';
import Content from './content.svelte';
import Header from './header.svelte';
import Title from './title.svelte';
import Description from './description.svelte';
import Footer from './footer.svelte';

export { Root as Dialog, Overlay, Content, Header, Title, Description, Footer };
export type { default as DialogTitle } from './dialog.svelte';
export type { default as DialogDescription } from './dialog.svelte';
