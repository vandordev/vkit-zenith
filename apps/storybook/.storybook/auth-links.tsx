import type { Decorator } from "@storybook/react";
import { usesMemoryRouter } from "./bo-router";

// Auth presentation uses plain anchors in production; in previews they must not leave the isolated story.
// Click covers mouse and keyboard (Enter) activation, including modified clicks; auxclick covers middle-click.
// Memory-routed titles (organization dashboard) use declarative Link and keep it working; only external anchors are blocked there.
const external = (anchor: Element) => /^https?:\/\//i.test(anchor.getAttribute("href") ?? "");
const keepInPreview = (routed: boolean) => (event: { preventDefault: () => void; target: EventTarget }) => {
	const anchor = event.target instanceof Element ? event.target.closest("a[href]") : null;
	if (anchor && (!routed || external(anchor))) event.preventDefault();
};

export const withAuthLinks: Decorator = (Story, context) => {
	if (!context.title.startsWith("Web/")) return <Story />;
	const handler = keepInPreview(usesMemoryRouter(context.title));
	return (
		<div style={{ display: "contents" }} onClick={handler} onAuxClick={handler}>
			<Story />
		</div>
	);
};
