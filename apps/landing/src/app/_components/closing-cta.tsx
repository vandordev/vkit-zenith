import { zenithMark } from "@repo/brand/logo-mark";
import { Button } from "@repo/ui/button";
import { DisplayHeading } from "@repo/ui/display-heading";
import { contactHref } from "./landing-content";

export function ClosingCta() {
	return (
		<div className="lp-wrap">
			<section className="lp-close tone tone-aurora">
				{/* The brand mark draws its outline as the section scrolls in (CSS), then fills. */}
				<svg className="lp-mark" viewBox={zenithMark.viewBox} width={56} height={56} aria-hidden="true">
					{zenithMark.paths.map((d) => (
						<path key={d} d={d} pathLength={1} />
					))}
				</svg>
				<DisplayHeading as="h2" className="lp-display">
					Siap memulai?
				</DisplayHeading>
				<p>Ganti ajakan ini dengan langkah pertama untuk pengguna produk Anda.</p>
				<Button asChild size="lg">
					<a href={contactHref}>Hubungi kami</a>
				</Button>
			</section>
		</div>
	);
}
