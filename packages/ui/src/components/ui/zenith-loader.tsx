"use client";

import { zenithMark } from "@repo/brand/logo-mark";
import { cn } from "cn";
import type { ComponentProps } from "react";
import { ZenithMark } from "./zenith-mark";

export type ZenithLoaderProps = Omit<ComponentProps<"div">, "children"> & {
	/** Accessible status text, visually hidden by default. */
	label?: string;
	hideLabel?: boolean;
	/** Screen fills the viewport in normal flow; container fills its parent's height. */
	mode?: "screen" | "container";
	size?: number;
};

/** Draw the brand-owned outlines, then reveal the theme-aware Zenith mark. */
export function ZenithLoader({
	label = "Memuat…",
	hideLabel = true,
	mode = "container",
	size = 72,
	className,
	...props
}: ZenithLoaderProps) {
	return (
		<div
			{...props}
			role="status"
			aria-live="polite"
			aria-atomic="true"
			data-slot="zenith-loader"
			data-mode={mode}
			className={cn(
				"flex w-full flex-col items-center justify-center gap-4 p-6",
				mode === "screen" ? "min-h-dvh" : "h-full min-h-40",
				className,
			)}
		>
			<div aria-hidden="true" className="relative shrink-0 text-primary" style={{ width: size, height: size }}>
				<svg
					viewBox={zenithMark.viewBox}
					fill="none"
					stroke="currentColor"
					strokeWidth={8}
					strokeLinecap="round"
					strokeLinejoin="round"
					focusable="false"
					className="absolute inset-0 size-full"
				>
					{zenithMark.paths.map((path) => (
						<path key={path} data-slot="zenith-loader-outline" d={path} pathLength={1} />
					))}
				</svg>
				<ZenithMark data-slot="zenith-loader-logo" size={size} className="absolute inset-0 size-full" />
			</div>
			<span className={hideLabel ? "sr-only" : "text-center text-sm text-muted-foreground"}>{label}</span>
		</div>
	);
}
