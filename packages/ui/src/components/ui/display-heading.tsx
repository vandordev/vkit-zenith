import { cn } from "cn";
import type { ReactNode } from "react";

export interface DisplayHeadingProps {
	as?: "h1" | "h2";
	children: ReactNode;
	className?: string;
}

/** Display headline (sans, semibold, tight) for auth layouts and landing titles. Keep it ≥ 18px. */
export const DisplayHeading = ({ as: Tag = "h1", children, className }: DisplayHeadingProps) => (
	<Tag
		data-slot="display-heading"
		className={cn("text-4xl leading-tight font-semibold tracking-tight text-balance wrap-break-word", className)}
	>
		{children}
	</Tag>
);
