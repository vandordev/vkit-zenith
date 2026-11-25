"use client";

import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { formatDate, type DateFormatOptions, type DateInput } from "./date";

const subscribe = () => () => {};
const clientSnapshot = () => true;
const serverSnapshot = () => false;

/** Stable fallback on SSR and initial hydration; browser-local formatting afterwards. */
export function useDateFormatter(options: DateFormatOptions = {}): (value: DateInput) => string {
	const client = useSyncExternalStore(subscribe, clientSnapshot, serverSnapshot);
	return (value) => (client ? formatDate(value, options) : (options.fallback ?? "—"));
}

/** Presentation-only minimum loading time; callers must bypass it for errors. */
export function useMinimumPending(pending: boolean, minimumMs = 900): boolean {
	const [held, setHeld] = useState(pending);
	const startedAt = useRef(0);
	useEffect(() => {
		if (pending) {
			startedAt.current = performance.now();
			setHeld(true);
			return;
		}
		const remaining = Math.max(0, minimumMs - (performance.now() - startedAt.current));
		const timer = setTimeout(() => setHeld(false), remaining);
		return () => clearTimeout(timer);
	}, [pending, minimumMs]);
	return pending || held;
}
