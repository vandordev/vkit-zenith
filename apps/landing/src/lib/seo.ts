import { brand } from "@repo/brand";
import type { Metadata } from "next";

export const siteUrl = "https://example.com";
export const siteTitle = `${brand.name} — Fondasi SaaS dari Vandor`;
export const siteDescription = "Akun, organisasi, anggota, backoffice, dan API key siap pakai untuk produk SaaS Anda.";
// Legal pages carry no search value; they stay crawlable so robots can read noindex.
export const noindexPaths = ["/privacy"];
export const socialImage = {
	url: "/images/social-preview.png",
	width: 1200,
	height: 630,
	alt: `${brand.name}: fondasi SaaS dari Vandor.`,
};

export function isIndexablePath(path: string): boolean {
	return !noindexPaths.some((excluded) => path === excluded || path.startsWith(`${excluded}/`));
}

export function pageMetadata(title: string, description: string, path: string): Metadata {
	return {
		title,
		description,
		alternates: { canonical: path },
		robots: {
			index: isIndexablePath(path),
			follow: true,
			googleBot: { index: isIndexablePath(path), follow: true, "max-image-preview": "large" },
		},
		openGraph: {
			title,
			description,
			url: path,
			type: "website",
			locale: "id_ID",
			siteName: brand.name,
			images: [socialImage],
		},
		twitter: { card: "summary_large_image", title, description, images: [socialImage] },
	};
}
