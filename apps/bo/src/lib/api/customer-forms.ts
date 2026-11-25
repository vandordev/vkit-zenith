import * as schema from "./generated/zod.gen";
import { formInput } from "./forms";
// Generated schemas are the only transport shape source. Strict projection rejects
// forged privileged extras instead of disguising a rejected write as success.
export function organizationCreateBody(value: unknown) {
	const body = formInput(
		schema.zBoV1CreateOrganizationBody
			.extend({ newOwner: schema.zBoV1CreateOrganizationBody.shape.newOwner.unwrap().strict().optional() })
			.strict(),
		value,
	);
	if (Boolean(body.ownerUserId) === Boolean(body.newOwner)) throw new Error("Periksa satu sumber pemilik awal. Data belum dikirim.");
	return body;
}
export const organizationProfileBody = (value: unknown) => formInput(schema.zBoV1UpdateOrganizationBody.strict(), value);
export const organizationStatusBody = (value: unknown) => formInput(schema.zBoV1SetOrganizationStatusBody.strict(), value);
export const memberAddBody = (value: unknown) => formInput(schema.zBoV1AddOrganizationMemberBody.strict(), value);
export const memberRemoveBody = (value: unknown) => formInput(schema.zBoV1RemoveOrganizationMemberBody.strict(), value);
export const memberRestoreBody = (value: unknown) => formInput(schema.zBoV1RestoreOrganizationMemberBody.strict(), value);
export const ownerBody = (value: unknown) => formInput(schema.zBoV1SetOrganizationMembershipRoleBody.strict(), value);
export const userCreateBody = (value: unknown) => formInput(schema.zBoV1CreateUserBody.strict(), value);
export const userProfileBody = (value: unknown) => formInput(schema.zBoV1UpdateUserBody.strict(), value);
export const userEmailBody = (value: unknown) => formInput(schema.zBoV1ChangeUserEmailBody.strict(), value);
export const userVerifyBody = (value: unknown) => formInput(schema.zBoV1VerifyUserEmailBody.strict(), value);
export const userStatusBody = (value: unknown) => formInput(schema.zBoV1SetUserStatusBody.strict(), value);
