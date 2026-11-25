import { useEffect, useMemo, useState } from "react";
import { Badge } from "@repo/ui/badge";
import { Button } from "@repo/ui/button";
import { DateTime } from "@repo/ui/date-time";
import { Input } from "@repo/ui/input";
import type { ApiKey, ApiKeyCreateInputBody, V1CreateApiKeyResult, V1Page } from "../../../../../../lib/api/generated/types.gen";
import { ActionDialog } from "../../../-components/action-dialog";
import { dateOnly } from "../../../-components/format";
import { FieldProblem } from "../../../-components/field-problem";
import { buildBody, FormProblem } from "../../../-components/form-input";
import { ListPage } from "../../../-components/list-page";
import { StatusBadge } from "../../../-components/status-badge";
import { AccessNote, type WriteAccess } from "../../-components/settings-access";
import { SecretResult } from "../../-components/secret-result";
import { SettingsFilters, type SettingsFilterValue } from "../../-components/settings-filters";
import { apiKeyNameMax, apiKeyProblems, createApiKeyBody, emptyApiKeyDraft, type ApiKeyDraft } from "./api-key-input";

export type ApiKeysListProps = {
	state: "ready" | "pending" | "error";
	keys: readonly ApiKey[];
	pagination?: V1Page;
	filters: SettingsFilterValue;
	navigationKey: string;
	access: WriteAccess;
	simulated?: boolean;
	/** Ids that exist only as a dummy create result: shown, but the fixtures cannot revoke them. */
	simulatedIds: ReadonlySet<string>;
	loading?: boolean;
	errorMessage?: string;
	describeError: (failure: unknown) => string;
	onFilter: (filters: SettingsFilterValue) => void;
	onPage: (page: number) => void;
	onRetry: () => void;
	/** Resolves with the secret too: it is shown by the creating dialog and handed to nothing else. */
	onCreate: (body: ApiKeyCreateInputBody) => Promise<V1CreateApiKeyResult>;
	onRevoke: (keyId: string) => Promise<ApiKey>;
	/** Metadata only, never the secret. */
	onResult: (key: ApiKey, kind: "created" | "revoked") => void;
};

const statusOptions = [
	{ value: "active", label: "Aktif" },
	{ value: "revoked", label: "Dicabut" },
	{ value: "expired", label: "Kedaluwarsa" },
] as const;

/** Props-only form fields (the dialog stories render these directly). */
export function ApiKeyFields({
	draft,
	onDraft,
	problems,
	disabled,
}: {
	draft: ApiKeyDraft;
	onDraft: (draft: ApiKeyDraft) => void;
	problems: Partial<Record<keyof ApiKeyDraft, string>>;
	disabled?: boolean;
}) {
	return (
		<>
			<div className="flex min-w-0 flex-col gap-1.5">
				<Input
					labelStyle="static"
					label="Nama kunci"
					value={draft.name}
					maxLength={apiKeyNameMax}
					disabled={disabled}
					autoComplete="off"
					aria-invalid={problems.name ? true : undefined}
					aria-describedby="api-key-name-problem"
					onChange={(event) => onDraft({ ...draft, name: event.target.value })}
				/>
				<FieldProblem id="api-key-name-problem">{problems.name}</FieldProblem>
			</div>
			<div className="flex min-w-0 flex-col gap-1.5">
				<Input
					labelStyle="static"
					label="Kedaluwarsa pada (opsional)"
					type="date"
					value={draft.date}
					disabled={disabled}
					aria-invalid={problems.date ? true : undefined}
					aria-describedby="api-key-date-problem"
					onChange={(event) => onDraft({ ...draft, date: event.target.value })}
				/>
				<FieldProblem id="api-key-date-problem">{problems.date}</FieldProblem>
				<p className="text-xs text-muted-foreground">Berlaku sampai akhir hari yang dipilih, sesuai zona waktu perangkat Anda.</p>
			</div>
		</>
	);
}

