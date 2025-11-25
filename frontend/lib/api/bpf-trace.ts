import { z } from "zod";

const BaseSchema = z.object({
  _id: z.string(),
  func_name: z.string(),
  pid: z.string(),
  timestamp: z.string(),
  duration_ns: z.string(),
  __v: z.number().optional(),
});

const EnterDataSchema = z.object({
  caller_name: z.string(),
  args_hex: z.array(z.string()).default([]),
});

const ExitDataSchema = z.object({
  back_to_name: z.string(),
  return_value: z.string(),
});

export const BpfTraceEventSchema = z.discriminatedUnion("event_type", [
  BaseSchema.extend({
    event_type: z.literal("enter"),
    data: EnterDataSchema,
  }),
  BaseSchema.extend({
    event_type: z.literal("exit"),
    data: ExitDataSchema,
  }),
]);

export type BpfTraceEvent = z.infer<typeof BpfTraceEventSchema>;
