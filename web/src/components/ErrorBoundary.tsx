import { AlertTriangle, RotateCcw } from "lucide-react";
import { Component, type ErrorInfo, type ReactNode } from "react";

export class ErrorBoundary extends Component<
  { children: ReactNode },
  { error: Error | null }
> {
  state = { error: null as Error | null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("Disk Insight interface error", error, info);
  }

  render() {
    if (!this.state.error) return this.props.children;
    return (
      <main className="fatal-error">
        <div className="fatal-error-card">
          <span className="fatal-error-icon">
            <AlertTriangle size={28} />
          </span>
          <p>Interface recovery / 界面恢复</p>
          <h1>页面组件发生错误</h1>
          <code>{this.state.error.message}</code>
          <button type="button" onClick={() => window.location.reload()}>
            <RotateCcw size={17} />
            重新加载
          </button>
        </div>
      </main>
    );
  }
}
