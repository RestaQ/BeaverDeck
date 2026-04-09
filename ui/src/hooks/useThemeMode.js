import { useEffect, useMemo, useState } from 'react';
import { detectSystemTheme } from '../lib/appUtils.js';

export default function useThemeMode() {
  const [themePreference, setThemePreference] = useState('auto');
  const [systemTheme, setSystemTheme] = useState(detectSystemTheme());

  const resolvedTheme = useMemo(
    () => (themePreference === 'auto' ? systemTheme : themePreference),
    [themePreference, systemTheme]
  );

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return undefined;
    }
    const media = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = () => {
      setSystemTheme(media.matches ? 'dark' : 'light');
    };
    handleChange();
    if (typeof media.addEventListener === 'function') {
      media.addEventListener('change', handleChange);
      return () => media.removeEventListener('change', handleChange);
    }
    media.addListener(handleChange);
    return () => media.removeListener(handleChange);
  }, []);

  useEffect(() => {
    document.body.dataset.theme = resolvedTheme;
    document.documentElement.dataset.theme = resolvedTheme;
    document.documentElement.style.colorScheme = resolvedTheme;
    return () => {
      delete document.body.dataset.theme;
      delete document.documentElement.dataset.theme;
      document.documentElement.style.colorScheme = '';
    };
  }, [resolvedTheme]);

  return {
    themePreference,
    setThemePreference,
    resolvedTheme
  };
}
