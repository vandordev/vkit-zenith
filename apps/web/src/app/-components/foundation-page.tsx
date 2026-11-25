import { brand } from "@repo/brand";

export function FoundationPage() {
	return (
		<main className="mx-auto flex min-h-[28rem] w-full max-w-3xl flex-col justify-center px-6 py-20 sm:px-10">
			<h1 className="text-4xl font-semibold tracking-tight">{brand.name}</h1>
			<p className="mt-5 max-w-xl text-lg leading-relaxed text-muted-foreground">
				A foundation for what comes next.
			</p>
			<p className="mt-3 max-w-xl leading-relaxed text-muted-foreground">
				A TypeScript web application and independent Go APIs.
			</p>
		</main>
	);
}
