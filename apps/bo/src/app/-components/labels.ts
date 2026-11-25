export const providerLabels: Record<string, string> = { cring: "Cring", xendit: "Xendit", midtrans: "Midtrans", manual: "Transfer manual" };
export const methodLabels: Record<string, string> = {
	card: "Kartu",
	virtual_account: "Virtual account",
	qris: "QRIS",
	ewallet: "E-wallet",
	retail_outlet: "Gerai ritel",
	direct_debit: "Debit langsung",
	paylater: "PayLater",
	bank_transfer: "Transfer bank",
};
const units: Record<string, string> = { day: "hari", week: "minggu", month: "bulan", year: "tahun" };
// Unknown values from a newer API render raw rather than failing.
export const providerLabel = (value: string) => providerLabels[value] ?? value;
export const methodLabel = (value: string) => methodLabels[value] ?? value;
export const intervalLabel = (interval: string | null, count: number | null) =>
	interval ? `Setiap ${count && count > 1 ? `${count} ` : ""}${units[interval] ?? interval}` : "Sekali bayar";
export const methodTypes = Object.keys(methodLabels);
