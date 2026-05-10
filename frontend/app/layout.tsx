import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Mycelo Console",
  description: "Operator console for Mycelo streams, destinations, delivery state, and DLQ replay.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
