import { useState, type ReactNode } from "react";
import { Autocomplete } from "@repo/ui/autocomplete";
import { Select } from "@repo/ui/select";
import { Button } from "@repo/ui/button";
import { Notice, type Feedback } from "../../../../../-components/common";
import { ContentState } from "../../../../../-components/content-state";
import type { CustomerOwnerChoices } from "../../../-components/organization-create";
import type {
	V1OrganizationMembershipGetBody,
	V1OrganizationMemberAddInputBody,
	V1OrganizationMemberRemoveInputBody,
	V1OrganizationMemberRestoreInputBody,
} from "@bo/lib/api/generated/types.gen";

export function MembershipAdd({
	choices,
	pending,
	feedback,
	onDirty,
	onSubmit,
}: {
	choices: CustomerOwnerChoices;
	pending?: boolean;
	feedback?: Feedback;
	onDirty: (value: boolean) => void;
	onSubmit: (body: V1OrganizationMemberAddInputBody) => void;
}) {
	const [role, setRole] = useState<"admin" | "member">("member");
	return (
		<form
			className="space-y-5"
			onSubmit={(e) => {
				e.preventDefault();
				if (!pending && !choices.loading && !choices.error && choices.value?.eligible) onSubmit({ userId: choices.value.id, role });
			}}
		>
			<Notice feedback={feedback} />
			<p className="max-w-[65ch] text-sm text-muted-foreground">
				Tambahkan pelanggan aktif sebagai admin atau anggota. Keanggotaan yang dihapus harus dipulihkan, bukan dibuat ulang. Tindakan ini
				tidak memberikan kepemilikan.
			</p>
			<Select
				label="Peran"
				value={role}
				data={[
					{ value: "member", label: "Anggota" },
					{ value: "admin", label: "Admin" },
				]}
				disabled={pending}
				onValueChange={(value) => {
					if (value === "admin" || value === "member") {
						setRole(value);
						onDirty(true);
					}
				}}
			/>
			<Autocomplete
				label="Pelanggan anggota"
				labelStyle="floating"
				mode="selection"
				items={choices.data?.items ?? []}
				value={choices.value}
				inputValue={choices.q}
				onInputValueChange={choices.onSearch}
				getItemLabel={(item) => `${item.displayName} · ${item.email}`}
				getItemValue={(item) => item.id}
				filter={null}
				isItemDisabled={(item) => !item.eligible}
				onValueChange={(value) => {
					choices.onSelected(value);
					onDirty(true);
				}}
				disabled={pending}
				loading={choices.loading}
				error={choices.error}
				onRetry={choices.onRetry}
				retryLabel="Muat ulang"
				emptyMessage="Tidak ada pelanggan yang dapat ditambahkan"
				loadingMessage="Memuat pelanggan…"
				contentProps={{ className: "z-[100]" }}
				clearable
				clearLabel="Kosongkan pelanggan"
				renderItem={(item) => (
					<div className="min-w-0">
						<p className="break-words">{item.displayName}</p>
						<p className="break-all text-xs text-muted-foreground">{item.email}</p>
						{!item.eligible && <p className="text-xs">{item.disabledReason}</p>}
					</div>
				)}
			/>
			{choices.value && (
				<p className="break-words text-sm">
					Terpilih: {choices.value.displayName} <span className="break-all">({choices.value.email})</span>
				</p>
			)}
			{choices.value && !choices.value.eligible && <Notice feedback={{ kind: "error", message: choices.value.disabledReason }} />}
			{choices.error && <Notice feedback={{ kind: "error", message: choices.error }} />}
			<div className="flex flex-wrap items-center gap-2">
				<Button
					type="button"
					variant="ghost"
					disabled={pending || choices.loading || choices.page <= 1}
					onClick={() => choices.onPage(choices.page - 1)}
				>
					Pilihan sebelumnya
				</Button>
				<span className="text-xs">Halaman {choices.page}</span>
				<Button
					type="button"
					variant="ghost"
					disabled={
						pending ||
						choices.loading ||
						!choices.data ||
						BigInt(choices.page) * BigInt(choices.data.pagination.pageSize) >= BigInt(choices.data.pagination.total)
					}
					onClick={() => choices.onPage(choices.page + 1)}
				>
					Pilihan berikutnya
				</Button>
			</div>
			<Button type="submit" isLoading={pending} disabled={!choices.value?.eligible || choices.loading || !!choices.error}>
				Tambah anggota
			</Button>
		</form>
	);
}

