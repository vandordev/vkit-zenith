/** Dummy write results the page shows until navigation; refetched data never replaces them. Keyed by record id (the dummy API reuses one id per kind). */
export type Simulation<T extends { id: string }> = { created: readonly T[]; changed: Readonly<Record<string, T>> };
export const noSimulation: Simulation<never> = { created: [], changed: {} };

export const simulateCreated = <T extends { id: string }>(simulation: Simulation<T>, item: T): Simulation<T> => ({
	...simulation,
	created: [item, ...simulation.created.filter((existing) => existing.id !== item.id)],
});
export const simulateChanged = <T extends { id: string }>(simulation: Simulation<T>, item: T): Simulation<T> => ({
	...simulation,
	changed: { ...simulation.changed, [item.id]: item },
});

/** Ids of records that exist only as a dummy create result: the fixtures cannot open or change them. */
export const createdIds = <T extends { id: string }>(simulation: Simulation<T>): ReadonlySet<string> =>
	new Set(simulation.created.map((item) => item.id));

/**
 * Server rows with changed records substituted in place, then the active filter applied to the final (overlaid) record: an archive under
 * status=active removes the row. Unchanged server rows are the server's own filter result and pass as is. Created records lead page 1.
 *
 * `statusFiltered`: a status filter is active, so the server's response is missing every record whose stored status differs from it, even
 * though a dummy write changed that record to match (revoke a key, then pick "Dicabut"). Such changed records, absent from the response and
 * accepted by the filter on their final state, are added to page 1 after the created ones and before the server rows. They are never added
 * on page > 1, never twice (a record in the response or created in this session is already shown) and never when no status filter is set
 * (then the server's list already holds the record, on some page).
 */
export function applySimulation<T extends { id: string }>(
	items: readonly T[],
	simulation: Simulation<T>,
	page: number,
	accepts: (item: T) => boolean,
	statusFiltered = false,
): T[] {
	const final = (item: T) => simulation.changed[item.id] ?? item;
	const shown = items.filter((item) => !(item.id in simulation.changed) || accepts(final(item))).map(final);
	const known = new Set(items.map((item) => item.id));
	if (page > 1) return shown;
	const created = simulation.created
		.filter((item) => !known.has(item.id))
		.map(final)
		.filter(accepts);
	const shownIds = new Set([...known, ...simulation.created.map((item) => item.id)]);
	const moved = statusFiltered ? Object.values(simulation.changed).filter((item) => !shownIds.has(item.id) && accepts(item)) : [];
	return [...created, ...moved, ...shown];
}
