---
name: merge
description: Merge task changes to the main branch using git or jj commands. Handles repo detection, initialization, conflict resolution, and cleanup.
---

# Merge Skill

You are a merge specialist. Your task is to merge the current task's changes into the main branch using the appropriate version control system.

## Prerequisites

- You have access to the workspace directory via existing tools
- You can use bash, read, edit, and other standard tools
- You should recall this skill by calling `recall-skill` with `name="merge"`

## Steps

### 1. Detect Repository Type

Check for version control in the workspace:

```bash
# Check for JJ repo (preferred)
ls -la .jj 2>/dev/null && echo "JJ_REPO"

# Check for Git repo
ls -la .git 2>/dev/null && echo "GIT_REPO"
```

**IMPORTANT - JJ Gotchas:**
- The project root is the directory that contains `.jj/` - don't run jj commands if `.jj` doesn't exist
- NEVER assume the current directory is the project root - we may operate either inside the project root or its parent
- JJ workspaces live in the parent of the project root and MUST NOT be created inside the project root
- Workspace naming convention: `jjws-<reponame>-<task>`
- Agent workspaces are temporary and should be deleted after merge
- If `@` (current change in main) already has changes, run `jj new` before `jj squash --from <workspace_change> --into @` to keep merges clean

### 2. Initialize Git if No VCS Found

If neither `.jj` nor `.git` exists:

```bash
git init
git add -A
git commit -m "Initial commit"
```

### 3. For Git Repositories

#### 3a. Identify the Task Branch

```bash
git branch --show-current
```

Note the branch name (usually starts with `agent/task-` or similar).

#### 3b. Checkout and Update Main

```bash
# Try main first, fall back to master
git checkout main || git checkout master
git pull origin main || git pull origin master
```

#### 3c. Merge the Task Branch

```bash
git merge <branch-name> --no-edit
```

#### 3d. Handle Conflicts (if any)

If merge fails with conflicts:

```bash
# See which files have conflicts
git status

# For each conflicted file, read and understand the conflict:
cat <conflicted-file>

# Resolve by editing - remove conflict markers (<<<<<<<, =======, >>>>>>>)
# Keep the correct code, remove the markers
# Then stage the resolved file:
git add <resolved-file>
```

**Conflict Resolution Tips:**
- Read the file to understand what both sides changed
- Look at context around the conflict markers
- Keep the best version or merge changes intelligently
- Always remove all conflict markers
- Test that the resolved code still works if possible

#### 3e. Complete the Merge

```bash
# If conflicts were resolved, commit them
git commit --no-edit

# Push to origin
git push origin main || git push origin master

# Cleanup: delete the task branch
git branch -d <branch-name>
git push origin --delete <branch-name>  # optional
```

### 4. For JJ (Jujutsu) Repositories

**CRITICAL JJ WORKSPACE RULES:**
- Agents MUST NOT edit in `main` - work only in their own workspace
- All work for a task MUST be in a single jj change
- If multiple changes exist, squash/fold into ONE final change with clear description before merge

#### 4a. Get the Change ID

```bash
# Get the current change ID
jj log -n 1 --template 'change_id'
```

#### 4b. Squash into Main

From the main workspace (not the task workspace):

```bash
# IMPORTANT: If @ already has changes in main, run jj new first
jj new

# Then squash the task change into main
jj squash --from <change-id> --into @
```

#### 4c. Push to Remote

```bash
# Set bookmark to main
jj bookmark set main -r @

# Push to origin
jj git push --bookmark main
```

#### 4d. Cleanup

```bash
# Delete the task bookmark
jj bookmark delete <task-bookmark-name>

# Remove the workspace
jj workspace forget <workspace-name>

# Delete the workspace directory (it's in parent of project root)
rm -rf ../jjws-<reponame>-<task>
```

**JJ Quick Reference:**
- `jj workspace list` - see workspaces and change IDs
- `jj status` - working copy changes
- `jj log -n 10` - recent changes with IDs
- `jj describe -m "type: message"` - set change description
- `jj diff` - review edits

## 5. Report Results

Always report:
- What repository type was detected (git/jj/none)
- Key commands executed and their outcomes
- Whether the merge succeeded or failed
- If conflicts were resolved, summarize what files were affected and how
- Any cleanup steps taken
- If using jj, confirm: `jj workspace list` shows workspace is gone

## Success Criteria

- For git: Changes are in main branch, pushed to origin, task branch deleted
- For jj: Change is squashed into @, pushed to origin, workspace cleaned up
- For either: No uncommitted changes, no lingering branches/bookmarks

## Error Handling

If any command fails:
1. Stop and report the error
2. Show the command output
3. Suggest what might have gone wrong
4. Ask user for guidance on how to proceed

Do not proceed blindly if commands are failing - this could corrupt the repository state.
