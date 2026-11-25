import { describe, expect, test } from "bun:test";

import { createMetadata } from "./metadata";

describe("createMetadata", () => {
	test("uses the centralized application brand", async () => {
		const source = await Bun.file(new URL("./metadata.ts", import.meta.url)).text();
		expect(source).toContain('import { appConfig } from "./config"');
		expect(source).toContain("appConfig.appName");
	});

	test("adds the brand and canonical metadata for a route", () => {
		const metadata = createMetadata({ title: "Pricing", description: "Plans", pathname: "/pricing", index: true });
		expect(metadata.meta).toEqual(
			expect.arrayContaining([
				{ title: "Pricing | Zenith" },
				{ name: "description", content: "Plans" },
				{ property: "og:title", content: "Pricing | Zenith" },
				{ property: "og:url", content: "https://app.example.com/pricing" },
				{ name: "robots", content: "index, follow" },
				{ name: "twitter:card", content: "summary" },
			]),
		);
		expect(metadata.links).toEqual([{ rel: "canonical", href: "https://app.example.com/pricing" }]);
	});

	test("does not duplicate the brand title and supports an image", () => {
		const metadata = createMetadata({ title: "Zenith", description: "Home", image: "/fixture.png", index: false });
		expect(metadata.meta).toEqual(
			expect.arrayContaining([
				{ title: "Zenith" },
				{ property: "og:image", content: "https://app.example.com/fixture.png" },
				{ name: "robots", content: "noindex, nofollow" },
				{ name: "twitter:card", content: "summary_large_image" },
			]),
		);
		expect(metadata.links).toEqual([]);
	});
});
