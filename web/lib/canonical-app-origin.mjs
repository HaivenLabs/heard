export const LOCAL_CANONICAL_APP_ORIGIN = "http://localhost:3010";

const localRuntimeEnvironments = new Set(["development", "local", "docker", "test"]);

export function canonicalOrigin(value) {
  let parsed;
  try {
    parsed = new URL(value);
  } catch {
    throw new Error("PUBLIC_APP_URL must be an absolute HTTP(S) origin");
  }

  if (
    (parsed.protocol !== "http:" && parsed.protocol !== "https:") ||
    !parsed.host ||
    parsed.username ||
    parsed.password ||
    (parsed.pathname !== "" && parsed.pathname !== "/") ||
    parsed.search ||
    parsed.hash
  ) {
    throw new Error("PUBLIC_APP_URL must be an absolute HTTP(S) origin without a path, query, or fragment");
  }

  return parsed.origin;
}

export function configuredCanonicalAppOrigin(environment, publicAppURL) {
  const configured = publicAppURL?.trim();
  if (configured) return canonicalOrigin(configured);
  if (localRuntimeEnvironments.has(environment)) return LOCAL_CANONICAL_APP_ORIGIN;
  throw new Error("PUBLIC_APP_URL is required outside local development");
}

export function canonicalRedirectURL(requestURL, origin) {
  const request = new URL(requestURL);
  const canonical = new URL(origin);
  if (request.origin === canonical.origin) return null;
  return new URL(`${request.pathname}${request.search}`, canonical);
}

export function browserFacingRequestURL(requestURL, host, forwardedProtocol) {
  const request = new URL(requestURL);
  const browserHost = host?.split(",")[0]?.trim();
  const protocol = forwardedProtocol?.split(",")[0]?.trim().toLowerCase();
  if (!browserHost) return request;
  if (protocol === "http" || protocol === "https") request.protocol = `${protocol}:`;
  request.host = browserHost;
  return request;
}
