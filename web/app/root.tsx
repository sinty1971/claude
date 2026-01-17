import type { LinksFunction, MetaFunction } from "@remix-run/node";
import { Links, Meta, Outlet, Scripts, ScrollRestoration } from "@remix-run/react";
import { Navigation } from "@/components/Navigation";
import { Providers } from "./providers";
import stylesheet from "./globals.css?url";

export const links: LinksFunction = () => [
  { rel: "stylesheet", href: stylesheet },
];

export const meta: MetaFunction = () => [
  { title: "Penguin フォルダー管理" },
  {
    name: "description",
    content: "ファイルと工事プロジェクトの情報管理を支援する社内向けダッシュボード",
  },
];

export default function App() {
  return (
    <html lang="ja">
      <head>
        <Meta />
        <Links />
      </head>
      <body>
        <Providers>
          <div className="app">
            <Navigation />
            <main className="main-content">
              <Outlet />
            </main>
          </div>
        </Providers>
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}
