"use client";

import { Radio as RadioPrimitive } from "@base-ui/react/radio";
import { RadioGroup as RadioGroupPrimitive } from "@base-ui/react/radio-group";
import { cn } from "cn";
import { motion as brandMotion } from "@repo/brand/motion";
import { motion, useReducedMotion } from "motion/react";
import type { ComponentProps } from "react";

export type RadioGroupProps<Value = string> = RadioGroupPrimitive.Props<Value>;

export const RadioGroup = <Value = string,>({ className, ...props }: RadioGroupProps<Value>) => (
	<RadioGroupPrimitive
		{...props}
		data-slot="radio-group"
		className={(state) => cn("grid gap-2", typeof className === "function" ? className(state) : className)}
	/>
);

export type RadioGroupItemProps = Omit<ComponentProps<typeof RadioPrimitive.Root>, "children"> & {
	/** Animate the indicator. Reduced-motion preferences take precedence. */
	animated?: boolean;
};

export const RadioGroupItem = ({ className, animated = true, ...props }: RadioGroupItemProps) => {
	const reducedMotion = useReducedMotion();
	const duration = animated && !reducedMotion ? brandMotion.fast : 0;

	return (
		<RadioPrimitive.Root
			{...props}
			data-slot="radio-group-item"
			className={(state) =>
				cn(
					"peer relative flex size-4 shrink-0 items-center justify-center rounded-full border border-input outline-none transition-colors duration-(--motion-fast) motion-reduce:transition-none after:absolute after:-inset-x-3 after:-inset-y-2 focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 data-disabled:cursor-not-allowed data-disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 data-checked:border-primary data-checked:bg-primary data-checked:text-primary-foreground vandor-dark:bg-input/30 vandor-dark:data-checked:bg-primary",
					typeof className === "function" ? className(state) : className,
				)
			}
		>
			<RadioPrimitive.Indicator
				keepMounted
				data-slot="radio-group-indicator"
				render={(indicatorProps, state) => (
					<span {...indicatorProps} aria-hidden="true" className="grid place-content-center">
						<motion.span
							className="block size-1.5 rounded-full bg-current"
							initial={false}
							animate={{ opacity: state.checked ? 1 : 0, scale: state.checked ? 1 : 0.4 }}
							transition={{ duration, ease: brandMotion.easeOut }}
						/>
					</span>
				)}
			/>
		</RadioPrimitive.Root>
	);
};
