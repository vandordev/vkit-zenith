import { useEffect, useRef, useState } from "react";
import type { FormStatus } from "./form-status";

export function statusMessage(status: FormStatus): string | undefined {
	switch (status.kind) {
		case "error":
			return status.message;
		case "rate-limited":
			return status.retryAfterSeconds
				? `Terlalu banyak percobaan. Coba lagi dalam ${status.retryAfterSeconds} detik.`
				: "Terlalu banyak percobaan. Coba lagi beberapa saat lagi.";
		case "unavailable":
			return "Layanan tidak tersedia. Coba lagi nanti.";
		default:
			return undefined;
	}
}

export function StatusNotice({ status }: { status: FormStatus }) {
	const ref = useRef<HTMLParagraphElement>(null);
	const message = statusMessage(status);
	useEffect(() => {
		if (message) ref.current?.focus();
	}, [message]);
	return message ? (
		<p ref={ref} tabIndex={-1} role="alert" className="rounded-md border border-border bg-muted p-3 text-sm leading-relaxed">
			{message}
		</p>
	) : null;
}

export function fieldError(status: FormStatus, name: string): string | undefined {
	return status.kind === "error" ? status.fields?.[name] : undefined;
}

export function FieldError({ id, message }: { id: string; message?: string }) {
	return message ? (
		<p id={id} className="text-xs leading-relaxed text-destructive">
			{message}
		</p>
	) : null;
}

/** Secret input state owned by the form: cleared whenever a pending submit resolves. */
export function useSecretField(status: FormStatus) {
	const [value, setValue] = useState("");
	const previous = useRef(status.kind);
	useEffect(() => {
		if (previous.current === "pending" && status.kind !== "pending") setValue("");
		previous.current = status.kind;
	}, [status.kind]);
	return [value, setValue] as const;
}
