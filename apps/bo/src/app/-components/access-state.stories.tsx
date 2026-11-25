import type { Meta, StoryObj } from "@storybook/react";
import { useEffect, useState } from "react";
import { useMinimumPending } from "@repo/ui-utils/react";
import { AccessState } from "./access-state";
import { noop } from "../-fixtures";

const meta = {
	title: "BO/Components/AccessState",
	component: AccessState,
	parameters: { layout: "fullscreen" },
	args: { kind: "pending" },
	argTypes: { onRetry: { control: false } },
} satisfies Meta<typeof AccessState>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Pending: Story = {};
function TimedAccess({ delay, fail = false }: { delay: number; fail?: boolean }) {
	const [pending, setPending] = useState(true);
	const held = useMinimumPending(pending);
	useEffect(() => {
		const timer = setTimeout(() => setPending(false), delay);
		return () => clearTimeout(timer);
	}, [delay]);
	if (!pending && fail) return <AccessState kind="config" />;
	return held ? <AccessState kind="pending" /> : <p>Halaman siap</p>;
}
export const FastReady: Story = { render: () => <TimedAccess delay={50} /> };
export const SlowReady: Story = { render: () => <TimedAccess delay={1400} /> };
export const FastError: Story = { render: () => <TimedAccess delay={50} fail /> };
export const ConfigError: Story = { args: { kind: "config", onRetry: noop } };
export const Denied: Story = { args: { kind: "denied" } };
export const SessionError: Story = { args: { kind: "session", onRetry: noop } };
