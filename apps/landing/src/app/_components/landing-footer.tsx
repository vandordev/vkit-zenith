import { ZenithMark, ZenithWordmark } from "@repo/ui/zenith-mark";
import { companyName, contactHref } from "./landing-content";

export function LandingFooter() {
	return (
		<footer className="lp-wrap lp-foot">
			<div className="lp-foot-grid">
				<div className="lp-foot-about">
					<a className="lp-brand" href="#">
						<ZenithMark size={24} />
						<ZenithWordmark />
					</a>
					<p>Fondasi SaaS dari {companyName}.</p>
				</div>
				<nav className="lp-foot-nav" aria-label="Jelajahi Zenith">
					<h3>Jelajahi</h3>
					<ul>
						<li>
							<a href="#fitur">Fitur</a>
						</li>
						<li>
							<a href="/privacy">Kebijakan privasi</a>
						</li>
					</ul>
				</nav>
				<div className="lp-foot-contact">
					<h3>Kontak</h3>
					<a href={contactHref}>hello@example.com</a>
				</div>
			</div>
			<div className="lp-foot-bottom">
				<span>© 2026 {companyName}</span>
			</div>
		</footer>
	);
}
