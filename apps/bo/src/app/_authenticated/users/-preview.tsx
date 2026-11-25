import { useState, type ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { EditorDrawer, Page } from "../../-components/common";
export function CustomerActionPreview({
	title,
	pageTitle = "Pengguna pelanggan",
	pending,
	children,
}: {
	title: string;
	pageTitle?: string;
	pending?: boolean;
	children: (onDirty: (value: boolean) => void, onSubmit: () => void) => ReactNode;
}) {
	const [open, setOpen] = useState(false);
	const [message, setMessage] = useState("");
	return (
		<Page title={pageTitle}>
			<Button onClick={() => setOpen(true)}>{title}</Button>
			{message && <p role="status">{message}</p>}
			<EditorDrawer title={title} open={open} pending={pending} onClose={() => setOpen(false)}>
				{(onDirty) =>
					children(onDirty, () => setMessage("Tindakan dikirim. Pratinjau lokal, bukan penyimpanan API. Kata sandi tidak ditampilkan."))
				}
			</EditorDrawer>
		</Page>
	);
}
