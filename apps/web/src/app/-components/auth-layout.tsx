import type { ReactNode } from "react";
import { brand } from "@repo/brand";
import { DisplayHeading } from "@repo/ui/display-heading";
import { ZenithMark, ZenithWordmark } from "@repo/ui/zenith-mark";
import { AuthArtwork } from "./auth-artwork";

export type AuthLayoutProps = { title: string; description?: string; children: ReactNode; footer?: ReactNode };

export function AuthLayout({ title, description, children, footer }: AuthLayoutProps) {
	return (
		<main
			lang="id"
			className="min-h-dvh bg-background text-foreground selection:bg-primary selection:text-primary-foreground [--ring:var(--primary)] [&_:focus-visible:not([data-slot=input])]:outline-2 [&_:focus-visible:not([data-slot=input])]:outline-solid [&_:focus-visible:not([data-slot=input])]:outline-offset-2 [&_:focus-visible:not([data-slot=input])]:outline-primary lg:bg-muted lg:p-3"
		>
			<div className="grid min-h-dvh lg:min-h-[calc(100dvh-1.5rem)] lg:grid-cols-2 lg:gap-3">
				<section className="flex min-w-0 flex-col bg-background px-6 py-8 sm:px-12 lg:rounded-lg lg:px-12 xl:px-20">
					<div className="flex items-center gap-2.5 text-xl">
						<ZenithMark size={32} className="size-8 shrink-0" />
						<ZenithWordmark />
					</div>
					<div className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-8 py-12 sm:py-16">
						<div className="flex flex-col gap-3">
							<DisplayHeading>{title}</DisplayHeading>
							{description && <p className="[overflow-wrap:anywhere] text-sm leading-relaxed text-muted-foreground">{description}</p>}
						</div>
						{children}
						{footer && (
							<div className="text-sm leading-relaxed text-muted-foreground [&_a]:font-medium [&_a]:text-foreground [&_a]:decoration-border [&_a]:underline-offset-4 [&_a:hover]:text-primary [&_a:hover]:underline [&_a:focus-visible]:outline-primary">
								{footer}
							</div>
						)}
					</div>
				</section>
				<aside className="relative hidden min-h-[40rem] min-w-0 flex-col justify-between overflow-hidden rounded-lg bg-foreground p-12 text-background lg:flex xl:p-16">
					<AuthArtwork />
					<div
						aria-hidden="true"
						className="pointer-events-none absolute inset-0 bg-linear-to-b from-foreground/85 via-transparent to-foreground/85"
					/>
					<p className="relative z-10 max-w-sm text-balance text-4xl font-medium leading-tight tracking-tight xl:text-5xl">
						Satu akun.
						<br />
						Terhubung ke {brand.name}.
					</p>
					<p className="relative z-10 mt-auto pt-80 text-sm leading-relaxed">Akun Anda untuk mengakses aplikasi Zenith yang terhubung.</p>
				</aside>
			</div>
		</main>
	);
}
