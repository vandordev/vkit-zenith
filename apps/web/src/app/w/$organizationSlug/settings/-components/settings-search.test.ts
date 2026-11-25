import { expect, test } from "bun:test";
import { applySimulation, simulateChanged, simulateCreated, noSimulation } from "../../-components/simulated";
import { matchesSettingsFilter, searchFields } from "./settings-search";

const key = { id: "k1", status: "active", name: "Kunci integrasi toko", prefix: "zk_abcd" };
const hit = (q: string, item = key, status?: string) => matchesSettingsFilter({ q, status }, item, searchFields.apiKey(item));

test("API key search covers name and prefix, case-insensitively, as api_keys_list.go does", () => {
	for (const q of ["integrasi", "TOKO", "zk_ab", "ABCD", ""]) expect(hit(q)).toBe(true);
	expect(hit("tidak ada")).toBe(false);
	// Each field is checked alone: text spanning the end of one field and the start of the next does not match.
	expect(hit("toko zk_")).toBe(false);
});

test("status filter applies to the final record, together with the text", () => {
	expect(hit("zk_ab", { ...key, status: "revoked" }, "active")).toBe(false);
	expect(hit("zk_ab", { ...key, status: "revoked" }, "revoked")).toBe(true);
	expect(hit("zzz", key, "active")).toBe(false);
});

test("a changed key leaves a filtered list only on a status mismatch, a created one joins only when the server would return it", () => {
	const search = { q: "abcd", status: undefined as string | undefined };
	const accepts = (item: typeof key) => matchesSettingsFilter(search, item, searchFields.apiKey(item));
	const revoked = { ...key, status: "revoked" };
	// No status filter: the prefix still matches, so the revoked row stays (and shows its new status).
	expect(applySimulation([key], simulateChanged(noSimulation, revoked), 1, accepts)).toEqual([revoked]);
	// Active filter: the final status no longer matches, so only then it drops.
	const acceptsActive = (item: typeof key) => matchesSettingsFilter({ q: "abcd", status: "active" }, item, searchFields.apiKey(item));
	expect(applySimulation([key], simulateChanged(noSimulation, revoked), 1, acceptsActive)).toEqual([]);
	// Created by prefix only, by name only, and a non-matching one.
	const byPrefix = { ...key, id: "p", name: "Lain", prefix: "zk_abcd9" };
	expect(applySimulation([], simulateCreated(noSimulation, byPrefix), 1, accepts)).toEqual([byPrefix]);
	const byName = { ...key, id: "n", name: "Kunci abcd", prefix: "zk_zzzz" };
	expect(applySimulation([], simulateCreated(noSimulation, byName), 1, accepts)).toEqual([byName]);
	expect(applySimulation([], simulateCreated(noSimulation, { ...byPrefix, prefix: "zk_zzzz" }), 1, accepts)).toEqual([]);
});
