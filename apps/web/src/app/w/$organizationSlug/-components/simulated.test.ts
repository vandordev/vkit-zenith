import { expect, test } from "bun:test";
import { applySimulation, createdIds, noSimulation, simulateChanged, simulateCreated } from "./simulated";

type Row = { id: string; name: string };
const server: Row[] = [{ id: "a", name: "A" }];
const none = noSimulation as typeof noSimulation & { created: readonly Row[]; changed: Record<string, Row> };

test("created records lead page 1 only and repeated ids keep the latest", () => {
	let simulation = simulateCreated(none, { id: "new", name: "Satu" });
	simulation = simulateCreated(simulation, { id: "new", name: "Dua" });
	expect(applySimulation(server, simulation, 1, () => true).map((row) => row.name)).toEqual(["Dua", "A"]);
	expect(applySimulation(server, simulation, 2, () => true)).toEqual(server);
	expect(applySimulation(server, simulation, 1, () => false)).toEqual(server);
});

test("changed records replace the refetched row in place", () => {
	const simulation = simulateChanged(none, { id: "a", name: "Diubah" });
	expect(applySimulation(server, simulation, 1, () => true)).toEqual([{ id: "a", name: "Diubah" }]);
});

test("the filter applies to the final record: an archive leaves an active-only list, a created row changed to archived leaves it too", () => {
	type Item = Row & { status: string };
	const rows: Item[] = [
		{ id: "a", name: "A", status: "active" },
		{ id: "b", name: "B", status: "active" },
	];
	const none = noSimulation as typeof noSimulation & { created: readonly Item[]; changed: Record<string, Item> };
	const active = (item: Item) => item.status === "active";
	const archived = simulateChanged(none, { id: "a", name: "A", status: "archived" });
	expect(applySimulation(rows, archived, 1, active).map((row) => row.id)).toEqual(["b"]);
	// Without a status filter the archived row stays, updated in place.
	expect(applySimulation(rows, archived, 1, () => true).map((row) => row.status)).toEqual(["archived", "active"]);
	const created = simulateChanged(simulateCreated(none, { id: "new", name: "N", status: "active" }), {
		id: "new",
		name: "N",
		status: "archived",
	});
	expect(applySimulation(rows, created, 1, active).map((row) => row.id)).toEqual(["a", "b"]);
	expect(applySimulation(rows, created, 1, () => true).map((row) => row.id)).toEqual(["new", "a", "b"]);
});

test("created ids are tracked so follow-ups the fixtures cannot serve are not offered", () => {
	expect(createdIds(simulateCreated(none, { id: "new", name: "N" })).has("new")).toBe(true);
	expect(createdIds(none).size).toBe(0);
});

// Revoke an active key, then pick the "revoked" filter: the fixture-backed server answers for revoked keys only, so the changed record is
// missing from its response and must be added from the overlay.
type Key = { id: string; name: string; status: string };
const keys = noSimulation as typeof noSimulation & { created: readonly Key[]; changed: Record<string, Key> };
const revokedFilter = (item: Key) => item.status === "revoked";
const revokedServer: Key[] = [{ id: "old", name: "Lama", status: "revoked" }];
const revokedNow = simulateChanged(keys, { id: "k1", name: "Aktif", status: "revoked" });

test("a record changed to match a status filter joins page 1 even though the server response omits it", () => {
	const shown = applySimulation(revokedServer, revokedNow, 1, revokedFilter, true);
	expect(shown.map((row) => row.id)).toEqual(["k1", "old"]);
	expect(shown[0]?.status).toBe("revoked");
});

test("it is not added on page > 1, without a status filter, or when the final record does not match", () => {
	expect(applySimulation(revokedServer, revokedNow, 2, revokedFilter, true)).toEqual(revokedServer);
	// No status filter (text-only or none): the server list already holds the record somewhere, so nothing is prepended.
	expect(applySimulation(revokedServer, revokedNow, 1, () => true, false)).toEqual(revokedServer);
	const stillActive = simulateChanged(keys, { id: "k1", name: "Aktif", status: "active" });
	expect(applySimulation(revokedServer, stillActive, 1, revokedFilter, true)).toEqual(revokedServer);
	// A search that the final record does not match excludes it as well.
	const textFilter = (item: Key) => item.status === "revoked" && item.name.includes("zzz");
	expect(applySimulation([], revokedNow, 1, textFilter, true)).toEqual([]);
});

test("never twice: a record already in the response, or created in this session, appears once", () => {
	const inResponse = [{ id: "k1", name: "Aktif", status: "revoked" }, ...revokedServer];
	expect(applySimulation(inResponse, revokedNow, 1, revokedFilter, true).map((row) => row.id)).toEqual(["k1", "old"]);
	// Created and then changed (same id): the created row carries the final state once.
	const createdThenRevoked = simulateChanged(simulateCreated(keys, { id: "n", name: "Baru", status: "active" }), {
		id: "n",
		name: "Baru",
		status: "revoked",
	});
	expect(applySimulation(revokedServer, createdThenRevoked, 1, revokedFilter, true).map((row) => row.id)).toEqual(["n", "old"]);
	// A created record updated later keeps its position (before moved records) with the latest values.
	const both = simulateChanged(simulateCreated(keys, { id: "n", name: "Baru", status: "revoked" }), {
		id: "k1",
		name: "Aktif",
		status: "revoked",
	});
	expect(applySimulation(revokedServer, both, 1, revokedFilter, true).map((row) => row.id)).toEqual(["n", "k1", "old"]);
});

test("a later change of the same record replaces the earlier one; flipping back to non-matching removes it", () => {
	const back = simulateChanged(revokedNow, { id: "k1", name: "Aktif", status: "active" });
	expect(applySimulation(revokedServer, back, 1, revokedFilter, true).map((row) => row.id)).toEqual(["old"]);
});
