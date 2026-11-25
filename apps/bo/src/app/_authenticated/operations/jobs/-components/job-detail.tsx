import type { Job } from "@bo/lib/api/generated/types.gen";
import { DateTime } from "@repo/ui/date-time";
import { OpsActions } from "../../-components/ops-actions";
import { JsonSection, OpsDetailFrame, statusMessage, useOpsRecord, type OpsDetailProps } from "../../-components/ops-detail";

export type JobDetailProps = OpsDetailProps<Job> & {
	onRetryJob: (reason: string) => Promise<Job>;
	onCancel: (reason: string) => Promise<Job>;
};

export function JobDetail({ item, onRetryJob, onCancel, ...frame }: JobDetailProps) {
	const { value, setOverlay, state, history } = useOpsRecord(item, frame.state, frame.history);
	const run = (call: (reason: string) => Promise<Job>) => async (reason: string) => {
		const next = await call(reason);
		setOverlay(next);
		return statusMessage("job", next.state);
	};
	return (
		<OpsDetailFrame
			{...frame}
			state={state}
			history={history}
			title="Job"
			section="jobs"
			sectionLabel="Job"
			id={value?.id ?? ""}
			statusKind="job"
			status={value?.state}
			actions={
				value && (
					<OpsActions
						ariaLabel="Tindakan job"
						actions={[
							{
								key: "retry",
								label: "Ulangi job",
								title: "Ulangi job",
								description: "Job akan dikembalikan ke antrean untuk dijalankan lagi.",
								visible: value.state === "retryable" || value.state === "discarded",
								onSubmit: run(onRetryJob),
							},
							{
								key: "cancel",
								label: "Batalkan job",
								title: "Batalkan job",
								description: "Job tidak akan dijalankan lagi.",
								tone: "destructive",
								visible: value.state === "available" || value.state === "scheduled" || value.state === "retryable",
								onSubmit: run(onCancel),
							},
						]}
					/>
				)
			}
			facts={[
				{ label: "Jenis", value: value && <span className="font-mono text-xs">{value.kind}</span> },
				{ label: "Antrean", value: value?.queue },
				{ label: "Percobaan", value: value && `${value.attempt}/${value.maxAttempts}` },
				{ label: "Dibuat", value: value && <DateTime value={value.createdAt} /> },
			]}
			related={
				value && (
					<>
						<section className="flex min-w-0 flex-col gap-3" aria-label="Kesalahan">
							<h2 className="font-semibold">Kesalahan</h2>
							{value.errors.length ? (
								<ol className="flex flex-col gap-2">
									{value.errors.map((message, index) => (
										<li key={`${index}:${message}`} className="rounded-md border border-border p-3 font-mono text-xs wrap-anywhere">
											{message}
										</li>
									))}
								</ol>
							) : (
								<p className="text-sm text-muted-foreground">Tidak ada kesalahan tercatat.</p>
							)}
						</section>
						<JsonSection heading="Argumen" label="Argumen job" value={value.args} />
					</>
				)
			}
		/>
	);
}
