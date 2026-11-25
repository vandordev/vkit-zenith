import { Button } from "@repo/ui/button";
import type { V1OrganizationListPagination } from "@bo/lib/api/generated/types.gen";

/** Server-side page controls; totals are exact decimal strings, so the maths is BigInt. */
export function Pager({
	pagination,
	noun,
	loading,
	onPage,
}: {
	pagination: V1OrganizationListPagination;
	noun: string;
	loading?: boolean;
	onPage: (page: number) => void;
}) {
	const total = BigInt(pagination.total);
	const size = BigInt(pagination.pageSize);
	const pages = total === 0n ? 1n : (total + size - 1n) / size;
	if (pages === 1n && pagination.page === 1) return null;
	return (
		<nav aria-label="Halaman" className="flex flex-wrap items-center justify-between gap-3 text-sm">
			<p role="status" className="text-muted-foreground">
				Halaman {pagination.page} dari {String(pages)} · {pagination.total} {noun}
			</p>
			<div className="flex gap-2">
				<Button
					type="button"
					variant="outline"
					disabled={loading || pagination.page <= 1}
					onClick={() => onPage(Math.max(1, Math.min(pagination.page - 1, Number(pages))))}
				>
					Sebelumnya
				</Button>
				<Button
					type="button"
					variant="outline"
					disabled={loading || BigInt(pagination.page) >= pages}
					onClick={() => onPage(pagination.page + 1)}
				>
					Berikutnya
				</Button>
			</div>
		</nav>
	);
}
