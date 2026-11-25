import { useState, type ComponentProps, type ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { EditForm, EditorDrawer, Page } from "./-components/common";

// Local preview state, not routing, fetching or a parallel page implementation.
export function EditorPreview({ title, ...args }: ComponentProps<typeof EditForm> & { title: string }) {
	const [open, setOpen] = useState(true);
	return (
		<Page title={title} action={<Button onClick={() => setOpen(true)}>Buka editor</Button>}>
			<EditorDrawer open={open} title={title} pending={args.pending} onClose={() => setOpen(false)}>
				{(onDirty) => <EditForm {...args} onDirty={onDirty} />}
			</EditorDrawer>
		</Page>
	);
}

type DialogConfig = { title: string; description: string; confirmLabel: string; tone: "default" | "destructive"; fields?: ReactNode };
/** The five states every adjustment dialog must pin, built from one domain config. */
export function adjustmentStates(config: DialogConfig, result: ReactNode) {
	const base = {
		...config,
		open: true,
		onOpenChange: () => {},
		reason: "Alasan contoh yang cukup panjang.",
		onReasonChange: () => {},
		showValidation: false,
		phase: "idle" as const,
		onConfirm: () => {},
	};
	return {
		Open: { args: { ...base, reason: "" } },
		ValidationError: { args: { ...base, reason: "Singkat", showValidation: true } },
		Submitting: { args: { ...base, phase: "submitting" as const } },
		Failed: { args: { ...base, phase: "failed" as const, error: "Permintaan ditolak. Periksa hak akses Anda lalu coba lagi." } },
		Success: { args: { ...base, phase: "success" as const, result } },
	};
}
