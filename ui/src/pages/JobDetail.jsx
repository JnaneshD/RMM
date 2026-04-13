import { Link, useParams } from "react-router-dom";
import StatusBadge from "../components/StatusBadge";
import { PageShell, Panel } from "../components/Shell";
import { useJobDetail } from "../lib/queries";
import { displayStatus, statusTone } from "../utils/format";
import { RefreshIcon } from "../components/icons";

export default function JobDetail() {
  const { id } = useParams();
  const { job, isLoading, isError, error, refetch } = useJobDetail(id);

  return (
    <PageShell
      title={job ? `Job #${job.id}` : "Job detail"}
      subtitle="Inspect command execution status and captured output."
      actions={
        <Link
          to={job ? `/clients/${job.client_id}` : "/"}
          className="rounded-full border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-sky-300 hover:text-sky-700"
        >
          {job ? "Back to client" : "Back"}
        </Link>
      }
    >
      {isLoading ? <Panel>Loading job...</Panel> : null}
      {isError ? <Panel className="text-rose-600">{error?.message ?? "Failed to load job."}</Panel> : null}

      {!isLoading && !isError && !job ? <Panel>Job not found.</Panel> : null}

      {!isLoading && !isError && job ? (
        <div className="grid gap-6 lg:grid-cols-[0.9fr_1.4fr]">
          <Panel>
            <div className="flex items-center justify-between gap-4">
              <div className="space-y-2">
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Job summary</p>
                <h2 className="text-2xl font-semibold text-slate-900">#{job.id}</h2>
              </div>
              <StatusBadge tone={statusTone(job.status)}>{displayStatus(job.status)}</StatusBadge>
            </div>

            <dl className="mt-6 space-y-4 text-sm">
              <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-4">
                <dt className="text-slate-500">Client ID</dt>
                <dd className="font-mono text-xs text-slate-900">{job.client_id}</dd>
              </div>
              <div className="space-y-2">
                <dt className="text-slate-500">Command</dt>
                <dd className="rounded-2xl bg-slate-950 px-4 py-3 font-mono text-xs text-slate-100">
                  {job.command}
                </dd>
              </div>
            </dl>
          </Panel>

          <Panel>
            <div className="flex items-center justify-between gap-4">
              <div className="space-y-2">
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Output</p>
                <h3 className="text-lg font-semibold text-slate-900">Execution log</h3>
              </div>
              <button
                type="button"
                onClick={() => refetch()}
                title="Refresh output"
                className="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-900 mx-2"
              >
                <RefreshIcon className="w-5 h-5" />
              </button>
            </div>

            <pre className="mt-4 min-h-80 overflow-auto rounded-2xl bg-slate-950 p-5 text-sm leading-6 text-slate-100">
              {job.output || "No output captured yet."}
            </pre>
          </Panel>
        </div>
      ) : null}
    </PageShell>
  );
}
