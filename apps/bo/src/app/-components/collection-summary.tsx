import { ChartNoAxesCombined } from "lucide-react";
import { StatsCard } from "@repo/ui/stats-card";

export type SummaryMetric = { key: string; label: string; value: string };
export function CollectionSummary({
	metrics,
	ariaLabel = "Ringkasan koleksi",
}: {
	metrics?: readonly SummaryMetric[];
	ariaLabel?: string;
}) {
	if (!metrics) return null;
	return (
		<StatsCard
			ariaLabel={ariaLabel}
			items={metrics.map((metric) => ({
				key: metric.key,
				title: metric.label,
				value: { display: metric.value },
				icon: ChartNoAxesCombined,
			}))}
		/>
	);
}
