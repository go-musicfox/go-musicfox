/// <reference path="../env.d.ts" />
import { tool } from "@opencode-ai/plugin"

const LABELS = [
  "windows",
  "linux",
  "macos",
  "perf",
  "bug",
  "enhancement",
  "docs",
  "question",
  "helpwanted",
  "player",
  "unblock",
  "lastfm",
  "ui",
  "build",
] as const

const ASSIGNEES = ["anhoder"] as const

function getIssueNumber(): number {
  const issue = parseInt(process.env.ISSUE_NUMBER ?? "", 10)
  if (!issue) throw new Error("ISSUE_NUMBER env var not set")
  return issue
}

async function githubFetch(endpoint: string, options: RequestInit = {}) {
  const response = await fetch(`https://api.github.com${endpoint}`, {
    ...options,
    headers: {
      Authorization: `Bearer ${process.env.GITHUB_TOKEN}`,
      Accept: "application/vnd.github+json",
      "Content-Type": "application/json",
      ...options.headers,
    },
  })
  if (!response.ok) {
    throw new Error(`GitHub API error: ${response.status} ${response.statusText}`)
  }
  return response.json()
}

export default tool({
  description: `Use this tool to assign and/or label a GitHub issue for go-musicfox.

Choose labels and assignee using the current triage policy.
Pick the most fitting labels for the issue and assign the maintainer.`,
  args: {
    assignee: tool.schema
      .enum(ASSIGNEES as [string, ...string[]])
      .describe("The username of the assignee")
      .default("anhoder"),
    labels: tool.schema
      .array(tool.schema.enum(LABELS))
      .describe("The label(s) to add to the issue")
      .default([]),
  },
  async execute(args) {
    const issue = getIssueNumber()
    const owner = "go-musicfox"
    const repo = "go-musicfox"

    const results: string[] = []
    const text = `${process.env.ISSUE_TITLE ?? ""}\n${process.env.ISSUE_BODY ?? ""}`.toLowerCase()

    const osLabels: string[] = []
    if (/\bwindows\b/.test(text)) osLabels.push("windows")
    if (/\blinux\b/.test(text)) osLabels.push("linux")
    if (/\bmacos\b|\bmac os\b/.test(text)) osLabels.push("macos")

    const allLabels = [...new Set([...args.labels, ...osLabels])]

    await githubFetch(`/repos/${owner}/${repo}/issues/${issue}/assignees`, {
      method: "POST",
      body: JSON.stringify({ assignees: [args.assignee] }),
    })
    results.push(`Assigned @${args.assignee} to issue #${issue}`)

    if (allLabels.length > 0) {
      await githubFetch(`/repos/${owner}/${repo}/issues/${issue}/labels`, {
        method: "POST",
        body: JSON.stringify({ labels: allLabels }),
      })
      results.push(`Added labels: ${allLabels.join(", ")}`)
    }

    return results.join("\n")
  },
})
