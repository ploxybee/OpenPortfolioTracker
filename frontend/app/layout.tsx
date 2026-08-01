import "./globals.css";

export const metadata = {
  title: "Open Portfolio Tracker",
  description: "Your portfolio, in perspective.",
};
// RootLayout provides the shared document structure for every page.
export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
