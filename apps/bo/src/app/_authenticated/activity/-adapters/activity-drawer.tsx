import type { RefObject } from "react";
import { Link } from "@tanstack/react-router";
import { ActivityDetail } from "../-components/activity-detail";
import { usePrivateQuery } from "@bo/lib/api/session";
import { boV1GetActivity } from "@bo/lib/api/generated/sdk.gen";
import { errorMessage } from "@bo/lib/api/client";
import { activityTarget } from "@bo/lib/api/activity-navigation";

export function ActivityDrawer({
	selected,
	onClose,
	returnFocus,
}: {
	selected?: string;
	onClose: () => void;
	returnFocus: RefObject<HTMLElement | null>;
}) {
	const detail = usePrivateQuery(
		["activity-detail", selected],
		async (t, signal) => (await boV1GetActivity({ client: t.client, signal, throwOnError: true, path: { id: selected ?? "" } })).data,
		!!selected,
	);
	const destination = detail.data && detail.data.id === selected ? activityTarget(detail.data) : undefined;
	return (
		<ActivityDetail
			selected={selected}
			data={detail.data}
			loading={detail.isPending}
			error={detail.error ? errorMessage(detail.error) : undefined}
			onClose={onClose}
			onRetry={() => {
				void detail.refetch();
			}}
			returnFocus={returnFocus}
			targetLink={
				destination ? (
					<Link className="underline" to={destination}>
						Buka target
					</Link>
				) : undefined
			}
		/>
	);
}
