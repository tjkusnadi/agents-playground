# AGENTS

## Instructions
- Search across the internet, stackoverflow, reddit, github issues, all sources including docs, but developer experience may varies and not exists in the docs, so prioritize developer experience
- If there is no agents specify on prompt, please stop and give all agents available in the folder and give the summary of the tasks folder (they should be linked!)
- If needed learn from each tasks, it might help

## Folder Structure
- agents
    - This folder would consists of the main of the agent task
    - For example, `agents/image-processing.md`, the instructions will be there
- tasks
    - Once the task is done, please create a md file here
    - The format is 
        - tasks/${agent}/YYYY-MM/T-YYYY-MM-${agent}-${counter}.md
        - tasks/image-processing/2025-10/T-2025-10-image-processing-1.md
    - Increase the counter if there is a new task, not from pull request changes

### Task example    
```
id: T-2025-10-image-processing-1.md
title: Use img-proxy.net for resizing
owner: image-processing
created_at: 2025-10-25T07:00:00Z

Summary
Using img-proxy to resize

Idea of improvement on image-processing
- Explore another library for processing

Agent: [image-processing](./agents/image-processing.md)
```
if there is a review / change requested please store it on reviews/
```
tasks/image-processing/2025-10/
    T-2025-10-image-processing-1.md
    reviews/ -> use round format, 01,02, ...
        01.md -> store who requested a change, what are the changes, etc
    evidence/ -> store proof of completion or testing
        test-result.txt
        screenshots.png
        etc
    improvement.txt -> add your thoughts about next feature
```

Finally update tasks/image-processing/README.md to explain all md