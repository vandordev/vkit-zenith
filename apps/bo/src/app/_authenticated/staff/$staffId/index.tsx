import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { StaffSummary } from "./-components/staff-summary";
import * as sdk from "../../../../lib/api/generated/sdk.gen";
import * as schemas from "../../../../lib/api/generated/zod.gen";
import { errorMessage } from "../../../../lib/api/client";
import { formInput, passwordValid } from "../../../../lib/api/forms";
import { usePrivateQuery, useProtected } from "../../../../lib/api/session";
import { useUnsaved, useWrite } from "../../../../lib/api/workspace";
import { staffDetailSearch, staffReturnSearch } from "@bo/lib/api/customer-navigation";

export const Route = createFileRoute("/_authenticated/staff/$staffId/")({ validateSearch: staffDetailSearch, component: StaffOverview });
function StaffOverview() {
	const { staff } = useProtected();
	const { staffId } = Route.useParams();
	const { returnTo } = Route.useSearch();
	// The layout denies admins; this guard keeps the account query and editor from mounting anyway.
	return staff.role === "superadmin" ? <Overview key={staffId} staffId={staffId} returnTo={returnTo} /> : null;
}
function Overview({ staffId, returnTo }: { staffId: string; returnTo?: string }) {
	const { staff, transport } = useProtected();
	const write = useWrite();
	const [dirty, setDirty] = useState(false);
	const [saves, setSaves] = useState(0);
	useUnsaved(dirty);
	const detail = usePrivateQuery(
		["staff-detail", staffId],
		async (t, signal) => (await sdk.boV1GetStaff({ client: t.client, signal, throwOnError: true, path: { id: staffId } })).data,
	);
	return (
		<StaffSummary
			staff={detail.data}
			error={detail.error ? errorMessage(detail.error) : undefined}
			loading={detail.isFetching}
			saves={saves}
			pending={write.pending}
			feedback={write.feedback}
			backLink={
				<Link className="text-sm underline" to="/staff" search={staffReturnSearch(returnTo)}>
					Kembali ke daftar staf
				</Link>
			}
			onDirty={setDirty}
			onRetry={() => {
				void detail.refetch();
			}}
			onReload={async () => {
				const result = await detail.refetch();
				if (result.error) throw result.error;
				if (result.data) write.setFeedback(undefined);
				return result.data;
			}}
			onSubmit={(values, revision, replacing) => {
				if (!detail.data) return;
				const current = detail.data;
				void write.run(
					async (t) => {
						if (replacing && !passwordValid(values.password ?? ""))
							throw new Error("Periksa kata sandi baru:15–128 karakter dan maksimal512 byte UTF-8.");
						const options = await t.writeOptions();
						if (replacing)
							await sdk.boV1ReplaceStaffPassword({
								...options,
								path: { id: current.id },
								body: formInput(schemas.zBoV1ReplaceStaffPasswordBody, { ...values, revision }),
							});
						else
							await sdk.boV1UpdateStaff({
								...options,
								path: { id: current.id },
								body: formInput(schemas.zBoV1UpdateStaffBody, { ...values, revision }),
							});
						if (current.id === staff.id && (replacing || values.role !== staff.role || values.status !== staff.status)) {
							transport.reset();
							window.location.replace("/login");
						}
					},
					() => {
						setDirty(false);
						setSaves((count) => count + 1);
					},
				);
			}}
		/>
	);
}
