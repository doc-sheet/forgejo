// @ts-check
import {test, expect} from '@playwright/test';
import {login_user, save_visual, load_logged_in_context} from './utils_e2e.js';

test.beforeAll(async ({browser}, workerInfo) => {
  await login_user(browser, workerInfo, 'user2');
});

test('Test Markdown Indentation', async ({browser}, workerInfo) => {
  const context = await load_logged_in_context(browser, workerInfo, 'user2');

  const initText = `* first\n* second\n* third\n* last`;

  const page = await context.newPage();
  const response = await page.goto('/user2/repo1/issues/new');
  await expect(response?.status()).toBe(200);

  const textarea = page.locator('textarea[name=content]');
  const tab = '    ';
  await textarea.fill(initText);
  await textarea.click(); // Tab handling is disabled until pointer event or input.

  // Indent, then unindent first line
  await textarea.evaluate((it) => it.setSelectionRange(0, 0));
  await textarea.press('Tab');
  await expect(textarea).toHaveValue(`${tab}* first\n* second\n* third\n* last`);
  await textarea.press('Shift+Tab');
  await expect(textarea).toHaveValue(initText);

  // Indent second line while somewhere inside of it
  await textarea.press('ArrowDown');
  await textarea.press('ArrowRight');
  await textarea.press('ArrowRight');
  await textarea.press('Tab');
  await expect(textarea).toHaveValue(`* first\n${tab}* second\n* third\n* last`);

  // Subsequently, select a chunk of 2nd and 3rd line and indent both, preserving the cursor position in relation to text
  await textarea.evaluate((it) => it.setSelectionRange(it.value.indexOf('cond'), it.value.indexOf('hird')));
  await textarea.press('Tab');
  const lines23 = `* first\n${tab}${tab}* second\n${tab}* third\n* last`;
  await expect(textarea).toHaveValue(lines23);
  await expect(textarea).toHaveJSProperty('selectionStart', lines23.indexOf('cond'));
  await expect(textarea).toHaveJSProperty('selectionEnd', lines23.indexOf('hird'));

  // Then unindent twice, erasing all indents.
  await textarea.press('Shift+Tab');
  await expect(textarea).toHaveValue(`* first\n${tab}* second\n* third\n* last`);
  await textarea.press('Shift+Tab');
  await expect(textarea).toHaveValue(initText);

  // Indent and unindent with cursor at the end of the line
  await textarea.evaluate((it) => it.setSelectionRange(it.value.indexOf('cond'), it.value.indexOf('cond')));
  await textarea.press('End');
  await textarea.press('Tab');
  await expect(textarea).toHaveValue(`* first\n${tab}* second\n* third\n* last`);
  await textarea.press('Shift+Tab');
  await expect(textarea).toHaveValue(initText);

  // Ensure textarea is blurred on Esc, and does not intercept Tab before input
  await textarea.press('Escape');
  await expect(textarea).not.toBeFocused();
  await textarea.focus();
  await textarea.press('Tab');
  await expect(textarea).toHaveValue(initText);
  await expect(textarea).not.toBeFocused(); // because tab worked as normal

  // Check that Tab does work after input
  await textarea.focus();
  await textarea.evaluate((it) => it.setSelectionRange(it.value.length, it.value.length));
  await textarea.press('Shift+Enter'); // Avoid triggering the prefix continuation feature
  await textarea.pressSequentially('* least');
  await textarea.press('Tab');
  await expect(textarea).toHaveValue(`* first\n* second\n* third\n* last\n${tab}* least`);

  // Check that partial indents are cleared
  await textarea.fill(initText);
  await textarea.evaluate((it) => it.setSelectionRange(it.value.indexOf('* second'), it.value.indexOf('* second')));
  await textarea.pressSequentially('  ');
  await textarea.press('Shift+Tab');
  await expect(textarea).toHaveValue(initText);
});
