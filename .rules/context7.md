Priority: High
Instruction: MUST follow Context7 Instructions to correctly ask for documentation

# Context7 Instructions

## Tool Usage Guidelines
1. When you need documentation for a specific library, format your prompts to clearly mention the library name.
2. If you need information about a specific feature or function, include relevant keywords to help Context7 fetch the most relevant documentation.
3. For complex queries involving multiple libraries, prioritize the main library in your prompt.
4. Remember that Context7 works best when you're specific about what you're trying to accomplish.

## Available Tools
Context7 provides two main tools for accessing documentation:

1. **`resolve-library-id`**: Resolves a general library name into a Context7-compatible library ID.
   * **Parameters**:
     * `libraryName` (required): The name of the library you want documentation for (e.g., "react", "nextjs", "postgres")

2. **`get-library-docs`**: Fetches documentation for a library using a Context7-compatible library ID.
   * **Parameters**:
     * `context7CompatibleLibraryID` (required): The library ID (can be obtained from `resolve-library-id`)
     * `topic` (optional): Focus the docs on a specific topic (e.g., "routing", "hooks", "authentication")
     * `tokens` (optional, default 10000): Maximum number of tokens to return.

## Best Practices
1. **Be Clear and Specific**: Mention the exact library, version (if applicable), and specific functionality you're interested in.
2. **Mention Use Cases**: Describe what you're trying to build or the problem you're trying to solve.
3. **One Task at a Time**: For complex projects, break down your requests into smaller, focused queries.
4. **Verify and Iterate**: If the documentation isn't quite what you needed, refine your prompt to be more specific.

## Example Workflow
1. **Write your prompt**: "How do I create a basic Express.js API with MongoDB connection? use context7"
2. **Behind the scenes**:
   - Context7 identifies Express.js and MongoDB as key libraries
   - It fetches up-to-date documentation for both
   - The documentation is injected into the prompt context
3. **The AI response**: Generated code and explanations based on current, accurate documentation
