---
max_turns: 10
timeout_seconds: 180
allowed_tools: [Skill, Read, Glob]
model: sonnet
runs: 3
plugins: [../../skills/craft/write]
---
Tighten these strings for our file-sharing app:

1. Upload error: "We're sorry, but unfortunately the file that you attempted to upload exceeds the maximum permitted file size of 25 MB. Please try again with a smaller file."
2. Empty state: "It looks like you don't currently have any files in this folder at the moment. You can get started by uploading your very first file using the button below."
3. Delete confirm: "Are you absolutely sure that you would like to proceed with permanently deleting this file? Please be aware that this action can't be undone."
