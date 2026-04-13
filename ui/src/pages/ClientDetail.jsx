import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import StatusBadge from "../components/StatusBadge";
import { PageShell, Panel } from "../components/Shell";
import { useClientDetail, useDeleteClientJobs, useDispatchCommand } from "../lib/queries";
import { displayStatus, formatTimestamp, onlineTone, statusTone, truncateText } from "../utils/format";
import { RefreshIcon } from "../components/icons";

export default function ClientDetail() {
  const { id } = useParams();
  const { client, jobs, isLoading, isError, error, refetch } = useClientDetail(id);
  const dispatchMutation = useDispatchCommand();
  const deleteJobsMutation = useDeleteClientJobs();
  const [command, setCommand] = useState("");
  const [jobSearch, setJobSearch] = useState("");
  const [jobStatusFilter, setJobStatusFilter] = useState("all");
  const [jobSortField, setJobSortField] = useState("id");
  const [jobSortDirection, setJobSortDirection] = useState("desc");

  const visibleJobs = useMemo(() => {
    const normalizedSearch = jobSearch.trim().toLowerCase();

    const filteredJobs = jobs.filter((job) => {
      const normalizedStatus = displayStatus(job.status).toLowerCase();

      const matchesStatus = jobStatusFilter === "all" || normalizedStatus === jobStatusFilter;
      const matchesSearch =
        !normalizedSearch ||
        String(job.id).includes(normalizedSearch) ||
        job.command?.toLowerCase().includes(normalizedSearch) ||
        job.output?.toLowerCase().includes(normalizedSearch);

      return matchesStatus && matchesSearch;
    });

    filteredJobs.sort((left, right) => {
      let comparison = 0;

      switch (jobSortField) {
        case "status":
          comparison = displayStatus(left.status).localeCompare(displayStatus(right.status));
          break;
        case "command":
          comparison = (left.command || "").localeCompare(right.command || "");
          break;
        case "output":
          comparison = (left.output || "").localeCompare(right.output || "");
          break;
        case "id":
        default:
          comparison = Number(left.id) - Number(right.id);
          break;
      }

      return jobSortDirection === "asc" ? comparison : -comparison;
    });

    return filteredJobs;
  }, [jobSearch, jobSortDirection, jobSortField, jobStatusFilter, jobs]);

  async function handleSubmit(event) {
    event.preventDefault();
    const value = command.trim();
    if (!value || !id) {
      return;
    }

    try {
      await dispatchMutation.mutateAsync({ clientId: id, command: value });
      setCommand("");
      refetch();
    } catch {
      // Error state is surfaced below via the mutation.
    }
  }

  async function handleDeleteJobs() {
    if (!id || !jobs.length) {
      return;
    }

    const confirmed = window.confirm(
      `Delete all ${jobs.length} job(s) for this client? This uses the current backend delete-all endpoint.`,
    );

    if (!confirmed) {
      return;
    }

    try {
      await deleteJobsMutation.mutateAsync(id);
      refetch();
    } catch {
      // Error surfaced via mutation state.
    }
  }

  return (
    <PageShell
      title={client?.hostname || "Client detail"}
      subtitle={client ? "Inspect recent jobs and run a command on this client." : "Loading client details..."}
      actions={
        <Link
          to="/"
          className="rounded-full border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-sky-300 hover:text-sky-700"
        >
          Back to clients
        </Link>
      }
    >
      {isLoading ? <Panel>Loading client...</Panel> : null}
      {isError ? <Panel className="text-rose-600">{error?.message ?? "Failed to load client."}</Panel> : null}

      {!isLoading && !isError && client ? (
        <div className="space-y-6">
          <Panel>
            <div className="flex flex-col gap-6 xl:flex-row xl:items-start xl:justify-between">
              <div className="space-y-3">
                <div className="flex flex-wrap items-center gap-3">
                  <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Client</p>
                  <StatusBadge tone={onlineTone(client.online)}>{client.online ? "Online" : "Offline"}</StatusBadge>
                </div>
                <h2 className="text-2xl font-semibold text-slate-900">{client.hostname || "Unknown host"}</h2>
              </div>

              <dl className="grid gap-4 text-sm sm:grid-cols-3 xl:min-w-[30rem]">
                <div className="rounded-2xl bg-slate-50 px-4 py-3">
                  <dt className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Operating system</dt>
                  <dd className="mt-2 font-medium text-slate-900">{client.operating_system || "Unknown"}</dd>
                </div>
                <div className="rounded-2xl bg-slate-50 px-4 py-3">
                  <dt className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Last seen</dt>
                  <dd className="mt-2 font-medium text-slate-900">{formatTimestamp(client.last_seen_at)}</dd>
                </div>
                <div className="rounded-2xl bg-slate-50 px-4 py-3">
                  <dt className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Created</dt>
                  <dd className="mt-2 font-medium text-slate-900">{formatTimestamp(client.created_at)}</dd>
                </div>
              </dl>
            </div>
          </Panel>

          <Panel>
            <div className="mb-4 flex items-center justify-between gap-4">
              <div className="space-y-1">
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Run command</p>
                <h3 className="text-base font-semibold text-slate-900">Dispatch a new job</h3>
              </div>
              <p className="hidden text-sm text-slate-500 lg:block">
                Submit a command and track it below.
              </p>
            </div>

            <form className="space-y-3" onSubmit={handleSubmit}>
              <div className="flex flex-col gap-3 lg:flex-row lg:items-end">
                <textarea
                  value={command}
                  onChange={(event) => setCommand(event.target.value)}
                  rows={3}
                  placeholder="hostname && whoami"
                  className="min-h-24 w-full rounded-2xl border border-slate-300 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-sky-500 focus:bg-white"
                />
                <button
                  type="submit"
                  disabled={!client.online || dispatchMutation.isPending}
                  className="rounded-full bg-sky-600 px-5 py-2.5 text-sm font-semibold whitespace-nowrap text-white transition hover:bg-sky-500 disabled:cursor-not-allowed disabled:bg-slate-300"
                >
                  {dispatchMutation.isPending ? "Sending..." : client.online ? "Run command" : "Client offline"}
                </button>
              </div>

              <div className="min-h-5">
                {dispatchMutation.isError ? (
                  <p className="text-sm text-rose-600">
                    {dispatchMutation.error?.message ?? "Failed to dispatch command."}
                  </p>
                ) : null}
                {dispatchMutation.isSuccess ? (
                  <p className="text-sm text-emerald-600">Command dispatched successfully.</p>
                ) : null}
              </div>
            </form>
          </Panel>

          <Panel className="space-y-4">
            <div className="flex items-center justify-between gap-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Jobs</p>
                <h3 className="text-lg font-semibold text-slate-900">Recent activity</h3>
              </div>
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={() => refetch()}
                  title="Refresh jobs"
                  className="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-900"
                >
                  <RefreshIcon className="w-5 h-5" />
                </button>
                <p className="hidden text-sm text-slate-500 sm:block">{visibleJobs.length} visible / {jobs.length} total</p>
                <button
                  type="button"
                  onClick={handleDeleteJobs}
                  disabled={!jobs.length || deleteJobsMutation.isPending}
                  className="rounded-full border border-rose-200 px-4 py-2 text-sm font-medium text-rose-700 transition hover:bg-rose-50 disabled:cursor-not-allowed disabled:border-slate-200 disabled:text-slate-400"
                >
                  {deleteJobsMutation.isPending ? "Deleting..." : "Delete all"}
                </button>
              </div>
            </div>

            <div className="grid gap-3 lg:grid-cols-[1.5fr_0.9fr_0.9fr_auto]">
              <input
                type="search"
                value={jobSearch}
                onChange={(event) => setJobSearch(event.target.value)}
                placeholder="Search by job ID, command, or output"
                className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-sky-500"
              />

              <select
                value={jobStatusFilter}
                onChange={(event) => setJobStatusFilter(event.target.value)}
                className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-sky-500"
              >
                <option value="all">All statuses</option>
                <option value="pending">Pending</option>
                <option value="running">Running</option>
                <option value="finished">Finished</option>
                <option value="failed">Failed</option>
              </select>

              <select
                value={jobSortField}
                onChange={(event) => setJobSortField(event.target.value)}
                className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-sky-500"
              >
                <option value="id">Sort by ID</option>
                <option value="status">Sort by status</option>
                <option value="command">Sort by command</option>
                <option value="output">Sort by output</option>
              </select>

              <button
                type="button"
                onClick={() => setJobSortDirection((current) => (current === "asc" ? "desc" : "asc"))}
                className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm font-medium text-slate-700 transition hover:border-sky-300 hover:text-sky-700"
              >
                {jobSortDirection === "asc" ? "Ascending" : "Descending"}
              </button>
            </div>

            {deleteJobsMutation.isError ? (
              <p className="text-sm text-rose-600">
                {deleteJobsMutation.error?.message ?? "Failed to delete jobs."}
              </p>
            ) : null}

            {deleteJobsMutation.isSuccess ? (
              <p className="text-sm text-emerald-600">Jobs deleted successfully.</p>
            ) : null}

            {jobs.length ? (
              <div className="overflow-hidden rounded-2xl border border-slate-200">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-slate-200 text-sm">
                    <thead className="bg-slate-50 text-left text-xs uppercase tracking-[0.2em] text-slate-500">
                      <tr>
                        <th className="px-4 py-3 font-medium">ID</th>
                        <th className="px-4 py-3 font-medium">Status</th>
                        <th className="px-4 py-3 font-medium">Command</th>
                        <th className="px-4 py-3 font-medium">Output</th>
                        <th className="px-4 py-3 font-medium">Open</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 bg-white">
                      {visibleJobs.map((job) => (
                        <tr key={job.id} className="hover:bg-slate-50/80">
                          <td className="px-4 py-3 font-mono text-xs text-slate-500">#{job.id}</td>
                          <td className="px-4 py-3">
                            <StatusBadge tone={statusTone(job.status)}>{displayStatus(job.status)}</StatusBadge>
                          </td>
                          <td className="max-w-72 px-4 py-3 font-medium text-slate-900">{truncateText(job.command, 72)}</td>
                          <td className="max-w-80 px-4 py-3 text-slate-500">{truncateText(job.output, 88)}</td>
                          <td className="px-4 py-3">
                            <Link
                              to={`/jobs/${job.id}`}
                              className="text-sm font-medium text-sky-700 transition hover:text-sky-500"
                            >
                              View
                            </Link>
                          </td>
                        </tr>
                      ))}
                      {!visibleJobs.length ? (
                        <tr>
                          <td colSpan="5" className="px-4 py-10 text-center text-sm text-slate-500">
                            No jobs match the current search and filters.
                          </td>
                        </tr>
                      ) : null}
                    </tbody>
                  </table>
                </div>
              </div>
            ) : (
              <div className="rounded-2xl border border-dashed border-slate-200 px-4 py-10 text-center text-sm text-slate-500">
                No jobs recorded for this client yet.
              </div>
            )}
          </Panel>
        </div>
      ) : null}
    </PageShell>
  );
}
