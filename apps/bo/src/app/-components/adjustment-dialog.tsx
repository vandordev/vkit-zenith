import { useEffect, useRef, useState, type ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { Dialog, DialogBody, DialogFooter } from "@repo/ui/dialog";
import { TextArea } from "@repo/ui/textarea";
import { errorMessage } from "@bo/lib/api/client";

export const reasonMin = 10;
export const reasonMax = 1024;
export const reasonProblem = (reason: string) =>
	reason.trim().length < reasonMin
		? `Alasan minimal ${reasonMin} karakter.`
		: reason.length > reasonMax
			? `Alasan maksimal ${reasonMax} karakter.`
			: undefined;

export type AdjustmentPhase = "idle" | "submitting" | "failed" | "success";
type Shared = {
	title: string;
	description: string;
	confirmLabel: string;
	tone: "default" | "destructive";
	fields?: ReactNode;
	/** Where focus goes on close; defaults to the opener, which actions may have removed. */
	finalFocus?: () => HTMLElement | null | undefined;
};

/** Props-only dialog. Every state is driven from outside so stories can pin each one. */
export function AdjustmentDialogView({
	open,
	onOpenChange,
	title,
	description,
	confirmLabel,
	tone,
	fields,
	finalFocus,
	reason,
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
	onReasonChange: (reason: string) => void;
	showValidation: boolean;
	phase: AdjustmentPhase;
	error?: string;
	result?: ReactNode;
	onConfirm: () => void;
}) {
	const problem = showValidation ? reasonProblem(reason) : undefined;
	const busy = phase === "submitting";
	const done = phase === "success";
	// The busy submit button disables, so focus drops to body. Success lands on Tutup. Failure restores
	// the submit button only if focus was lost (body or the submit itself): a control the user moved to
	// while the request ran (reason, extra fields) keeps focus. No footer means the dialog is closed.
	const footerRef = useRef<HTMLDivElement>(null);
	useEffect(() => {
		if (phase !== "success" && phase !== "failed") return;
		const target = () => {
			const footer = footerRef.current;
			if (!footer?.isConnected) return;
			if (phase === "success") return footer.querySelector("button")?.focus();
			const submit = footer.querySelector<HTMLButtonElement>("button[type=submit]");
			const active = document.activeElement;
			if (!active || active === document.body || active === submit) submit?.focus();
		};
		target();
		// Base UI re-evaluates focus after the swap; check again on the next frame.
		const frame = requestAnimationFrame(target);
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
			contentProps={{ closeButtonLabel: "Tutup dialog", finalFocus }}
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
							<p className="text-xs text-muted-foreground">Data dummy — belum tersimpan</p>
						</div>
					) : (
						<>
							{fields}
							<div className="flex flex-col gap-1.5">
								<TextArea
									label="Alasan"
									value={reason}
									onChange={(event) => onReasonChange(event.target.value)}
									maxLength={reasonMax}
									rows={3}
									disabled={busy}
									required
									aria-invalid={problem ? true : undefined}
									aria-describedby="adjustment-reason-help"
								/>
								<div id="adjustment-reason-help" className="flex items-start justify-between gap-3 text-xs">
									<span role={problem ? "alert" : undefined} className={problem ? "text-destructive-text" : "text-muted-foreground"}>
										{problem ?? "Wajib diisi dan tercatat bersama tindakan ini."}
									</span>
									<span className="shrink-0 tabular-nums text-muted-foreground">
										{reason.length}/{reasonMax}
									</span>
								</div>
							</div>
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
							<Button type="submit" variant={tone === "destructive" ? "destructive" : "default"} isLoading={busy}>
								{confirmLabel}
							</Button>
						</>
					)}
				</DialogFooter>
			</form>
		</Dialog>
	);
}

/** Owns the submit lifecycle: validation, busy/double-submit guard, inline failure and result. */
export function AdjustmentDialog<TResult>({
	open,
	onOpenChange,
	onSubmit,
	renderResult,
	...shared
}: Shared & {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Resolves with the would-be result (dummy); rejects with an error shown inline. */
	onSubmit: (reason: string) => Promise<TResult>;
	renderResult: (result: TResult) => ReactNode;
}) {
	const [reason, setReason] = useState("");
	const [showValidation, setShowValidation] = useState(false);
	const [phase, setPhase] = useState<AdjustmentPhase>("idle");
	const [error, setError] = useState<string>();
	const [result, setResult] = useState<{ value: TResult }>();
	const inFlight = useRef(false);
	useEffect(() => {
		if (open) return;
		inFlight.current = false;
		setReason("");
		setShowValidation(false);
		setPhase("idle");
		setError(undefined);
		setResult(undefined);
	}, [open]);
	function confirm() {
		if (inFlight.current) return;
		setShowValidation(true);
		if (reasonProblem(reason)) return;
		inFlight.current = true;
		setPhase("submitting");
		setError(undefined);
		onSubmit(reason.trim()).then(
			(value) => {
				inFlight.current = false;
				setResult({ value });
				setPhase("success");
			},
			(failure: unknown) => {
				inFlight.current = false;
				setError(errorMessage(failure));
				setPhase("failed");
			},
		);
	}
	return (
		<AdjustmentDialogView
			{...shared}
			open={open}
			onOpenChange={onOpenChange}
			reason={reason}
			onReasonChange={setReason}
			showValidation={showValidation}
			phase={phase}
			error={error}
			result={result ? renderResult(result.value) : undefined}
			onConfirm={confirm}
		/>
	);
}
