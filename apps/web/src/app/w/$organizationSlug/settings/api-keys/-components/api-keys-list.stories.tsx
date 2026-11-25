import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { Button } from "@repo/ui/button";
import { ActionDialogView } from "../../../-components/action-dialog";
import {
	applySimulation,
	createdIds,
	noSimulation,
	simulateChanged,
	simulateCreated,
	type Simulation,
} from "../../../-components/simulated";
import { delayed, describeError, failing, noop, pageOf } from "../../../-components/-story-fixtures";
import { organization } from "../../../-components/-fixtures";
import { ShellPreview } from "../../../-components/-shell-preview";
import { SecretResult } from "../../-components/secret-result";
import { matchesSettingsFilter, searchFields } from "../../-components/settings-search";
import { apiKeys, createdApiKey, dummySecret, longApiKey } from "../../-components/-settings-fixtures";
import type { ApiKey } from "../../../../../../lib/api/generated/types.gen";
import { apiKeyProblems, emptyApiKeyDraft, type ApiKeyDraft } from "./api-key-input";
import { ApiKeyFields, ApiKeysList } from "./api-keys-list";

const path = "/w/kopi-senja/settings/api-keys";

const meta = {
	title: "Web/Pages/Dashboard/Settings/ApiKeys",
	component: ApiKeysList,
	parameters: { layout: "fullscreen", routerInitialPath: path },
	args: {
		state: "ready",
		keys: apiKeys,
		pagination: pageOf(apiKeys.length),
		filters: { q: "", status: "" },
		navigationKey: "1",
		access: "allowed",
		simulatedIds: new Set<string>(),
		describeError,
		onFilter: noop,
		onPage: noop,
		onRetry: noop,
		onCreate: () => delayed(createdApiKey()),
		onRevoke: () => delayed({ ...apiKeys[0]!, status: "revoked", revokedAt: "2026-11-19T00:00:00Z" }),
		onResult: noop,
	},
	argTypes: { keys: { control: false }, pagination: { control: false }, filters: { control: false } },
} satisfies Meta<typeof ApiKeysList>;
export default meta;
type Story = StoryObj<typeof meta>;

/** Owner: active (used and never used), expired and revoked keys; only the prefix is ever shown. */
export const Owner: Story = {};
/** Admin manages developer keys too (same view as the owner). */
export const Admin: Story = {};
export const Member: Story = { args: { access: "no-permission" } };
export const Suspended: Story = { args: { access: "suspended" } };
export const NoMatches: Story = { args: { keys: [], pagination: pageOf(0), filters: { q: "zzz", status: "expired" } } };
export const OutOfRange: Story = { args: { keys: [], pagination: pageOf(9, 3) } };
export const LongNames: Story = { args: { keys: [longApiKey], pagination: pageOf(1) } };
export const Pending: Story = { args: { state: "pending", keys: [], pagination: undefined } };
export const Error: Story = {
	args: { state: "error", keys: [], pagination: undefined, errorMessage: "Layanan sedang tidak tersedia. Coba lagi sebentar lagi." },
};
export const WritesFail: Story = { args: { onCreate: () => failing(422), onRevoke: () => failing(409) } };
export const Mobile: Story = { globals: { viewport: { value: "mobile1", isRotated: false } } };
export const InShell: Story = {
	render: (args) => (
		<ShellPreview org={organization("admin")}>
			<ApiKeysList {...args} />
		</ShellPreview>
	),
};
/** Real page state: a created key is marked and cannot be revoked; a revoked one flips in place, under the Aktif filter it leaves the list. */
export const SimulatedFlow: Story = {
	render: (args) => {
		const [simulation, setSimulation] = useState<Simulation<ApiKey>>(noSimulation);
		return (
			<ApiKeysList
				{...args}
				keys={applySimulation(apiKeys, simulation, 1, () => true)}
				simulatedIds={createdIds(simulation)}
				simulated={simulation.created.length > 0 || Object.keys(simulation.changed).length > 0}
				onRevoke={(id) => delayed({ ...apiKeys.find((key) => key.id === id)!, status: "revoked", revokedAt: "2026-11-19T00:00:00Z" })}
				onResult={(key, kind) =>
					setSimulation((current) => (kind === "created" ? simulateCreated(current, key) : simulateChanged(current, key)))
				}
			/>
		);
	},
};

/**
 * Secret lifecycle with the real list: the secret exists only in the open creating dialog. Closing it, switching organization, and a
 * new identity (the buttons remount the page exactly as the route's key does) leave no trace of it in the page; the table only ever gets
 * the key's metadata. Create a key, then try each button and search the page text for "DUMMY".
 */
