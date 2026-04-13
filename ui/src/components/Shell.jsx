export function PageShell({ title, subtitle, actions, children }) {
  return (
    <section className="space-y-6">
      <header className="flex flex-col gap-4 rounded-3xl border border-slate-200 bg-white/90 p-6 shadow-sm shadow-slate-200/60 lg:flex-row lg:items-end lg:justify-between">
        <div className="space-y-2">
          <p className="text-xs font-semibold uppercase tracking-[0.28em] text-sky-700">
            Remote Control
          </p>
          <h1 className="text-3xl font-semibold tracking-tight text-slate-900">{title}</h1>
          {subtitle ? <p className="max-w-2xl text-sm text-slate-600">{subtitle}</p> : null}
        </div>
        {actions ? <div className="flex flex-wrap gap-3">{actions}</div> : null}
      </header>
      {children}
    </section>
  );
}

export function Panel({ className = "", children }) {
  return (
    <div className={`rounded-3xl border border-slate-200 bg-white p-6 shadow-sm shadow-slate-200/60 ${className}`}>
      {children}
    </div>
  );
}
