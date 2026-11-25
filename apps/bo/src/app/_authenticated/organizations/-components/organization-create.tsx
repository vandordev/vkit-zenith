import { useId, useState } from "react";
import { Autocomplete } from "@repo/ui/autocomplete";
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { Select } from "@repo/ui/select";
import { Notice, type Feedback } from "../../../-components/common";
import type { V1OrganizationCreateInputBody, V1UserChoiceListBody, V1UserChoiceListItem } from "@bo/lib/api/generated/types.gen";

export type CustomerOwnerChoices = {
	data?: V1UserChoiceListBody;
	value: V1UserChoiceListItem | null;
	q: string;
	page: number;
	loading?: boolean;
	error?: string;
	onSelected: (value: V1UserChoiceListItem | null) => void;
	onSearch: (q: string) => void;
	onPage: (page: number) => void;
	onRetry: () => void;
};

export function OrganizationCreate({
	choices,
	pending,
	feedback,
	initialMode = "existing",
	onDirty,
	onSubmit,
}: {
	choices: CustomerOwnerChoices;
	pending?: boolean;
	feedback?: Feedback;
	initialMode?: "existing" | "new";
	onDirty: (dirty: boolean) => void;
	onSubmit: (body: V1OrganizationCreateInputBody) => void;
}) {
	const [name, setName] = useState("");
	const [slug, setSlug] = useState("");
	const [mode, setMode] = useState(initialMode);
	const [displayName, setDisplayName] = useState("");
	const [email, setEmail] = useState("");
	const help = useId();
	return (
		<form
			className="flex min-w-0 flex-col gap-5"
			onSubmit={(event) => {
				event.preventDefault();
				if (pending || (mode === "existing" && (!choices.value?.eligible || choices.loading || choices.error))) return;
				onSubmit(mode === "new" ? { name, slug, newOwner: { displayName, email } } : { name, slug, ownerUserId: choices.value?.id });
			}}
		>
			<Notice feedback={feedback} />
			<Input
				label="Nama organisasi"
				name="name"
				value={name}
				maxLength={200}
				required
				disabled={pending}
				onChange={(event) => {
					setName(event.target.value);
					onDirty(true);
				}}
			/>
			<Input
				label="Slug organisasi"
				name="slug"
				value={slug}
				maxLength={64}
				pattern="[a-z0-9]+(-[a-z0-9]+)*"
				required
				disabled={pending}
				autoComplete="off"
				aria-describedby={`${help}-slug`}
				onChange={(event) => {
					setSlug(event.target.value);
					onDirty(true);
				}}
			/>
			<p id={`${help}-slug`} className="text-xs text-muted-foreground">
				Huruf kecil, angka dan tanda hubung. Slug harus unik.
			</p>
			<div className="space-y-5 border-t border-border pt-5">
				<h2 className="font-semibold">Pemilik awal</h2>
				<Select
					label="Sumber pemilik"
					value={mode}
					disabled={pending}
					data={[
						{ value: "existing", label: "Pilih pelanggan yang sudah ada" },
						{ value: "new", label: "Buat pelanggan baru" },
					]}
					onValueChange={(next) => {
						if ((next !== "existing" && next !== "new") || next === mode || pending) return;
						setMode(next);
						setDisplayName("");
						setEmail("");
						choices.onSelected(null);
						choices.onSearch("");
						onDirty(true);
					}}
				/>
				{mode === "existing" ? (
					<div className="space-y-3">
						<Autocomplete
							label="Pelanggan pemilik"
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
							disabled={pending}
							onValueChange={(value) => {
								choices.onSelected(value);
								onDirty(true);
							}}
							loading={choices.loading}
							error={choices.error}
							onRetry={choices.onRetry}
							retryLabel="Muat ulang"
							loadingMessage="Memuat pelanggan…"
							emptyMessage="Tidak ada pelanggan yang cocok"
							clearable
							clearLabel="Kosongkan pelanggan"
							contentProps={{ className: "z-[100]" }}
							inputProps={{ maxLength: 100, autoComplete: "off" }}
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
								Terpilih: {choices.value.displayName} <span className="break-all text-muted-foreground">({choices.value.email})</span>
							</p>
						)}
						{choices.value && !choices.value.eligible && (
							<Notice
								feedback={{ kind: "error", message: choices.value.disabledReason || "Pelanggan ini tidak dapat menjadi pemilik awal." }}
							/>
						)}
						{choices.error && <Notice feedback={{ kind: "error", message: choices.error }} />}
						<div className="flex flex-wrap items-center gap-2 text-xs">
							<Button
								type="button"
								variant="ghost"
								size="sm"
								disabled={pending || choices.loading || choices.page <= 1}
								onClick={() => choices.onPage(choices.page - 1)}
							>
								Pilihan sebelumnya
							</Button>
							<span>Halaman {choices.page}</span>
							<Button
								type="button"
								variant="ghost"
								size="sm"
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
					</div>
				) : (
					<div className="space-y-5">
						<Input
							label="Nama pelanggan baru"
							name="displayName"
							value={displayName}
							required
							maxLength={200}
							disabled={pending}
							onChange={(event) => {
								setDisplayName(event.target.value);
								onDirty(true);
							}}
						/>
						<Input
							label="Email pelanggan baru"
							name="email"
							type="email"
							value={email}
							required
							maxLength={254}
							disabled={pending}
							onChange={(event) => {
								setEmail(event.target.value);
								onDirty(true);
							}}
						/>
					</div>
				)}
				<p className="max-w-[65ch] text-sm leading-relaxed text-muted-foreground">
					Organisasi dan keanggotaan pemilik dibuat bersama. Pelanggan baru berstatus aktif dengan email belum terverifikasi, tanpa kata
					sandi atau akses masuk. Memilih pelanggan lama tidak mengubah verifikasi emailnya.
				</p>
			</div>
			<Button
				type="submit"
				isLoading={pending}
				disabled={mode === "existing" && (!choices.value?.eligible || choices.loading || !!choices.error)}
			>
				Buat organisasi
			</Button>
		</form>
	);
}
