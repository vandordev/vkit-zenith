import { brand } from "@repo/brand";
import { LandingPage } from "./_components/landing-page";
import { pageMetadata, siteDescription, siteTitle, siteUrl } from "../lib/seo";

export const metadata = pageMetadata(siteTitle, siteDescription, "/");

const structuredData = {
	"@context": "https://schema.org",
	"@graph": [
		{
			"@type": "Organization",
			"@id": `${siteUrl}/#organization`,
			name: brand.name,
			url: siteUrl,
			logo: `${siteUrl}/images/logo.png`,
		},
		{
			"@type": "WebSite",
			"@id": `${siteUrl}/#website`,
			name: brand.name,
			url: siteUrl,
			inLanguage: "id-ID",
			publisher: { "@id": `${siteUrl}/#organization` },
		},
	],
};

export default function Page() {
	return (
		<>
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData).replace(/</g, "\\u003c") }} />
			<LandingPage />
		</>
	);
}
