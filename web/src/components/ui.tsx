import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import type {
  ButtonHTMLAttributes,
  HTMLAttributes,
  PropsWithChildren,
  ReactNode,
} from "react";
import { cn } from "../lib/cn";

export function Button({
  className,
  variant = "primary",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost" | "danger";
}) {
  return (
    <button
      className={cn("button", `button-${variant}`, className)}
      {...props}
    />
  );
}

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <section className={cn("card", className)} {...props} />;
}

export function CardHeader({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <header className="card-header">
      <div>
        <h2>{title}</h2>
        {description && <p>{description}</p>}
      </div>
      {action && <div className="card-action">{action}</div>}
    </header>
  );
}

export function Segmented<T extends string>({
  value,
  options,
  onChange,
  label,
}: {
  value: T;
  options: Array<{ value: T; label: string }>;
  onChange: (value: T) => void;
  label: string;
}) {
  return (
    <fieldset className="segmented">
      <legend className="sr-only">{label}</legend>
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          aria-pressed={value === option.value}
          className={value === option.value ? "active" : ""}
          onClick={() => onChange(option.value)}
        >
          {option.label}
        </button>
      ))}
    </fieldset>
  );
}

export function Tooltip({
  content,
  children,
}: PropsWithChildren<{ content: ReactNode }>) {
  return (
    <TooltipPrimitive.Provider delayDuration={300}>
      <TooltipPrimitive.Root>
        <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content className="tooltip" sideOffset={8}>
            {content}
            <TooltipPrimitive.Arrow className="tooltip-arrow" />
          </TooltipPrimitive.Content>
        </TooltipPrimitive.Portal>
      </TooltipPrimitive.Root>
    </TooltipPrimitive.Provider>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={cn("skeleton", className)} aria-hidden="true" />;
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="empty-state">
      <div className="empty-orbit">
        <span />
      </div>
      <h2>{title}</h2>
      <p>{description}</p>
      {action}
    </div>
  );
}

export function ErrorState({
  error,
  onRetry,
  retryLabel,
}: {
  error: Error;
  onRetry: () => void;
  retryLabel: string;
}) {
  return (
    <div className="error-state" role="alert">
      <strong>{error.message}</strong>
      <Button variant="secondary" onClick={onRetry}>
        {retryLabel}
      </Button>
    </div>
  );
}
