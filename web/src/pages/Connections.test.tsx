// Copyright 2026 Zyvor AI Labs · https://zyvor.dev
// SPDX-License-Identifier: LicenseRef-Zyvor-Production-1.0

import { renderToStaticMarkup } from 'react-dom/server';
import { expect, test } from 'vitest';
import { Connections } from './Traces';

test('host connections always name the QEMU process and not the guest', () => {
  const html = renderToStaticMarkup(<Connections />);
  expect(html).toContain('QEMU process');
  expect(html).toContain('not the guest');
  expect(html).toContain('guest_attributed is false');
});
