import type { ReactNode } from "react";

/** Inline message under a field; the field links to it with aria-describedby. */
export function FieldProblem({ id, children }: { id: string; children?: ReactNode }) {
	return children ? (
		<p id={id} className="text-xs text-destructive-text">
			{children}
		</p>
	) : null;
}
export const formProblemMessage = "Periksa kembali isian yang ditandai.";
