"use client";

import {useMutation, useQueryClient} from "@tanstack/react-query";
import {useRouter} from "next/navigation";
import {useToast} from "../../components/ToastProvider";

type MutationOptions<TData, TVariables> = {
  onSuccess?: (data: TData, variables: TVariables) => void;
  onError?: (error: Error, variables: TVariables) => void;
  successMessage?: string;
  queryKeysToInvalidate?: string[][];
};

export function useApiMutation<TData, TVariables>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options: MutationOptions<TData, TVariables> = {}
) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation<TData, Error, TVariables>({
    mutationFn,
    onSuccess: (data, variables) => {
      // Invalidate all query keys that changed
      if (options.queryKeysToInvalidate) {
        options.queryKeysToInvalidate.forEach((key) => {
          queryClient.invalidateQueries({queryKey: key});
        });
      }

      // Show success toast
      if (options.successMessage) {
        toast.success(options.successMessage);
      }

      // Call custom onSuccess if provided
      options.onSuccess?.(data, variables);

      // Refresh the router for server components
      router.refresh();
    },
    onError: (error, variables) => {
      toast.error(error.message || "An error occurred");
      options.onError?.(error, variables);
    },
  });
}