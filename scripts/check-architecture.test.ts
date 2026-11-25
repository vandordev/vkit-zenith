import { expect, test } from "bun:test";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { checkArchitecture, checkAuthArchitecture } from "./check-architecture";
import { mkdtemp, mkdir, writeFile } from "node:fs/promises";

async function fixture() {
	const root = await mkdtemp("/tmp/opencode/zenith-architecture-");
	await mkdir(join(root, "apps/web/src"), { recursive: true });
	return root;
}

test("permits public config and type imports but rejects browser env and Zod 3", async () => {
	const root = await fixture();
	const page = join(root, "apps/web/src/page.tsx");
	await writeFile(page, 'import "@repo/config/public"; import type { X } from "@repo/config/server"; const dev = import.meta.env.DEV;');
	expect(checkArchitecture(root)).toEqual([]);
	await writeFile(page, 'import "@repo/config/server"; const url = import.meta.env.VITE_API_URL; import "zod/v3";');
	expect(checkArchitecture(root)).toEqual([
		"apps/web/src/page.tsx: Zod 3 contracts are forbidden",
		"apps/web/src/page.tsx: browser code must not import server-only modules",
		"apps/web/src/page.tsx: browser deployment environment is forbidden",
	]);
});

test("permits independent Go ownership and rejects cross-service internals", async () => {
	const root = await fixture();
	await mkdir(join(root, "apps/web-api"), { recursive: true });
	const main = join(root, "apps/web-api/main.go");
	await writeFile(main, 'package main\nimport "example/apps/web-api/internal/server"');
	expect(checkArchitecture(root)).toEqual([]);
	await writeFile(main, 'package main\nimport "example/apps/bo-api/internal/server"');
	expect(checkArchitecture(root)).toEqual(["apps/web-api/main.go: service-internal imports must stay local"]);
});

test("rejects retired imports and source owners", async () => {
	const root = await fixture();
	await writeFile(join(root, "apps/web/src/page.tsx"), 'import "@repo/db";');
	await mkdir(join(root, "apps/scheduler"), { recursive: true });
	await writeFile(join(root, "apps/scheduler/main.go"), "package main");
	expect(checkArchitecture(root)).toEqual([
		"apps/scheduler: retired source owner is forbidden",
		"apps/web/src/page.tsx: retired runtime imports are forbidden",
	]);
});

test("forbids tRPC, browser file readers and presentation SDK calls", async () => {
	const root = await fixture();
	await mkdir(join(root, "apps/web/src/app/-components"), { recursive: true });
	await writeFile(join(root, "apps/web/src/app/-components/page.tsx"), 'import "@/lib/api/client";');
	await writeFile(join(root, "apps/web/src/page.tsx"), 'import "node:fs"; fetch("/trpc/config.public");');
	expect(checkArchitecture(root)).toEqual([
		"apps/web/src/app/-components/page.tsx: presentation must not fetch API or runtime configuration",
		"apps/web/src/page.tsx: browser code must not import server-only modules",
		"apps/web/src/page.tsx: tRPC adapters are forbidden",
	]);
});

test("admits native database/application owners only, retaining retired boundaries", () => {
	const root = mkdtempSync("/tmp/opencode/architecture-db-");
	const put = (path: string, content: string) => {
		mkdirSync(join(root, path, ".."), { recursive: true });
		writeFileSync(join(root, path), content);
	};
	try {
		put("packages/db/doc.go", "package db\n");
		put("packages/application/doc.go", "package application\n");
		expect(checkArchitecture(root)).toEqual([]);
		put("packages/db/package.json", "{}");
		put("packages/application/package.json", "{}");
		put("apps/web/legacy.ts", 'import "@repo/db";');
		put("packages/queue/doc.go", "package queue\n");
		const violations = checkArchitecture(root).join("\n");
		expect(violations).toContain("packages/db: JavaScript wrapper");
		expect(violations).toContain("packages/application: JavaScript wrapper");
		expect(violations).toContain("retired runtime imports");
		expect(violations).toContain("packages/queue: retired source owner");
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

test("BO presentation admits DTO-only imports but rejects actual SDK imports", async () => {
	const root = await fixture();
	const directory = join(root, "apps/bo/src/app/-components");
	try {
		await mkdir(directory, { recursive: true });
		const page = join(directory, "page.tsx");
		await writeFile(page, 'import type { StaffDto } from "../../lib/api/generated/types.gen";');
		expect(checkArchitecture(root)).toEqual([]);
		await writeFile(page, 'import { boV1GetMe } from "../../lib/api/generated/sdk.gen";');
		expect(checkArchitecture(root)).toEqual([
			"apps/bo/src/app/-components/page.tsx: presentation must not fetch API or runtime configuration",
		]);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

test("Auth runtime rejects test resources and obvious secret logging, tools remain test-owned", async () => {
	const root = await fixture();
	try {
		await mkdir(join(root, "apps/web-api/internal/identityhttp"), { recursive: true });
		await mkdir(join(root, "packages/tools/authtest"), { recursive: true });
		await writeFile(
			join(root, "packages/tools/authtest/fixture.go"),
			'package authtest\nimport "github.com/vandordev/vkit-zenith/packages/tools/dbtool"',
		);
		await writeFile(
			join(root, "apps/web-api/internal/identityhttp/delivery.go"),
			'package identityhttp\nimport "github.com/vandordev/vkit-zenith/packages/tools/authtest"\nfunc bad(){log.Printf("%v", delivery.URL)}',
		);
		expect(checkArchitecture(root)).toEqual([
			"apps/web-api/internal/identityhttp/delivery.go: identity secrets must not enter ordinary logs",
			"apps/web-api/internal/identityhttp/delivery.go: runtime must not import test/tooling resources",
		]);
		expect(checkAuthArchitecture(root)).toEqual(checkArchitecture(root));
		await mkdir(join(root, "apps/bo/src/app/-components"), { recursive: true });
		await writeFile(join(root, "apps/bo/src/app/-components/baseline.tsx"), 'import "@/lib/api/client";');
		expect(checkArchitecture(root).length).toBe(3);
		expect(checkAuthArchitecture(root).length).toBe(2);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});
