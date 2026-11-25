import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { test } from "node:test";
import { pageMetadata, siteUrl } from "./seo";
import { sitemapPaths } from "./sitemap";

test("static page discovery excludes privacy, private, parallel and dynamic routes", () => {
	const directory = mkdtempSync(join(tmpdir(), "zenith-sitemap-"));
	try {
		for (const page of [
			"page.tsx",
			"about/page.tsx",
			"(public)/guide/page.tsx",
			"privacy/page.tsx",
			"privacy/archive/page.tsx",
			"_private/page.tsx",
			"@modal/page.tsx",
			"blog/[slug]/page.tsx",
		]) {
			const filename = join(directory, page);
			mkdirSync(dirname(filename), { recursive: true });
			writeFileSync(filename, "");
		}
		assert.deepEqual(sitemapPaths(directory), ["/", "/about", "/guide"]);
	} finally {
		rmSync(directory, { recursive: true, force: true });
	}
});

test("page metadata uses the canonical host, social image and shared indexing policy", () => {
	assert.equal(siteUrl, "https://example.com");
	const home = pageMetadata("Home", "Description", "/");
	assert.deepEqual(home.alternates, { canonical: "/" });
	assert.ok(home.robots && typeof home.robots === "object" && home.robots.index === true && home.robots.follow === true);
	const privacy = pageMetadata("Privacy", "Description", "/privacy");
	assert.ok(privacy.robots && typeof privacy.robots === "object" && privacy.robots.index === false && privacy.robots.follow === true);
	assert.ok(home.twitter && "card" in home.twitter && home.twitter.card === "summary_large_image");
});
