export function formatTimestamp(value) {
  if (!value) {
    return "Unavailable";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function statusTone(status) {
  switch (normalizeStatus(status)) {
    case "finished":
    case "succeeded":
      return "success";
    case "failed":
      return "danger";
    case "running":
      return "info";
    case "pending":
      return "warning";
    default:
      return "neutral";
  }
}

export function onlineTone(isOnline) {
  return isOnline ? "success" : "neutral";
}

export function normalizeStatus(status) {
  if (typeof status === "number") {
    switch (status) {
      case 0:
        return "pending";
      case 1:
        return "running";
      case 2:
        return "finished";
      case 3:
        return "failed";
      default:
        return "unknown";
    }
  }

  if (typeof status === "string") {
    return status.toLowerCase();
  }

  return "unknown";
}

export function displayStatus(status) {
  const normalized = normalizeStatus(status);
  return normalized.charAt(0).toUpperCase() + normalized.slice(1);
}

export function truncateText(value, maxLength = 64) {
  if (!value) {
    return "No output captured yet.";
  }

  if (value.length <= maxLength) {
    return value;
  }

  return `${value.slice(0, maxLength - 1)}...`;
}
