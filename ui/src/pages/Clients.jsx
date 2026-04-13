import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import StatusBadge from "../components/StatusBadge";
import { PageShell, Panel } from "../components/Shell";
import { useClientsQuery } from "../lib/queries";
import { formatTimestamp, onlineTone } from "../utils/format";
import { RefreshIcon } from "../components/icons";

export default function Clients() {
  const { data, isLoading, isError, error, refetch, isFetching } = useClientsQuery();
  const clients = data?.clients ?? [];
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [sortField, setSortField] = useState("id");
  const [sortDirection, setSortDirection] = useState("desc");

  const filteredClients = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();

    const nextClients = clients.filter((client) => {
      const matchesStatus =
        statusFilter === "all" ||
        (statusFilter === "online" && client.online) ||
        (statusFilter === "offline" && !client.online);

      const matchesSearch =
        !normalizedSearch ||
        client.id?.toLowerCase().includes(normalizedSearch) ||
        client.hostname?.toLowerCase().includes(normalizedSearch) ||
        client.operating_system?.toLowerCase().includes(normalizedSearch);

      return matchesStatus && matchesSearch;
    });

    nextClients.sort((left, right) => {
      let comparison = 0;

      switch (sortField) {
        case "hostname":
          comparison = (left.hostname || "").localeCompare(right.hostname || "");
          break;
        case "status":
          comparison = Number(left.online) - Number(right.online);
          break;
        case "last_seen":
          comparison = new Date(left.last_seen_at || 0).getTime() - new Date(right.last_seen_at || 0).getTime();
          break;
        case "created_at":
          comparison = new Date(left.created_at || 0).getTime() - new Date(right.created_at || 0).getTime();
          break;
        case "id":
        default:
          comparison = (left.id || "").localeCompare(right.id || "");
          break;
      }

      return sortDirection === "asc" ? comparison : -comparison;
    });

    return nextClients;
  }, [clients, search, sortDirection, sortField, statusFilter]);

  return (
    <PageShell
      title="Clients"
      subtitle="See which agents are online, when they were last seen, and jump into their command history."
    >
      <Panel className="overflow-hidden p-0">
        <div className="space-y-4 border-b border-slate-200 px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-slate-900">Fleet Overview</h2>
              <p className="text-sm text-slate-500">{filteredClients.length} visible client(s)</p>
            </div>
            <div className="flex items-center gap-4">
              <p className="text-sm text-slate-500 hidden sm:block">{clients.length} registered total</p>
              <button
                type="button"
                onClick={() => refetch()}
                disabled={isFetching}
                title="Refresh clients"
                className="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-900 disabled:opacity-50 disabled:cursor-wait"
              >
                <RefreshIcon className={`w-5 h-5 ${isFetching ? 'animate-spin text-sky-500' : ''}`} />
              </button>
            </div>
          </div>

          <div className="grid gap-3 lg:grid-cols-[1.6fr_0.8fr_0.9fr_auto]">
            <input
              type="search"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search by hostname, client ID, or OS"
              className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-sky-500"
            />

            <select
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value)}
              className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-sky-500"
            >
              <option value="all">All statuses</option>
              <option value="online">Online only</option>
              <option value="offline">Offline only</option>
            </select>

            <select
              value={sortField}
              onChange={(event) => setSortField(event.target.value)}
              className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-sky-500"
            >
              <option value="id">Sort by ID</option>
              <option value="hostname">Sort by hostname</option>
              <option value="status">Sort by status</option>
              <option value="last_seen">Sort by last seen</option>
              <option value="created_at">Sort by created</option>
            </select>

            <button
              type="button"
              onClick={() => setSortDirection((current) => (current === "asc" ? "desc" : "asc"))}
              className="rounded-2xl border border-slate-300 bg-white px-4 py-3 text-sm font-medium text-slate-700 transition hover:border-sky-300 hover:text-sky-700"
            >
              {sortDirection === "asc" ? "Ascending" : "Descending"}
            </button>
          </div>
        </div>

        {isLoading ? (
          <div className="px-6 py-10 text-sm text-slate-500">Loading clients...</div>
        ) : null}

        {isError ? (
          <div className="px-6 py-10 text-sm text-rose-600">{error?.message ?? "Failed to load clients."}</div>
        ) : null}

        {!isLoading && !isError ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-200 text-sm">
              <thead className="bg-slate-50 text-left text-xs uppercase tracking-[0.2em] text-slate-500">
                <tr>
                  <th className="px-6 py-4 font-medium">Status</th>
                  <th className="px-6 py-4 font-medium">Hostname</th>
                  <th className="px-6 py-4 font-medium">Client ID</th>
                  <th className="px-6 py-4 font-medium">OS</th>
                  <th className="px-6 py-4 font-medium">Last Seen</th>
                  <th className="px-6 py-4 font-medium">Created</th>
                  <th className="px-6 py-4 font-medium">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 bg-white">
                {filteredClients.map((client) => (
                  <tr key={client.id} className="transition hover:bg-slate-50/80">
                    <td className="px-6 py-4">
                      <StatusBadge tone={onlineTone(client.online)}>
                        {client.online ? "Online" : "Offline"}
                      </StatusBadge>
                    </td>
                    <td className="px-6 py-4 font-medium text-slate-900">{client.hostname || "Unknown host"}</td>
                    <td className="px-6 py-4 font-mono text-xs text-slate-600">{client.id}</td>
                    <td className="px-6 py-4 text-slate-600">{client.operating_system || "Unknown"}</td>
                    <td className="px-6 py-4 text-slate-600">{formatTimestamp(client.last_seen_at)}</td>
                    <td className="px-6 py-4 text-slate-600">{formatTimestamp(client.created_at)}</td>
                    <td className="px-6 py-4">
                      <Link
                        to={`/clients/${client.id}`}
                        className="inline-flex rounded-full border border-slate-300 px-4 py-2 font-medium text-slate-700 transition hover:border-sky-300 hover:text-sky-700"
                      >
                        Open
                      </Link>
                    </td>
                  </tr>
                ))}
                {!filteredClients.length ? (
                  <tr>
                    <td colSpan="7" className="px-6 py-10 text-center text-slate-500">
                      No clients match the current search and filters.
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        ) : null}
      </Panel>
    </PageShell>
  );
}
