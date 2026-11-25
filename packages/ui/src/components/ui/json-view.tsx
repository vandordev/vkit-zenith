"use client";

import { cn } from "cn";
import { useLayoutEffect, useRef, useState } from "react";

import { Button } from "./button";
import { CopyButton } from "./copyable-id";

export interface JsonViewProps {
	value: unknown;
	/** Accessible name of the region and the copy button. */
	label: string;
	/** Collapse past this height in px (default 240). */
	maxHeight?: number;
	className?: string;
}

const stringify = (value: unknown) => {
	try {
		return JSON.stringify(value, null, 2) ?? "null";
	} catch {
		return "—"; // Circular or non-serializable input.
	}
};

export const JsonView = ({ value, label, maxHeight = 240, className }: JsonViewProps) => {
	const text = stringify(value);
	const ref = useRef<HTMLPreElement>(null);
	const [overflowing, setOverflowing] = useState(false);
	const [expanded, setExpanded] = useState(false);
	useLayoutEffect(() => {
		setOverflowing((ref.current?.scrollHeight ?? 0) > maxHeight);
	}, [text, maxHeight]);
	const collapsed = overflowing && !expanded;
	return (
		<div data-slot="json-view" role="group" aria-label={label} className={cn("relative min-w-0 rounded-md border border-border bg-muted/50", className)}>
			<div className="absolute top-1 right-1 z-10">
				<CopyButton value={text} label={label} />
			</div>
			<pre
				ref={ref}
				tabIndex={0}
				style={collapsed ? { maxHeight } : undefined}
				className={cn(
					"overflow-x-auto p-3 pr-10 font-mono text-xs leading-relaxed outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50",
					collapsed && "overflow-y-hidden",
				)}
			>
				{text}
			</pre>
			{collapsed && (
				<div aria-hidden="true" className="pointer-events-none absolute inset-x-0 bottom-9 h-10 bg-linear-to-t from-muted to-transparent" />
			)}
			{overflowing && (
				<div className="border-t border-border p-1">
					<Button type="button" variant="ghost" size="xs" aria-expanded={expanded} onClick={() => setExpanded(!expanded)}>
						{expanded ? "Ciutkan" : "Tampilkan semua"}
					</Button>
				</div>
			)}
		</div>
	);
};
