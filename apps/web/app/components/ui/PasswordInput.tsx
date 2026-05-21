"use client";

import { useState, forwardRef } from "react";
import { Eye, EyeOff } from "lucide-react";

interface PasswordInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  inputClassName?: string;
}

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  function PasswordInput({ className, inputClassName, ...props }, ref) {
    const [visible, setVisible] = useState(false);

    return (
      <div className={`password-input-wrapper${className ? ` ${className}` : ""}`}>
        <input
          ref={ref}
          type={visible ? "text" : "password"}
          className={`password-input-field${inputClassName ? ` ${inputClassName}` : ""}`}
          {...props}
        />
        <button
          type="button"
          className="password-input-toggle"
          onClick={() => setVisible((v) => !v)}
          aria-label={visible ? "Hide password" : "Show password"}
          tabIndex={-1}
        >
          {visible ? <EyeOff size={16} /> : <Eye size={16} />}
        </button>
      </div>
    );
  },
);
