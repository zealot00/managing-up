'use client';

import {useState, useEffect} from 'react';
import {useLocale} from 'next-intl';

export default function LanguageSwitcher() {
  const locale = useLocale();
  const [current, setCurrent] = useState(locale);

  useEffect(() => {
    setCurrent(locale);
  }, [locale]);

  function switchLocale(newLocale: string) {
    document.cookie = `NEXT_LOCALE=${newLocale};path=/;max-age=31536000;samesite=lax`;
    window.location.reload();
  }

  return (
    <div className="lang-switcher">
      <button
        className={`lang-btn ${current === 'zh' ? 'lang-btn-active' : ''}`}
        onClick={() => switchLocale('zh')}
        title="中文"
      >
        中文
      </button>
      <span className="lang-separator">/</span>
      <button
        className={`lang-btn ${current === 'en' ? 'lang-btn-active' : ''}`}
        onClick={() => switchLocale('en')}
        title="English"
      >
        EN
      </button>
    </div>
  );
}
