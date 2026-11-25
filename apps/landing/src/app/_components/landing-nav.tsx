import { Button } from "@repo/ui/button";
import { ZenithMark, ZenithWordmark } from "@repo/ui/zenith-mark";
import { contactHref } from "./landing-content";

export function LandingNav() {
	return (
		<header className="lp-navbar">
			<div className="lp-wrap lp-nav">
				<a className="lp-brand" href="#">
					<ZenithMark size={24} />
					<ZenithWordmark />
				</a>
				<nav aria-label="Utama">
					<ul>
						<li>
							<a href="#fitur">Fitur</a>
						</li>
					</ul>
				</nav>
				<Button asChild size="sm">
					<a href={contactHref}>Hubungi kami</a>
				</Button>
			</div>
		</header>
	);
}
