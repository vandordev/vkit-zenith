import { useEffect, useRef, useState, type ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { InputPassword } from "@repo/ui/input-password";
import { PageContainer } from "@repo/ui/page-container";
import { PageHeader } from "@repo/ui/page-header";
import { Select } from "@repo/ui/select";
import { Drawer, DrawerBody } from "@repo/ui/drawer";
import { useDraftState } from "./draft-state";

export type Feedback = { kind: "error" | "success"; message: string } | undefined;
export function Notice({ feedback }: { feedback?: Feedback }) {
	const ref = useRef<HTMLParagraphElement>(null);
	useEffect(() => {
		if (feedback?.kind === "error") ref.current?.focus();
	}, [feedback]);
	return feedback ? (
		<p
			ref={ref}
			tabIndex={-1}
			role={feedback.kind === "error" ? "alert" : "status"}
			className="rounded-md border border-border bg-muted p-3 text-sm leading-relaxed"
		>
			{feedback.message}
		</p>
	) : null;
}
export function Page({
	title,
	description,
	action,
	children,
}: {
	title: string;
	description?: string;
	action?: ReactNode;
	children: ReactNode;
}) {
	return (
		<PageContainer className="flex flex-col gap-6">
			<PageHeader title={title} description={description} actions={action} />
			{children}
		</PageContainer>
	);
}
export type Field = {
	name: string;
	label: string;
	type?: "text" | "email" | "password";
	help?: string;
	options?: readonly { value: string; label: string }[];
	readOnly?: boolean;
	required?: boolean;
	maxLength?: number;
	autoComplete?: string;
};
export function EditForm({
	fields,
	initial = {},
	pending,
	submitDisabled,
	feedback,
	submitLabel = "Simpan perubahan",
	onSubmit,
	children,
	onDirty,
	labelStyle = "floating",
}: {
	fields: readonly Field[];
	initial?: Record<string, string>;
	pending?: boolean;
	submitDisabled?: boolean;
	feedback?: Feedback;
	submitLabel?: string;
	onSubmit: (values: Record<string, string>) => void;
	children?: ReactNode;
	onDirty?: (dirty: boolean) => void;
	labelStyle?: "static" | "floating";
}) {
	const [values, setValues] = useState(initial);
	const form = useRef<HTMLFormElement>(null);
	function change(name: string, value: string) {
		setValues((current) => ({ ...current, [name]: value }));
		onDirty?.(true);
	}
	return (
		<form
			ref={form}
			className="flex flex-col gap-5"
			onSubmit={(event) => {
				event.preventDefault();
				if (pending || submitDisabled) return;
				onSubmit(values);
				// Never retain submitted passwords as recoverable draft data.
				setValues((current) =>
					Object.fromEntries(
						Object.entries(current).map(([key, value]) => [
							key,
							fields.some((field) => field.name === key && field.type === "password") ? "" : value,
						]),
					),
				);
			}}
		>
			<Notice feedback={feedback} />
			{fields.map((field) => (
				<div key={field.name} className="flex min-w-0 flex-col gap-2">
					{field.options ? (
						<>
							{labelStyle === "static" && (
								<span id={`label-${field.name}`} className="text-sm font-medium">
									{field.label}
								</span>
							)}
							<Select
								aria-labelledby={labelStyle === "static" ? `label-${field.name}` : undefined}
								label={labelStyle === "floating" ? field.label : undefined}
								labelStyle={labelStyle}
								value={values[field.name] ?? ""}
								onValueChange={(value) => change(field.name, value ?? "")}
								disabled={pending || field.readOnly}
								data={field.options}
								className="w-full"
							/>
						</>
					) : field.type === "password" ? (
						<InputPassword
							label={field.label}
							labelStyle={labelStyle}
							name={field.name}
							value={values[field.name] ?? ""}
							onChange={(event) => change(field.name, event.target.value)}
							required={field.required ?? true}
							maxLength={field.maxLength ?? 512}
							readOnly={field.readOnly}
							disabled={pending}
							autoComplete={field.autoComplete ?? "new-password"}
							aria-describedby={field.help ? `help-${field.name}` : undefined}
							showPasswordLabel="Tampilkan kata sandi"
							hidePasswordLabel="Sembunyikan kata sandi"
						/>
					) : (
						<Input
							label={field.label}
							labelStyle={labelStyle}
							name={field.name}
							type={field.type ?? "text"}
							value={values[field.name] ?? ""}
							onChange={(event) => change(field.name, event.target.value)}
							required={field.required ?? true}
							maxLength={field.maxLength ?? 200}
							readOnly={field.readOnly}
							disabled={pending}
							autoComplete={field.autoComplete ?? "off"}
							aria-describedby={field.help ? `help-${field.name}` : undefined}
						/>
					)}
					{field.help && (
						<p id={`help-${field.name}`} className="text-xs leading-relaxed text-muted-foreground">
							{field.help}
						</p>
					)}
				</div>
			))}
			{children}
			<Button type="submit" isLoading={pending} disabled={submitDisabled}>
				{submitLabel}
			</Button>
		</form>
	);
}
export function EditorDrawer({
	title,
	children,
	onClose,
	pending,
	open = true,
	editorKey = "editor",
}: {
	title: string;
	children: (onDirty: (value: boolean) => void) => ReactNode;
	onClose: () => void;
	pending?: boolean;
	open?: boolean;
	editorKey?: string;
}) {
	const [dirty, setDirty] = useState(false);
	const retained = useRef<{ title: string; children: typeof children; editorKey: string } | undefined>(undefined);
	if (open) retained.current = { title, children, editorKey };
	const content = retained.current;
	const draft = useDraftState();
	useEffect(() => {
		if (open) setDirty(false);
	}, [open, editorKey]);
	useEffect(() => {
		if (draft) draft.current = open && dirty;
		return () => {
			if (draft) draft.current = false;
		};
	}, [draft, dirty, open]);
	return (
		<Drawer
			title={content?.title ?? title}
			open={open}
			onOpenChangeComplete={(nextOpen) => {
				if (!nextOpen) retained.current = undefined;
			}}
			onOpenChange={(nextOpen, details) => {
				if (!nextOpen && (pending || (dirty && !window.confirm("Buang perubahan yang belum disimpan?")))) {
					details.cancel();
					return;
				}
				if (!nextOpen) {
					if (draft) draft.current = false;
					onClose();
				}
			}}
			contentProps={{ closeButtonLabel: "Tutup editor" }}
		>
			<DrawerBody key={content?.editorKey}>{content?.children(setDirty)}</DrawerBody>
		</Drawer>
	);
}
export const labels: Record<string, string> = {
	active: "Aktif",
	inactive: "Nonaktif",
	disabled: "Dinonaktifkan",
	admin: "Admin",
	superadmin: "Superadmin",
};
