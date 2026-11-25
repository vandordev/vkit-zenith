import type { ReactNode } from "react";
import { CopyableId } from "@repo/ui/copyable-id";

/**
 * The one-time secret view, shared by API key and webhook creation. It only renders what its parent dialog holds in local state;
 * nothing here stores, logs or navigates with the value.
 */
export function SecretResult({ title, label, secret }: { title: ReactNode; label: string; secret: string }) {
	return (
		<div className="flex min-w-0 flex-col gap-3">
			<p className="text-sm">{title}</p>
			<div className="min-w-0 rounded-md border border-border bg-muted/40 p-3">
				<CopyableId value={secret} label={label} />
			</div>
			<p role="alert" className="rounded-md border border-warning/30 bg-warning/10 p-3 text-sm text-warning-text">
				Salin {label.toLowerCase()} ini sekarang. Nilai ini tidak akan ditampilkan lagi setelah dialog ditutup.
			</p>
		</div>
	);
}
