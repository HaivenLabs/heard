import { NextResponse, type NextRequest } from "next/server";
import { browserFacingRequestURL, canonicalRedirectURL, configuredCanonicalAppOrigin } from "./lib/canonical-app-origin.mjs";

const runtimeEnvironment = process.env.APP_ENV ?? process.env.NEXT_PUBLIC_APP_ENV ?? process.env.NODE_ENV;
const canonicalAppOrigin = configuredCanonicalAppOrigin(runtimeEnvironment, process.env.PUBLIC_APP_URL);

export function middleware(request: NextRequest) {
  const browserRequestURL = browserFacingRequestURL(
    request.nextUrl.toString(),
    request.headers.get("x-forwarded-host") ?? request.headers.get("host"),
    request.headers.get("x-forwarded-proto")
  );
  const redirect = canonicalRedirectURL(browserRequestURL, canonicalAppOrigin);
  return redirect ? NextResponse.redirect(redirect, 307) : NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"]
};
