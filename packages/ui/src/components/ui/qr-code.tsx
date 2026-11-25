import qrcode from "qrcode-generator";
import { useMemo } from "react";
import type { ComponentProps } from "react";

export type QrCodeProps = Omit<ComponentProps<"svg">, "children" | "viewBox" | "role" | "aria-label"> & {
	/** Text to encode. Encoded locally; nothing is sent over the network. */
	value: string;
	/** Rendered width and height in pixels. */
	size?: number;
	/** Accessible name describing what the code opens or pays. */
	label: string;
};

// The library default truncates UTF-16 code units to one byte; encode Byte mode as UTF-8 instead.
qrcode.stringToBytes = (s) => Array.from(new TextEncoder().encode(s));

const QUIET_ZONE = 4;

export const QrCode = ({ value, size = 192, label, className, ...props }: QrCodeProps) => {
	const { count, path } = useMemo(() => {
		const qr = qrcode(0, "M");
		qr.addData(value, "Byte");
		qr.make();
		const total = qr.getModuleCount();
		let d = "";
		for (let row = 0; row < total; row += 1) {
			for (let col = 0; col < total; col += 1) {
				if (qr.isDark(row, col)) d += `M${col + QUIET_ZONE} ${row + QUIET_ZONE}h1v1h-1z`;
			}
		}
		return { count: total + QUIET_ZONE * 2, path: d };
	}, [value]);

	return (
		<svg
			{...props}
			data-slot="qr-code"
			role="img"
			aria-label={label}
			viewBox={`0 0 ${count} ${count}`}
			width={size}
			height={size}
			shapeRendering="crispEdges"
			className={className}
		>
			{/* Fixed colours, not theme tokens: scanners need dark modules on a light field in both themes. */}
			<rect width={count} height={count} fill="#fff" />
			<path d={path} fill="#000" />
		</svg>
	);
};
