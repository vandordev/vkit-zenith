import { useEffect, useState } from "react";
import { useBlocker } from "@tanstack/react-router";
import { z } from "zod";
import { errorMessage, type BOTransport } from "./client";
import type { Feedback } from "../../app/-components/common";
import { useDraftState } from "../../app/-components/draft-state";
import { useInvalidate, useProtected } from "./session";
export const listSearch = z.object({
	q: z.string().max(100).catch(""),
	status: z.string().catch(""),
	page: z.coerce.number().int().min(1).max(1000000).catch(1),
	selected: z.string().uuid().optional(),
	create: z.boolean().optional(),
});
export function useWrite() {
	const draft = useDraftState();
	const { transport, writing } = useProtected();
	const invalidate = useInvalidate();
	const [pending, setPending] = useState(false);
	useBlocker({ shouldBlockFn: () => writing.current, enableBeforeUnload: () => writing.current });
	const [feedback, setFeedback] = useState<Feedback>();
	function run(write: (transport: BOTransport) => Promise<unknown>, done?: () => void) {
		if (writing.current) return;
		writing.current = true;
		setPending(true);
		setFeedback(undefined);
		return (async () => {
			try {
				await write(transport);
				if (draft) draft.current = false;
				await invalidate();
				setFeedback({ kind: "success", message: "Perubahan berhasil disimpan." });
				writing.current = false;
				done?.();
			} catch (error) {
				setFeedback({
					kind: "error",
					message:
						error instanceof Error && !("status" in error) && error.message.startsWith("Periksa") ? error.message : errorMessage(error),
				});
			} finally {
				writing.current = false;
				setPending(false);
			}
		})();
	}
	return { pending, feedback, setFeedback, run };
}
export function useUnsaved(dirty: boolean) {
	const draft = useDraftState();
	// The shell owns both navigation and logout confirmation. Register dedicated
	// workspaces there too, rather than protecting router navigation alone.
	useEffect(() => {
		if (draft) draft.current = dirty;
		return () => {
			if (draft) draft.current = false;
		};
	}, [draft, dirty]);
}
