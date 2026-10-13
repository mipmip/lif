import type { Plugin } from '@opencode-ai/plugin';

/**
 * Lif Prime Plugin for OpenCode
 *
 * This plugin injects the output of `lif prime` into OpenCode's system prompt,
 * giving the AI context about your project's items (issues/tasks). It runs on:
 *
 * - Chat session start: Adds item context to the system prompt
 * - Session compaction: Re-injects context when the session is compacted
 *
 * Plugin Location Options:
 *
 * 1. Project-local (current): .opencode/plugin/lif-prime.ts
 *    - Only available in this project
 *    - Committed to version control, shared with collaborators
 *
 * 2. User-global: ~/.opencode/plugin/lif-prime.ts
 *    - Available in all your projects that use lif
 *    - Personal configuration, not shared
 */

export const LifPrimePlugin: Plugin = async ({ $, directory }) => {
  // Check if lif CLI exists and project has a lif config
  let prime = undefined;

  try {
    // Both conditions must be true:
    // 1. lif CLI is installed
    // 2. Project has .lif.yml config (or the legacy .beans.yml)
    const hasLif = await $`which lif`.quiet();
    // `.beans.yml` is the legacy config filename: legacy data names stay
    // readable (DESIGN §2, R-05), so a project not yet migrated still counts.
    const hasConfig = await $`test -f ${directory}/.lif.yml || test -f ${directory}/.beans.yml`.quiet();

    if (hasLif.exitCode === 0 && hasConfig.exitCode === 0) {
      const result = await $`lif prime`.cwd(directory).quiet();
      prime = result.stdout.toString();
    }
  } catch (e) {
    // lif not available or not configured - silently skip
  }

  return {
    'experimental.chat.system.transform': async (_, output) => {
      if (prime) {
        output.system.push(prime);
      }
    },
    'experimental.session.compacting': async (_, output) => {
      if (prime) {
        output.context.push(prime);
      }
    },
  };
};

export default LifPrimePlugin;
