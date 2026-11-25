import { brand } from "@repo/brand";

export const appConfig = {
	appName: brand.name,
	defaultTitle: brand.name,
	defaultDescription: "Fondasi SaaS multi-organisasi dari Vandor: akun, organisasi, anggota, dan API key siap pakai.",
	favicon: "/favicon.ico",
	repositoryUrl: "https://github.com/vandordev/vkit-zenith",
	// shortcut: a constant approved by the CTO, not deployment config; move it to the YAML public config if staging needs its own origin.
	siteOrigin: "https://app.example.com",
	locale: "id_ID",
} as const;
