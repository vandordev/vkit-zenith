import { appConfig } from "./config";

export type MetadataInput = {
	title: string;
	description: string;
	/** Absolute-path of this page ("/login"); becomes the canonical and og:url under the approved origin. */
	pathname?: string;
	image?: string;
	/** Every page states its policy: only the public auth pages pass `true`. */
	index: boolean;
};

const absolute = (pathname: string) => new URL(pathname, appConfig.siteOrigin).toString();

export function createMetadata({ title, description, pathname, image, index }: MetadataInput) {
	const fullTitle = title === appConfig.appName ? title : `${title} | ${appConfig.appName}`;
	const url = pathname ? absolute(pathname) : undefined;
	const imageUrl = image ? absolute(image) : undefined;

	return {
		meta: [
			{ title: fullTitle },
			{ name: "description", content: description },
			{ name: "robots", content: index ? "index, follow" : "noindex, nofollow" },
			{ property: "og:type", content: "website" },
			{ property: "og:site_name", content: appConfig.appName },
			{ property: "og:locale", content: appConfig.locale },
			{ property: "og:title", content: fullTitle },
			{ property: "og:description", content: description },
			...(url ? [{ property: "og:url", content: url }] : []),
			...(imageUrl ? [{ property: "og:image", content: imageUrl }] : []),
			{ name: "twitter:card", content: imageUrl ? "summary_large_image" : "summary" },
			{ name: "twitter:title", content: fullTitle },
			{ name: "twitter:description", content: description },
			...(imageUrl ? [{ name: "twitter:image", content: imageUrl }] : []),
		],
		links: url ? [{ rel: "canonical", href: url }] : [],
	};
}
