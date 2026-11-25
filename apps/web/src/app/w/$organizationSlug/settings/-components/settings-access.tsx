import type { ReactNode } from "react";

/**
 * Whether the viewer may change settings here. A role without the permission sees a short note instead of the controls; a suspended
 * organization accepts no writes at all (the banner above says why), whatever the role.
 */
export type WriteAccess = "allowed" | "no-permission" | "suspended";
export const writeAccess = (hasPermission: boolean, organizationStatus: "active" | "suspended"): WriteAccess =>
	!hasPermission ? "no-permission" : organizationStatus === "suspended" ? "suspended" : "allowed";

/** Replaces the controls a role cannot use; renders nothing when the viewer may write. */
export function AccessNote({ access, children }: { access: WriteAccess; children: ReactNode }) {
	if (access === "allowed") return null;
	return (
		<p role="note" className="rounded-md border border-border bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
			{access === "suspended" ? "Organisasi ditangguhkan: perubahan dinonaktifkan sampai organisasi diaktifkan kembali." : children}
		</p>
	);
}
