'use client';

import { useState, useEffect } from 'react';
import { useLocale, useTranslations } from 'next-intl';

export default function LanguageSwitcher() {
  const locale = useLocale();
  const t = useTranslations('languageSwitcher');
  const [current, setCurrent] = useState(locale);
  const [showWarning, setShowWarning] = useState(false);
  const [pendingLocale, setPendingLocale] = useState<string | null>(null);

  useEffect(() => {
    setCurrent(locale);
  }, [locale]);

  function handleClick(newLocale: string) {
    if (newLocale === current) return;
    setPendingLocale(newLocale);
    setShowWarning(true);
  }

  function confirmSwitch() {
    if (!pendingLocale) return;
    document.cookie = `NEXT_LOCALE=${pendingLocale};path=/;max-age=31536000;samesite=lax`;
    window.location.reload();
  }

  function cancelSwitch() {
    setShowWarning(false);
    setPendingLocale(null);
  }

  return (
    <div className="lang-switcher">
      <button
        className={`lang-btn ${current === 'zh' ? 'lang-btn-active' : ''}`}
        onClick={() => handleClick('zh')}
        title="中文"
      >
        中文
      </button>
      <span className="lang-separator">/</span>
      <button
        className={`lang-btn ${current === 'en' ? 'lang-btn-active' : ''}`}
        onClick={() => handleClick('en')}
        title="English"
      >
        EN
      </button>
      {showWarning && (
        <div className="lang-switcher-warning" role="alert">
          <span className="lang-switcher-warning-text">{t('reloadWarning')}</span>
          <button className="lang-switcher-warning-btn" onClick={confirmSwitch}>OK</button>
          <button className="lang-switcher-warning-btn" onClick={cancelSwitch}>Cancel</button>
        </div>
      )}
    </div>
  );
}
