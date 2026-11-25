import type { ReactNode } from "react";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { ContentState } from "./content-state";

export type SectionState = "ready" | "pending" | "error" | "empty";

/** Header + loading/error/empty composition shared by every organization context page. */
export function OrgSection({
	title,
	description,
	status,
	actions,
	breadcrumb,
	state,
	error,
	emptyTitle,
	emptyDescription,
	onRetry,
	pager,
	children,
}: {
	title: string;
	description?: ReactNode;
	status?: ReactNode;
	actions?: ReactNode;
	breadcrumb?: ReactNode;
	state: SectionState;
	error?: string;
	emptyTitle?: string;
	emptyDescription?: string;
	onRetry?: () => void;
	/** Page controls under the content; shown whenever the list is ready, even on an empty later page. */
	pager?: ReactNode;
	children?: ReactNode;
}) {
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			<PageHeader title={title} description={description} status={status} actions={actions} breadcrumb={breadcrumb} />
			{state === "pending" && <ContentState kind="loading" title={`Memuat ${title.toLowerCase()}…`} />}
			{state === "error" && <ContentState kind="error" title={error ?? "Data tidak dapat dimuat."} onRetry={onRetry} />}
			{state === "empty" && emptyTitle && <ContentState kind="empty" title={emptyTitle} description={emptyDescription} />}
			{state === "ready" && children}
			{state === "ready" && pager}
		</PageContainer>
	);
}

export const dash = "—";

/** Empty means the whole list is empty; an out-of-range later page keeps its pager. */
export const listState = (state: SectionState, count: number, total?: string): SectionState =>
	state === "ready" && (total === undefined ? count === 0 : total === "0") ? "empty" : state;
