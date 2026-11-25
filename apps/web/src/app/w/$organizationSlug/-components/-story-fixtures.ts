import type { V1Page } from "../../../../lib/api/generated/types.gen";
import { orgFailureMessage } from "../../../../lib/organization-pure";

// Story-only helpers: deterministic pagination and simulated API outcomes. No secrets.
export const pageOf = (total: number, page = 1, size = 20): V1Page => ({ page, size, total: String(total) });
export const noop = () => {};
/** A failed call as the organization helpers read it: any error carrying an HTTP `status`, without importing the API client. */
export const statusFailure = (status: number, message: string) => Object.assign(new Error(message), { status });
export const delayed = <T>(value: T, ms = 700) => new Promise<T>((resolve) => setTimeout(() => resolve(value), ms));
export const failing = (status: number, ms = 700) =>
	new Promise<never>((_, reject) => setTimeout(() => reject(statusFailure(status, "SERVER-DETAIL-SENTINEL")), ms));
export const describeError = orgFailureMessage;
