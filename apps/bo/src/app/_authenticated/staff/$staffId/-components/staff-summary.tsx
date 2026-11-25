import type { ComponentProps, ReactNode } from "react";
import { Badge } from "@repo/ui/badge";
import { DateTime } from "@repo/ui/date-time";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { labels } from "../../../../-components/common";
import { ContentState } from "../../../../-components/content-state";
import { StaffGovernance } from "../../-components/staff-governance";
import type { BoV1GetStaffResponse } from "@bo/lib/api/generated/types.gen";

/** Ringkasan of one staff account: the former drawer content (real governance form) as a page. */
export function StaffSummary({
	staff,
	error,
	loading,
	saves = 0,
	backLink,
	onRetry,
	...governance
}: {
	staff?: BoV1GetStaffResponse;
	error?: string;
	loading?: boolean;
	/** Bumped after a successful save so the form restarts from the fresh revision. */
	saves?: number;
	backLink: ReactNode;
	onRetry: () => void;
} & Pick<ComponentProps<typeof StaffGovernance>, "pending" | "feedback" | "onSubmit" | "onDirty" | "onReload">) {
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			{backLink}
			{!staff ? (
				error ? (
					<ContentState kind="error" title={error} onRetry={onRetry} />
				) : (
					<ContentState kind="loading" title="Memuat staf…" />
				)
			) : (
				<>
					<PageHeader
						title={staff.displayName}
						description={<span className="break-all">{staff.email}</span>}
						status={
							<span className="flex flex-wrap items-center gap-2">
								<Badge variant="secondary">{labels[staff.role] ?? staff.role}</Badge>
								<Badge variant="secondary">{labels[staff.status] ?? staff.status}</Badge>
							</span>
						}
					/>
					<p className="text-xs text-muted-foreground">
						Diperbarui <DateTime value={staff.updatedAt} />. Perubahan peran, status dan kata sandi mengakhiri sesi staf terkait.
					</p>
					<div className="w-full min-w-0 max-w-xl">
						<StaffGovernance
							key={`${staff.id}:${staff.revision}:${saves}`}
							staff={staff}
							loadError={error}
							refreshing={loading}
							onRetry={onRetry}
							{...governance}
						/>
					</div>
				</>
			)}
		</PageContainer>
	);
}
