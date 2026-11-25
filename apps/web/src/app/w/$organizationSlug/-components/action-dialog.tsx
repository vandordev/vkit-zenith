import { useEffect, useLayoutEffect, useRef, useState, type ReactNode, type Ref } from "react";
import { Button } from "@repo/ui/button";
import { Dialog, DialogBody, DialogFooter } from "@repo/ui/dialog";
import { TextArea } from "@repo/ui/textarea";

export const reasonMin = 10;
export const reasonMax = 1024;
export const reasonProblem = (reason: string) =>
	reason.trim().length < reasonMin
		? `Alasan minimal ${reasonMin} karakter.`
		: reason.length > reasonMax
			? `Alasan maksimal ${reasonMax} karakter.`
			: undefined;

export type ActionPhase = "idle" | "submitting" | "failed" | "success";
type Shared = {
	title: string;
	description: string;
	confirmLabel: string;
	tone?: "default" | "destructive";
	/** Reason (10–1024) is required for actions that need an audit justification. */
	reasonRequired?: boolean;
	fields?: ReactNode;
	/** The action cannot succeed as things stand (e.g. nothing to pick); the dialog says why in `fields`. */
	confirmDisabled?: boolean;
};

/** Props-only dialog: every state is driven from outside so stories can pin each one. */
export function ActionDialogView({
	open,
	onOpenChange,
	title,
	description,
	confirmLabel,
	tone = "default",
	reasonRequired,
	fields,
	confirmDisabled,
	reason,
	reasonRef,
	onReasonChange,
	showValidation,
	phase,
	error,
	result,
	onConfirm,
}: Shared & {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	reason: string;
	/** Receives focus when an invalid submit is attempted. */
	reasonRef?: Ref<HTMLTextAreaElement>;
	onReasonChange: (reason: string) => void;
	showValidation: boolean;
	phase: ActionPhase;
	error?: string;
	/** The dummy result, shown as "Simulasi (data contoh)". */
	result?: ReactNode;
	onConfirm: () => void;
}) {
	const problem = reasonRequired && showValidation ? reasonProblem(reason) : undefined;
	const busy = phase === "submitting";
	const done = phase === "success";
	const footerRef = useRef<HTMLDivElement>(null);
	// The busy submit button disables, so focus drops to body: success lands on Tutup, failure on submit again.
	useEffect(() => {
		if (phase !== "success" && phase !== "failed") return;
		const frame = requestAnimationFrame(() => {
			const footer = footerRef.current;
			if (!footer?.isConnected) return;
			const target = phase === "success" ? footer.querySelector("button") : footer.querySelector<HTMLButtonElement>("button[type=submit]");
			const active = document.activeElement;
			if (phase === "success" || !active || active === document.body) target?.focus();
		});
		return () => cancelAnimationFrame(frame);
	}, [phase]);
	return (
		<Dialog
			open={open}
			title={title}
			description={description}
			onOpenChange={(next, details) => {
				if (!next && busy) details.cancel();
				else onOpenChange(next);
			}}
			contentProps={{ closeButtonLabel: "Tutup dialog" }}
		>
			<form
				className="flex min-h-0 flex-1 flex-col"
				noValidate
				onSubmit={(event) => {
					event.preventDefault();
					if (!done) onConfirm();
				}}
			>
				<DialogBody className="flex flex-col gap-4">
					{done ? (
						<div className="flex flex-col gap-3" role="status">
							{result}
							<p className="text-xs text-muted-foreground">Simulasi (data contoh): belum tersimpan.</p>
						</div>
					) : (
						<>
							{fields}
							{reasonRequired && (
								<div className="flex flex-col gap-1.5">
									<TextArea
										ref={reasonRef}
										label="Alasan"
										value={reason}
										onChange={(event) => onReasonChange(event.target.value)}
										maxLength={reasonMax}
										rows={3}
										disabled={busy}
										required
										aria-invalid={problem ? true : undefined}
										aria-describedby="action-reason-help"
									/>
									<div id="action-reason-help" className="flex items-start justify-between gap-3 text-xs">
										<span role={problem ? "alert" : undefined} className={problem ? "text-destructive-text" : "text-muted-foreground"}>
											{problem ?? "Wajib diisi dan tercatat bersama tindakan ini."}
										</span>
										<span className="shrink-0 tabular-nums text-muted-foreground">
											{reason.length}/{reasonMax}
										</span>
									</div>
								</div>
							)}
							{phase === "failed" && error && (
								<p role="alert" className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive-text">
									{error}
								</p>
							)}
						</>
					)}
				</DialogBody>
				<DialogFooter ref={footerRef}>
					{done ? (
						<Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
							Tutup
						</Button>
					) : (
						<>
							<Button type="button" variant="outline" disabled={busy} onClick={() => onOpenChange(false)}>
								Batal
							</Button>
							<Button
								type="submit"
								variant={tone === "destructive" ? "destructive" : "default"}
								isLoading={busy}
								disabled={confirmDisabled}
							>
								{confirmLabel}
							</Button>
						</>
					)}
				</DialogFooter>
			</form>
		</Dialog>
	);
}

