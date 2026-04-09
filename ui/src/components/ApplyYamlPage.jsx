import React from 'react';

export default function ApplyYamlPage({
  selectedTemplate,
  loadTemplate,
  applyTemplates,
  yamlText,
  setYamlText,
  safe,
  applyYaml,
  primaryNamespace,
  permissionInfo
}) {
  const applyPermission = [
    { allowed: Boolean(primaryNamespace), reason: 'Select namespace first' },
    permissionInfo('apply', 'edit', primaryNamespace)
  ].find((item) => !item.allowed) || { allowed: true, reason: '' };

  return (
    <>
      <div className="toolbar fixed-toolbar">
        <select value={selectedTemplate} onChange={(e) => loadTemplate(e.target.value)}>
          <option value="">Load template...</option>
          {Object.keys(applyTemplates).map((name) => (
            <option key={name} value={name}>{name}</option>
          ))}
        </select>
      </div>
      <textarea
        className="code-textarea"
        rows={12}
        value={yamlText}
        onChange={(e) => setYamlText(e.target.value)}
        placeholder={'---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: demo'}
      />
      <div className="toolbar compact fixed-toolbar">
        <button className="warn" onClick={() => safe(() => applyYaml(true))} disabled={!applyPermission.allowed} title={applyPermission.reason}>
          Dry-run
        </button>
        <button onClick={() => safe(() => applyYaml(false))} disabled={!applyPermission.allowed} title={applyPermission.reason}>
          Apply
        </button>
      </div>
    </>
  );
}
