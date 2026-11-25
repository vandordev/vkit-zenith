import { useRef, useState, type RefObject } from "react";
import { Button } from "@repo/ui/button";
import { AdjustmentDialog } from "../../../-components/adjustment-dialog";

export type OpsActionSpec = {
	key: string;
	label: string;
	title: string;
	description: string;
	tone?: "default" | "destructive";
	/** The backend's guard state; the mounted dialog outlives a state change until dismissed. */
	visible: boolean;
	onSubmit: (reason: string) => Promise<string>;
};

/** Staff actions of an operations record, each behind the shared reason-required dummy dialog. */
export function OpsActions({ ariaLabel, actions }: { ariaLabel: string; actions: readonly OpsActionSpec[] }) {
	const group = useRef<HTMLDivElement>(null);
	return (
		<div
			ref={group}
			role="group"
			aria-label={ariaLabel}
			tabIndex={-1}
			className="flex flex-wrap items-center gap-2 rounded-md p-0.5 outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
		>
			{actions.map(({ key, ...action }) => (
				<OpsAction key={key} fallbackFocus={group} {...action} />
			))}
			{actions.every((action) => !action.visible) && <p className="text-sm text-muted-foreground">Tidak ada tindakan untuk status ini.</p>}
		</div>
	);
}

/**
 * One reason-required action. Focus returns to the opener while it still exists, else to the action
 * group: a successful action can remove its own trigger.
 */
function OpsAction({
	label,
	title,
	description,
	tone = "default",
	visible,
	onSubmit,
	fallbackFocus,
}: Omit<OpsActionSpec, "key"> & { fallbackFocus: RefObject<HTMLElement | null> }) {
	const [open, setOpen] = useState(false);
	const opener = useRef<HTMLElement | null>(null);
	return (
		<>
			{visible && (
				<Button
					type="button"
					variant={tone === "destructive" ? "destructive" : "outline"}
					onClick={(event) => {
						opener.current = event.currentTarget;
						setOpen(true);
					}}
				>
					{label}
				</Button>
			)}
			<AdjustmentDialog
				finalFocus={() => (opener.current?.isConnected ? opener.current : fallbackFocus.current)}
				open={open}
				onOpenChange={setOpen}
				title={title}
				description={description}
				confirmLabel={label}
				tone={tone}
				onSubmit={onSubmit}
				renderResult={(result) => <p>{String(result)}</p>}
			/>
		</>
	);
}
