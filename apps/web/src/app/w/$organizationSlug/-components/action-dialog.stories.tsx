import { Fragment, useLayoutEffect, useRef, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { Button } from "@repo/ui/button";
import { Money } from "@repo/ui/money";
import { statusFailure } from "./-story-fixtures";
import { orgFailureMessage } from "../../../../lib/organization-pure";
import { ActionDialog, ActionDialogView } from "./action-dialog";

const noop = () => {};
const base = {
	open: true,
	onOpenChange: noop,
	title: "Tolak transfer",
	description: "Pembeli akan melihat pembayaran ini gagal.",
	confirmLabel: "Tolak transfer",
	tone: "destructive",
	reasonRequired: true,
	reason: "",
	onReasonChange: noop,
	showValidation: false,
	phase: "idle",
	onConfirm: noop,
} as const;

const meta = {
	title: "Web/Components/Dashboard/ActionDialog",
	component: ActionDialogView,
	parameters: { layout: "centered" },
	args: base,
	argTypes: { fields: { control: false }, result: { control: false } },
} satisfies Meta<typeof ActionDialogView>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Open: Story = {};
export const ValidationError: Story = { args: { reason: "pendek", showValidation: true } };
export const Submitting: Story = { args: { reason: "Bukti transfer tidak sesuai dengan pesanan.", phase: "submitting" } };
export const Success: Story = {
	args: {
		reason: "Bukti transfer tidak sesuai dengan pesanan.",
		phase: "success",
		result: (
			<p className="text-sm">
				Pembayaran <Money amount="1500000" currency="IDR" /> ditandai gagal.
			</p>
		),
	},
};
export const Failure: Story = {
	args: {
		reason: "Bukti transfer tidak sesuai dengan pesanan.",
		phase: "failed",
		error: "Tindakan tidak dapat dilakukan pada kondisi saat ini.",
	},
};
/** Without a reason field (e.g. confirm transfer): no textarea, confirm goes straight through. */
export const WithoutReason: Story = {
	args: {
		title: "Konfirmasi transfer",
		description: "Pembayaran akan ditandai berhasil.",
		confirmLabel: "Konfirmasi",
		tone: "default",
		reasonRequired: false,
	},
};

/** The real lifecycle: validation, pending, then either the simulated result or an inline failure. */
export const Interactive: Story = {
	parameters: { layout: "padded" },
	render: () => {
		const [open, setOpen] = useState(false);
		const [fail, setFail] = useState(false);
		return (
			<div className="flex gap-2">
				<Button
					type="button"
					onClick={() => {
						setFail(false);
						setOpen(true);
					}}
				>
					Buka (berhasil)
				</Button>
				<Button
					type="button"
					variant="outline"
					onClick={() => {
						setFail(true);
						setOpen(true);
					}}
				>
					Buka (gagal)
				</Button>
				<ActionDialog
					open={open}
					onOpenChange={setOpen}
					title="Tolak transfer"
					description="Pembeli akan melihat pembayaran ini gagal."
					confirmLabel="Tolak transfer"
					tone="destructive"
					reasonRequired
					describeError={orgFailureMessage}
					onSubmit={(reason) =>
						new Promise<string>((resolve, reject) =>
							setTimeout(() => (fail ? reject(statusFailure(503, "SERVER-DETAIL-SENTINEL: db password=hunter2")) : resolve(reason)), 700),
						)
					}
					renderResult={() => <p className="text-sm">Pembayaran ditandai gagal.</p>}
				/>
			</div>
		);
	},
};

// ---- Lifecycle harness -------------------------------------------------------------------------------------
// The real ActionDialog with a deferred onSubmit that the buttons below settle by hand. The parent (outside the
// keyed subtree, like the identity-keyed shell) records every callback, so a stale completion is observable.
type Deferred = { id: number; resolve: (value: number) => void; reject: (failure: unknown) => void };

function LifecycleHarness() {
	const deferreds = useRef<Deferred[]>([]);
	const [identity, setIdentity] = useState(0);
	const [results, setResults] = useState<number[]>([]);
	const [errorCopies, setErrorCopies] = useState(0);
	const [submittedCount, setSubmittedCount] = useState(0);
	const last = () => deferreds.current[deferreds.current.length - 1];
	// "After commit" actions settle the pending submission in a microtask queued from a layout-effect of the commit
	// that unmounts/closes the dialog: after the commit, before React's passive effects. A timer drives the update so
	// it is not a discrete event (discrete commits flush passive effects synchronously and hide the window).
	const armed = useRef<"resolve" | "reject" | null>(null);
	const closeRef = useRef<() => void>(() => {});
	const fire = () => {
		const mode = armed.current;
		armed.current = null;
		if (mode)
			queueMicrotask(() =>
				mode === "resolve" ? last()?.resolve(last()?.id ?? 0) : last()?.reject(statusFailure(503, "SERVER-DETAIL-SENTINEL")),
			);
	};
	const afterCommit = (mode: "resolve" | "reject", action: () => void) => {
		armed.current = mode;
		setTimeout(action, 0);
	};
	return (
		<div className="flex flex-col gap-3">
			<div className="flex flex-wrap gap-2">
				<Button type="button" variant="outline" onClick={() => setIdentity((value) => value + 1)}>
					Ganti identitas
				</Button>
				<Button type="button" variant="outline" onClick={() => afterCommit("resolve", () => setIdentity((value) => value + 1))}>
					Ganti identitas, selesaikan setelah commit
				</Button>
				<Button type="button" variant="outline" onClick={() => afterCommit("reject", () => setIdentity((value) => value + 1))}>
					Ganti identitas, tolak setelah commit
				</Button>
				<Button type="button" variant="outline" onClick={() => afterCommit("resolve", () => closeRef.current())}>
					Tutup dari luar, selesaikan setelah commit
				</Button>
				<Button type="button" variant="outline" onClick={() => last()?.resolve(last()?.id ?? 0)}>
					Selesaikan terakhir
				</Button>
				<Button type="button" variant="outline" onClick={() => last()?.reject(statusFailure(503, "SERVER-DETAIL-SENTINEL"))}>
					Tolak terakhir
				</Button>
			</div>
			<output data-harness="calls" className="rounded-md border border-border p-2 text-xs">
				{`onResult=[${results.join(",")}] describeError=${errorCopies} submitted=${submittedCount} identity=${identity}`}
			</output>
			<Fragment key={identity}>
				<LifecyclePage
					onCommit={fire}
					registerClose={(close) => (closeRef.current = close)}
					onSubmit={() =>
						new Promise<number>((resolve, reject) => {
							deferreds.current.push({ id: deferreds.current.length + 1, resolve, reject });
							setSubmittedCount(deferreds.current.length);
						})
					}
					onResult={(value) => setResults((previous) => [...previous, value])}
					describeError={(failure) => {
						setErrorCopies((value) => value + 1);
						return orgFailureMessage(failure);
					}}
				/>
			</Fragment>
		</div>
	);
}

/** Reports the commit that closes or unmounts the dialog, from a layout effect (runs inside the commit). */
function CommitProbe({ open, onCommit }: { open: boolean; onCommit: () => void }) {
	const latest = useRef(onCommit);
	latest.current = onCommit;
	useLayoutEffect(() => {
		if (!open) latest.current();
	}, [open]);
	useLayoutEffect(
		() => () => {
			latest.current();
		},
		[],
	);
	return null;
}

function LifecyclePage({
	onSubmit,
	onResult,
	describeError,
	onCommit,
	registerClose,
}: {
	onCommit: () => void;
	registerClose: (close: () => void) => void;
	onSubmit: () => Promise<number>;
	onResult: (value: number) => void;
	describeError: (failure: unknown) => string;
}) {
	const [open, setOpen] = useState(false);
	registerClose(() => setOpen(false));
	return (
		<>
			<CommitProbe open={open} onCommit={onCommit} />
			<div className="flex gap-2">
				<Button type="button" onClick={() => setOpen(true)}>
					Buka dialog
				</Button>
				<Button type="button" variant="outline" onClick={() => setOpen(false)}>
					Tutup dari luar
				</Button>
			</div>
			<ActionDialog
				open={open}
				onOpenChange={setOpen}
				title="Konfirmasi transfer"
				description="Siklus hidup dialog."
				confirmLabel="Konfirmasi"
				onSubmit={onSubmit}
				onResult={onResult}
				describeError={describeError}
				renderResult={(value) => <p>{`Hasil ${value}`}</p>}
			/>
		</>
	);
}

const settle = (ms = 60) => new Promise((resolve) => setTimeout(resolve, ms));
async function until(check: () => boolean, what: string) {
	for (let attempt = 0; attempt < 60; attempt++) {
		if (check()) return;
		await settle(50);
	}
	throw new Error(`Timed out waiting for: ${what}`);
}
const button = (text: string) => [...document.querySelectorAll("button")].find((candidate) => candidate.textContent?.trim() === text);
async function press(text: string) {
	await until(() => Boolean(button(text)), `button "${text}"`);
	button(text)?.click();
	await settle();
}
const calls = () => document.querySelector("[data-harness=calls]")?.textContent ?? "";
const expectCalls = (expected: string, step: string) => {
	const output = document.querySelector<HTMLElement>("[data-harness=calls]");
	// Observed callback counts per step stay readable on the page for manual verification.
	if (output) output.dataset.log = `${output.dataset.log ?? ""}${step} => ${calls()}\n`;
	if (!calls().includes(expected)) throw new Error(`${step}: expected "${expected}" in "${calls()}"`);
};
const submitted = () => Number(calls().match(/submitted=(\d+)/)?.[1] ?? 0);
// A leftover dialog still animating out would swallow the next click, so wait until it is gone, and prove each
// submit really started (otherwise a "nothing leaked" assertion could pass vacuously).
async function openDialog() {
	await until(() => !document.querySelector("[role=dialog]"), "previous dialog removed");
	await press("Buka dialog");
	await until(() => Boolean(button("Konfirmasi")), "dialog open");
}
async function submit() {
	const before = submitted();
	await press("Konfirmasi");
	await until(() => submitted() > before, "submission started");
}
const dialogText = () => document.querySelector("[role=dialog]")?.textContent ?? "";

/**
 * Plays itself in the browser (Storybook interactions, not a test runner): a completion that arrives after the
 * dialog unmounted (identity change), was closed, or belongs to an earlier open cycle must set no state and call
 * neither onResult nor describeError. Failure throws, which Storybook shows as a failed play function.
 */
export const Lifecycle: Story = {
	parameters: { layout: "padded" },
	render: () => <LifecycleHarness />,
	play: async () => {
		const done = (ok: boolean) => document.querySelector("[data-harness=calls]")?.setAttribute("data-result", ok ? "passed" : "failed");
		try {
			// 1. identity change (unmount) while pending, then resolve
			await openDialog();
			await submit();
			await press("Ganti identitas");
			await press("Selesaikan terakhir");
			await settle(150);
			expectCalls("onResult=[]", "resolve after unmount");
			// 2. closed from outside while pending, then resolve
			await openDialog();
			await submit();
			await press("Tutup dari luar");
			await press("Selesaikan terakhir");
			await settle(150);
			expectCalls("onResult=[]", "resolve after close");
			// 3. identity change while pending, then reject: no error copy computed or shown
			await openDialog();
			await submit();
			await press("Ganti identitas");
			await press("Tolak terakhir");
			await settle(150);
			expectCalls("describeError=0", "reject after unmount");
			// 4. close, reopen, then the OLD submission settles: it must not leak into the new cycle
			await openDialog();
			await submit();
			const stale = calls().match(/submitted=(\d+)/)?.[1];
			await press("Tutup dari luar");
			await settle(500);
			await openDialog();
			await until(() => Boolean(button("Konfirmasi")), "reopened dialog");
			await press("Selesaikan terakhir");
			await settle(150);
			expectCalls("onResult=[]", `old submission #${stale} after reopen`);
			if (dialogText().includes("Hasil")) throw new Error("old result rendered in the reopened dialog");
			// 5. the current cycle still completes normally
			await submit();
			await press("Selesaikan terakhir");
			await until(() => dialogText().includes("Hasil"), "current result");
			if (!/onResult=\[\d+\]/.test(calls())) throw new Error(`current result not delivered: ${calls()}`);
			await press("Tutup");
			await settle(500);
			// 6. current-cycle rejection shows fixed copy only
			await openDialog();
			await submit();
			await press("Tolak terakhir");
			await until(() => dialogText().includes("Layanan sedang tidak tersedia"), "fixed failure copy");
			if (dialogText().includes("SENTINEL")) throw new Error("server detail rendered");
			expectCalls("describeError=1", "current rejection");
			await press("Tutup dari luar");
			await settle(500);
			// 7. the stale completion lands right after the unmount commit, before passive effects (resolve)
			const before = calls().match(/onResult=\[[^\]]*\]/)?.[0];
			await openDialog();
			await submit();
			await press("Ganti identitas, selesaikan setelah commit");
			await settle(200);
			expectCalls(before ?? "", "resolve between unmount commit and passive cleanup");
			// 8. same for a rejection: no error copy is computed
			const errorsBefore = calls().match(/describeError=\d+/)?.[0];
			await openDialog();
			await submit();
			await press("Ganti identitas, tolak setelah commit");
			await settle(200);
			expectCalls(errorsBefore ?? "", "reject between unmount commit and passive cleanup");
			// 9. and for a close driven by the parent
			await openDialog();
			await submit();
			await press("Tutup dari luar, selesaikan setelah commit");
			await settle(200);
			expectCalls(before ?? "", "resolve between close commit and passive cleanup");
			done(true);
		} catch (failure) {
			done(false);
			throw failure;
		}
	},
};
