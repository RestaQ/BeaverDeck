import React, { useState } from 'react';

function EyeIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className="password-eye-icon">
      <path
        d="M2 12c2.6-4.2 6.1-6.3 10-6.3S19.4 7.8 22 12c-2.6 4.2-6.1 6.3-10 6.3S4.6 16.2 2 12Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
      />
      <circle cx="12" cy="12" r="3.1" fill="none" stroke="currentColor" strokeWidth="1.8" />
    </svg>
  );
}

export default function PasswordField({ value, onChange, placeholder, className = '', inputClassName = '', ...rest }) {
  const [revealed, setRevealed] = useState(false);

  const show = (event) => {
    event?.preventDefault?.();
    setRevealed(true);
  };
  const hide = (event) => {
    event?.preventDefault?.();
    setRevealed(false);
  };

  return (
    <div className={`password-field ${className}`.trim()}>
      <input
        {...rest}
        type={revealed ? 'text' : 'password'}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        className={inputClassName}
      />
      <button
        type="button"
        className="password-eye-button"
        aria-label={revealed ? 'Hide value' : 'Show value'}
        aria-pressed={revealed}
        onMouseDown={show}
        onMouseUp={hide}
        onMouseLeave={hide}
        onTouchStart={show}
        onTouchEnd={hide}
        onTouchCancel={hide}
        onKeyDown={(event) => {
          if (event.key === ' ' || event.key === 'Enter') show(event);
        }}
        onKeyUp={(event) => {
          if (event.key === ' ' || event.key === 'Enter') hide(event);
        }}
      >
        <EyeIcon />
      </button>
    </div>
  );
}
