import { z } from "zod";
import * as schema from "@bo/lib/api/generated/zod.gen";

const page = z.coerce.number().int().min(1).max(1000000).catch(1);
const q = z.string().max(100).catch("");
const empty = <T extends string>(value: T | undefined) => value ?? ("" as const);
// Jobs belong to no organization.
export const jobsSearch = z.object({
	q,
	page,
	status: schema.zBoV1ListJobsQuery.shape.status.catch("").transform(empty),
	stateGroup: schema.zBoV1ListJobsQuery.shape.stateGroup.catch("").transform(empty),
});
