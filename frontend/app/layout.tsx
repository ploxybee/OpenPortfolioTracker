import "./globals.css";

export const metadata = {
  title: "Open Portfolio Tracker",
  description: "Your portfolio, in perspective.",
};
export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
