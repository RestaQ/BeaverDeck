import React, { useMemo } from 'react';

const WARN_RE = /\bwarn(?:ing)?\b/i;
const ERR_RE = /\berr(?:or)?\b/i;

export default function LogViewer({
  text,
  className = 'mono-block',
  search = '',
  showWarnings = false,
  showErrors = false
}) {
  const lines = useMemo(() => String(text || '').split('\n'), [text]);
  const query = search.trim().toLowerCase();
  const filteredLines = useMemo(() => {
    const useSeverityFilter = showWarnings || showErrors;
    return lines.filter((line) => {
      const matchesSearch = !query || line.toLowerCase().includes(query);
      if (!matchesSearch) return false;
      if (!useSeverityFilter) return true;
      const isError = ERR_RE.test(line);
      const isWarning = WARN_RE.test(line);
      return (showErrors && isError) || (showWarnings && isWarning);
    });
  }, [lines, query, showWarnings, showErrors]);

  return (
    <div className={`${className} log-block`}>
      {filteredLines.map((line, idx) => {
        let lineClass = '';
        if (ERR_RE.test(line)) {
          lineClass = 'log-line-error';
        } else if (WARN_RE.test(line)) {
          lineClass = 'log-line-warn';
        }
        return <div key={idx} className={lineClass}>{line || '\u00A0'}</div>;
      })}
    </div>
  );
}
