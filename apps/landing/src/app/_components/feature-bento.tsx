import { DisplayHeading } from "@repo/ui/display-heading";

const features = [
	{ title: "Akun dan sesi.", body: "Registrasi, verifikasi email, login, dan reset kata sandi.", tone: "tone tone-sky wide" },
	{ title: "Organisasi.", body: "Anggota dengan peran owner, admin, dan member.", tone: "" },
	{ title: "Backoffice.", body: "Staf terpisah dengan audit untuk setiap perubahan.", tone: "tone tone-lilac" },
	{ title: "API untuk developer.", body: "API key per organisasi dan dokumentasi OpenAPI.", tone: "tone tone-lime wide" },
] as const;

// Placeholder features: a product built on Zenith lists its own.
export function FeatureBento() {
	return (
		<section id="fitur" className="lp-section">
			<div className="lp-wrap">
				<div className="lp-head">
					<div>
						<DisplayHeading as="h2" className="lp-display">
							Fondasi yang sudah jadi. <span className="lp-sub">Fokus ke produk Anda.</span>
						</DisplayHeading>
					</div>
					<p className="lp-body">Ganti bagian ini dengan fitur utama produk Anda.</p>
				</div>
				<div className="lp-bento">
					{features.map((feature) => (
						<div key={feature.title} className={`lp-tile ${feature.tone}`.trim()}>
							<p className="cap">
								<b>{feature.title}</b> <span>{feature.body}</span>
							</p>
						</div>
					))}
				</div>
			</div>
		</section>
	);
}
