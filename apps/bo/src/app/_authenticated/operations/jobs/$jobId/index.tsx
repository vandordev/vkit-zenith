import { createFileRoute } from "@tanstack/react-router";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { usePrivateQuery } from "@bo/lib/api/session";
import { collectionState, useOpsAction } from "../../../-adapters/global-list";
import { JobDetail } from "../-components/job-detail";

export const Route = createFileRoute("/_authenticated/operations/jobs/$jobId/")({ component: Job });
function Job() {
	const { jobId: id } = Route.useParams();
	const act = useOpsAction();
	const detail = usePrivateQuery(["ops-jobs", "detail", id], async (t, signal) => (await sdk.boV1GetJob({ client: t.client, signal, throwOnError: true, path: { id } })).data);
	const { state, error, onRetry } = collectionState(detail);
	return (
		<JobDetail
			item={detail.data?.item}
			history={detail.data?.history ?? []}
			state={detail.data ? "ready" : state === "ready" ? "empty" : state}
			error={error}
			onRetry={onRetry}
			onRetryJob={async (reason) => (await act("ops-jobs", (t) => sdk.boV1RetryJob({ client: t.client, throwOnError: true, path: { id }, body: { reason } }))).data.result}
			onCancel={async (reason) => (await act("ops-jobs", (t) => sdk.boV1CancelJob({ client: t.client, throwOnError: true, path: { id }, body: { reason } }))).data.result}
		/>
	);
}
