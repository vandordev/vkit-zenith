import { useRef } from "react";
import type { V1ActivityStaffChoice } from "@bo/lib/api/generated/types.gen";
import { Autocomplete } from "@repo/ui/autocomplete";
import { usePrivateChoices } from "@bo/lib/api/choice-query";
import { usePrivateQuery } from "@bo/lib/api/session";
import { errorMessage } from "@bo/lib/api/client";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import type { ActivityFilterValue } from "../-components/activity-filters";

export function ActivityChoices({ value, onChange }: { value: ActivityFilterValue; onChange: (value: ActivityFilterValue) => void }) {
	const staff = usePrivateQuery(
		["activity-staff-choices", "selected", value.actorStaffId],
		async (t, signal) =>
			(
				await sdk.boV1ListActivityStaffChoices({
					client: t.client,
					signal,
					throwOnError: true,
					query: { page: 1, pageSize: 25, selectedId: value.actorStaffId },
				})
			).data,
		!!value.actorStaffId,
	);
	const staffChoices = usePrivateChoices(
		["activity-staff-choices", "infinite", value.actorStaffId],
		async (t, signal, page, q) =>
			(
				await sdk.boV1ListActivityStaffChoices({
					client: t.client,
					signal,
					throwOnError: true,
					query: { q, page, pageSize: 25, selectedId: value.actorStaffId },
				})
			).data,
		(data) => data.items,
	);
	const staffLabel = useRef<V1ActivityStaffChoice | null>(null);
	if (staff.data?.selected) staffLabel.current = staff.data.selected;
	const currentStaff = staffLabel.current;
	const staffChoice =
		currentStaff && currentStaff.id === value.actorStaffId
			? { ...currentStaff, key: currentStaff.status === "disabled" ? "Dinonaktifkan" : "Aktif" }
			: null;
	return (
		<div className="grid min-w-0 gap-5 lg:grid-cols-2">
			<div>
				<Autocomplete
					{...staffChoices.props}
					mode="selection"
					multiple={false}
					label="Staf pelaku (nama saat ini)"
					error={staff.error ? errorMessage(staff.error) : staffChoices.props.error}
					onRetry={() => {
						if (value.actorStaffId) void staff.refetch();
						void staffChoices.props.onRetry?.();
					}}
					value={staffChoice}
					onValueChange={(item) => {
						staffLabel.current = item;
						onChange({ ...value, actorStaffId: item?.id });
					}}
				/>
				{value.actorStaffId && !staffChoice && (
					<p className="break-all text-xs text-muted-foreground">Staf terpilih: {value.actorStaffId}. Label saat ini belum tersedia.</p>
				)}
			</div>
		</div>
	);
}
