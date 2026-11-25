import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";

/** Quiet period after the last keystroke before a text filter reaches the URL. */
export const searchDelay = 300;

export type SearchLocation = { q: string; navigationKey: string };
export type QueryEcho = { from: string; to: string };
export const isOwnQueryEcho = (echo: QueryEcho | undefined, previous: SearchLocation, next: SearchLocation) =>
	echo?.from === previous.q && echo.to === next.q;
export const searchLocationChanged = (previous: SearchLocation, next: SearchLocation) =>
	previous.q !== next.q || previous.navigationKey !== next.navigationKey;
export const shouldResetSearchDraft = (previous: SearchLocation, next: SearchLocation, ownQueryEcho: boolean) =>
	(previous.q !== next.q && !ownQueryEcho) || (previous.q === next.q && previous.navigationKey !== next.navigationKey);

/**
 * Draft of a text filter that applies itself: typing commits once after `searchDelay`, URL changes
 * (back/forward, clear) replace the draft, and a pending timer never outlives the component.
 * `settle()` ends the quiet period now and returns the text a select/clear change should submit with.
 */
export function useSearchDraft(q: string, commit: (q: string) => void, navigationKey = "", delay = searchDelay) {
	const [draft, setDraft] = useState(q);
	const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
	const location = useRef<SearchLocation>({ q, navigationKey });
	const ownEcho = useRef<QueryEcho | undefined>(undefined);
	// Older URL echoes from this input must not replace a newer pending draft.
	// The timer reads the latest callback, never the one captured when typing started.
	const latest = useRef(commit);
	latest.current = commit;
	const cancel = useCallback(() => {
		clearTimeout(timer.current);
		timer.current = undefined;
	}, []);
	useLayoutEffect(() => {
		const previous = location.current;
		const next = { q, navigationKey };
		location.current = { q, navigationKey };
		const echo = ownEcho.current;
		const isOwnEcho = isOwnQueryEcho(echo, previous, next);
		ownEcho.current = undefined;
		if (!shouldResetSearchDraft(previous, next, isOwnEcho)) return;
		cancel();
		setDraft(q);
	}, [q, navigationKey, cancel]);
	useEffect(() => cancel, [cancel]);
	const change = (next: string) => {
		setDraft(next);
		cancel();
		if (next === q) return;
		timer.current = setTimeout(() => {
			timer.current = undefined;
			ownEcho.current = { from: location.current.q, to: next };
			latest.current(next);
		}, delay);
	};
	const settle = (next = draft) => {
		cancel();
		setDraft(next);
		if (next !== q) ownEcho.current = { from: location.current.q, to: next };
		return next;
	};
	return { draft, change, settle };
}
