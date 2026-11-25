import type { ReactNode } from "react";
import { Cog, Globe, KeyRound, ShieldCheck, UserRound } from "lucide-react";
import { Badge } from "@repo/ui/badge";
import type { V1Actor } from "@bo/lib/api/generated/types.gen";

const roles: Record<string, string> = { owner: "Pemilik", admin: "Admin", member: "Anggota", superadmin: "Superadmin" };

function describe(actor: V1Actor): { label: string; icon: ReactNode; detail?: string } {
	// `kind` can grow in a newer API: a value outside the known union renders raw.
	switch (actor.kind as string) {
		case "staff":
			return { label: "Staf Zenith", icon: <ShieldCheck />, detail: join(actor.name, role(actor.role)) };
		case "user":
			return { label: "Pengguna organisasi", icon: <UserRound />, detail: join(actor.name, role(actor.role)) };
		case "client":
			return {
				label: "API key",
				icon: <KeyRound />,
				detail: join(actor.name, actor.apiKeyPrefix ? `${actor.apiKeyPrefix}…` : null),
			};
		case "system":
			return { label: "Sistem", icon: <Cog />, detail: actor.system ?? undefined };
		case "anonymous":
			return { label: "Pembeli / publik", icon: <Globe /> };
		default:
			return { label: "Pelaku lain", icon: <Cog />, detail: actor.kind };
	}
}
const role = (value: string | null) => (value ? (roles[value] ?? value) : null);
const join = (...parts: (string | null)[]) => parts.filter(Boolean).join(" · ") || undefined;

export function ActorChip({ actor }: { actor: V1Actor }) {
	const { label, icon, detail } = describe(actor);
	return (
		<span
			data-slot="actor-chip"
			data-actor-kind={actor.kind}
			className="inline-flex min-w-0 max-w-full flex-wrap items-center gap-x-2 gap-y-1"
		>
			<Badge variant="outline" className="shrink-0 gap-1 [&_svg]:size-3">
				{icon}
				{label}
			</Badge>
			{detail && <span className="min-w-0 text-xs text-muted-foreground wrap-anywhere">{detail}</span>}
		</span>
	);
}
