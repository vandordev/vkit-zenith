import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { z } from "zod";
import { LoginForm } from "./-components/login-form";
import { AccessState } from "../-components/access-state";
import type { Feedback } from "../-components/common";
import { useSession } from "../../lib/api/session";
import { boV1Login } from "../../lib/api/generated/sdk.gen";
import { zBoV1LoginBody } from "../../lib/api/generated/zod.gen";
import { errorMessage } from "../../lib/api/client";
import { safeReturnTo } from "../../lib/api/redirect";
export const Route = createFileRoute("/login/")({ validateSearch: z.object({ returnTo: z.string().optional() }), component: Login });
function Login() {
	const { api, query } = useSession();
	const navigate = useNavigate();
	const { returnTo } = Route.useSearch();
	const [pending, setPending] = useState(false);
	const [feedback, setFeedback] = useState<Feedback>();
	useEffect(() => {
		if (query.data) window.location.replace(safeReturnTo(returnTo));
	}, [query.data, returnTo]);
	if (api.status !== "ready")
		return (
			<AccessState
				kind={api.status === "error" ? "config" : "pending"}
				onRetry={api.status === "error" ? () => window.location.reload() : undefined}
			/>
		);
	if (!api.invalidated && query.isPending) return <AccessState kind="pending" />;
	return (
		<LoginForm
			pending={pending}
			feedback={feedback}
			onSubmit={(values) => {
				if (pending) return;
				setPending(true);
				setFeedback(undefined);
				void (async () => {
					try {
						const body = zBoV1LoginBody.parse(values);
						await boV1Login({ ...(await api.transport.writeOptions()), body });
						api.transport.reset();
						// A full internal navigation gives the new identity a fresh memory/cache boundary.
						window.location.replace(safeReturnTo(returnTo));
					} catch (error) {
						setFeedback({ kind: "error", message: errorMessage(error) });
					} finally {
						setPending(false);
					}
				})();
			}}
		/>
	);
}
