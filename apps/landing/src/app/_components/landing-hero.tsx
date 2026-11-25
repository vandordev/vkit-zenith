import { Button } from "@repo/ui/button";
import { DisplayHeading } from "@repo/ui/display-heading";
import { contactHref } from "./landing-content";

// Placeholder copy: a product built on Zenith replaces it with its own story.
export function LandingHero() {
	return (
		<div className="lp-hero">
			<div className="lp-wrap">
				<DisplayHeading className="lp-display">
					Produk Anda, di atas fondasi yang rapi. <span className="lp-second">Ganti teks ini dengan cerita produk Anda.</span>
				</DisplayHeading>
				<p className="lp-lede">
					Zenith menyiapkan akun, organisasi, anggota, backoffice, dan API key, sehingga tim Anda langsung mengerjakan fitur yang membedakan
					produk.
				</p>
				<div className="lp-acts">
					<Button asChild size="lg">
						<a href={contactHref}>Hubungi kami</a>
					</Button>
					<Button asChild size="lg" variant="secondary">
						<a href="#fitur">Lihat fitur</a>
					</Button>
				</div>
			</div>
		</div>
	);
}
