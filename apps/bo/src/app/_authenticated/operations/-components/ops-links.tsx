import { Link } from "@tanstack/react-router";

/** Record pages without children: a collection row is their label and its whole surface opens them. */
export const jobLink = (id: string) => <Link to="/operations/jobs/$jobId" params={{ jobId: id }} />;

export const operationId = (id: string) => <span className="font-mono text-xs break-all">{id}</span>;
