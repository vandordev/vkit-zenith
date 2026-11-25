import type { ComponentProps, ReactNode } from "react";
import { Page } from "../../../-components/common";
import { StaffCollection } from "./staff-collection";
export function StaffPage({
	action,
	children,
	...props
}: ComponentProps<typeof StaffCollection> & { action?: ReactNode; children?: ReactNode }) {
	return (
		<Page
			title="Staf"
			description="Kelola akun internal. Perubahan peran, status dan kata sandi mengakhiri sesi staf terkait."
			action={action}
		>
			<StaffCollection {...props} />
			{children}
		</Page>
	);
}
