import { redirect } from "next/navigation";
import type { Route } from "next";

export default function SetupPage() {
  redirect("/onboarding" as Route);
}
