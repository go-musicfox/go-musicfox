import mri from "mri";
import { getFolderSizeBin } from "./bin";

function printUsage() {
  console.log(`\ngo-get-folder-size 

Get the size of a folder by recursively iterating through all its sub(files && folders). Use go, so high-speed.

usage:
  go-get-folder-size [options]

options:
  -h, --help            check help
  -l  --loose           ignore permission error
  -p, --pretty          pretty bytes (default true)
  -b, --base            target base path (default cwd)\n`);
}

async function main() {
  const _argv = process.argv.slice(2);
  const argv = mri(_argv, {
    default: {
      help: false,
      pretty: true,
      loose: false,
      base: process.cwd(),
    },
    string: ["base"],
    boolean: ["pretty", "help", "loose"],
    alias: {
      h: ["help"],
      b: ["base"],
      p: ["pretty"],
      l: ["loose"],
    },
  });

  if (argv.help) {
    printUsage();
  } else {
    const size = await getFolderSizeBin(
      argv.base,
      argv.pretty,
      { loose: argv.loose },
    );
    console.log(size);
  }
}

main();
