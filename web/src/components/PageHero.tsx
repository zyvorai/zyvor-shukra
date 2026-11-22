// Copyright 2026 Zyvor AI Labs · https://zyvor.dev
// SPDX-License-Identifier: LicenseRef-Zyvor-Production-2.0

export type HeroTint = 'green' | 'amber' | 'purple' | 'red';

export default function PageHero({
  eyebrow,
  title,
  lede,
  tint,
}: {
  eyebrow: string;
  title: string;
  lede: string;
  tint?: HeroTint;
}) {
  return (
    <header className={tint ? `page-hero hero-tint-${tint}` : 'page-hero'}>
      <p className="eyebrow">{eyebrow}</p>
      <h1>{title}</h1>
      <p>{lede}</p>
    </header>
  );
}
