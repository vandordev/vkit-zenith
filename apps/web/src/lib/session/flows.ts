import type { AuthTransport } from "../api/client";
import { AuthError, toFormStatus } from "../api/errors";
import {
	authV1ConfirmEmailAction,
	authV1CurrentSession,
	authV1InspectEmailAction,
	authV1Login,
	authV1Logout,
	authV1Register,
	authV1RequestPasswordAction,
	authV1ResendVerification,
	authV1TouchSession,
} from "../api/generated/sdk.gen";
import type { FormStatus } from "../../app/-components/form-status";
import type { CheckEmailState } from "../../app/check-email/-components/check-email-panel";
import type { LinkPageState } from "../../app/-components/link-outcome";
import type { LoginValues } from "../../app/login/-components/login-form";
import type { RegisterValues } from "../../app/register/-components/register-form";

export type Purpose = "verify_email" | "setup_password" | "reset_password";
export type HomeState =
	| { kind: "signed-in"; displayName: string; email: string }
	| { kind: "restricted"; displayName: string; email: string }
	| { kind: "anonymous" }
	| { kind: "unavailable" };
type Outcome = Extract<LinkPageState, { kind: "outcome" }>;
const unavailableLink: Outcome = { kind: "outcome", outcome: "unavailable" };

const read = (t: AuthTransport) => ({ client: t.client, throwOnError: true as const });
const write = async (t: AuthTransport) => ({ ...read(t), headers: { "X-CSRF-Token": await t.csrf() } });
const statusOf = (error: unknown) => (error instanceof AuthError ? error.status : undefined);

export const toE164 = (digits: string) => (digits ? `+${digits}` : "");

export function afterIdentityChange(t: AuthTransport, csrfToken?: string) {
	t.identityChanged();
	if (csrfToken) t.setCsrf(csrfToken);
}

const sessionChanged = () => new AuthError(401, "session_changed", "Sesi berubah.");
const isSessionChanged = (error: unknown) => error instanceof AuthError && error.code === "session_changed";

/** Multi-step flows read under one identity; call the returned check between steps and before any mutation. */
function fence(t: AuthTransport) {
	const accepted = t.epoch();
	return () => {
		if (t.epoch() !== accepted) throw sessionChanged();
	};
}

/** Renews the central cookie; ordinary failures never hide a valid session, an identity change always aborts. */
async function touch(t: AuthTransport) {
	try {
		await authV1TouchSession(await write(t));
	} catch (error) {
		if (isSessionChanged(error)) throw error;
		// shortcut: other touch failures (429/503/csrf) are ignored, the next visit renews again
	}
}

export async function resolveHome(t: AuthTransport): Promise<HomeState> {
	const unchanged = fence(t);
	try {
		const { data } = await authV1CurrentSession(read(t));
		unchanged();
		if (data.kind === "restricted") return { kind: "restricted", displayName: data.displayName, email: data.email };
		await touch(t);
		unchanged();
		return { kind: "signed-in", displayName: data.displayName, email: data.email };
	} catch (error) {
		// a stale identity is never reported; the caller's next query runs under the new epoch
		if (isSessionChanged(error)) return { kind: "unavailable" };
		return statusOf(error) === 401 ? { kind: "anonymous" } : { kind: "unavailable" };
	}
}

export async function login(t: AuthTransport, values: LoginValues): Promise<{ kind: "full" } | { kind: "restricted" } | FormStatus> {
	try {
		const { data } = await authV1Login({ ...(await write(t)), body: values });
		afterIdentityChange(t, data.csrfToken);
		return { kind: data.kind };
	} catch (error) {
		return toFormStatus(error);
	}
}

export async function register(t: AuthTransport, values: RegisterValues): Promise<{ kind: "accepted" } | FormStatus> {
	try {
		const { displayName, email, password, telephoneDigits } = values;
		await authV1Register({ ...(await write(t)), body: { displayName, email, password, telephone: toE164(telephoneDigits) } });
		afterIdentityChange(t);
		return { kind: "accepted" };
	} catch (error) {
		return toFormStatus(error);
	}
}

export async function checkEmailState(t: AuthTransport): Promise<CheckEmailState | { kind: "unavailable" }> {
	try {
		const { data } = await authV1CurrentSession(read(t));
		return data.kind === "restricted" ? { kind: "restricted", email: data.email, displayName: data.displayName } : { kind: "generic" };
	} catch (error) {
		return statusOf(error) === 401 ? { kind: "generic" } : { kind: "unavailable" };
	}
}

async function done(run: () => Promise<unknown>): Promise<{ kind: "done" } | FormStatus> {
	try {
		await run();
		return { kind: "done" };
	} catch (error) {
		return toFormStatus(error);
	}
}

export const requestPasswordLink = (t: AuthTransport, email: string) =>
	done(async () => authV1RequestPasswordAction({ ...(await write(t)), body: { email } }));

export const resend = (t: AuthTransport) => done(async () => authV1ResendVerification(await write(t)));

export const logout = (t: AuthTransport) =>
	done(async () => {
		await authV1Logout(await write(t));
		afterIdentityChange(t);
	});

export async function inspectForRoute(t: AuthTransport, token: string, expected: Purpose): Promise<LinkPageState> {
	try {
		const { data } = await authV1InspectEmailAction({ ...read(t), query: { token } });
		if (data.purpose !== expected) return unavailableLink;
		if (data.requiresLogin) return { kind: "cross-browser" };
		return { kind: "ready", expiresAt: data.expiresAt, ...(data.displayName && { displayName: data.displayName }) };
	} catch (error) {
		// 404/422: unknown, used, expired or malformed look identical. Anything else is a real failure.
		if (statusOf(error) === 404 || statusOf(error) === 422) return unavailableLink;
		throw error;
	}
}

export async function confirm(
	t: AuthTransport,
	token: string,
	values?: { password?: string; telephoneDigits?: string },
): Promise<Outcome | FormStatus> {
	try {
		const body = {
			token,
			...(values?.password !== undefined && { password: values.password }),
			...(values?.telephoneDigits !== undefined && { telephone: toE164(values.telephoneDigits) }),
		};
		const { data } = await authV1ConfirmEmailAction({ ...(await write(t)), body });
		afterIdentityChange(t, data.csrfToken);
		if (!data.completed) return unavailableLink;
		return { kind: "outcome", outcome: data.csrfToken ? "completed-signed-in" : "completed-login" };
	} catch (error) {
		return statusOf(error) === 404 ? unavailableLink : toFormStatus(error);
	}
}
