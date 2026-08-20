"use client";

import { createContext, ReactNode, useContext } from "react";
import { Session } from "../lib/api";

const AdminSessionContext = createContext<Session | null>(null);

export function AdminSessionProvider({ children, session }: { children: ReactNode; session: Session }) {
  return <AdminSessionContext.Provider value={session}>{children}</AdminSessionContext.Provider>;
}

export function useAdminSession() {
  const session = useContext(AdminSessionContext);
  if (!session) {
    throw new Error("useAdminSession must be used within the admin layout");
  }
  return session;
}
