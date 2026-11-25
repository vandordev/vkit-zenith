import type { ReactNode } from "react";
import { Link, type LinkProps } from "@tanstack/react-router";
import { ListChecks } from "lucide-react";
import { StatsCard, type StatsCardItem } from "@repo/ui/stats-card";
import type { V1HomeCountersBody, V1HomeCounter } from "@bo/lib/api/generated/types.gen";
import { Page } from "../../-components/common";
import { ContentState } from "../../-components/content-state";
import type { SectionState } from "../../-components/org-section";
import { jobsSearch } from "../-adapters/global-search";

const targets = {
	"/operations/jobs": jobsSearch,
} as const;

/**
 * The counter's API `href` carries filters in URL vocabulary. It becomes a typed link only
 * when the destination route's own validator accepts every value unchanged, so a link can never
 * silently open a list that differs from what the count measured.
 */
export function counterTarget(href: string): Pick<LinkProps, "to" | "search"> | undefined {
	const url = new URL(href, "https://bo.invalid");
	const schema = (targets as Record<string, (typeof targets)[keyof typeof targets]>)[url.pathname];
	if (!schema || url.origin !== "https://bo.invalid") return undefined;
	const raw = Object.fromEntries(url.searchParams);
	const parsed = schema.safeParse(raw);
	if (!parsed.success) return undefined;
	const search = parsed.data as Record<string, unknown>;
	// Any value the validator changed is rejected.
	if (Object.entries(raw).some(([key, value]) => String(search[key] ?? "") !== value)) return undefined;
	return { to: url.pathname, search } as Pick<LinkProps, "to" | "search">;
}

const definitions = [{ key: "failedJobs", title: "Job gagal atau akan diulang", icon: ListChecks }] as const satisfies readonly {
	key: keyof V1HomeCountersBody;
	title: string;
	icon: StatsCardItem["icon"];
}[];

export type HomeOverviewProps = {
	counters?: V1HomeCountersBody;
	state: SectionState;
	error?: string;
	onRetry: () => void;
	/** Real recent activity (superadmin only); absent for other roles. */
	children?: ReactNode;
};

function CounterLink({ counter }: { counter: V1HomeCounter }) {
	const target = counterTarget(counter.href);
	return target ? (
		<Link {...target} className="underline">
			Lihat daftar
		</Link>
	) : null;
}

/** Beranda: operational counters (each linking to its filtered list), then the existing activity. */
export function HomeOverview({ counters, state, error, onRetry, children }: HomeOverviewProps) {
	return (
		<Page title="Beranda" description="Hal yang perlu perhatian staf.">
			<section aria-labelledby="counters-heading" className="flex min-w-0 flex-col gap-4">
				<h2 id="counters-heading" className="text-base font-semibold">
					Perlu ditindaklanjuti
				</h2>
				{state === "pending" && <ContentState kind="loading" title="Memuat penghitung…" />}
				{state === "error" && <ContentState kind="error" title={error ?? "Penghitung tidak dapat dimuat."} onRetry={onRetry} />}
				{state === "ready" && counters && (
					<StatsCard
						ariaLabel="Penghitung operasional"
						items={definitions.map(({ key, title, icon }) => ({
							key,
							title,
							icon,
							value: { display: counters[key].count },
							caption: <CounterLink counter={counters[key]} />,
						}))}
					/>
				)}
			</section>
			{children}
		</Page>
	);
}
