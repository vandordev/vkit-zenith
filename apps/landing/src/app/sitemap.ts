import { join } from "node:path";
import type { MetadataRoute } from "next";
import { siteUrl } from "../lib/seo";
import { sitemapPaths } from "../lib/sitemap";

export const dynamic = "force-static";

export default function sitemap(): MetadataRoute.Sitemap {
	return sitemapPaths(join(process.cwd(), "src/app")).map((path) => ({ url: new URL(path, siteUrl).href }));
}
