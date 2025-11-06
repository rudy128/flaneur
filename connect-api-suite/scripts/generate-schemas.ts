import { z } from 'zod';

// Read OpenAPI URL from CLI args or env, fallback to local backend server
const argv = Bun.argv.slice(2);
let openApiUrl = process.env.OPENAPI_URL;
for (let i = 0; i < argv.length; i++) {
  const arg = argv[i];
  if (!arg.startsWith('-') && !openApiUrl) {
    openApiUrl = arg;
    break;
  }
  if (arg === '--url' || arg === '-u') {
    openApiUrl = argv[i + 1];
    break;
  }
  const match = arg.match(/^--url=(.+)$/);
  if (match) {
    openApiUrl = match[1];
    break;
  }
}
openApiUrl = openApiUrl || 'http://localhost:8080/swagger/doc.json';

console.log(`📡 Fetching OpenAPI spec from: ${openApiUrl}`);

try {
  const res = await fetch(openApiUrl);
  
  if (!res.ok) {
    throw new Error(`Failed to fetch OpenAPI spec: ${res.status} ${res.statusText}`);
  }
  
  const openApiDoc = await res.json();

  console.log(`✅ OpenAPI spec fetched successfully`);
  console.log(`📝 API Title: ${openApiDoc.info?.title || 'Unknown'}`);
  console.log(`📝 API Version: ${openApiDoc.info?.version || 'Unknown'}`);
  console.log(`🔨 Generating TypeScript types and Zod schemas...`);

  // Generate schemas from definitions
  const definitions = openApiDoc.definitions || {};
  
  // Generate schemas.ts file
  let schemasContent = `/**
 * Auto-generated Zod schemas from OpenAPI spec
 * Generated: ${new Date().toISOString()}
 */

import { z } from "zod";

// =============================================================================
// Auth Schemas
// =============================================================================

`;

  // SignupRequest Schema
  if (definitions['schemas.SignupRequest']) {
    const signupDef = definitions['schemas.SignupRequest'];
    schemasContent += `/**
 * Signup Request Schema
 */
export const signupSchema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Invalid email address"),
  password: z.string().min(${signupDef.properties.password.minLength || 6}, "Password must be at least ${signupDef.properties.password.minLength || 6} characters"),
});

export type SignupRequest = z.infer<typeof signupSchema>;

/**
 * Signup Form Schema (with password confirmation)
 */
export const signupFormSchema = signupSchema.extend({
  confirmPassword: z.string().min(${signupDef.properties.password.minLength || 6}, "Password must be at least ${signupDef.properties.password.minLength || 6} characters"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

export type SignupFormData = z.infer<typeof signupFormSchema>;

`;
  }

  // LoginRequest Schema
  if (definitions['schemas.LoginRequest']) {
    schemasContent += `/**
 * Login Request Schema
 */
export const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(1, "Password is required"),
});

export type LoginRequest = z.infer<typeof loginSchema>;

`;
  }

  // LoginResponse Schema
  if (definitions['schemas.LoginResponse']) {
    schemasContent += `/**
 * Login Response Schema
 */
export const loginResponseSchema = z.object({
  token: z.string(),
  email: z.string(),
});

export type LoginResponse = z.infer<typeof loginResponseSchema>;

`;
  }

  // MessageResponse Schema
  if (definitions['schemas.MessageResponse']) {
    schemasContent += `/**
 * Message Response Schema
 */
export const messageResponseSchema = z.object({
  message: z.string(),
});

export type MessageResponse = z.infer<typeof messageResponseSchema>;

`;
  }

  // Error Response Schema
  schemasContent += `/**
 * Error Response Schema
 */
export const errorResponseSchema = z.object({
  error: z.string(),
});

export type ErrorResponse = z.infer<typeof errorResponseSchema>;

// =============================================================================
// Twitter Schemas
// =============================================================================

`;

  // TwitterAccountRequest Schema
  if (definitions['schemas.TwitterAccountRequest']) {
    schemasContent += `/**
 * Twitter Account Request Schema
 */
export const twitterAccountSchema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

export type TwitterAccountRequest = z.infer<typeof twitterAccountSchema>;

`;
  }

  // GetTweetsRequest Schema
  if (definitions['schemas.GetTweetsRequest']) {
    schemasContent += `/**
 * Get Tweets Request Schema
 */
export const getTweetsSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetTweetsRequest = z.infer<typeof getTweetsSchema>;

`;
  }

  // GetLikesRequest Schema
  if (definitions['schemas.GetLikesRequest']) {
    schemasContent += `/**
 * Get Likes Request Schema
 */
export const getLikesSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetLikesRequest = z.infer<typeof getLikesSchema>;

`;
  }

  // GetQuotesRequest Schema
  if (definitions['schemas.GetQuotesRequest']) {
    schemasContent += `/**
 * Get Quotes Request Schema
 */
export const getQuotesSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetQuotesRequest = z.infer<typeof getQuotesSchema>;

`;
  }

  // GetCommentsRequest Schema
  if (definitions['schemas.GetCommentsRequest']) {
    schemasContent += `/**
 * Get Comments Request Schema
 */
export const getCommentsSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetCommentsRequest = z.infer<typeof getCommentsSchema>;

`;
  }

  // GetRepostsRequest Schema
  if (definitions['schemas.GetRepostsRequest']) {
    schemasContent += `/**
 * Get Reposts Request Schema
 */
export const getRepostsSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetRepostsRequest = z.infer<typeof getRepostsSchema>;

`;
  }

  // TwitterLoginRequest Schema
  if (definitions['schemas.TwitterLoginRequest']) {
    schemasContent += `/**
 * Twitter Login Request Schema
 */
export const twitterLoginSchema = z.object({
  username: z.string().min(1, "Username is required"),
});

export type TwitterLoginRequest = z.infer<typeof twitterLoginSchema>;

`;
  }

  // Add Models section
  schemasContent += `
// =============================================================================
// Models
// =============================================================================

`;

  // ApiCallLog Model
  if (definitions['models.ApiCallLog']) {
    schemasContent += `/**
 * API Call Log Model
 */
export const apiCallLogSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  twitter_username: z.string(),
  endpoint: z.string(),
  method: z.string(),
  request_url: z.string(),
  status_code: z.number(),
  success: z.boolean(),
  error_message: z.string(),
  response_time: z.number(), // in milliseconds
  ip_address: z.string(),
  user_agent: z.string(),
  created_at: z.string(),
  user: z.object({
    id: z.string(),
    name: z.string(),
    email: z.string(),
    twitter_reqs: z.number(),
  }).optional(),
});

export type ApiCallLog = z.infer<typeof apiCallLogSchema>;

`;
  }

  // API Stats Response
  schemasContent += `/**
 * API Call Statistics Response
 */
export const apiStatsSchema = z.object({
  total_calls: z.number(),
  successful_calls: z.number(),
  failed_calls: z.number(),
  success_rate: z.number(),
  average_response_time: z.number(),
  calls_by_endpoint: z.record(z.number()),
});

export type ApiStats = z.infer<typeof apiStatsSchema>;

`;

  // Write schemas.ts file
  await Bun.write('./src/lib/schemas.ts', schemasContent);

  console.log(`✨ Generated: src/lib/schemas.ts`);
  console.log(`📦 Includes:`);
  console.log(`   ✅ Auth schemas (signup, login)`);
  console.log(`   ✅ Twitter schemas (account, tweets)`);
  console.log(`   ✅ Response schemas`);
  console.log(`   ✅ Model schemas (ApiCallLog)`);
  console.log(`   ✅ TypeScript types`);
  console.log(`   ✅ Zod validation schemas`);
  console.log(`🎉 Done! Import from '@/lib/schemas'`);

} catch (error) {
  console.error(`❌ Error generating schemas:`, error);
  process.exit(1);
}
