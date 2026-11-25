import { HeadContent, Scripts, createRootRoute, Outlet, Link } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { brand } from "@repo/brand";
import { BOProvider } from "../lib/api/provider";
import { themeInitScript } from "../lib/theme";
import appCss from "../styles.css?url";
export const Route = createRootRoute({
	head: () => ({
		meta: [
			{ charSet: "utf-8" },
			{ name: "viewport", content: "width=device-width, initial-scale=1" },
			{ title: `${brand.name} Backoffice` },
			{ name: "robots", content: "noindex, nofollow" },
		],
		scripts: [{ children: themeInitScript }],
		links: [
			{ rel: "stylesheet", href: appCss },
			{ rel: "icon", href: "/favicon.svg", type: "image/svg+xml" },
			{ rel: "icon", href: "/favicon-32x32.png", type: "image/png", sizes: "32x32" },
			{ rel: "apple-touch-icon", href: "/apple-touch-icon.png" },
		],
	}),
	shellComponent: Document,
	component: Outlet,
	errorComponent: () => (
		<main className="p-8">
			<h1>Halaman tidak tersedia</h1>
			<p>Muat ulang halaman atau kembali ke beranda.</p>
			<Link to="/">Beranda</Link>
		</main>
	),
	notFoundComponent: () => (
		<main className="p-8">
			<h1>Halaman tidak ditemukan</h1>
			<Link to="/">Beranda</Link>
		</main>
	),
});
function Document({ children }: { children: ReactNode }) {
	return (
		<html lang="id" suppressHydrationWarning>
			<head>
				<HeadContent />
			</head>
			<body className="bg-background text-foreground">
				<BOProvider>{children}</BOProvider>
				<Scripts />
			</body>
		</html>
	);
}
