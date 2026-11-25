import type { Meta, StoryObj } from "@storybook/react";
import { AdjustmentDialogView } from "../../../../../-components/adjustment-dialog";
import { adjustmentStates } from "../../../../../-story-helpers";
import { apiKeys } from "../../-setup-fixtures";
import { revokeApiKeyDialog, revokeApiKeyResult } from "./api-key-list";

const meta = {
	title: "BO/Pages/Organizations/ApiKeyDialogs",
	component: AdjustmentDialogView,
	parameters: { layout: "fullscreen" },
	argTypes: { fields: { control: false }, result: { control: false } },
} satisfies Meta<typeof AdjustmentDialogView>;
export default meta;
type Story = StoryObj<typeof meta>;
const key = apiKeys[0]!;
const states = adjustmentStates(
	revokeApiKeyDialog(key),
	revokeApiKeyResult({ dummy: true, result: { ...key, status: "revoked", revokedAt: "2026-11-15T08:00:00Z" } }),
);
export const RevokeOpen: Story = states.Open;
export const RevokeValidationError: Story = states.ValidationError;
export const RevokeSubmitting: Story = states.Submitting;
export const RevokeFailed: Story = states.Failed;
export const RevokeSuccess: Story = states.Success;
