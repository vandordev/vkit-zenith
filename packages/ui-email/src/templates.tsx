import { Body, Button, Container, Head, Heading, Html, Link, Preview, Text } from "@react-email/components";
import { brand } from "@repo/brand";

export type EmailPurpose = "verify_email" | "setup_password" | "reset_password";
export interface EmailFields {
	DisplayName: string;
	ActionURL: string;
	ExpiresAt: string;
}
export interface EmailTheme {
	background: string;
	foreground: string;
	primary: string;
	primaryForeground: string;
	mutedForeground: string;
}

export const families = {
	verify_email: {
		subject: `Verifikasi email ${brand.name}`,
		heading: "Konfirmasi alamat email Anda",
		instruction: "Konfirmasi ini membantu melindungi akun Anda. Lanjutkan dengan browser tempat Anda mendaftar atau masuk.",
		action: "Verifikasi email",
		closing: "Jika Anda tidak membuat akun ini, jangan konfirmasi tautan ini. Anda dapat mengabaikan email ini.",
	},
	setup_password: {
		subject: `Buat kata sandi ${brand.name}`,
		heading: "Siapkan kata sandi pertama Anda",
		instruction: "Lengkapi nomor telepon dan buat kata sandi untuk akun Anda. Setelah selesai, masuk dengan kata sandi baru.",
		action: "Buat kata sandi",
		closing: "Jika Anda tidak meminta pembuatan kata sandi, abaikan email ini. Akun Anda tidak berubah.",
	},
	reset_password: {
		subject: `Atur ulang kata sandi ${brand.name}`,
		heading: "Pilih kata sandi baru",
		instruction: "Setelah kata sandi diperbarui, sesi sebelumnya akan dicabut. Anda perlu masuk kembali.",
		action: "Atur ulang kata sandi",
		closing: "Jika Anda tidak meminta pengaturan ulang, abaikan email ini. Kata sandi Anda tidak berubah.",
	},
} as const;

export function AuthEmail({ purpose, fields, theme }: { purpose: EmailPurpose; fields: EmailFields; theme: EmailTheme }) {
	const family = families[purpose];
	return (
		<Html lang="id">
			<Head />
			<Preview>{family.heading}</Preview>
			<Body style={{ margin: 0, backgroundColor: theme.background, color: theme.foreground, fontFamily: "Arial, Helvetica, sans-serif" }}>
				<Container style={{ maxWidth: 520, margin: "0 auto", padding: "40px 24px" }}>
					<Text style={{ fontSize: 20, fontWeight: 700, margin: "0 0 36px", color: theme.primary }}>{brand.name}</Text>
					<Heading as="h1" style={{ fontSize: 26, lineHeight: "34px", fontWeight: 700, margin: "0 0 24px" }}>
						{family.heading}
					</Heading>
					<Text style={{ fontSize: 16, lineHeight: "26px", margin: "0 0 16px" }}>Halo {fields.DisplayName},</Text>
					<Text style={{ fontSize: 16, lineHeight: "26px", margin: "0 0 24px" }}>{family.instruction}</Text>
					<Button
						href={fields.ActionURL}
						style={{
							backgroundColor: theme.primary,
							color: theme.primaryForeground,
							padding: "14px 22px",
							borderRadius: 6,
							fontSize: 16,
							fontWeight: 600,
							textDecoration: "none",
						}}
					>
						{family.action}
					</Button>
					<Text style={{ fontSize: 14, lineHeight: "23px", margin: "24px 0 0" }}>
						Tautan berlaku sampai {fields.ExpiresAt}. Membuka tautan saja tidak menyelesaikan tindakan.
					</Text>
					<Text style={{ fontSize: 14, lineHeight: "23px", margin: "16px 0 0" }}>
						Jika tombol tidak dapat dibuka, salin tautan berikut ke browser Anda:
					</Text>
					<Link href={fields.ActionURL} style={{ color: theme.primary, fontSize: 14, lineHeight: "23px", wordBreak: "break-all" }}>
						{fields.ActionURL}
					</Link>
					<Text style={{ color: theme.mutedForeground, fontSize: 13, lineHeight: "21px", margin: "32px 0 0" }}>{family.closing}</Text>
				</Container>
			</Body>
		</Html>
	);
}
