import { makeApi, Zodios, type ZodiosOptions } from "@zodios/core";
import { z } from "zod";

const endpoints = makeApi([
  {
    method: "post",
    path: "/twitter/account",
    alias: "postTwitteraccount",
    description: `Add a Twitter account for data extraction (requires JWT authentication)`,
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
      {
        status: 500,
        description: `Internal Server Error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/twitter/post",
    alias: "postTwitterpost",
    description: `Fetch tweet data including media (requires Twitter token authentication)`,
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
      {
        status: 429,
        description: `Too Many Requests`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal Server Error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/twitter/post/comments",
    alias: "postTwitterpostcomments",
    description: `Fetch comments/replies for a tweet (requires Twitter token authentication)`,
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
      {
        status: 429,
        description: `Too Many Requests`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal Server Error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/twitter/post/likes",
    alias: "postTwitterpostlikes",
    description: `Fetch users who liked a tweet (requires Twitter token authentication)`,
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
      {
        status: 429,
        description: `Too Many Requests`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal Server Error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/twitter/post/quotes",
    alias: "postTwitterpostquotes",
    description: `Fetch quote tweets for a tweet (requires Twitter token authentication)`,
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
      {
        status: 429,
        description: `Too Many Requests`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal Server Error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/twitter/post/reposts",
    alias: "postTwitterpostreposts",
    description: `Fetch users who reposted a tweet (requires Twitter token authentication)`,
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
      {
        status: 429,
        description: `Too Many Requests`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal Server Error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/twitter/regenerate-token",
    alias: "postTwitterregenerateToken",
    description: `Generate a new authentication token for a Twitter account (requires JWT authentication)`,
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
      {
        status: 404,
        description: `Not Found`,
        schema: z.void(),
      },
    ],
  },
]);

export const TwitterApi = new Zodios(endpoints);

export function createApiClient(baseUrl: string, options?: ZodiosOptions) {
  return new Zodios(baseUrl, endpoints, options);
}
