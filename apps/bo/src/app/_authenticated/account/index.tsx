import { createFileRoute } from "@tanstack/react-router";
import { AccountPage } from "./-components/account-page";
import { useProtected } from "../../../lib/api/session";
import { boV1ChangePassword } from "../../../lib/api/generated/sdk.gen";
import { zBoV1ChangePasswordBody } from "../../../lib/api/generated/zod.gen";
import { field, passwordValid } from "../../../lib/api/forms";
import { useWrite } from "../../../lib/api/workspace";
export const Route = createFileRoute("/_authenticated/account/")({ component: Account });
function Account() {
	const { staff, transport } = useProtected();
	const { pending, feedback, setFeedback, run } = useWrite();
	return (
		<AccountPage
			staff={staff}
			pending={pending}
			feedback={feedback}
			onSubmit={(values) => {
				if (pending) return;
				const newPassword = field(values, "newPassword");
				if (!passwordValid(newPassword) || newPassword !== values.confirmPassword || newPassword === values.currentPassword) {
					setFeedback({
						kind: "error",
						message: "Periksa batas kata sandi, kesamaan konfirmasi dan pastikan berbeda dari kata sandi saat ini.",
					});
					return;
				}
				run(async () => {
					const body = zBoV1ChangePasswordBody.parse({ currentPassword: values.currentPassword, newPassword });
					await boV1ChangePassword({ ...(await transport.writeOptions()), body });
					transport.reset();
					window.location.replace("/login");
				});
			}}
		/>
	);
}
