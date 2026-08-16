import { NextRequest } from "next/server";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

const FORWARDED_REQUEST_HEADERS = [
  "authorization",
  "content-type",
  "traceparent",
  "x-heard-tenant-id",
  "x-request-id"
];

async function proxyToHeardAPI(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  const apiBaseURL = process.env.API_INTERNAL_BASE_URL ?? "http://localhost:8080";
  const target = new URL(`/api/${path.map(encodeURIComponent).join("/")}`, apiBaseURL);
  target.search = request.nextUrl.search;

  const headers = new Headers();
  for (const name of FORWARDED_REQUEST_HEADERS) {
    const value = request.headers.get(name);
    if (value) {
      headers.set(name, value);
    }
  }

  try {
    const upstream = await fetch(target, {
      method: request.method,
      headers,
      body: request.method === "GET" || request.method === "HEAD" ? undefined : await request.arrayBuffer(),
      cache: "no-store",
      redirect: "manual"
    });
    const responseHeaders = new Headers();
    for (const name of ["content-type", "location", "traceparent", "x-request-id"]) {
      const value = upstream.headers.get(name);
      if (value) {
        responseHeaders.set(name, value);
      }
    }
    return new Response(upstream.body, {
      status: upstream.status,
      headers: responseHeaders
    });
  } catch (error) {
    console.error("heard api gateway unavailable", {
      path: request.nextUrl.pathname,
      error: error instanceof Error ? error.message : "unknown error"
    });
    return Response.json(
      { error: { message: "heard API is temporarily unavailable", status: 502 } },
      { status: 502 }
    );
  }
}

export const GET = proxyToHeardAPI;
export const POST = proxyToHeardAPI;
export const PUT = proxyToHeardAPI;
export const PATCH = proxyToHeardAPI;
export const DELETE = proxyToHeardAPI;
export const OPTIONS = proxyToHeardAPI;
