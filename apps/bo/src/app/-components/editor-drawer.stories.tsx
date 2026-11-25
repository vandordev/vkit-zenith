import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";
import { Button } from "@repo/ui/button";
import { EditForm, EditorDrawer, Page } from "./common";

const meta = {
	title: "BO/Components/EditorDrawer",
	component: EditorDrawer,
	parameters: { layout: "fullscreen" },
	args: { title: "Ubah profil", onClose: () => {}, children: () => null },
	argTypes: { children: { control: false }, onClose: { control: false } },
	render: function DrawerLifecycle(args) {
		const [selection, setSelection] = useState<string>();
		return (
			<Page title="Profil" action={<Button onClick={() => setSelection("profile")}>Ubah profil</Button>}>
				<EditorDrawer
					{...args}
					title={selection ? "Ubah profil" : "Tambah profil"}
					open={selection !== undefined}
					editorKey={selection}
					onClose={() => setSelection(undefined)}
				>
					{(onDirty) =>
						selection ? (
							<EditForm
								fields={[{ name: "name", label: "Nama" }]}
								initial={{ name: "Nadia Prameswari" }}
								pending={args.pending}
								onDirty={onDirty}
								onSubmit={() => setSelection(undefined)}
							/>
						) : (
							<p>Tidak ada profil terpilih.</p>
						)
					}
				</EditorDrawer>
			</Page>
		);
	},
} satisfies Meta<typeof EditorDrawer>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
export const Pending: Story = { args: { pending: true } };
