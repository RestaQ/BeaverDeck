import React from 'react';

export default function AuditPage({ audit }) {
  return (
    <pre className="mono-block">
      {audit.map((a) => `${a.time} ${a.action} ${a.resource}/${a.name} ns=${a.namespace} dryRun=${a.dry_run} ok=${a.success} ${a.message}`).join('\n')}
    </pre>
  );
}
