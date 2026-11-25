import { brand } from "@repo/brand";
import { DisplayHeading } from "@repo/ui/display-heading";
import { companyName, contactHref } from "../../_components/landing-content";

const updated = "18 November 2026";

export function PrivacyPage() {
	return (
		<div className="lp">
			<main className="lp-wrap lp-doc">
				<a className="lp-doc-back" href="/">
					Kembali ke {brand.name}
				</a>
				<DisplayHeading className="lp-display">Kebijakan privasi</DisplayHeading>
				<p className="lp-doc-intro">
					Halaman ini menjelaskan data apa yang diproses saat Anda mengunjungi situs {brand.name}. Situs ini dikelola oleh {companyName}.
				</p>
				<section>
					<h2>Data yang kami kumpulkan lewat situs ini</h2>
					<p>
						Situs ini tidak memakai cookie, analytics, iklan, maupun formulir. Kami tidak meminta atau menyimpan data pribadi melalui situs
						ini.
					</p>
				</section>
				<section>
					<h2>Log hosting</h2>
					<p>
						Situs disajikan oleh penyedia hosting kami, Vercel. Seperti layanan web pada umumnya, penyedia hosting dapat mencatat data
						teknis permintaan, seperti alamat IP dan jenis peramban, untuk keamanan dan keandalan layanan.
					</p>
				</section>
				<section>
					<h2>Saat Anda menghubungi kami</h2>
					<p>Pesan yang Anda kirim lewat email hanya kami gunakan untuk menanggapi pertanyaan Anda.</p>
				</section>
				<section>
					<h2>Pertanyaan dan permintaan penghapusan</h2>
					<p>
						Untuk pertanyaan tentang privasi atau meminta kami menghapus percakapan Anda, hubungi kami di{" "}
						<a href={contactHref}>hello@example.com</a>.
					</p>
				</section>
				<p className="lp-doc-note">Kebijakan ini dapat diperbarui. Terakhir diperbarui {updated}.</p>
			</main>
		</div>
	);
}
