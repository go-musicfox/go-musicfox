#!/usr/bin/env bun

const repo = "go-musicfox/go-musicfox"
const days = 90
const msg =
  "This issue has been automatically closed because there has been no activity for 90 days. If you still want this issue to be addressed, please open a new issue."

const token = process.env.GITHUB_TOKEN
if (!token) {
  console.error("GITHUB_TOKEN environment variable is required")
  process.exit(1)
}

const cutoff = new Date(Date.now() - days * 24 * 60 * 60 * 1000)

type Issue = {
  number: number
  updated_at: string
}

const headers = {
  Authorization: `Bearer ${token}`,
  "Content-Type": "application/json",
  Accept: "application/vnd.github+json",
  "X-GitHub-Api-Version": "2022-11-28",
}

async function close(num: number) {
  const base = `https://api.github.com/repos/${repo}/issues/${num}`

  const comment = await fetch(`${base}/comments`, {
    method: "POST",
    headers,
    body: JSON.stringify({ body: msg }),
  })
  if (!comment.ok) throw new Error(`Failed to comment #${num}: ${comment.status} ${comment.statusText}`)

  const patch = await fetch(base, {
    method: "PATCH",
    headers,
    body: JSON.stringify({ state: "closed", state_reason: "not_planned" }),
  })
  if (!patch.ok) throw new Error(`Failed to close #${num}: ${patch.status} ${patch.statusText}`)

  console.log(`Closed https://github.com/${repo}/issues/${num}`)
}

async function main() {
  let page = 1
  let closed = 0

  while (true) {
    const res = await fetch(
      `https://api.github.com/repos/${repo}/issues?state=open&sort=updated&direction=asc&per_page=100&page=${page}`,
      { headers },
    )
    if (!res.ok) throw new Error(res.statusText)

    const all = (await res.json()) as Issue[]
    if (all.length === 0) break
    console.log(`Fetched page ${page} ${all.length} issues`)

    const stale: number[] = []
    for (const i of all) {
      const updated = new Date(i.updated_at)
      if (updated < cutoff) {
        stale.push(i.number)
      } else {
        console.log(`\nFound fresh issue #${i.number}, stopping`)
        if (stale.length > 0) {
          for (const num of stale) {
            await close(num)
            closed++
          }
        }
        console.log(`Closed ${closed} issues total`)
        return
      }
    }

    if (stale.length > 0) {
      for (const num of stale) {
        await close(num)
        closed++
      }
    }

    page++
  }

  console.log(`Closed ${closed} issues total`)
}

main().catch((err) => {
  console.error("Error:", err)
  process.exit(1)
})
