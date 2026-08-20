import { NextRequest } from "next/server";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

const FORWARDED_REQUEST_HEADERS = [
  "authorization",
  "cookie",
  "content-type",
  "idempotency-key",
  "traceparent",
  "x-heard-tenant-id",
  "x-request-id"
];

const MAX_PROXY_REQUEST_BODY_BYTES = 64 * 1024;

class ProxyBodyTooLargeError extends Error {}

async function readBoundedRequestBody(request: NextRequest): Promise<ArrayBuffer | undefined> {
  if (request.method === "GET" || request.method === "HEAD" || !request.body) return undefined;
  const contentLength = Number(request.headers.get("content-length") ?? "0");
  if (Number.isFinite(contentLength) && contentLength > MAX_PROXY_REQUEST_BODY_BYTES) {
    throw new ProxyBodyTooLargeError("request body exceeds gateway limit");
  }

  const reader = request.body.getReader();
  const chunks: Uint8Array[] = [];
  let total = 0;
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    total += value.byteLength;
    if (total > MAX_PROXY_REQUEST_BODY_BYTES) {
      await reader.cancel();
      throw new ProxyBodyTooLargeError("request body exceeds gateway limit");
    }
    chunks.push(value);
  }
  const body = new Uint8Array(total);
  let offset = 0;
  for (const chunk of chunks) {
    body.set(chunk, offset);
    offset += chunk.byteLength;
  }
  return body.buffer;
}

async function proxyToHeardAPI(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  const apiBaseURL = process.env.API_INTERNAL_BASE_URL ?? "http://localhost:8082";
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
    const body = await readBoundedRequestBody(request);
    const upstream = await fetch(target, {
      method: request.method,
      headers,
      body,
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
    const upstreamHeaders = upstream.headers as Headers & { getSetCookie?: () => string[] };
    const setCookies = upstreamHeaders.getSetCookie?.() ?? [];
    if (setCookies.length > 0) {
      for (const cookie of setCookies) responseHeaders.append("set-cookie", cookie);
    } else {
      const setCookie = upstream.headers.get("set-cookie");
      if (setCookie) responseHeaders.set("set-cookie", setCookie);
    }
    return new Response(upstream.body, {
      status: upstream.status,
      headers: responseHeaders
    });
  } catch (error) {
    if (error instanceof ProxyBodyTooLargeError) {
      return Response.json(
        { error: { message: "request body exceeds 64 KiB", status: 413 } },
        { status: 413 }
      );
    }
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
