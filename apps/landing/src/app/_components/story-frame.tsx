import type { Decorator } from "@storybook/react";
import "../landing.css";

/** Storybook only: the landing scope (light theme only, ADR-0019). */
export const withLandingFrame: Decorator = (Story) => (
	<div className="lp">
		<Story />
	</div>
);
