import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { ContentState } from "./content-state";

const meta = {
	title: "BO/Components/ContentState",
	component: ContentState,
	args: { kind: "loading", title: "Memuat data…" },
	parameters: { layout: "padded" },
} satisfies Meta<typeof ContentState>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Loading: Story = {};
export const Error: Story = {
	args: { kind: "error", title: "Data tidak dapat dimuat", description: "Coba muat ulang data." },
	render: function RetryPreview() {
		const [retried, setRetried] = useState(false);
		return (
			<>
				<ContentState kind="error" title="Data tidak dapat dimuat" onRetry={() => setRetried(true)} />
				{retried && <p role="status">Muat ulang diminta</p>}
			</>
		);
	},
};
export const Empty: Story = { args: { kind: "empty", title: "Belum ada data" } };
