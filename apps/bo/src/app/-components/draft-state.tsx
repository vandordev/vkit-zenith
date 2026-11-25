import { createContext, useContext, type MutableRefObject } from "react";
// Presentation-only dirty signal. Routing and confirmation live in the adapter.
export const DraftState = createContext<MutableRefObject<boolean> | undefined>(undefined);
export const useDraftState = () => useContext(DraftState);
