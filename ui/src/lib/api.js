const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "/api";

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers ?? {}),
    },
    ...options,
  });

  let data = null;
  const contentType = response.headers.get("content-type") ?? "";

  if (contentType.includes("application/json")) {
    data = await response.json();
  } else {
    const text = await response.text();
    data = text ? { message: text } : null;
  }

  if (!response.ok) {
    const message = data?.error || data?.message || `Request failed with status ${response.status}`;
    throw new Error(message);
  }

  return data;
}

export function fetchClients() {
  return request("/clients");
}

export function fetchClient(clientId) {
  return request(`/clients/${clientId}`);
}

export function fetchJobs(clientId = null) {
  const url = clientId ? `/jobs?clientId=${clientId}` : "/jobs";
  return request(url);
}

export function dispatchCommand(clientId, command, shell_type) {
  return request(`/push/${clientId}`, {
    method: "POST",
    body: JSON.stringify({ command, shell_type }),
  });
}

export function deleteClientJobs(clientId) {
  const params = new URLSearchParams({ clientId });
  return request(`/delete/jobs?${params.toString()}`, {
    method: "DELETE",
  });
}
