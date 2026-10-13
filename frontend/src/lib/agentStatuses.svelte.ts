import { pipe, subscribe } from 'wonka';
import { client } from './graphqlClient';
import { ActiveAgentStatusesDocument, AgentSessionStatus } from './graphql/generated';

class AgentStatusesStore {
  runningItemIds = $state<Set<string>>(new Set());

  #unsubscribe: (() => void) | null = null;

  subscribe(): void {
    if (this.#unsubscribe) return;

    const { unsubscribe } = pipe(
      client.subscription(ActiveAgentStatusesDocument, {}),
      subscribe(
        (result) => {
          if (result.error) {
            console.error('Agent statuses subscription error:', result.error);
            return;
          }

          const statuses = result.data?.activeAgentStatuses;
          if (statuses) {
            this.runningItemIds = new Set(
              statuses.filter((s) => s.status === AgentSessionStatus.Running).map((s) => s.itemId)
            );
          }
        }
      )
    );

    this.#unsubscribe = unsubscribe;
  }

  unsubscribe(): void {
    if (this.#unsubscribe) {
      this.#unsubscribe();
      this.#unsubscribe = null;
    }
  }

  isRunning(itemId: string): boolean {
    return this.runningItemIds.has(itemId);
  }
}

export const agentStatusesStore = new AgentStatusesStore();
