import { NextResponse } from "next/server";

export async function POST() {
  const response = NextResponse.json({ ok: true });
  for (const name of ["heard_session", "heard_oauth_state", "heard_oauth_verifier", "heard_oauth_return"]) {
    response.cookies.set({ name, value: "", path: "/", maxAge: 0, httpOnly: true, sameSite: "lax" });
    if (name !== "heard_session") response.cookies.set({ name, value: "", path: "/api/v1/auth", maxAge: 0, httpOnly: true, sameSite: "lax" });
  }
  return response;
}
