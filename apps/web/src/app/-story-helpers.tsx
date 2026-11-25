import { useState } from "react";
import type { FormStatus } from "./-components/form-status";

export const noop = () => {};

/** Local preview only: submit goes pending, then settles on `result`. Never calls an API. */
export function useSubmitCycle(result: FormStatus) {
	const [status, setStatus] = useState<FormStatus>({ kind: "idle" });
	const submit = () => {
		setStatus({ kind: "pending" });
		setTimeout(() => setStatus(result), 600);
	};
	return [status, submit] as const;
}