function SecretLifecycle() {
	const [org, setOrg] = useState(1);
	const [epoch, setEpoch] = useState(1);
	const [simulation, setSimulation] = useState<Simulation<ApiKey>>(noSimulation);
	return (
		<div className="flex flex-col gap-4">
			<div className="flex flex-wrap gap-2 px-6 pt-4">
				<Button type="button" variant="outline" onClick={() => setOrg(org + 1)}>
					Ganti organisasi ({org})
				</Button>
				<Button type="button" variant="outline" onClick={() => setEpoch(epoch + 1)}>
					Identitas baru ({epoch})
				</Button>
			</div>
			<ApiKeysList
				key={`${epoch}:${org}`}
				state="ready"
				keys={applySimulation([], noSimulation, 1, () => true)}
				pagination={pageOf(0)}
				filters={{ q: "", status: "" }}
				navigationKey="1"
				access="allowed"
				simulatedIds={createdIds(noSimulation)}
				describeError={describeError}
				onFilter={noop}
				onPage={noop}
				onRetry={noop}
				onCreate={() => delayed(createdApiKey())}
				onRevoke={() => delayed(apiKeys[0]!)}
				onResult={(key, kind) =>
					setSimulation((current) => (kind === "created" ? simulateCreated(current, key) : simulateChanged(current, key)))
				}
			/>
			<p className="px-6 pb-4 text-xs text-muted-foreground" data-harness="page-created">
				Kunci tercatat di halaman: {simulation.created.map((key) => `${key.name} (${key.prefix})`).join(", ") || "belum ada"}
			</p>
		</div>
	);
}
export const SecretLifecycle_: Story = { name: "SecretLifecycle", render: () => <SecretLifecycle /> };

// Dialog states with the real fields; the live lifecycle is the harness above.
function Dialog({
	initial,
	show = false,
	...state
}: { initial: ApiKeyDraft; show?: boolean } & Partial<Parameters<typeof ActionDialogView>[0]>) {
	const [draft, setDraft] = useState(initial);
	return (
		<ActionDialogView
			open
			onOpenChange={noop}
			reason=""
			onReasonChange={noop}
			onConfirm={noop}
			title="Buat API key"
			description="Kunci baru berlaku segera. Nilainya hanya ditampilkan satu kali."
			confirmLabel="Buat API key"
			showValidation={false}
			phase="idle"
			fields={
				<ApiKeyFields
					draft={draft}
					onDraft={setDraft}
					problems={show ? apiKeyProblems(draft, new Date(2026, 10, 19)) : {}}
					disabled={state.phase === "submitting"}
				/>
			}
			{...state}
		/>
	);
}
export const DialogOpen: Story = { render: () => <Dialog initial={emptyApiKeyDraft} />, parameters: { layout: "centered" } };
export const DialogInvalid: Story = {
	render: () => <Dialog show initial={{ name: " ", date: "2026-11-01" }} phase="failed" error="Periksa kembali isian yang ditandai." />,
	parameters: { layout: "centered" },
};
export const DialogSubmitting: Story = {
	render: () => <Dialog initial={{ name: "Kunci integrasi", date: "2026-12-31" }} phase="submitting" />,
	parameters: { layout: "centered" },
};
export const DialogFailure: Story = {
	render: () => (
		<Dialog initial={{ name: "Kunci integrasi", date: "" }} phase="failed" error="Peran Anda tidak memiliki izin untuk tindakan ini." />
	),
	parameters: { layout: "centered" },
};
const secretDialog = () => (
	<Dialog
		initial={emptyApiKeyDraft}
		phase="success"
		result={
			<SecretResult
				title={
					<>
						API key <strong>Kunci integrasi baru</strong> dibuat.
					</>
				}
				label="API key"
				secret={dummySecret}
			/>
		}
	/>
);
/** The one-time secret (the API's unusable constant), copy control and the warning. */
export const Secret: Story = { render: () => secretDialog(), parameters: { layout: "centered" } };

/**
 * Revoke, then filter: the "server" is a fixture-backed handler that filters the stored keys by status and text exactly like the list
 * handler; the page overlays dummy results with the shared helper. Revoke an active key (unfiltered list), then choose the Dicabut filter:
 * the handler still stores the key as active, yet the revoked row appears first with the simulation notice. Choosing Aktif again drops it.
 */
export const RevokeThenRevokedFilter: Story = {
	render: (args) => {
		const [filters, setFilters] = useState({ q: "", status: "" });
		const [simulation, setSimulation] = useState<Simulation<ApiKey>>(noSimulation);
		const search = { q: filters.q || undefined, status: filters.status || undefined };
		const accepts = (item: ApiKey) => matchesSettingsFilter(search, item, searchFields.apiKey(item));
		const server = apiKeys.filter(accepts);
		return (
			<ApiKeysList
				{...args}
				filters={filters}
				keys={applySimulation(server, simulation, 1, accepts, search.status !== undefined)}
				pagination={pageOf(server.length)}
				simulatedIds={createdIds(simulation)}
				simulated={simulation.created.length > 0 || Object.keys(simulation.changed).length > 0}
				onFilter={setFilters}
				onRevoke={(id) => delayed({ ...apiKeys.find((key) => key.id === id)!, status: "revoked", revokedAt: "2026-11-19T00:00:00Z" })}
				onResult={(key, kind) =>
					setSimulation((current) => (kind === "created" ? simulateCreated(current, key) : simulateChanged(current, key)))
				}
			/>
		);
	},
};
