import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "PhotoTransfer LAN",
  description: "LAN photo and video transfer for Windows"
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}

