import { cn } from "cn";

export interface MoneyProps {
	/** Integer minor units as an exact decimal string, e.g. "125000000". */
	amount: string;
	/** ISO 4217 code. */
	currency: string;
	/** "negative" renders the magnitude as a deduction (refunds); "auto" follows the amount's own sign. */
	sign?: "auto" | "negative";
	className?: string;
}

const placeholder = "—";

// ADR-0005: IDR is whole rupiah regardless of the ICU table.
/** Minor-unit exponent of a currency (IDR 0, USD 2); null when unsupported. */
export const exponentOf = (currency: string): number | null => {
	if (currency === "IDR") return 0;
	try {
		return new Intl.NumberFormat("id-ID", { style: "currency", currency }).resolvedOptions().maximumFractionDigits ?? null;
	} catch {
		return null;
	}
};

export const formatMoney = (amount: string, currency: string, sign: "auto" | "negative" = "auto"): string => {
	if (!/^-?\d+$/.test(amount)) return placeholder;
	const exponent = exponentOf(currency);
	if (exponent === null) return placeholder;
	const minor = BigInt(amount);
	const magnitude = minor < 0n ? -minor : minor;
	const scale = 10n ** BigInt(exponent);
	// Exact decimal string: Intl formats strings without a float round-trip.
	const decimal = exponent === 0 ? `${magnitude}` : `${magnitude / scale}.${String(magnitude % scale).padStart(exponent, "0")}`;
	const text = new Intl.NumberFormat("id-ID", {
		style: "currency",
		currency,
		minimumFractionDigits: exponent,
		maximumFractionDigits: exponent,
	}).format(decimal as unknown as number).replace(/\u00a0/g, ""); // spec: Rp1.250.000, no gap after the symbol
	const negative = magnitude !== 0n && (sign === "negative" || minor < 0n);
	return negative ? `−${text}` : text;
};

export const Money = ({ amount, currency, sign = "auto", className }: MoneyProps) => (
	<span data-slot="money" className={cn("tabular-nums whitespace-nowrap", className)}>
		{formatMoney(amount, currency, sign)}
	</span>
);
