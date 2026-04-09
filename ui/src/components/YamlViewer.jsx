import React, { useMemo } from 'react';

export default function YamlViewer({ text }) {
  const lines = useMemo(() => (text || '').split('\n'), [text]);
  return (
    <pre className="mono-block yaml-block">
      {lines.map((line, idx) => {
        const keyMatch = line.match(/^(\s*)([A-Za-z0-9_.-]+):(.*)$/);
        if (line.trim().startsWith('#')) {
          return <div key={idx} className="yaml-comment">{line}</div>;
        }
        if (keyMatch) {
          const [, indent, key, rest] = keyMatch;
          return (
            <div key={idx}>
              <span>{indent}</span>
              <span className="yaml-key">{key}</span>
              <span className="yaml-colon">:</span>
              <span className="yaml-value">{rest}</span>
            </div>
          );
        }
        return <div key={idx}>{line}</div>;
      })}
    </pre>
  );
}
