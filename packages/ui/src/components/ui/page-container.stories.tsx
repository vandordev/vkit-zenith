import type { Meta, StoryObj } from "@storybook/react";

import { PageContainer } from "./page-container";

const meta = {
	argTypes: {
		children: { control: "text" },
		className: { control: "text" },
		ref: { control: false },
		size: { control: "select", options: ["sm", "md", "lg", "xl", "full"] },
	},
	args: { children: "Page content", size: "full" },
	component: PageContainer,
	parameters: {
		layout: "fullscreen",
		docs: {
			description: {
				component: "Page padding is 16px on mobile, 24px from sm, and 48px from lg. The shell does not add another padding layer.",
			},
		},
	},
	render: ({ children, ...args }) => (
		<PageContainer {...args}>
			<div className="bg-muted py-8 text-sm text-foreground">{children}</div>
		</PageContainer>
	),
	title: "UI/PageContainer",
} satisfies Meta<typeof PageContainer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
/** Full width is the default; constrain reading-heavy pages such as forms. */
export const Constrained: Story = { args: { size: "md" } };
export const LongContent: Story = {
	args: {
		children:
			"Page content stays within the available width. Long prose wraps naturally; tables and unbroken values should manage overflow inside the page, not on the container.",
		size: "sm",
	},
};
export const Sizes: Story = {
	argTypes: { children: { control: false }, size: { control: false } },
	render: (args) => (
		<div className="flex flex-col gap-6 py-6">
			{(["sm", "md", "lg", "xl", "full"] as const).map((size) => (
				<PageContainer {...args} key={size} size={size}>
					<div className="bg-muted py-4 text-sm text-foreground">{size}</div>
				</PageContainer>
			))}
		</div>
	),
};
