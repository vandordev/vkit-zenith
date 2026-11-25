import { Link } from "@tanstack/react-router";
import { buttonVariants } from "@repo/ui/button";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyTitle } from "@repo/ui/empty";

/** One state for an unknown slug, a removed membership or a non-member: the three are indistinguishable by design. */
export function OrganizationNotFound() {
	return (
		<main className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
			<Empty>
				<EmptyHeader>
					<EmptyTitle>Organisasi tidak ditemukan</EmptyTitle>
					<EmptyDescription>Organisasi ini tidak ada atau Anda bukan anggotanya.</EmptyDescription>
				</EmptyHeader>
				<EmptyContent>
					<Link to="/w" className={buttonVariants({ variant: "outline" })}>
						Kembali ke dashboard
					</Link>
				</EmptyContent>
			</Empty>
		</main>
	);
}
