import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "heard · Guest feedback your team can act on",
  description: "Collect private restaurant feedback, recover unhappy guests, and turn every response into operational insight."
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
