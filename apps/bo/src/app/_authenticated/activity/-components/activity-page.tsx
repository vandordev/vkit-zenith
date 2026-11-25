import type { ComponentProps, ReactNode, RefObject } from "react";
import { Page } from "@bo/app/-components/common";
import { ContentState } from "@bo/app/-components/content-state";
import { ActivityCollection } from "./activity-collection";

export function ActivityPage({
	denied,
	filters,
	children,
	focusRef,
	...collection
}: ComponentProps<typeof ActivityCollection> & {
	denied?: boolean;
	filters?: ReactNode;
	children?: ReactNode;
	focusRef?: RefObject<HTMLDivElement | null>;
}) {
	return (
		<Page title="Log Aktivitas" description="Riwayat staf dan operasi pelanggan yang telah tercatat, bukan pemantauan keamanan lengkap.">
			{denied ? (
				<ContentState
					kind="error"
					title="Akses hanya untuk Superadmin"
					description="Tidak ada data aktivitas yang dimuat untuk peran ini."
				/>
			) : (
				<div ref={focusRef} tabIndex={-1} className="min-w-0 space-y-6 outline-none">
					{filters}
					<ActivityCollection {...collection} />
					{children}
				</div>
			)}
		</Page>
	);
}
