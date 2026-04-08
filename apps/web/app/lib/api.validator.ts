import { ZodSchema } from "zod";

export function validateResponse<T>(schema: ZodSchema<T>, data: unknown): T {
  const result = schema.safeParse(data);

  if (!result.success) {
    const issues = result.error.issues
      .map(i => `${i.path.join(".")}: ${i.message}`)
      .join(", ");
    console.error("[API Validation Error]", result.error.issues);
    throw new Error(`API response validation failed: ${issues}`);
  }

  return result.data;
}