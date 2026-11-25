import { brand } from "@repo/brand";
import { pageMetadata } from "../../lib/seo";
import { PrivacyPage } from "./_components/privacy-page";

const title = `Kebijakan privasi | ${brand.name}`;
const description = `Data apa yang diproses saat Anda mengunjungi situs ${brand.name} atau menghubungi kami.`;

export const metadata = pageMetadata(title, description, "/privacy");

export default function Page() {
	return <PrivacyPage />;
}
