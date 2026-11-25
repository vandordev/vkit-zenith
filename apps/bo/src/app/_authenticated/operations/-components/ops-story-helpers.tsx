import { useState, type ComponentType } from "react";
import { opsHistory } from "../../-global-fixtures";
import { noop } from "../../../-fixtures";

/** Detail story props shared by operations screens. */
export const detailArgs = <T extends { id: string }>(item: T, targetType: string) => ({
	item,
	history: opsHistory(targetType, item.id),
	state: "ready" as const,
	onRetry: noop,
});

/** Wraps the actual detail with a visible request counter: more than 1 after one submit is a double submit. */
export function counted<P extends object>(Detail: ComponentType<P>, args: P, handlers: (count: () => void) => Partial<P>) {
	return function Counted() {
		const [calls, setCalls] = useState(0);
		return (
			<div>
				<p role="status" className="mx-auto max-w-7xl px-4 pt-4 text-sm">
					Permintaan terkirim: {calls}
				</p>
				<Detail {...args} {...handlers(() => setCalls((n) => n + 1))} />
			</div>
		);
	};
}
