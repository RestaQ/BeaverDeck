import React from 'react';
import LogViewer from './LogViewer.jsx';

export default function WarningPopover({ warningPopover, cancelHide, scheduleHide }) {
  if (!warningPopover) {
    return null;
  }

  return (
    <div
      className="warning-popover"
      style={{ top: `${warningPopover.top}px`, left: `${warningPopover.left}px` }}
      onMouseEnter={cancelHide}
      onMouseLeave={() => scheduleHide(warningPopover.key)}
    >
      <div className="warning-popover-title">{warningPopover.title}</div>
      {warningPopover.loading ? (
        <pre className="warning-popover-body">Loading...</pre>
      ) : warningPopover.title?.toLowerCase().includes('log') ? (
        <LogViewer text={warningPopover.text} className="warning-popover-body" />
      ) : (
        <pre className="warning-popover-body">{warningPopover.text}</pre>
      )}
    </div>
  );
}
