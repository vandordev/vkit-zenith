import { useState } from "react";
import { BoringAvatar } from "@repo/ui/boring-avatar";
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { ErrorState, ErrorStateActions, ErrorStateContent, ErrorStateHeader, ErrorStateTitle } from "@repo/ui/error-state";
import { Loading } from "@repo/ui/loading";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import type { LogoResult, LogoUploadInputBody, V1OrganizationProfile } from "../../../../../../lib/api/generated/types.gen";
import { ActionDialog } from "../../../-components/action-dialog";
import { FieldProblem } from "../../../-components/field-problem";
import { buildBody, FormProblem } from "../../../-components/form-input";
import { SimulatedBanner } from "../../../-components/list-page";
import { StatusBadge } from "../../../-components/status-badge";
import { sizeText } from "../../../-components/format";
import { AccessNote, type WriteAccess } from "../../-components/settings-access";
import { logoBody, logoContentTypes, logoProblem, type LogoChoice } from "./logo-input";

export type ProfileViewProps = {
	state: "ready" | "pending" | "error";
	profile?: V1OrganizationProfile;
	access: WriteAccess;
	/** The dummy upload result, shown until navigation. It is metadata only: no file was uploaded and the preview does not change. */
	simulatedLogo?: LogoResult;
	errorMessage?: string;
	describeError: (failure: unknown) => string;
	onRetry: () => void;
	onUploadLogo: (body: LogoUploadInputBody) => Promise<LogoResult>;
	onLogoResult: (result: LogoResult) => void;
};

const logoLabels: Record<string, string> = { "image/png": "PNG", "image/jpeg": "JPEG", "image/svg+xml": "SVG" };

