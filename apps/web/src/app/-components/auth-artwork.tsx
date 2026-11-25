import { useEffect, useState } from "react";
import { GrainGradient } from "@paper-design/shaders-react";

export function AuthArtwork() {
	const [palette, setPalette] = useState<{ light: string; primary: string } | null>(null);
	const [desktop, setDesktop] = useState(false);
	const [reducedMotion, setReducedMotion] = useState(true);

	useEffect(() => {
		const wideScreen = window.matchMedia("(min-width: 64rem)");
		const motionPreference = window.matchMedia("(prefers-reduced-motion: reduce)");
		const updateScreen = () => setDesktop(wideScreen.matches);
		const updateMotion = () => setReducedMotion(motionPreference.matches);
		updateScreen();
		updateMotion();
		wideScreen.addEventListener("change", updateScreen);
		motionPreference.addEventListener("change", updateMotion);

		// Paper accepts RGB/hex only; resolve the brand's CSS colors through canvas.
		const context = document.createElement("canvas").getContext("2d");
		const probe = document.createElement("canvas").getContext("webgl2");
		if (context && probe) {
			const tokens = getComputedStyle(document.documentElement);
			const readColor = (token: string) => {
				context.fillStyle = tokens.getPropertyValue(token).trim();
				context.fillRect(0, 0, 1, 1);
				const [red, green, blue] = context.getImageData(0, 0, 1, 1).data;
				return `rgb(${red}, ${green}, ${blue})`;
			};
			setPalette({ light: readColor("--background"), primary: readColor("--primary") });
		}
		probe?.getExtension("WEBGL_lose_context")?.loseContext();
		return () => {
			wideScreen.removeEventListener("change", updateScreen);
			motionPreference.removeEventListener("change", updateMotion);
		};
	}, []);

	if (!desktop || !palette) return null;
	return (
		<GrainGradient
			aria-hidden="true"
			className="pointer-events-none absolute inset-0"
			speed={reducedMotion ? 0 : 1}
			frame={2854.5}
			scale={1}
			rotation={0}
			offsetX={0}
			offsetY={0}
			softness={0.5}
			intensity={0.5}
			noise={0.25}
			shape="corners"
			colors={[palette.light, palette.primary, palette.primary, palette.light]}
			colorBack="rgba(0, 0, 0, 0)"
		/>
	);
}
