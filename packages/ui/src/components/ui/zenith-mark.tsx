import { brand } from "@repo/brand";
import { zenithMark } from "@repo/brand/logo-mark";
import { cn } from "cn";
import type { ComponentProps } from "react";

/** The brand name in the sans family, semibold and slightly larger than its context to balance the mark. */
export const zenithWordmarkClassName = "text-[1.1em] leading-tight font-semibold tracking-[-0.02em]";

/**
 * Without `surface`: the brand name. With `surface` (e.g. "Backoffice", "Docs"): "/ Surface" for
 * use beside the mark, which already says Zenith; screen readers still hear "Zenith Surface".
 */
export function ZenithWordmark({ surface, className, ...props }: Omit<ComponentProps<"span">, "children"> & { surface?: string }) {
	return (
		<span data-slot="zenith-wordmark" {...props} className={cn(zenithWordmarkClassName, className)}>
			{surface === undefined ? (
				brand.name
			) : (
				<>
					<span className="sr-only">{brand.name} </span>
					<span aria-hidden="true" className="me-2 text-muted-foreground">
						/
					</span>
					{surface}
				</>
			)}
		</span>
	);
}

export type ZenithMarkProps = Omit<ComponentProps<"svg">, "children"> & {
	size?: number;
	/** Accessible name; without it the mark is decorative. */
	title?: string;
};

/** Transparent theme-aware logo: ink in light mode, cream in dark mode. */
export function ZenithMark({ size = 32, title, ...props }: ZenithMarkProps) {
	return (
		<svg
			data-slot="zenith-mark"
			{...props}
			viewBox={zenithMark.viewBox}
			width={size}
			height={size}
			role={title ? "img" : undefined}
			aria-hidden={title ? undefined : true}
			focusable="false"
		>
			{title ? <title>{title}</title> : null}
			<g style={{ fill: "var(--primary)" }}>
				{zenithMark.paths.map((path) => (
					<path key={path} d={path} />
				))}
			</g>
		</svg>
	);
}
