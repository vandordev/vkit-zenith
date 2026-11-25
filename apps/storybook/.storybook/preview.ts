import type { Decorator, Preview } from "@storybook/react";
import { withAuthLinks } from "./auth-links";
import { withBORouter } from "./bo-router";
import "./styles.css";

// Explicit light/dark: tokens switch on the `.dark` class of an ancestor (the document root).
const withTheme: Decorator = (Story, context) => {
	document.documentElement.classList.toggle("dark", context.globals.theme === "dark");
	return Story();
};

const preview: Preview = {
	decorators: [withTheme, withBORouter, withAuthLinks],
	globalTypes: {
		theme: { description: "Color scheme", toolbar: { title: "Theme", icon: "circlehollow", items: ["light", "dark"], dynamicTitle: true } },
	},
	initialGlobals: { theme: "light" },
	parameters: {
		layout: "padded",
		controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
	},
};

export default preview;
