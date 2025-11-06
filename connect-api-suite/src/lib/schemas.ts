/**
 * Auto-generated Zod schemas from OpenAPI spec
 * Generated: 2025-11-06T10:49:28.314Z
 */

import { z } from "zod";

// =============================================================================
// Auth Schemas
// =============================================================================

/**
 * Signup Request Schema
 */
export const signupSchema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Invalid email address"),
  password: z.string().min(6, "Password must be at least 6 characters"),
});

export type SignupRequest = z.infer<typeof signupSchema>;

/**
 * Signup Form Schema (with password confirmation)
 */
export const signupFormSchema = signupSchema.extend({
  confirmPassword: z.string().min(6, "Password must be at least 6 characters"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

export type SignupFormData = z.infer<typeof signupFormSchema>;

/**
 * Login Request Schema
 */
export const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(1, "Password is required"),
});

export type LoginRequest = z.infer<typeof loginSchema>;

/**
 * Login Response Schema
 */
export const loginResponseSchema = z.object({
  token: z.string(),
  email: z.string(),
});

export type LoginResponse = z.infer<typeof loginResponseSchema>;

/**
 * Message Response Schema
 */
export const messageResponseSchema = z.object({
  message: z.string(),
});

export type MessageResponse = z.infer<typeof messageResponseSchema>;

/**
 * Error Response Schema
 */
export const errorResponseSchema = z.object({
  error: z.string(),
});

export type ErrorResponse = z.infer<typeof errorResponseSchema>;

/**
 * Change Password Request Schema
 */
export const changePasswordSchema = z.object({
  current_password: z.string().min(1, "Current password is required"),
  new_password: z.string().min(6, "New password must be at least 6 characters"),
});

export type ChangePasswordRequest = z.infer<typeof changePasswordSchema>;

/**
 * Change Password Form Schema (with confirmation)
 */
export const changePasswordFormSchema = changePasswordSchema.extend({
  confirm_password: z.string().min(6, "Password must be at least 6 characters"),
}).refine((data) => data.new_password === data.confirm_password, {
  message: "Passwords don't match",
  path: ["confirm_password"],
});

export type ChangePasswordFormData = z.infer<typeof changePasswordFormSchema>;

// =============================================================================
// Twitter Schemas
// =============================================================================

/**
 * Twitter Account Request Schema
 */
export const twitterAccountSchema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

export type TwitterAccountRequest = z.infer<typeof twitterAccountSchema>;

/**
 * Get Tweets Request Schema
 */
export const getTweetsSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetTweetsRequest = z.infer<typeof getTweetsSchema>;

/**
 * Get Likes Request Schema
 */
export const getLikesSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetLikesRequest = z.infer<typeof getLikesSchema>;

/**
 * Get Quotes Request Schema
 */
export const getQuotesSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetQuotesRequest = z.infer<typeof getQuotesSchema>;

/**
 * Get Comments Request Schema
 */
export const getCommentsSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetCommentsRequest = z.infer<typeof getCommentsSchema>;

/**
 * Get Reposts Request Schema
 */
export const getRepostsSchema = z.object({
  url: z.string().url("Invalid URL").refine(
    (url) => url.includes("twitter.com") || url.includes("x.com"),
    "Must be a Twitter/X URL"
  ),
});

export type GetRepostsRequest = z.infer<typeof getRepostsSchema>;

/**
 * Twitter Login Request Schema
 */
export const twitterLoginSchema = z.object({
  username: z.string().min(1, "Username is required"),
});

export type TwitterLoginRequest = z.infer<typeof twitterLoginSchema>;


// =============================================================================
// Models
// =============================================================================

/**
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

/**
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