export function ApiKeysList({
	access,
	describeError,
	onCreate,
	onRevoke,
	onResult,
	filters,
	keys,
	simulatedIds,
	...page
}: ApiKeysListProps) {
	const [dialog, setDialog] = useState<"create" | "revoke">();
	const [target, setTarget] = useState<ApiKey>();
	const [draft, setDraft] = useState<ApiKeyDraft>(emptyApiKeyDraft);
	const [show, setShow] = useState(false);
	// Fixed per open dialog: the "not already past" check judges one instant.
	const now = useMemo(() => new Date(), [dialog]);
	const canManage = access === "allowed";
	useEffect(() => {
		if (dialog === "create") return;
		setDraft(emptyApiKeyDraft);
		setShow(false);
	}, [dialog]);
	const close = (open: boolean) => !open && setDialog(undefined);
	const create = canManage ? (
		<Button type="button" onClick={() => setDialog("create")}>
			Buat API key
		</Button>
	) : undefined;
	const problems = show ? apiKeyProblems(draft, now) : {};
	return (
		<>
			<ListPage
				title="API Key"
				description="Kunci untuk memanggil API Zenith dari sistem Anda. Setiap kunci diawali zk_."
				primaryAction={create}
				filters={
					<>
						<AccessNote access={access}>Hanya owner atau admin yang bisa mengelola API key.</AccessNote>
						<SettingsFilters
							filters={filters}
							onFilter={page.onFilter}
							searchLabel="Cari nama atau awalan kunci"
							statusOptions={statusOptions}
							navigationKey={page.navigationKey}
						/>
					</>
				}
				items={keys}
				getKey={(key) => key.id}
				columns={[
					{
						id: "name",
						header: "Kunci",
						cell: (key) => (
							<span className="flex min-w-0 flex-col gap-0.5">
								<span className="flex flex-wrap items-center gap-2 font-medium">
									{key.name}
									{simulatedIds.has(key.id) && <Badge variant="info-light">Simulasi</Badge>}
								</span>
								<span className="font-mono text-xs text-muted-foreground wrap-anywhere">{key.prefix}…</span>
							</span>
						),
					},
					{
						id: "created",
						header: "Dibuat",
						inlineLabel: true,
						cell: (key) => <DateTime value={key.createdAt} format={dateOnly} className="whitespace-nowrap" />,
					},
					{
						id: "used",
						header: "Terakhir dipakai",
						inlineLabel: true,
						cell: (key) =>
							key.lastUsedAt ? <DateTime value={key.lastUsedAt} format={dateOnly} className="whitespace-nowrap" /> : "Belum pernah dipakai",
					},
					{
						id: "expires",
						header: "Kedaluwarsa",
						inlineLabel: true,
						cell: (key) =>
							key.expiresAt ? <DateTime value={key.expiresAt} format={dateOnly} className="whitespace-nowrap" /> : "Tanpa kedaluwarsa",
					},
					{ id: "status", header: "Status", cell: (key) => <StatusBadge resource="apiKey" status={key.status} /> },
					...(canManage
						? [
								{
									id: "actions",
									header: "Tindakan",
									slot: "actions" as const,
									cell: (key: ApiKey) =>
										key.status === "active" && !simulatedIds.has(key.id) ? (
											<Button
												type="button"
												variant="outline"
												size="sm"
												onClick={() => {
													setTarget(key);
													setDialog("revoke");
												}}
											>
												Cabut
											</Button>
										) : null,
								},
							]
						: []),
				]}
				pagination={page.pagination}
				noun="API key"
				state={page.state}
				filtered={Boolean(filters.q || filters.status)}
				emptyTitle="Belum ada API key"
				emptyDescription="Buat API key pertama untuk menghubungkan sistem Anda ke Zenith."
				emptyAction={create}
				errorMessage={page.errorMessage}
				simulated={page.simulated}
				loading={page.loading}
				onPage={page.onPage}
				onRetry={page.onRetry}
			/>
			{canManage && (
				<>
					<ActionDialog<V1CreateApiKeyResult>
						open={dialog === "create"}
						onOpenChange={close}
						title="Buat API key"
						description="Kunci baru berlaku segera. Nilainya hanya ditampilkan satu kali."
						confirmLabel="Buat API key"
						fields={<ApiKeyFields draft={draft} onDraft={setDraft} problems={problems} />}
						onSubmit={async () => {
							setShow(true);
							return onCreate(buildBody(() => createApiKeyBody(draft, now)));
						}}
						// The page gets the key's metadata only; the secret stays in this dialog's own result state.
						onResult={(result) => onResult(result.key, "created")}
						describeError={(failure) => (failure instanceof FormProblem ? failure.message : describeError(failure))}
						renderResult={(result) => (
							<SecretResult
								title={
									<>
										API key <strong className="wrap-anywhere">{result.key.name}</strong> dibuat.
									</>
								}
								label="API key"
								secret={result.secret}
							/>
						)}
					/>
					<ActionDialog<ApiKey>
						open={dialog === "revoke"}
						onOpenChange={close}
						title="Cabut API key?"
						description={`Kunci "${target?.name ?? ""}" (${target?.prefix ?? ""}…) langsung berhenti berlaku dan tidak bisa dipulihkan. Sistem yang memakainya akan ditolak.`}
						confirmLabel="Cabut API key"
						tone="destructive"
						onSubmit={() => onRevoke(target?.id ?? "")}
						onResult={(key) => onResult(key, "revoked")}
						describeError={describeError}
						renderResult={(key) => (
							<p className="text-sm">
								API key <strong className="wrap-anywhere">{key.name}</strong> dicabut.
							</p>
						)}
					/>
				</>
			)}
		</>
	);
}
