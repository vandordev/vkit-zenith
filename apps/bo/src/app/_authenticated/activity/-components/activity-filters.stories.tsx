import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { ActivityFilters, type ActivityFilterValue } from "./activity-filters";
import { actions, noop } from "../-fixtures";
import { Autocomplete } from "@repo/ui/autocomplete";
import { useChoiceFixture } from "../../../-fixtures/choice-autocomplete";
import { activityFilterSearch, parseActivitySearch } from "@bo/lib/api/activity-navigation";
function initialSearch(value: ActivityFilterValue) {
	return parseActivitySearch({ ...value, page: 3, selected: "00000000-0000-4000-8000-000000000002", invalid: true });
}
const meta = {
	title: "BO/Pages/Activity/Filters",
	component: ActivityFilters,
	args: { value: {}, actions, onChange: noop, onReset: noop },
	argTypes: { onChange: { control: false }, onReset: { control: false }, choices: { control: false } },
	parameters: { layout: "padded" },
} satisfies Meta<typeof ActivityFilters>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {
	render: function Preview(args) {
		const [value, setValue] = useState(() => initialSearch(args.value));
		return (
			<>
				<ActivityFilters
					{...args}
					value={value}
						onChange={(next) => setValue(activityFilterSearch({ ...next, selected: value.selected }))}
						onReset={() => setValue(activityFilterSearch({ selected: value.selected }))}
				/>
				<output className="mt-4 block break-all text-sm">{JSON.stringify(value)}</output>
			</>
		);
	},
};
export const Populated: Story = {
	...Default,
	args: { value: { from: "2026-10-05", to: "2026-10-12", action: "staff.update", outcome: "success" } },
};
export const Invalid: Story = {
	...Default,
	args: {
		value: { from: "2026-10-12", to: "2026-10-05" },
		error: "Rentang tanggal tidak valid. Tidak ada permintaan aktivitas yang dikirim.",
	},
};
export const BoundedChoices: Story = {
	render: function Preview(args) {
		const [value, setValue] = useState(() =>
			initialSearch({ from: "2026-10-05", to: "2026-10-12", action: "staff.update", outcome: "success" }),
		);
		const choice = { id: "00000000-0000-4000-8000-000000000001", name: "Nadia Prameswari (dinonaktifkan)", key: "Staf" };
		const staffChoices = useChoiceFixture([choice]);
		return (
			<>
				<ActivityFilters
					{...args}
					value={value}
						onChange={(next) => setValue(activityFilterSearch({ ...next, selected: value.selected }))}
						onReset={() => setValue(activityFilterSearch({ selected: value.selected }))}
					choices={
						<>
							<Autocomplete
								{...staffChoices}
								mode="selection"
								multiple={false}
								label="Staf pelaku (nama saat ini)"
								value={value.actorStaffId ? choice : null}
								onValueChange={(item) => setValue(activityFilterSearch({ ...value, actorStaffId: item?.id, selected: value.selected }))}
							/>
						</>
					}
				/>
				<output className="break-all">{JSON.stringify({ value })}</output>
			</>
		);
	},
};
