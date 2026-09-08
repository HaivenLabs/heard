export declare const LOCAL_CANONICAL_APP_ORIGIN: string;
export declare function canonicalOrigin(value: string): string;
export declare function configuredCanonicalAppOrigin(environment: string, publicAppURL?: string): string;
export declare function canonicalRedirectURL(requestURL: string | URL, origin: string): URL | null;
export declare function browserFacingRequestURL(requestURL: string | URL, host?: string | null, forwardedProtocol?: string | null): URL;
