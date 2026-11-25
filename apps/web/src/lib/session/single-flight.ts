/** Ignores calls while one is pending so a double click sends one request. */
export function singleFlight<A extends unknown[]>(fn: (...a: A) => Promise<void>): (...a: A) => void {
	let pending = false;
	return (...a) => {
		if (pending) return;
		pending = true;
		void fn(...a).finally(() => {
			pending = false;
		});
	};
}
