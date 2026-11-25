import type { ReactNode } from "react";
import { DatePicker } from "@repo/ui/date-picker";
import { dateRangeIndonesianLocale } from "@repo/ui/date-range-picker";
import { FilterBar } from "@repo/ui/filter-bar";
import { Button } from "@repo/ui/button";
import type { V1ActivityAction } from "@bo/lib/api/generated/types.gen";
import { filterLabels } from "../../../-components/filter-labels";

export type ActivityFilterValue = {
	from?: string;
	to?: string;
	action?: string;
	outcome?: "success" | "failure" | "denied";
	actorStaffId?: string;
};
function calendarDate(value?: string) {
	if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return;
	const date = new Date(0);
	const [year, month, day] = value.split("-").map(Number);
	if (year === undefined || month === undefined || day === undefined) return;
	date.setFullYear(year, month - 1, day);
	date.setHours(0, 0, 0, 0);
	return date.getFullYear() === year && date.getMonth() === month - 1 && date.getDate() === day ? date : undefined;
}
function dateOnly(date?: Date) {
	return date
		? `${String(date.getFullYear()).padStart(4, "0")}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`
		: undefined;
}
export function ActivityFilters({
	value,
	actions,
	choices,
	error,
	onChange,
	onReset,
}: {
	value: ActivityFilterValue;
	actions: V1ActivityAction[];
	choices?: ReactNode;
	error?: string;
	onChange: (value: ActivityFilterValue) => void;
	onReset: () => void;
}) {
	const from = calendarDate(value.from);
	const to = calendarDate(value.to);
	return (
		<section aria-label="Filter aktivitas" className="space-y-4">
			<FilterBar
				filters={[
					{
						id: "action",
						label: "Tindakan tercatat",
						options: actions,
						value: value.action ?? "",
						onChange: (action) => onChange({ ...value, action: action || undefined }),
					},
					{
						id: "outcome",
						label: "Hasil tercatat",
						options: [
							{ value: "success", label: "Berhasil" },
							{ value: "failure", label: "Gagal" },
							{ value: "denied", label: "Ditolak" },
						],
						value: value.outcome ?? "",
						onChange: (outcome) =>
							onChange({ ...value, outcome: outcome === "success" || outcome === "failure" || outcome === "denied" ? outcome : undefined }),
					},
				]}
				labels={filterLabels}
			/>
			{/* Dates and the staff choice are not single choices, so they stay as their own controls. */}
			<div className="grid min-w-0 items-end gap-4 sm:grid-cols-2 md:max-w-2xl">
				<DatePicker
					label="Dari tanggal"
					placeholder="Semua tanggal"
					locale={dateRangeIndonesianLocale}
					value={from}
					onValueChange={(date) =>
						onChange({
							...value,
							from: date && to && date > to ? value.to : dateOnly(date),
							to: date && to && date > to ? dateOnly(date) : value.to,
						})
					}
				/>
				<DatePicker
					label="Sampai tanggal"
					placeholder="Semua tanggal"
					locale={dateRangeIndonesianLocale}
					value={to}
					onValueChange={(date) =>
						onChange({
							...value,
							from: date && from && date < from ? dateOnly(date) : value.from,
							to: date && from && date < from ? value.from : dateOnly(date),
						})
					}
				/>
			</div>
			{choices}
			{(value.from || value.to) && (
				<p className="text-xs text-muted-foreground">
					Tanggal kalender lokal browser: {value.from || "tanpa batas awal"} sampai {value.to || "tanpa batas akhir"}, hari terakhir
					termasuk.
				</p>
			)}
			{error && (
				<p role="alert" className="text-sm">
					{error}
				</p>
			)}
			<Button type="button" variant="ghost" onClick={onReset}>
				Kosongkan filter
			</Button>
		</section>
	);
}