/** Read-only: name, slug and status are staff-owned; the object key is never exposed, only `hasLogo`. */
export function ProfileView({
	state,
	profile,
	access,
	simulatedLogo,
	errorMessage,
	describeError,
	onRetry,
	onUploadLogo,
	onLogoResult,
}: ProfileViewProps) {
	const [uploading, setUploading] = useState(false);
	const [choice, setChoice] = useState<LogoChoice>();
	const [show, setShow] = useState(false);
	const problem = show && choice ? logoProblem(choice) : show ? "Pilih berkas logo terlebih dulu." : undefined;
	return (
		<PageContainer size="md" className="flex min-w-0 flex-col gap-8">
			<PageHeader title="Profil & Branding" description="Identitas organisasi seperti yang dilihat pembeli saat membayar." />
			{state === "pending" && (
				<div className="flex min-h-32 items-center justify-center py-8" aria-busy="true">
					<Loading variant="skeleton" size={48} aria-label="Memuat profil" />
				</div>
			)}
			{state === "error" && (
				<ErrorState role="alert">
					<ErrorStateContent>
						<ErrorStateHeader>
							<ErrorStateTitle>{errorMessage ?? "Profil tidak dapat dimuat"}</ErrorStateTitle>
						</ErrorStateHeader>
						<ErrorStateActions>
							<Button type="button" variant="outline" onClick={onRetry}>
								Muat ulang
							</Button>
						</ErrorStateActions>
					</ErrorStateContent>
				</ErrorState>
			)}
			{state === "ready" && profile && (
				<>
					{simulatedLogo && <SimulatedBanner />}
					<dl className="grid gap-x-6 gap-y-4 sm:grid-cols-2">
						<div className="flex min-w-0 flex-col gap-1">
							<dt className="text-xs text-muted-foreground">Nama organisasi</dt>
							<dd className="min-w-0 text-sm font-medium wrap-anywhere">{profile.name}</dd>
						</div>
						<div className="flex min-w-0 flex-col gap-1">
							<dt className="text-xs text-muted-foreground">Slug</dt>
							<dd className="min-w-0 text-sm wrap-anywhere">{profile.slug}</dd>
						</div>
						<div className="flex min-w-0 flex-col gap-1">
							<dt className="text-xs text-muted-foreground">Status</dt>
							<dd>
								<StatusBadge resource="organization" status={profile.status} />
							</dd>
						</div>
					</dl>
					<p className="text-sm text-muted-foreground">Nama dan slug dikelola oleh tim Zenith. Hubungi kami untuk mengubahnya.</p>
					<section aria-labelledby="logo-heading" className="flex flex-col gap-3">
						<h2 id="logo-heading" className="text-lg font-semibold">
							Logo
						</h2>
						<div className="flex min-w-0 items-center gap-4">
							{/* Real logo rendering and upload arrive with M4 (presigned storage): a generated marble from the stable id is the fallback. */}
							<BoringAvatar name={profile.id} variant="marble" size={56} square className="overflow-hidden rounded-lg" />
							<p className="text-sm text-muted-foreground">
								{profile.hasLogo ? "Logo tersimpan, tampilan logo tersedia segera." : "Belum ada logo. Avatar otomatis dipakai sementara."}
							</p>
						</div>
						{simulatedLogo && (
							<p role="status" className="text-sm">
								Logo {logoLabels[simulatedLogo.contentType] ?? simulatedLogo.contentType} ({sizeText(simulatedLogo.size)}) dicatat sebagai
								simulasi. Tidak ada berkas yang diunggah, jadi avatar otomatis tetap dipakai.
							</p>
						)}
						<AccessNote access={access}>Hanya owner yang bisa mengubah logo.</AccessNote>
						{access === "allowed" && (
							<Button type="button" variant="outline" className="self-start" onClick={() => setUploading(true)}>
								Unggah logo
							</Button>
						)}
					</section>
					<section aria-labelledby="preview-heading" className="flex flex-col gap-3">
						<h2 id="preview-heading" className="text-lg font-semibold">
							Pratinjau identitas
						</h2>
						<div className="flex min-w-0 items-center gap-3 rounded-lg border border-border bg-card p-4">
							<BoringAvatar name={profile.id} variant="marble" size={36} square className="overflow-hidden rounded-md" />
							<span className="min-w-0 truncate font-medium">{profile.name}</span>
						</div>
					</section>
				</>
			)}
			{access === "allowed" && (
				<ActionDialog<LogoResult>
					open={uploading}
					onOpenChange={(open) => {
						setUploading(open);
						if (!open) {
							setChoice(undefined);
							setShow(false);
						}
					}}
					title="Unggah logo"
					description="Simulasi: hanya jenis dan ukuran berkas yang dicatat. Berkasnya sendiri tidak diunggah atau disimpan."
					confirmLabel="Unggah logo"
					fields={
						<div className="flex min-w-0 flex-col gap-1.5">
							<Input
								labelStyle="static"
								label="Berkas logo"
								type="file"
								accept={logoContentTypes.join(",")}
								aria-invalid={problem ? true : undefined}
								aria-describedby="logo-problem"
								onChange={(event) => {
									const file = event.target.files?.[0];
									setChoice(file && { contentType: file.type, size: file.size });
									setShow(false);
								}}
							/>
							<FieldProblem id="logo-problem">{problem}</FieldProblem>
							<p className="text-xs text-muted-foreground">PNG, JPEG atau SVG, maksimal 1 MB.</p>
						</div>
					}
					onSubmit={async () => {
						setShow(true);
						return onUploadLogo(
							buildBody(() => {
								const body = choice && logoBody(choice);
								if (!body) throw new FormProblem("Pilih berkas logo yang valid.");
								return body;
							}),
						);
					}}
					onResult={onLogoResult}
					describeError={(failure) => (failure instanceof FormProblem ? failure.message : describeError(failure))}
					renderResult={(result) => (
						<p className="text-sm">
							Logo {logoLabels[result.contentType] ?? result.contentType} ({sizeText(result.size)}) dicatat. Tidak ada berkas yang diunggah.
						</p>
					)}
				/>
			)}
		</PageContainer>
	);
}
