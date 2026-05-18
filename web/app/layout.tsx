import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Heard",
  description: "Flyer survey campaigns, guest feedback, and recovery workflows."
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
