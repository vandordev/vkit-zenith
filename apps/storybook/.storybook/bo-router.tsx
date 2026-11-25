import type { Decorator } from "@storybook/react";
import { createMemoryHistory, createRootRoute, createRoute, createRouter, RouterProvider } from "@tanstack/react-router";
import { createContext, useContext, useState, type ReactNode } from "react";

const PreviewContent = createContext<ReactNode>(null);
function StoryContent() {
	return useContext(PreviewContent);
}

// Presentation that links declaratively (BO and the Web organization dashboard) gets the in-memory router.
const routed = ["BO/", "Web/Pages/Dashboard/", "Web/Layouts/OrgShell", "Web/Components/Dashboard/"];
export const usesMemoryRouter = (title: string) => routed.some((prefix) => title.startsWith(prefix));

// Never import the app route tree, loaders, auth providers or API clients.
// Every in-memory destination renders the same isolated story.
export const withBORouter: Decorator = (Story, context) => {
	if (!usesMemoryRouter(context.title)) return <Story />;
	const initialPath = typeof context.parameters.routerInitialPath === "string" ? context.parameters.routerInitialPath : "/";
	return <BORouter key={`${context.id}:${initialPath}`} story={Story} initialPath={initialPath} />;
};

function BORouter({ story: Story, initialPath }: { story: Parameters<Decorator>[0]; initialPath: string }) {
	const [router] = useState(() => {
		const root = createRootRoute({ component: StoryContent });
		const destination = createRoute({ getParentRoute: () => root, path: "$", component: () => null });
		return createRouter({
			routeTree: root.addChildren([destination]),
			history: createMemoryHistory({ initialEntries: [initialPath] }),
			defaultPreload: false,
		});
	});
	return (
		<PreviewContent.Provider value={<Story />}>
			<RouterProvider router={router} />
		</PreviewContent.Provider>
	);
}
