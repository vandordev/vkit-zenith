import { createContext, useContext } from "react";
import type { OrgContext } from "./navigation";

// Adapters provide it from the verified membership; presentation components never read it.
export const OrgContextValue = createContext<OrgContext | undefined>(undefined);
export function useOrg(): OrgContext {
	const value = useContext(OrgContextValue);
	if (!value) throw new Error("useOrg must be used inside the organization route");
	return value;
}
