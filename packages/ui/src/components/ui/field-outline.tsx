import { cn } from "cn";

/** A real border notch, so floating labels need no opaque background patch. */
export function FieldOutline({ label, raised, className }: { label: string; raised: boolean; className?: string }) {
	return (
		<fieldset
			aria-hidden="true"
			data-slot="field-outline"
			className={cn("pointer-events-none absolute inset-x-0 -top-1.5 bottom-0 min-w-0 rounded-md border border-input px-2", className)}
		>
			<legend
				className={cn("invisible block h-3 overflow-hidden whitespace-nowrap p-0 text-xs leading-3", raised ? "max-w-full" : "max-w-0")}
			>
				<span className="px-1">{label}</span>
			</legend>
		</fieldset>
	);
}
