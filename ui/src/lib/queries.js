import React, { useState, useEffect, useCallback } from "react";
import { deleteClientJobs, dispatchCommand, fetchClient, fetchClients, fetchJobs } from "./api";

function useFetch(fetchFn) {
  const [data, setData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isFetching, setIsFetching] = useState(true);
  const [isError, setIsError] = useState(false);
  const [error, setError] = useState(null);

  const refetch = useCallback(async () => {
    setIsFetching(true);
    setIsError(false);
    setError(null);
    try {
      const result = await fetchFn();
      setData(result);
    } catch (err) {
      setIsError(true);
      setError(err);
    } finally {
      setIsLoading(false);
      setIsFetching(false);
    }
  }, [fetchFn]);

  const hasFetched = React.useRef(false);

  useEffect(() => {
    if (!hasFetched.current) {
        hasFetched.current = true;
        refetch().catch(() => {});
    }
  }, [refetch]);

  return { data, isLoading, isError, error, refetch, isFetching };
}

export function useClientsQuery() {
  return useFetch(fetchClients);
}

export function useJobsQuery() {
  return useFetch(fetchJobs);
}

export function useClientDetail(clientId) {
  const fetchClientFn = useCallback(() => {
    if (!clientId) return Promise.resolve(null);
    return fetchClient(clientId);
  }, [clientId]);

  const fetchJobsFn = useCallback(() => {
    if (!clientId) return Promise.resolve([]);
    return fetchJobs(clientId);
  }, [clientId]);

  const clientQuery = useFetch(fetchClientFn);
  const jobsQuery = useFetch(fetchJobsFn);

  const refetch = useCallback(() => {
    clientQuery.refetch().catch(() => {});
    jobsQuery.refetch().catch(() => {});
  }, [clientQuery.refetch, jobsQuery.refetch]);

  return {
    client: clientQuery.data ?? null,
    jobs: jobsQuery.data ?? [],
    isLoading: clientQuery.isLoading || jobsQuery.isLoading,
    isError: clientQuery.isError || jobsQuery.isError,
    error: clientQuery.error ?? jobsQuery.error ?? null,
    refetch
  };
}

export function useJobDetail(jobId) {
  const query = useJobsQuery();
  const job = Object.values(query.data?.jobs ?? {})
    .flat()
    .find((j) => String(j.id) === String(jobId)) ?? null;

  return { job, ...query };
}

function useMutation(mutationFn) {
  const [isPending, setIsPending] = useState(false);
  const [isError, setIsError] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState(null);

  const mutateAsync = async (...args) => {
    setIsPending(true);
    setIsError(false);
    setIsSuccess(false);
    setError(null);
    try {
      const result = await mutationFn(...args);
      setIsSuccess(true);
      return result;
    } catch (err) {
      setIsError(true);
      setError(err);
      throw err;
    } finally {
      setIsPending(false);
    }
  };

  return { mutateAsync, isPending, isError, isSuccess, error };
}

export function useDispatchCommand() {
  return useMutation(({ clientId, command }) => dispatchCommand(clientId, command));
}

export function useDeleteClientJobs() {
  return useMutation((clientId) => deleteClientJobs(clientId));
}