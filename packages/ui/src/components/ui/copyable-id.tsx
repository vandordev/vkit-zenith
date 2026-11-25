"use client";

import { cn } from "cn";
import { CheckIcon, CopyIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { Button } from "./button";

const COPIED_MS = 1500;

export interface CopyButtonProps {
	value: string;
	/** Names what is copied; becomes the accessible label. */
	label: string;
	className?: string;
}

export const CopyButton = ({ value, label, className }: CopyButtonProps) => {
	const [copied, setCopied] = useState(false);
	const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
	useEffect(() => () => clearTimeout(timer.current), []);
	const copy = () => {
		navigator.clipboard?.writeText(value).then(
			() => {
				setCopied(true);
				clearTimeout(timer.current);
				timer.current = setTimeout(() => setCopied(false), COPIED_MS);
			},
			() => undefined, // Clipboard denied: stay silent, the value remains selectable.
		);
	};
	return (
		<>
			<Button type="button" variant="ghost" size="icon-xs" aria-label={copied ? `${label} disalin` : `Salin ${label}`} onClick={copy} className={className}>
				{copied ? <CheckIcon aria-hidden="true" /> : <CopyIcon aria-hidden="true" />}
			</Button>
			<span role="status" className="sr-only">
				{copied ? `${label} disalin` : ""}
			</span>
		</>
	);
};

export interface CopyableIdProps {
	value: string;
	label: string;
	truncate?: "middle" | "none";
	/** Shown instead of `value` (the copied text is always `value`). */
	display?: string;
}

const middle = (text: string) => (text.length > 20 ? `${text.slice(0, 10)}…${text.slice(-6)}` : text);

export const CopyableId = ({ value, label, truncate = "none", display }: CopyableIdProps) => {
	const shown = display ?? value;
	return (
		<span data-slot="copyable-id" className="inline-flex max-w-full min-w-0 items-center gap-1 align-middle">
			<span title={value} className={cn("min-w-0 font-mono text-xs", truncate === "none" && "wrap-anywhere")}>
				{truncate === "middle" ? middle(shown) : shown}
			</span>
			<CopyButton value={value} label={label} />
		</span>
	);
};
