import { createFileRoute, Link } from "@tanstack/react-router";
import { useRef } from "react";
import { ActivityPage } from "./-components/activity-page";
import { ActivityFilters, type ActivityFilterValue } from "./-components/activity-filters";
import { ActivityDrawer } from "./-adapters/activity-drawer";
import { ActivityChoices } from "./-adapters/activity-choices";
import { ContentState } from "@bo/app/-components/content-state";
import { useProtected, usePrivateQuery } from "@bo/lib/api/session";
import * as sdk from "@bo/lib/api/generated/sdk.gen";
import { activityRange } from "@bo/lib/api/activity-range";
import { parseActivitySearch, activityFilterSearch } from "@bo/lib/api/activity-navigation";
import { errorMessage } from "@bo/lib/api/client";
export const Route = createFileRoute("/_authenticated/activity/")({ validateSearch: parseActivitySearch, component: Activity });
const noop = () => {};
function Activity() {
	const { staff, epoch } = useProtected();
	// Admin never mounts audit query owners, including selected/choice reads.
	return staff.role === "superadmin" ? (
		<AuthorizedActivity key={`${staff.id}:${epoch}`} />
	) : (
		<ActivityPage denied onPage={noop} onReload={noop} onReset={noop} renderLink={() => null} />
	);
}
function AuthorizedActivity() {
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const fallback = useRef<HTMLDivElement>(null);
	const focus = useRef<HTMLElement | null>(null);
	let range: ReturnType<typeof activityRange> = {};
	let validation = search.invalid ? "Filter aktivitas tidak valid. Kosongkan filter untuk melanjutkan." : undefined;
	try {
		range = activityRange(search.from, search.to);
	} catch {
		validation = "Rentang tanggal tidak valid. Periksa tanggal atau kosongkan filter.";
	}
	const actions = usePrivateQuery(
		["activity-actions"],
		async (t, signal) => (await sdk.boV1ListActivityActions({ client: t.client, signal, throwOnError: true })).data,
	);
	if (search.action && actions.data && !actions.data.items.some((item) => item.value === search.action))
		validation = "Tindakan tercatat tidak didukung. Kosongkan filter untuk melanjutkan.";
	const enabled = !validation && (!search.action || !!actions.data);
	const query = {
		...range,
		action: search.action,
		outcome: search.outcome,
		actorStaffId: search.actorStaffId,
		page: search.page,
		pageSize: 25,
	};
	const list = usePrivateQuery(
		["activity", query],
		async (t, signal) => (await sdk.boV1ListActivity({ client: t.client, signal, throwOnError: true, query })).data,
		enabled,
	);
	const change = (filters: ActivityFilterValue) => {
		void navigate({ search: activityFilterSearch({ ...filters, selected: search.selected }) });
	};
	const reset = () => change({});
	return (
		<ActivityPage
			focusRef={fallback}
			data={enabled ? list.data : undefined}
			actions={actions.data?.items}
			loading={!validation && ((!enabled && !actions.error) || list.isPending)}
			error={
				validation ??
				(!enabled && actions.error ? "Pilihan tindakan tidak dapat dimuat." : list.error ? errorMessage(list.error) : undefined)
			}
			onPage={(page) => {
				void navigate({ search: { ...search, page } });
			}}
			onReset={reset}
			onReload={() => {
				if (enabled) void list.refetch();
			}}
			filters={
				<>
					<ActivityFilters
						value={search}
						actions={actions.data?.items ?? []}
						error={validation}
						onChange={change}
						onReset={reset}
						choices={<ActivityChoices value={search} onChange={change} />}
					/>
					{actions.error && (
						<ContentState
							kind="error"
							title="Pilihan tindakan tidak dapat dimuat"
							onRetry={() => {
								void actions.refetch();
							}}
						/>
					)}
				</>
			}
			renderLink={(row) => (
				<Link
					to="/activity"
					search={{ ...search, selected: row.id }}
					onClick={(event) => {
						focus.current = event.currentTarget;
					}}
				/>
			)}
		>
			<ActivityDrawer
				selected={!validation ? search.selected : undefined}
				onClose={() => {
					void navigate({ search: { ...search, selected: undefined }, replace: true });
				}}
				returnFocus={{
					get current() {
						return focus.current?.isConnected ? focus.current : fallback.current;
					},
				}}
			/>
		</ActivityPage>
	);
}
