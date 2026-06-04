import type { ReactNode } from "react"

interface Props {
  children: ReactNode
  params: Promise<{ locale: string }>
}

export default async function LocaleLayout({ children }: Props): Promise<React.JSX.Element> {
  return (
    <html>
      <body>{children}</body>
    </html>
  )
}