export function MembershipEditor({
	record,
	loading,
	error,
	feedback,
	pending,
	userLink,
	onRetry,
	onRemove,
	onRestore,
	onOwner,
}: {
	record?: V1OrganizationMembershipGetBody;
	loading?: boolean;
	error?: string;
	feedback?: Feedback;
	pending?: boolean;
	userLink?: ReactNode;
	onRetry: () => void;
	onRemove: (body: V1OrganizationMemberRemoveInputBody) => void;
	onRestore: (body: V1OrganizationMemberRestoreInputBody) => void;
	onOwner: () => void;
}) {
	return (
		<div className="space-y-5">
			<Notice feedback={feedback} />
			{error && <ContentState kind="error" title={error} description="Muat ulang detail sebelum melakukan tindakan." onRetry={onRetry} />}
			{!record ? (
				!error && <ContentState kind="loading" title="Memuat keanggotaan…" />
			) : (
				<>
					<div className="space-y-1">
						<h2 className="break-words text-lg font-semibold">{record.userName}</h2>
						<p className="break-all text-sm text-muted-foreground">{record.userEmail}</p>
						{userLink}
					</div>
					<dl className="grid grid-cols-2 gap-3 text-sm">
						<dt>Pengguna global</dt>
						<dd>{record.userStatus === "active" ? "Aktif" : "Dinonaktifkan"}</dd>
						<dt>Organisasi</dt>
						<dd className="break-words">
							{record.organizationName} ({record.organizationStatus === "active" ? "aktif" : "ditangguhkan"})
						</dd>
						<dt>Keanggotaan</dt>
						<dd>{record.status === "active" ? "Aktif" : "Dihapus"}</dd>
						<dt>Peran</dt>
						<dd>{record.role === "owner" ? "Pemilik" : record.role === "admin" ? "Admin" : "Anggota"}</dd>
					</dl>
					<section className="space-y-3 border-t border-border pt-4" aria-label="Tindakan keanggotaan">
						{record.status === "active" ? (
							<>
								<Button
									variant="outline"
									disabled={pending || loading || !!error || !record.canRemove}
									onClick={() => onRemove({ revision: record.revision })}
								>
									Hapus keanggotaan
								</Button>
								{!record.canRemove && <p className="text-sm text-muted-foreground">{record.removeBlockedReason}</p>}
							</>
						) : (
							<>
								<Button
									variant="outline"
									disabled={pending || loading || !!error || !record.canRestore}
									onClick={() => onRestore({ revision: record.revision })}
								>
									Pulihkan keanggotaan
								</Button>
								{!record.canRestore && <p className="text-sm text-muted-foreground">{record.restoreBlockedReason}</p>}
							</>
						)}
						<p className="max-w-[65ch] text-xs text-muted-foreground">
							Penghapusan tidak menghapus pelanggan. Pemulihan memakai keanggotaan yang sama tanpa memulihkan kepemilikan.
						</p>
					</section>
					<section className="space-y-3 border-t border-border pt-4" aria-label="Tindakan kepemilikan">
						<Button variant="outline" disabled={pending || loading || !!error || !record.canChangeRole} onClick={onOwner}>
							{"Ubah peran"}
						</Button>
						{!record.canChangeRole && <p className="text-sm text-muted-foreground">{record.roleBlockedReason}</p>}
						<p className="max-w-[65ch] text-xs text-muted-foreground">
							Cabut kepemilikan melalui tindakan terpisah sebelum menghapus anggota. Pencabutan tidak menghapus keanggotaan biasa.
						</p>
					</section>
				</>
			)}
		</div>
	);
}
