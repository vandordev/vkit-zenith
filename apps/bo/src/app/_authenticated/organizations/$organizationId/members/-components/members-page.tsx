import type { ReactNode } from "react";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { ContentState } from "../../../../../-components/content-state";

/** Header frame for the real membership list; the list and its drawers stay unchanged. */
export function MembersPage({
	error,
	backLink,
	onRetry,
	children,
}: {
	error?: string;
	backLink?: ReactNode;
	onRetry: () => void;
	children: ReactNode;
}) {
	return (
		<PageContainer className="flex min-w-0 flex-col gap-6">
			{backLink}
			<PageHeader title="Anggota" description="Keanggotaan pengguna pada organisasi ini, beserta peran dan statusnya." />
			{error && <ContentState kind="error" title={error} onRetry={onRetry} />}
			{children}
		</PageContainer>
	);
}
