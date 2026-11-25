import { createFileRoute, Outlet, useLocation, useBlocker } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useMinimumPending } from "@repo/ui-utils/react";
import { DraftState } from "../-components/draft-state";
import { AccessState } from "../-components/access-state";
import { BOShell } from "./-components/bo-shell";
import { buildShellNavigation, buildSidebarContext } from "./-components/navigation-links";
import { resolveNavigation } from "./-navigation";
import { ShellContextProvider, useShellContextValue } from "../-components/shell-context";
import { Notice, type Feedback } from "../-components/common";
import { ContentState } from "../-components/content-state";
import { SessionProvider, useSession } from "../../lib/api/session";
import { BOError, errorMessage } from "../../lib/api/client";
import { boV1Logout } from "../../lib/api/generated/sdk.gen";
import { safeReturnTo } from "../../lib/api/redirect";
import { useTheme } from "../../lib/theme";
import { isActivityQuery } from "../../lib/api/activity-navigation";
import { isStaffGovernanceQuery } from "../../lib/api/private-cache";
import { parseCustomerReturn, parseStaffReturn } from "@bo/lib/api/customer-navigation";
export const Route = createFileRoute("/_authenticated")({ component: Protected });
function Protected() {
	return (
		<ShellContextProvider>
			<ProtectedContent />
		</ShellContextProvider>
	);
}
function ProtectedContent() {
	const shellContext = useShellContextValue();
	const navigate = Route.useNavigate();
	const [theme, setTheme] = useTheme();
	const draft = useRef(false);
	const writing = useRef(false);
	useBlocker({
		shouldBlockFn: () => draft.current && !window.confirm("Buang perubahan yang belum disimpan?"),
		enableBeforeUnload: () => draft.current,
	});
	const { api, query } = useSession();
	const queries = useQueryClient();
	useEffect(() => {
		if (query.data?.role === "superadmin") return;
		const remove = (entry: { queryKey: readonly unknown[] }) => isActivityQuery(entry) || isStaffGovernanceQuery(entry);
		void queries.cancelQueries({ predicate: remove }).then(() => queries.removeQueries({ predicate: remove }));
	}, [queries, query.data?.role]);
	const showPending = useMinimumPending(api.status !== "ready" || query.isPending);
	const location = useLocation();
	const [pending, setPending] = useState(false);
	const [feedback, setFeedback] = useState<Feedback>();
	const unauthorized = (api.status === "ready" && api.invalidated) || (query.error instanceof BOError && query.error.status === 401);
	useEffect(() => {
		if (unauthorized) window.location.replace(`/login?returnTo=${encodeURIComponent(safeReturnTo(location.href))}`);
	}, [unauthorized, location.href]);
	if (api.status !== "ready")
		return (
			<AccessState
				kind={api.status === "error" ? "config" : "pending"}
				onRetry={api.status === "error" ? () => window.location.reload() : undefined}
			/>
		);
	if (query.isPending || unauthorized) return <AccessState kind="pending" />;
	if (!query.data)
		return (
			<AccessState
				kind="session"
				onRetry={() => {
					void query.refetch();
				}}
			/>
		);
	if (showPending && !query.error) return <AccessState kind="pending" />;
	const staff = query.data;
	const navigation = resolveNavigation({
		pathname: location.pathname,
		search: {
			customerReturnTo: parseCustomerReturn((location.search as { returnTo?: unknown }).returnTo),
			returnTo: parseStaffReturn((location.search as { returnTo?: unknown }).returnTo),
		},
		role: staff.role,
		context: shellContext ?? undefined,
	});
	return (
		<SessionProvider value={{ staff, transport: api.transport, epoch: api.epoch, writing }}>
			<DraftState.Provider value={draft}>
				<BOShell
					staff={staff}
					pending={pending}
					theme={theme}
					onThemeChange={setTheme}
					onAccount={() => {
						void navigate({ to: "/account" });
					}}
					navigation={buildShellNavigation(navigation.items)}
					navigationContext={navigation.panel}
					sidebarContext={buildSidebarContext(navigation.sidebarContext)}
					onLogout={() => {
						if (pending || writing.current) return;
						if (draft.current && !window.confirm("Keluar dan buang perubahan yang belum disimpan?")) return;
						draft.current = false;
						writing.current = true;
						setPending(true);
						setFeedback(undefined);
						void (async () => {
							try {
								await boV1Logout({ ...(await api.transport.writeOptions()) });
								api.transport.reset();
								window.location.replace("/login");
							} catch (error) {
								setFeedback({ kind: "error", message: errorMessage(error) });
							} finally {
								writing.current = false;
								setPending(false);
							}
						})();
					}}
				>
					<Notice feedback={feedback} />
					{query.error && (
						<ContentState
							kind="error"
							title={errorMessage(query.error)}
							onRetry={() => {
								void query.refetch();
							}}
						/>
					)}
					<Outlet />
				</BOShell>
			</DraftState.Provider>
		</SessionProvider>
	);
}
