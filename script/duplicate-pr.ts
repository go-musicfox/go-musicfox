#!/usr/bin/env bun

import path from "path"
import { parseArgs } from "util"

async function main() {
  const { values, positionals } = parseArgs({
    args: Bun.argv.slice(2),
    options: {
      file: { type: "string", short: "f" },
      help: { type: "boolean", short: "h", default: false },
    },
    allowPositionals: true,
  })

  if (values.help) {
    console.log(`
Usage: bun script/duplicate-pr.ts [options] <message>

Options:
  -f, --file <path>   File to attach to the prompt (reads content and includes in prompt)
  -h, --help          Show this help message

Examples:
  bun script/duplicate-pr.ts -f pr_info.txt "Check the attached file for PR details"
`)
    process.exit(0)
  }

  const message = positionals.join(" ")
  if (!message) {
    console.error("Error: message is required")
    process.exit(1)
  }

  let fullPrompt = message

  if (values.file) {
    const resolved = path.resolve(process.cwd(), values.file)
    const file = Bun.file(resolved)
    if (!(await file.exists())) {
      console.error(`Error: file not found: ${values.file}`)
      process.exit(1)
    }
    const content = await file.text()
    fullPrompt = `${message}\n\n\`\`\`\n${content}\n\`\`\``
  }

  const proc = Bun.spawn(["opencode", "run", "--agent", "duplicate-pr"], {
    stdin: "pipe",
    stdout: "pipe",
    stderr: "inherit",
  })

  if (proc.stdin) {
    await proc.stdin.write(fullPrompt)
    proc.stdin.end()
  }

  const output = await new Response(proc.stdout).text()
  console.log(output.trim())
}

main()
