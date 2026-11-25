import { ClosingCta } from "./closing-cta";
import { FeatureBento } from "./feature-bento";
import { LandingFooter } from "./landing-footer";
import { LandingHero } from "./landing-hero";
import { LandingNav } from "./landing-nav";

export function LandingPage() {
	return (
		<div className="lp">
			<LandingNav />
			<main>
				<LandingHero />
				<FeatureBento />
				<ClosingCta />
			</main>
			<LandingFooter />
		</div>
	);
}
