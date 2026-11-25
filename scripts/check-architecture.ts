import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join, relative } from "node:path";

function sourceFiles(root: string, directory: string): string[] {
	const absolute = join(root, directory);
	if (!existsSync(absolute)) return [];
	return readdirSync(absolute, { withFileTypes: true }).flatMap((entry) => {
		const path = join(absolute, entry.name);
		if (entry.isDirectory())
			return ["node_modules", "dist", ".output", ".vite", ".turbo", "storybook-static"].includes(entry.name)
				? []
				: sourceFiles(root, relative(root, path));
		return /\.(?:ts|tsx|go)$/.test(entry.name) && !/\.(?:test|spec)\.tsx?$|_test\.go$/.test(entry.name) ? [path] : [];
	});
}

function goImportPaths(content: string): string[] {
	const source = content.replace(/\/\*[\s\S]*?\*\/|\/\/[^\n]*/g, "");
	const paths: string[] = [];
	for (const block of source.matchAll(/\bimport\s*\(([\s\S]*?)\)/g)) {
		const body = block[1];
		if (body === undefined) continue;
		for (const spec of body.matchAll(/(?:^|[\n;])\s*(?:[._A-Za-z]\w*\s+)?["`]([^"`]+)["`]/g)) {
			const path = spec[1];
			if (path !== undefined) paths.push(path);
		}
	}
	const singles = source.replace(/\bimport\s*\([\s\S]*?\)/g, "");
	for (const spec of singles.matchAll(/\bimport\s+(?:[._A-Za-z]\w*\s+)?["`]([^"`]+)["`]/g)) {
		const path = spec[1];
		if (path !== undefined) paths.push(path);
	}
	return paths;
}

// A conservative source guard, not a substitute for bundle/runtime verification.
export function checkArchitecture(root = process.cwd()): string[] {
	const violations: string[] = [];
	for (const file of [...sourceFiles(root, "apps"), ...sourceFiles(root, "packages")]) {
		const content = readFileSync(file, "utf8");
		const label = relative(root, file).replaceAll("\\", "/");
		const reject = (condition: boolean, reason: string) => {
			if (condition) violations.push(`${label}: ${reason}`);
		};
		reject(/zod\/v3|zod@3\.|ZodTypeAny/.test(content), "Zod 3 contracts are forbidden");
		reject(
			/["']@repo\/(?:application|db|query|queue|realtime|storage)(?:[\/"'])|["'](?:@trpc\/|elysia|@elysiajs\/|@prisma\/|bullmq|ioredis|socket\.io)/.test(
				content,
			),
			"retired runtime imports are forbidden",
		);
		const frontend = /^apps\/(web|bo|auth)\//.exec(label)?.[1];
		if (frontend) {
			reject(
				/(?:import\.meta\.env|process\.env)\s*(?:\.\s*(?:VITE_|PUBLIC_)|\[\s*["'](?:VITE_|PUBLIC_))/.test(content) ||
					/import\.meta\.env(?!\.(?:DEV|SSR|PROD|MODE|BASE_URL)\b)/.test(content),
				"browser deployment environment is forbidden",
			);
			reject(label === `apps/${frontend}/vite.config.ts` && /\bdefine\s*:/.test(content), "browser build injection is forbidden");
			if (label.startsWith(`apps/${frontend}/src/`)) {
				const serverModule =
					label === `apps/${frontend}/src/server.ts` ||
					label.startsWith(`apps/${frontend}/src/server/`) ||
					label === `apps/${frontend}/src/app/health/index.ts`;
				const valueImports = content.replace(/(?:import|export)\s+type\b[^;]*?from\s*["'][^"']+["'];?/g, "");
				reject(
					!serverModule &&
						(/["']@repo\/config(?:["']|\/(?!public["']))/.test(valueImports) ||
							/process\.env|["']node:(?:fs|net|tls|child_process)/.test(content)),
					"browser code must not import server-only modules",
				);
				reject(label.includes("/trpc/") || /["'`]\/trpc(?:[\/"'`])/.test(content), "tRPC adapters are forbidden");
				reject(
					label.includes("/-components/") && /(?:lib\/api|server\/public-config)/.test(valueImports),
					"presentation must not fetch API or runtime configuration",
				);
			}
		}
		if (label.endsWith(".go")) {
			const runtimeOwner = label.startsWith("apps/") || /^packages\/(application|db|email|config)\//.test(label);
			const toolImports = goImportPaths(content).filter(
				(path) => path === "github.com/vandordev/vkit-zenith/packages/tools" || path.startsWith("github.com/vandordev/vkit-zenith/packages/tools/"),
			);
			const blockedToolImports = toolImports;
			reject(
				runtimeOwner && blockedToolImports.length > 0,
				"runtime must not import test/tooling resources",
			);
			const identityBoundary =
				/\/(?:identity|identityhttp)\//.test(label) || label.startsWith("packages/email/");
			reject(
				identityBoundary &&
					/(?:log\.(?:Print\w*|Fatal\w*|Panic\w*)|slog\.(?:Debug|Info|Warn|Error)|fmt\.Print\w*)\([^\n]*(?:\.Password|\.Token|\.Secret|\.Recipient|\.URL|ActionDelivery)/.test(
						content,
					),
				"identity secrets must not enter ordinary logs",
			);
			const service = /^apps\/(api|web-api|bo-api)\//.exec(label)?.[1];
			if (service) {
				reject(
					["api", "web-api", "bo-api"].some((other) => other !== service && content.includes(`/apps/${other}/internal/`)),
					"service-internal imports must stay local",
				);
				reject(label.endsWith("/routes.go"), "routes.go aggregators are forbidden");
				reject(
					!label.includes("/internal/server/") && /os\.(?:Getenv|LookupEnv|Environ)\(/.test(content),
					"feature configuration must use typed loaders",
				);
			}
		}
	}
	for (const owner of ["packages/db", "packages/application"]) {
		if (existsSync(join(root, owner, "package.json"))) violations.push(`${owner}: JavaScript wrapper is forbidden`);
		if (sourceFiles(root, owner).some((file) => !file.endsWith(".go"))) violations.push(`${owner}: only native Go source is permitted`);
	}
	for (const owner of [
		"apps/scheduler",
		"apps/migrate",
		"apps/realtime",
		"packages/query",
		"packages/queue",
		"packages/realtime",
	]) {
		// Ignored local dependency/build leftovers are not source ownership.
		if (existsSync(join(root, owner, "package.json")) || sourceFiles(root, owner).length)
			violations.push(`${owner}: retired source owner is forbidden`);
	}
	return violations.sort();
}

// Focused Auth gates do not silently absorb unrelated BO presentation checks.
// The unfiltered repository checker above retains its original behavior.
export function checkAuthArchitecture(root: string): string[] {
	return checkArchitecture(root).filter(
		(violation) =>
			/^(?:apps\/(?:auth|web|api|web-api)\/|packages\/(?:application\/identity|email|ui-email|config)\/)/.test(violation) ||
			/runtime must not import test\/tooling resources|identity secrets must not enter ordinary logs/.test(violation),
	);
}

if (import.meta.main) {
	const root = process.argv[2] ?? process.cwd();
	const violations = process.argv[3] === "--auth" ? checkAuthArchitecture(root) : checkArchitecture(root);
	if (violations.length) {
		console.error(violations.join("\n"));
		process.exitCode = 1;
	} else console.log("Architecture boundaries passed.");
}
