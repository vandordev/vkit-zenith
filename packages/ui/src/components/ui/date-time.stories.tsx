import type { Meta, StoryObj } from "@storybook/react";
import { DateTime } from "./date-time";
import { useEffect, useRef, useState } from "react";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";

const meta = {
	title: "UI/DateTime",
	component: DateTime,
	args: { value: "2026-10-24T08:00:00Z" },
	parameters: { layout: "centered" },
} satisfies Meta<typeof DateTime>;
export default meta;
type Story = StoryObj<typeof meta>;
export const BrowserLocal: Story = {};
export const CustomFormat: Story = {
	args: { locale: "en-GB", timeZone: "Asia/Jakarta", format: { dateStyle: "long", timeStyle: "short" } },
};
export const DateOnly: Story = { args: { format: { day: "numeric", month: "long", year: "numeric" } } };
export const Missing: Story = { args: { value: null } };
export const Invalid: Story = { args: { value: "invalid" } };

function HydrationPreview() {
	const container = useRef<HTMLDivElement>(null);
	const [serverText, setServerText] = useState("");
	const [error, setError] = useState("");
	useEffect(() => {
		const element = container.current;
		if (!element) return;
		const content = <DateTime value="2026-10-24T08:00:00Z" />;
		element.innerHTML = renderToString(content);
		setServerText(element.textContent ?? "");
		const root = hydrateRoot(element, content, {
			onRecoverableError: (cause) => setError(String(cause)),
		});
		return () => root.unmount();
	}, []);
	return (
		<div className="grid gap-3">
			<p>
				Server fallback: <span data-server-fallback>{serverText}</span>
			</p>
			<div ref={container} data-hydrated-date />
			<p role="status">{error || "No hydration errors"}</p>
		</div>
	);
}

/** Real renderToString → hydrateRoot boundary with local fixtures, no backend. */
export const Hydration: Story = { render: () => <HydrationPreview /> };
