import { ZenithLoader } from "@repo/ui/zenith-loader";

export function GlobalPending() {
	return (
		<main className="bg-background text-foreground">
			<ZenithLoader mode="screen" label="Loading" />
		</main>
	);
}
