import { describe, expect, test } from "bun:test";

import { appConfig } from "./config";

describe("web app config", () => {
	test("exposes the default brand configuration", () => {
		expect(appConfig).toEqual({
			appName: "Zenith",
			defaultTitle: "Zenith",
			defaultDescription: "Fondasi SaaS multi-organisasi dari Vandor: akun, organisasi, anggota, dan API key siap pakai.",
			favicon: "/favicon.ico",
			repositoryUrl: "https://github.com/vandordev/vkit-zenith",
			siteOrigin: "https://app.example.com",
			locale: "id_ID",
		});
	});
});
