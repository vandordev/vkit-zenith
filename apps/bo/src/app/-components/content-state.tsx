import type { ReactNode } from "react";
import { Button } from "@repo/ui/button";
import { Loading } from "@repo/ui/loading";
import {
	ErrorState,
	ErrorStateContent,
	ErrorStateHeader,
	ErrorStateTitle,
	ErrorStateDescription,
	ErrorStateActions,
} from "@repo/ui/error-state";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyContent } from "@repo/ui/empty";

export type ContentStateProps =
	| { kind: "loading"; title: string }
	| { kind: "error"; title: string; description?: string; onRetry?: () => void }
	| { kind: "empty"; title: string; description?: string; action?: ReactNode };

export function ContentState(props: ContentStateProps) {
	if (props.kind === "loading")
		return (
			<div className="flex min-h-32 items-center justify-center py-8" aria-busy="true">
				<Loading variant="skeleton" size={48} aria-label={props.title} />
			</div>
		);
	if (props.kind === "error")
		return (
			<ErrorState role="alert">
				<ErrorStateContent>
					<ErrorStateHeader>
						<ErrorStateTitle>{props.title}</ErrorStateTitle>
						{props.description && <ErrorStateDescription>{props.description}</ErrorStateDescription>}
					</ErrorStateHeader>
					{props.onRetry && (
						<ErrorStateActions>
							<Button type="button" variant="outline" onClick={props.onRetry}>
								Muat ulang
							</Button>
						</ErrorStateActions>
					)}
				</ErrorStateContent>
			</ErrorState>
		);
	return (
		<Empty>
			<EmptyHeader>
				<EmptyTitle>{props.title}</EmptyTitle>
				{props.description && <EmptyDescription>{props.description}</EmptyDescription>}
			</EmptyHeader>
			{props.action && <EmptyContent>{props.action}</EmptyContent>}
		</Empty>
	);
}
