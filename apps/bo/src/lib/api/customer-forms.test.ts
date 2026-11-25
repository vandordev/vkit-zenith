import { describe, expect, test } from "bun:test";
import { createBOTransport, BOError } from "./client";
import * as sdk from "./generated/sdk.gen";
import {
	organizationCreateBody,
	organizationProfileBody,
	memberAddBody,
	memberRemoveBody,
	memberRestoreBody,
	ownerBody,
	organizationStatusBody,
	userCreateBody,
	userProfileBody,
	userEmailBody,
	userVerifyBody,
	userStatusBody,
} from "./customer-forms";
const id = "00000000-0000-4000-8000-000000000101";
const revision = "a".repeat(64);
describe("customer form payloads", () => {
	test("user SDK commands keep current target revision and separate email/verify/status bodies", async () => {
		const calls: Request[] = [];
		const fetcher = Object.assign(
			async (input: Parameters<typeof fetch>[0]) => {
				const request = new Request(input);
				calls.push(request);
				return Response.json(request.url.endsWith("/csrf") ? { token: "fixture-csrf" } : {});
			},
			{ preconnect: fetch.preconnect },
		);
		const t = createBOTransport({ boApiBaseUrl: "https://api.invalid" }, () => {}, fetcher);
		const confirmation = { revision, reason: "Safe draft", currentPassword: "fixture-secret" };
		await sdk.boV1ChangeUserEmail({
			...(await t.writeOptions()),
			path: { id },
			body: userEmailBody({ ...confirmation, email: "next@example.test" }),
		});
		await sdk.boV1VerifyUserEmail({ ...(await t.writeOptions()), path: { id }, body: userVerifyBody(confirmation) });
		await sdk.boV1SetUserStatus({
			...(await t.writeOptions()),
			path: { id },
			body: userStatusBody({ ...confirmation, status: "disabled" }),
		});
		const writes = calls.filter((r) => r.method === "POST");
		expect(writes.map((r) => new URL(r.url).pathname)).toEqual([
			`/v1/users/${id}/email`,
			`/v1/users/${id}/email/verify`,
			`/v1/users/${id}/status`,
		]);
		expect(await Promise.all(writes.map((r) => r.json()))).toEqual([
			{ ...confirmation, email: "next@example.test" },
			confirmation,
			{ ...confirmation, status: "disabled" },
		]);
		expect(writes.every((r) => r.cache === "no-store" && r.credentials === "include")).toBe(true);
	});
	test("customer SDK reads cancel/no-store and protected write sends exact scoped command once", async () => {
		const calls: Request[] = [];
		const fetcher = Object.assign(
			async (input: Parameters<typeof fetch>[0]) => {
				const request = new Request(input);
				calls.push(request);
				if (request.url.endsWith("/csrf")) return Response.json({ token: "fixture-csrf" });
				if (request.method === "POST") return Response.json({ type: "urn:zenith:bo:conflict", detail: "Revisi berubah." }, { status: 409 });
				return Response.json({});
			},
			{ preconnect: fetch.preconnect },
		);
		const transport = createBOTransport({ boApiBaseUrl: "https://api.invalid" }, () => {}, fetcher);
		const controller = new AbortController();
		await sdk.boV1GetOrganizationMembership({
			client: transport.client,
			path: { organizationId: id, id },
			signal: controller.signal,
			throwOnError: true,
		});
		expect(calls[0]?.url).toBe(`https://api.invalid/v1/organizations/${id}/memberships/${id}`);
		expect(calls[0]?.cache).toBe("no-store");
		expect(calls[0]?.credentials).toBe("include");
		controller.abort();
		expect(calls[0]?.signal.aborted).toBe(true);
		const body = ownerBody({ revision, reason: "Safe reason", currentPassword: "fixture-secret", role: "member" });
		await expect(
			sdk.boV1SetOrganizationMembershipRole({ ...(await transport.writeOptions()), path: { organizationId: id, id }, body }),
		).rejects.toBeInstanceOf(BOError);
		expect(calls.filter((r) => r.method === "POST")).toHaveLength(1);
		const request = calls.at(-1);
		expect(request?.url).toBe(`https://api.invalid/v1/organizations/${id}/memberships/${id}/role`);
		expect(request?.headers.get("X-CSRF-Token")).toBe("fixture-csrf");
		expect(await request?.json()).toEqual(body);
	});
	test("user commands carry exactly one action and current revision, not profile overrides", () => {
		expect(userCreateBody({ displayName: "Nadia", email: "nadia@example.test" })).toEqual({
			displayName: "Nadia",
			email: "nadia@example.test",
		});
		expect(userProfileBody({ displayName: "Nadia", revision })).toEqual({ displayName: "Nadia", revision });
		for (const extra of [{ email: "next@example.test" }, { status: "disabled" }, { verified: true }])
			expect(() => userProfileBody({ displayName: "Nadia", revision, ...extra })).toThrow();
		const confirmation = { revision, reason: "Operations", currentPassword: "fixture-secret" };
		expect(userEmailBody({ ...confirmation, email: "next@example.test" })).toEqual({ ...confirmation, email: "next@example.test" });
		expect(userVerifyBody(confirmation)).toEqual(confirmation);
		expect(userStatusBody({ ...confirmation, status: "disabled" })).toEqual({ ...confirmation, status: "disabled" });
		expect(() => userVerifyBody({ ...confirmation, email: "other@example.test" })).toThrow();
		expect(() => userEmailBody({ ...confirmation, email: "next@example.test", verified: true })).toThrow();
		expect(() => userStatusBody({ ...confirmation, status: "active", revision: "old" })).toThrow();
	});
	test("atomic organization owner mode is exclusive and incompatible input cannot leak", () => {
		expect(organizationCreateBody({ name: "One", slug: "one", ownerUserId: id })).toEqual({ name: "One", slug: "one", ownerUserId: id });
		expect(organizationCreateBody({ name: "One", slug: "one", newOwner: { displayName: "Nadia", email: "nadia@example.test" } })).toEqual({
			name: "One",
			slug: "one",
			newOwner: { displayName: "Nadia", email: "nadia@example.test" },
		});
		for (const extras of [
			{},
			{ ownerUserId: id, newOwner: { displayName: "Nadia", email: "nadia@example.test" } },
			{ ownerUserId: id, status: "suspended" },
			{ newOwner: { displayName: "Nadia", email: "nadia@example.test", verified: true } },
		])
			expect(() => organizationCreateBody({ name: "One", slug: "one", ...extras })).toThrow();
	});
	test("profile cannot carry privileged override and scoped membership action payloads are explicit", () => {
		expect(organizationProfileBody({ name: "Name", slug: "slug", revision })).toEqual({ name: "Name", slug: "slug", revision });
		expect(() => organizationProfileBody({ name: "Name", slug: "slug", revision, status: "suspended" })).toThrow();
		expect(memberAddBody({ userId: id, role: "admin" })).toEqual({ userId: id, role: "admin" });
		expect(() => memberAddBody({ userId: id, role: "owner" })).toThrow();
		expect(memberRemoveBody({ revision })).toEqual({ revision });
		expect(memberRestoreBody({ revision })).toEqual({ revision });
		expect(() => memberRemoveBody({ revision, organizationId: id })).toThrow();
		const sensitive = { revision, reason: "Operational repair", currentPassword: "fixture-secret" };
		expect(ownerBody({ ...sensitive, role: "member" })).toEqual({ ...sensitive, role: "member" });
		expect(organizationStatusBody({ ...sensitive, status: "suspended" })).toEqual({ ...sensitive, status: "suspended" });
		expect(() => ownerBody({ ...sensitive, role: "member", status: "active" })).toThrow();
	});
});
