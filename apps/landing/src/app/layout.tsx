import { brand } from "@repo/brand";
import type { Metadata } from "next";
import type { ReactNode } from "react";
import { siteUrl } from "../lib/seo";
import "./globals.css";

export const metadata: Metadata = {
	metadataBase: new URL(siteUrl),
	applicationName: brand.name,
	icons: { icon: "/favicon.svg", apple: "/apple-touch-icon.png" },
};

export default function RootLayout({ children }: { children: ReactNode }) {
	return (
		<html lang="id">
			<body>{children}</body>
		</html>
	);
}
