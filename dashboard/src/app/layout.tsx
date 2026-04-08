import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Tribunal Security Dashboard",
  description: "Enterprise-grade CI/CD code analysis and risk mitigation platform.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased dark`}
      style={{ colorScheme: 'dark' }}
    >
      <body className="min-h-full flex flex-col bg-[#020617] text-gray-100 font-sans selection:bg-indigo-500/30 selection:text-indigo-200">
        <div className="flex h-screen overflow-hidden">
          {/* Add a sleek subtle radial gradient in background */}
          <div className="absolute top-0 left-0 w-full h-125 bg-[radial-gradient(ellipse_at_top,var(--tw-gradient-stops))] from-indigo-900/20 via-[#020617] to-transparent pointer-events-none -z-10" />
          {children}
        </div>
      </body>
    </html>
  );
}
