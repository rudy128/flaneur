import { makeApi, Zodios, type ZodiosOptions } from "@zodios/core";
import { z } from "zod";

const endpoints = makeApi([
  {
    method: "post",
    path: "/auth/login",
    alias: "postAuthlogin",
    description: `Authenticate user and return JWT token`,
    requestFormat: "json",
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Bad Request`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/auth/signup",
    alias: "postAuthsignup",
    description: `Register a new user account`,
    requestFormat: "json",
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Bad Request`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Conflict`,
        schema: z.void(),
      },
    ],
  },
]);

export const AuthApi = new Zodios(endpoints);

export function createApiClient(baseUrl: string, options?: ZodiosOptions) {
  return new Zodios(baseUrl, endpoints, options);
}
