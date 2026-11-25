import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { AdjustmentDialog, AdjustmentDialogView } from "./adjustment-dialog";
import { StatusBadge } from "./status-badge";
import { noop } from "../-fixtures";

const base = {
	open: true,
	onOpenChange: noop,
	title: "Cabut API key",
	description: "Kunci zk_ab12… berhenti berlaku seketika. Tindakan ini tidak dapat dibatalkan.",
	confirmLabel: "Cabut API key",
	tone: "destructive" as const,
	reason: "Kunci bocor di repositori publik.",
	onReasonChange: noop,
	showValidation: false,
	phase: "idle" as const,
	onConfirm: noop,
};
const resultNode = (
	<p className="flex items-center gap-2 text-sm">
		Status API key menjadi <StatusBadge kind="apiKey" status="revoked" />
	</p>
);
const meta = {
	title: "BO/Components/AdjustmentDialog",
	component: AdjustmentDialogView,
	parameters: { layout: "fullscreen" },
	args: base,
	argTypes: { fields: { control: false }, result: { control: false } },
} satisfies Meta<typeof AdjustmentDialogView>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Open: Story = { args: { reason: "" } };
export const ValidationError: Story = { args: { reason: "Bocor", showValidation: true } };
export const Submitting: Story = { args: { phase: "submitting" } };
export const Failed: Story = { args: { phase: "failed", error: "Permintaan ditolak. Konfirmasi keamanan berubah, coba kirim kembali." } };
export const Success: Story = { args: { phase: "success", result: resultNode } };
export const DefaultTone: Story = {
	args: {
		tone: "default",
		title: "Aktifkan provider",
		confirmLabel: "Aktifkan",
		description: "Provider akan kembali menerima pembayaran.",
	},
};
export const WithFields: Story = {
	args: {
		fields: <p className="rounded-md border border-border bg-muted p-3 text-sm">Kolom tambahan khusus tindakan tampil di atas alasan.</p>,
	},
};
export const LongReason: Story = { args: { reason: "Alasan yang sangat panjang ".repeat(40).slice(0, 1024) } };
// Real lifecycle: validation, busy guard, then the returned result or an inline failure.
export const Interactive: Story = {
	render: function Preview() {
		const [open, setOpen] = useState(false);
		const [fail, setFail] = useState(false);
		return (
			<div className="flex gap-2 p-6">
				<Button
					onClick={() => {
						setFail(false);
						setOpen(true);
					}}
				>
					Buka (berhasil)
				</Button>
				<Button
					variant="outline"
					onClick={() => {
						setFail(true);
						setOpen(true);
					}}
				>
					Buka (gagal)
				</Button>
				<AdjustmentDialog
					{...base}
					open={open}
					onOpenChange={setOpen}
					onSubmit={(reason) =>
						new Promise<string>((resolve, reject) => setTimeout(() => (fail ? reject(new Error("gagal")) : resolve(reason)), 900))
					}
					renderResult={(reason) => (
						<div>
							{resultNode}
							<p className="mt-2 text-xs text-muted-foreground">Alasan: {reason}</p>
						</div>
					)}
				/>
			</div>
		);
	},
};
