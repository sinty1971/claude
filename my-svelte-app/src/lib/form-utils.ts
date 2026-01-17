/**
 * Enterキーが押された際に、次のフォーカス可能な要素に移動するハンドラー
 * @param e キーボードイベント
 */
export function handleEnterKeyNavigation(
  e: KeyboardEvent & { currentTarget: EventTarget & (HTMLInputElement | HTMLSelectElement) }
): void {
  if (e.key === 'Enter') {
    e.preventDefault();
    const form = e.currentTarget.form;
    if (form) {
      const elements = Array.from(form.elements);
      const index = elements.indexOf(e.currentTarget);
      const nextElement = elements[index + 1] as HTMLElement | undefined;
      if (nextElement && 'focus' in nextElement) {
        nextElement.focus();
      }
    }
  }
}
