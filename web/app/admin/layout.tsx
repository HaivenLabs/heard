"use client";

import { ReactNode } from "react";
import { AdminSessionProvider } from "../../components/admin-session";
import { AdminShell } from "../../components/admin-shell";
import { AuthGate } from "../../components/auth-gate";

export default function AdminLayout({ children }: { children: ReactNode }) {
  return (
    <div data-haiven-product="heard">
    <AuthGate requireOnboardingComplete>
      {(session) => (
        <AdminSessionProvider session={session}>
          <AdminShell session={session}>{children}</AdminShell>
        </AdminSessionProvider>
      )}
    </AuthGate>
    </div>
  );
}