/**
 * Owns the submit lifecycle: validation, double-submit guard, inline failure and result.
 * `onSubmit` resolves with the dummy result (handed to the page via `onResult`) or rejects; a rejection is shown only
 * through `describeError` (fixed copy), never the raw rejection text. A completion that arrives after the dialog
 * closed, reopened or unmounted (identity change remounts the shell) is dropped: no state, no `onResult`.
 * The page keeps showing the result as "Simulasi (data contoh)" until navigation.
 */
export function ActionDialog<TResult>({
	open,
	onOpenChange,
	onSubmit,
	onResult,
	describeError,
	renderResult,
	...shared
}: Shared & {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSubmit: (reason: string) => Promise<TResult>;
	onResult?: (result: TResult) => void;
	/** Maps a rejection to fixed, safe copy (see orgFailureMessage). */
	describeError: (failure: unknown) => string;
	renderResult: (result: TResult) => ReactNode;
}) {
	const [reason, setReason] = useState("");
	const [showValidation, setShowValidation] = useState(false);
	const [phase, setPhase] = useState<ActionPhase>("idle");
	const [error, setError] = useState<string>();
	const [result, setResult] = useState<{ value: TResult }>();
	const inFlight = useRef(false);
	const reasonRef = useRef<HTMLTextAreaElement>(null);
	// Bumped synchronously on every close request, close commit and unmount: a submission only completes in the
	// cycle it started in. Passive-effect cleanup runs after the commit, so a promise settling in between would slip
	// through; layout effects (and the close request itself) run before any later microtask can.
	const cycle = useRef(0);
	useLayoutEffect(
		() => () => {
			cycle.current += 1;
		},
		[],
	);
	useLayoutEffect(() => {
		if (open) return;
		cycle.current += 1;
		inFlight.current = false;
		setReason("");
		setShowValidation(false);
		setPhase("idle");
		setError(undefined);
		setResult(undefined);
	}, [open]);
	const requestOpenChange = (next: boolean) => {
		if (!next) cycle.current += 1;
		onOpenChange(next);
	};
	function confirm() {
		if (inFlight.current) return;
		setShowValidation(true);
		if (shared.reasonRequired && reasonProblem(reason)) {
			requestAnimationFrame(() => reasonRef.current?.focus());
			return;
		}
		inFlight.current = true;
		const started = cycle.current;
		setPhase("submitting");
		setError(undefined);
		onSubmit(reason.trim()).then(
			(value) => {
				if (cycle.current !== started) return;
				inFlight.current = false;
				setResult({ value });
				setPhase("success");
				onResult?.(value);
			},
			(failure: unknown) => {
				if (cycle.current !== started) return;
				inFlight.current = false;
				setError(describeError(failure));
				setPhase("failed");
			},
		);
	}
	return (
		<ActionDialogView
			{...shared}
			open={open}
			onOpenChange={requestOpenChange}
			reason={reason}
			reasonRef={reasonRef}
			onReasonChange={setReason}
			showValidation={showValidation}
			phase={phase}
			error={error}
			result={result && renderResult(result.value)}
			onConfirm={confirm}
		/>
	);
}
